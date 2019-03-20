package spreadsheet

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"

	"github.com/360EntSecGroup-Skylar/excelize"

	"github.com/materials-commons/mcetl/internal/spreadsheet/model"
)

// Load will load the given excel file. This assumes that each process is in a separate
// worksheet and the process will take on the name of the worksheet. The way that Load
// works is it transforms the spreadsheet into a data structure that can be more easily
// understood and worked with. This is encompassed in the model.Process data structure.
func Load(path string) ([]*model.Process, error) {
	var processes []*model.Process
	xlsx, err := excelize.OpenFile(path)
	if err != nil {
		return processes, err
	}

	var savedErrs *multierror.Error
	for index, name := range xlsx.GetSheetMap() {
		process, err := loadWorksheet(xlsx, name, index)
		if err != nil {
			savedErrs = multierror.Append(savedErrs, err)
			continue
		}
		processes = append(processes, process)
	}

	if err := validateParents(processes); err != nil {
		savedErrs = multierror.Append(savedErrs, err)
	}

	return processes, savedErrs.ErrorOrNil()
}

// loadWorksheet will load the given worksheet into the model.Process data structure. The spreadsheet
// must have the follow format:
//   1st row is composed of headers as follows:
//     |sample|parent sample|optional process attribute columns|<blank>|optional sample attribute columns|
// Examples:
//    This example has no process attributes
//         |sample|parent sample||sample attr1(unit)|sample attr2(unit)|
//    This example has process attributes and no sample attributes
//         |sample|parent sample|process attr1(unit)|process attr2|
//    This example has 1 process attribute and 2 sample attributes
//         |sample|parent sample|process attr1(unit)||sample attr1(unit)|sample attr2(unit)|
func loadWorksheet(xlsx *excelize.File, worksheetName string, index int) (*model.Process, error) {
	rows, err := xlsx.Rows(worksheetName)
	if err != nil {
		return nil, err
	}

	rowProcessor := newRowProcessor(worksheetName, index)
	row := 0

	// First row is the header row that contains all the attributes, process this first
	// so that we don't have to special case the loop to check for the first row each time
	if rows.Next() {
		row++
		rowProcessor.processHeaderRow(rows)
	}

	// Loop through the rest of the rows processing the samples, sample attributes and process
	// attributes associated with a sample
	for rows.Next() {
		row++
		rowProcessor.processSampleRow(rows, row)
	}

	return rowProcessor.process, nil
}

type rowProcessor struct {
	process *model.Process

	// Sample attributes start at column 4 or greater. Remember the format is:
	// |sample|parent sample|<blank>|sample attr
	// That is there must be a blank column before sample attributes start. The
	// column that this starts can be greater than 4 if there are process attributes.
	// For example the following example would have startingSampleAttrsCol = 6
	//      1        2             3           4         5   6
	//   |sample|parent sample|process attr|process attr||sample attr
	// where 5 is our blank column
	startingSampleAttrsCol int
}

func newRowProcessor(processName string, index int) *rowProcessor {
	return &rowProcessor{
		process: &model.Process{
			Name:  processName,
			Index: index,
		},
		startingSampleAttrsCol: 4,
	}
}

// processHeaderRow processes the first row in the spreadsheet. This row is the header row and contains
// the names of all the process and sample attributes.
// Attributes have the following format: name(unit)
// For example:
//    temperature(c)   - Attribute name temperature with unit c
//    quartile         - Attribute quartile with no units
func (r *rowProcessor) processHeaderRow(row *excelize.Rows) {
	column := 0
	inProcessAttrs := true
	for _, colCell := range row.Columns() {
		colCell = strings.TrimSpace(colCell)
		column++
		if column < 3 {
			// column one is sample name
			// column two is parent sample
			continue
		}

		if colCell == "" && inProcessAttrs {
			// The first blank column denotes the end of the process attributes, after that we
			// are processing sample attributes
			inProcessAttrs = false
			r.startingSampleAttrsCol = column + 1
		} else if inProcessAttrs {
			// We haven't encountered a blank column yet so still reading process attributes
			name, unit := cell2NameAndUnit(colCell)
			attr := model.NewAttribute(name, unit, column)
			r.process.AddAttribute(attr)
		} else {
			// A blank column was previously encountered so we are reading sample attributes
			name, unit := cell2NameAndUnit(colCell)
			attr := model.NewAttribute(name, unit, column)
			r.process.AddSampleAttr(attr)
		}
	}
}

