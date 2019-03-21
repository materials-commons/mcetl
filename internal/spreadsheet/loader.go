package spreadsheet

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"

	"github.com/360EntSecGroup-Skylar/excelize"

	"github.com/materials-commons/mcetl/internal/spreadsheet/model"
)

type ColumnAttribute int

const (
	SampleAttributeColumn = iota + 1
	ProcessAttributeColumn
	FileAttributeColumn
)

var SampleAttributeKeywords = map[string]bool{
	"s":                true,
	"sample":           true,
	"sample attribute": true,
}

var ProcessAttributeKeywords = map[string]bool{
	"p":       true,
	"process": true,
}

var FileAttributeKeywords = map[string]bool{
	"f":     true,
	"file":  true,
	"files": true,
}

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
//     |sample|parent process for sample|keyword attribute columns|
// Examples:
//    This example has no process attributes
//         |sample|parent sample|s:sample attr1(unit)|s:sample attr2(unit)|
//    This example has process attributes and no sample attributes
//         |sample|parent sample|p:process attr1(unit)|p:process attr2|
//    This example has 1 process attribute and 2 sample attributes
//         |sample|parent sample|p:process attr1(unit)|s:sample attr1(unit)|s:sample attr2(unit)|
//
// Keywords are stored in the file level variables SampleAttributeKeywords, ProcessAttributeKeywords
// and FileAttributeKeywords
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
	process    *model.Process
	columnType map[int]ColumnAttribute
}

func newRowProcessor(processName string, index int) *rowProcessor {
	return &rowProcessor{
		process: &model.Process{
			Name:  processName,
			Index: index,
		},
		columnType: make(map[int]ColumnAttribute),
	}
}

// processHeaderRow processes the first row in the spreadsheet. This row is the header row and contains
// the names of all the process, sample and file attributes. The type of an attribute is determined
// by looking at its keyword prefix.
func (r *rowProcessor) processHeaderRow(row *excelize.Rows) {
	column := 0
	for _, colCell := range row.Columns() {
		colCell = strings.TrimSpace(colCell)
		column++
		if column < 3 {
			// column one is sample name
			// column two is parent sample
			continue
		}

		if isProcessAttributeHeader(colCell) {
			name, unit := cell2NameAndUnit(colCell)
			attr := model.NewAttribute(name, unit, column)
			r.columnType[column] = ProcessAttributeColumn
			r.process.AddAttribute(attr)
		} else if isSampleAttributeHeader(colCell) {
			name, unit := cell2NameAndUnit(colCell)
			attr := model.NewAttribute(name, unit, column)
			r.columnType[column] = SampleAttributeColumn
			r.process.AddSampleAttr(attr)
		} else if isFileAttributeHeader(colCell) {
			// ignore for the moment
			r.columnType[column] = FileAttributeColumn
		}
	}
}

func isSampleAttributeHeader(cell string) bool {
	return findIn(cell, SampleAttributeKeywords)
}

func isProcessAttributeHeader(cell string) bool {
	return findIn(cell, ProcessAttributeKeywords)
}

func isFileAttributeHeader(cell string) bool {
	return findIn(cell, FileAttributeKeywords)
}

func findIn(cell string, keywords map[string]bool) bool {
	cell = strings.ToLower(cell)
	i := strings.Index(cell, ":")
	if i == -1 {
		return false
	}

	keyword := cell[:i]
	_, ok := keywords[keyword]
	return ok
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
		} else {
			colType, ok := r.columnType[column]
			switch {
			case !ok: // Couldn't find column type
				continue
			case colType == SampleAttributeColumn:
				attr := findAttr(r.process.SampleAttrs, column)
				sampleAttr := model.NewAttribute(attr.Name, attr.Unit, attr.Column)
				if colCell != "" {
					val := make(map[string]interface{})
					if err := json.Unmarshal([]byte(fmt.Sprintf(`{"value": "%s"}`, colCell)), &val); err == nil {
						sampleAttr.Value = val
					} else {
						fmt.Printf("json.Unmarhal of %s failed: %s\n", colCell, err)
					}
				}
				currentSample.AddAttribute(sampleAttr)
			case colType == ProcessAttributeColumn:
				attr := findAttr(r.process.Attributes, column)
				processAttr := model.NewAttribute(attr.Name, attr.Unit, attr.Column)
				if colCell != "" {
					val := make(map[string]interface{})
					if err := json.Unmarshal([]byte(fmt.Sprintf(`{"value": "%s"}`, colCell)), &val); err == nil {
						processAttr.Value = val
					} else {
						fmt.Printf("json.Unmarhal of %s failed: %s\n", colCell, err)
					}
				}
				currentSample.AddProcessAttribute(processAttr)
			case colType == FileAttributeColumn:
			}
		}
	}
}

func findAttr(attributes []*model.Attribute, column int) *model.Attribute {
	for _, attr := range attributes {
		if attr.Column == column {
			return attr
		}
	}

	return nil
}

// cell2NameAndUnit takes a string of the form <keyword:>name(unit), where the (unit) part is optional,
// splits it up and returns the name and unit. The <keyword:> is optional. Examples:
//   temperature(c) => temperature, c
//   quadrant       => quadrant, ""
//   length(m       => length, m   // As a special case handles units specified without a closing paren
//   s:length(mm    => length, mm // This entry contains a keyword
func cell2NameAndUnit(cell string) (name, unit string) {
	name = ""
	unit = ""

	// Check for the default case of an empty cell
	if cell == "" {
		return name, unit
	}

	// Handle the case where there is a keyword
	i := strings.Index(cell, ":")
	if i != -1 {
		// Given a string like:
		//   sample:time(h) => "time(h)"
		//   sample:  time(h) => "time(h)"
		cell = strings.TrimSpace(cell[i+1:])
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
