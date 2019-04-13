package spreadsheet

import (
	"fmt"

	"github.com/hashicorp/go-multierror"

	"github.com/360EntSecGroup-Skylar/excelize"

	"github.com/materials-commons/mcetl/internal/spreadsheet/model"
)

// Load will load the given excel file. This assumes that each process is in a separate
// worksheet and the process will take on the name of the worksheet. The way that Load
// works is it transforms the spreadsheet into a data structure that can be more easily
// understood and worked with. This is encompassed in the model.Worksheet data structure.
func Load(path string, headerRow int) ([]*model.Worksheet, error) {
	var worksheets []*model.Worksheet

	// Make sure the keywords are valid before we start processing the spreadsheet,
	// otherwise we can't reliably load the spreadsheet because the same keyword
	// could be used for different attribute types.
	if err := ValidateKeywords(); err != nil {
		return worksheets, err
	}

	xlsx, err := excelize.OpenFile(path)
	if err != nil {
		return worksheets, err
	}

	var savedErrs *multierror.Error

	// Loop through each of the worksheets in the excel file creating a list
	// of loading errors so we can report back all the load/parsing errors
	// to the user.
	for index, name := range xlsx.GetSheetMap() {
		worksheet, err := loadWorksheet(xlsx, headerRow, name, index)
		if err != nil {
			savedErrs = multierror.Append(savedErrs, err)
			continue
		}
		worksheets = append(worksheets, worksheet)
	}

	// To build the workflow column 2 in a worksheet is the parent column. It points to
	// the sheet to that is sending a sample into this step. Validate that the parents
	// were correctly specified.
	if err := validateParents(worksheets); err != nil {
		savedErrs = multierror.Append(savedErrs, err)
	}

	return worksheets, savedErrs.ErrorOrNil()
}

// loadWorksheet will load the given worksheet into the model.Worksheet data structure. The spreadsheet
// must have the follow format:
//   1st row is composed of headers as follows:
//     |sample|parent process for sample|keyword attribute columns|
// Examples:
//    This example has no process attributes
//         |sample|parent sample|s:sample attr1(unit)|s:sample attr2(unit)|
//    This example has process attributes and no sample attributes
//         |sample|parent sample|p:process attr1(unit)|p:process attr2|
//    This example has 1 process attribute and 2 sample attributes
//         |sample|parent sample|p:process attr1(unit)|s:sample attr1(unit)|s:sample attr2(unit)|
//
// The rows after the header row contain the data. Columns 1 and 2 are special as they are reserved
// for the sample name and the parent worksheet.
func loadWorksheet(xlsx *excelize.File, headerRow int, worksheetName string, index int) (*model.Worksheet, error) {
	rows, err := xlsx.Rows(worksheetName)
	if err != nil {
		return nil, err
	}

	rowProcessor := newRowProcessor(worksheetName, index)
	row := 0

	// skip specified rows to header
	for i := 0; i < headerRow; i++ {
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