// processSampleRow processes a row that has a sample on it. This row has the same format as above
// except that now it is reading values for attributes as opposed to attribute names. These values
// can be arbitrary strings. They will be turned into JSON strings that look like {value: column},
// For example:
//   cell: [0,1,2,3], becomes the string: {value: [0,1,2,3]}
//   cell: {edge: 1, angle: 2}, becomes the string; {value: {edge: 1, angle: 2}}
// The reason for the conversion is that these cell values will be stored in the database a JSON objects
// with a top level value key.
func (r *rowProcessor) processSampleRow(row *excelize.Rows, rowIndex int) {
	column := 0
	inProcessAttrs := true
	var currentSample *model.Sample = nil

	for _, colCell := range row.Columns() {
		colCell = strings.TrimSpace(colCell)
		column++
		if column == 1 {
			// Sample
			currentSample = model.NewSample(colCell, rowIndex)
			r.process.AddSample(currentSample)
		} else if column == 2 {
			// parent sample
			currentSample.Parent = colCell
		} else if colCell == "" && inProcessAttrs {
			// Blank column - switch from process attributes to sample attributes
			inProcessAttrs = false
		} else if inProcessAttrs {
			// No blank column seed so still reading process attributes
			attr := r.process.Attributes[column-3]
			processAttr := model.NewAttribute(attr.Name, attr.Unit, attr.Column)
			if colCell != "" {
				fmt.Println("process attribute", colCell)
				val := make(map[string]interface{})
				if err := json.Unmarshal([]byte(fmt.Sprintf(`{"value": %s}`, colCell)), &val); err == nil {
					processAttr.Value = val
				} else {
					fmt.Println("  json.Unmarshal returned error:", err)
				}
			}
			currentSample.AddProcessAttribute(processAttr)
		} else {
			// saw a blank column so now reading sample attributes
			attr := r.process.SampleAttrs[column-r.startingSampleAttrsCol]
			sampleAttr := model.NewAttribute(attr.Name, attr.Unit, attr.Column)
			if colCell != "" {
				val := make(map[string]interface{})
				if err := json.Unmarshal([]byte(fmt.Sprintf("{value: %s}", colCell)), val); err == nil {
					sampleAttr.Value = val
				}
			}
			currentSample.AddAttribute(sampleAttr)
		}
	}
}

// cell2NameAndUnit takes a string of the form name(unit), where the (unit) part is optional,
// splits it up and returns the name and unit. Examples:
//   temperature(c) => temperature, c
//   quadrant       => quadrant, ""
//   length(m       => length, m   // As a special case handles units specified without a closing paren
func cell2NameAndUnit(cell string) (name, unit string) {
	name = ""
	unit = ""
	if cell == "" {
		return name, unit
	}

	indexOpeningParen := strings.Index(cell, "(")
	indexClosingParen := strings.Index(cell, ")")
	if indexOpeningParen == -1 {
		// No units specified so return the cell value as the name
		return cell, ""
	}

	// If we are here then there is a unit. There are two situations, either the user has a closing
	// paren ')' or they don't. We treat a missing closing paren as a correctable problem by just taking
	// everything after the open '(' to the end of the string as the unit if this occurs. Otherwise we
	// take the value between the parens.
	// Examples of parsing
	//   cell := "abc(u)"
	//   indexOpeningParen := strings.Index(str, "(")
	//   indexClosingParen := strings.Index(str, ")")
	//   fmt.Println(str[:indexOpeningParen]) => abc
	//   fmt.Println(str[indexOpeningParen+1:indexClosingParen]) => u
	//
	//   cell = "abcd(u"
	//   indexOpeningParen = strings.Index(str2, "(")
	//   fmt.Println(str2[indexOpeningParen+1:]) => u

	switch {
	case indexClosingParen != -1:
		name = cell[:indexOpeningParen]
		unit = cell[indexOpeningParen+1 : indexClosingParen]
		return name, unit
	default:
		// indexClosingParen == -1, which means we have a string like: abc(c
		// that has no closing paren
		name = cell[:indexOpeningParen]
		unit = cell[indexOpeningParen+1:]
		return name, unit
	}
}

// validateParents goes through the list of samples and checks each of their
// Parent attributes. If Parent is not blank then it must contain a reference
// to a known process. Additionally that process cannot be the current process.
// This determination is done by name. Remember processes have the name of their
// worksheet, so we check that a non blank Parent is equal to a known process
// that isn't the process the sample is in. validateParent returns a multierror
// containing all the errors encountered.
func validateParents(processes []*model.Process) error {
	knownProcesses := createKnownProcessesMap(processes)
	var foundErrors *multierror.Error
	for _, process := range processes {
		for _, sample := range process.Samples {
			if sample.Parent != "" {
				switch {
				case sample.Parent == process.Name:
					e := fmt.Errorf("process '%s' has Sample '%s' who's parent is the current process", process.Name, sample.Name)
					foundErrors = multierror.Append(foundErrors, e)
				default:
					if _, ok := knownProcesses[sample.Parent]; !ok {
						// Parent is set to a non-existent process
						e := fmt.Errorf("sample '%s' in process '%s' has parent '%s' that does not exist",
							sample.Name, process.Name, sample.Parent)
						foundErrors = multierror.Append(foundErrors, e)
					}
				}
			}
		}
	}

	return foundErrors.ErrorOrNil()
}

// createKnownProcessesMap creates a map of [process.Name] => Process
func createKnownProcessesMap(processes []*model.Process) map[string]*model.Process {
	knownProcesses := make(map[string]*model.Process)
	for _, process := range processes {
		knownProcesses[process.Name] = process
	}

	return knownProcesses
}
