package spreadsheet

import (
	"fmt"

	"github.com/hashicorp/go-multierror"

	"github.com/360EntSecGroup-Skylar/excelize"

	mcapi "github.com/materials-commons/gomcapi"
	"github.com/materials-commons/mcetl/internal/spreadsheet/model"
)

type Loader struct {
	HasParent bool
	HeaderRow int
	Paths     []string
}

func NewLoader(hasParent bool, headerRow int, paths []string) *Loader {
	return &Loader{
		HasParent: hasParent,
		HeaderRow: headerRow,
		Paths:     paths,
	}
}

// Load will load the given excel file. This assumes that each process is in a separate
// worksheet and the process will take on the name of the worksheet. The way that Load
// works is it transforms the spreadsheet into a data structure that can be more easily
// understood and worked with. This is encompassed in the model.Worksheet data structure.
// The header row parameter is the starting row for the header. Rows before that will
// be skipped.
func (l *Loader) Load() ([]*model.Worksheet, error) {
	var worksheets []*model.Worksheet

	// Make sure the keywords are valid before we start processing the spreadsheet,
	// otherwise we can't reliably load the spreadsheet because the same keyword
	// could be used for different attribute types.
	if err := ValidateKeywords(); err != nil {
		return worksheets, err
	}

	var savedErrs *multierror.Error

	// Loop through each file and build up the list of worksheets across all of the files
	for _, file := range l.Paths {
		xlsx, err := excelize.OpenFile(file)
		if err != nil {
			return worksheets, err
		}

		// Loop through each of the worksheets in the excel file creating a list
		// of loading errors so we can report back all the load/parsing errors
		// to the user.
		for index, name := range xlsx.GetSheetMap() {
			worksheet, err := l.loadWorksheet(xlsx, name, index)
			if err != nil {
				savedErrs = multierror.Append(savedErrs, err)
				continue
			}
			worksheets = append(worksheets, worksheet)
		}
	}

	// To build the workflow column 2 in a worksheet is the parent column. It points to
	// the sheet to that is sending a sample into this step. Validate that the parents
	// were correctly specified. This step is only needed when column 2 points to other
	// worksheets.
	if l.HasParent {
		if err := validateParents(worksheets); err != nil {
			savedErrs = multierror.Append(savedErrs, err)
		}
	}

	return worksheets, savedErrs.ErrorOrNil()
}

// ValidateFilesExistInProject will check that all the files in a given spreadsheet exist. It is broken out as
// a separate method from Load as checking can be expensive and the Load method is used both during
// checking and during the process where the spreadsheet is used to create data on the server. In
// this way the user of the API can decide when this potentially expensive step should be run.
func (l *Loader) ValidateFilesExistInProject(worksheets []*model.Worksheet, projectID string, c *mcapi.Client) error {
	uniqueFilePaths := make(map[string]bool)

	// Construct a list of all the unique file paths so we don't check a path multiple times. This could
	// occur because the same file path is used in multiple samples.
	for _, worksheet := range worksheets {
		for _, sample := range worksheet.Samples {
			for _, file := range sample.Files {
				uniqueFilePaths[file.Path] = true
			}
		}
	}

	var savedErrors *multierror.Error

	for path := range uniqueFilePaths {
		if _, err := c.GetFileByPathInProject(path, projectID); err != nil {
			savedErrors = multierror.Append(savedErrors, fmt.Errorf("warning: file '%s' not found in project", path))
		}
	}

	return savedErrors.ErrorOrNil()
}

// loadWorksheet will load the given worksheet into the model.Worksheet data structure. The spreadsheet
// must have the follow format:
//   1st row is composed of headers as follows:
//     |sample|parent process for sample|keyword attribute columns|
//   Note: 2nd column (parent process for sample) is optional, it may also be a normal attribute column.
// Examples:
//    This example has no process attributes
//         |sample|parent sample|s:sample attr1(unit)|s:sample attr2(unit)|
//    This example has process attributes and no sample attributes
//         |sample|parent sample|p:process attr1(unit)|p:process attr2|
//    This example has 1 process attribute and 2 sample attributes
//         |sample|parent sample|p:process attr1(unit)|s:sample attr1(unit)|s:sample attr2(unit)|
//
// The rows after the header row contain the data. Column 1 is special and column 2 may be special (if HasParent is true
// then column 2 is a special column). Column 1 is the sample name, and column 2, if it is special is the worksheet that
// is the parent process for this step.
func (l *Loader) loadWorksheet(xlsx *excelize.File, worksheetName string, index int) (*model.Worksheet, error) {
	rows, err := xlsx.Rows(worksheetName)
	if err != nil {
		return nil, err
	}

	rowProcessor := newRowProcessor(worksheetName, l.HasParent, index)
	row := 0

	// skip specified rows to header
	for i := 0; i < l.HeaderRow; i++ {
		rows.Next()
	}

	// First row is the header row that contains all the attributes. We process this first
	// outside of the loop that processes each of the sample rows.
	if rows.Next() {
		row++
		rowProcessor.processHeaderRow(rows)
	}

	// Loop through the rest of the rows processing the samples, and their process, sample and file attributes.
	for rows.Next() {
		row++
		if err := rowProcessor.processSampleRow(rows, row); err != nil {
			return nil, err
		}
	}

	return rowProcessor.worksheet, nil
}

// validateParents goes through all the samples in the worksheets and checks
// each of their Parent attributes. If Parent is not blank then it must contain
// a reference to a known process. Additionally that process cannot be the
// current process. This determination is done by name. Remember processes have
// the name of their worksheet, so we check that a non blank Parent is equal to
// a known process that isn't the process the sample is in. validateParent returns
// a multierror containing all the errors encountered.
func validateParents(worksheets []*model.Worksheet) error {
	knownProcesses := createKnownProcessesMap(worksheets)
	var foundErrors *multierror.Error
	for _, worksheet := range worksheets {
		for _, sample := range worksheet.Samples {
			if sample.Parent != "" {
				switch {
				case sample.Parent == worksheet.Name:
					e := fmt.Errorf("process '%s' has Sample '%s' who's parent is the current process", worksheet.Name, sample.Name)
					foundErrors = multierror.Append(foundErrors, e)
				default:
					if _, ok := knownProcesses[sample.Parent]; !ok {
						// Parent is set to a non-existent process
						e := fmt.Errorf("sample '%s' in process '%s' has parent '%s' that does not exist",
							sample.Name, worksheet.Name, sample.Parent)
						foundErrors = multierror.Append(foundErrors, e)
					}
				}
			}
		}
	}

	return foundErrors.ErrorOrNil()
}

// createKnownProcessesMap creates a map of [process.Name] => Worksheet
func createKnownProcessesMap(processes []*model.Worksheet) map[string]*model.Worksheet {
	knownProcesses := make(map[string]*model.Worksheet)
	for _, process := range processes {
		knownProcesses[process.Name] = process
	}

	return knownProcesses
}
