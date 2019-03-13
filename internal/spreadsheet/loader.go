package spreadsheet

import (
	"github.com/360EntSecGroup-Skylar/excelize"

	"github.com/materials-commons/mcetl/internal/spreadsheet/model"
)

// Load will load the given excel file. This assumes that each process is in a separate
// worksheet and the process will take on the name of the worksheet.
func Load(path string) ([]*model.Process, error) {
	var processes []*model.Process
	xlsx, err := excelize.OpenFile(path)
	if err != nil {
		return processes, err
	}

	var savedErr error
	for index, name := range xlsx.GetSheetMap() {
		process, err := loadWorksheet(xlsx, name, index)
		if err != nil {
			savedErr = err
			continue
		}
		processes = append(processes, process)
	}

	return processes, savedErr
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
//
// Attributes have the following format: name(unit)
// For example:
//    temperature(c)   - Attribute name temperature with unit c
//    length           - Attribute length with no unit
func loadWorksheet(xlsx *excelize.File, worksheetName string, index int) (*model.Process, error) {
	process := &model.Process{
		Name:  worksheetName,
		Index: index,
	}
	rows, err := xlsx.Rows(process.Name)
	if err != nil {
		return process, err
	}

	rowHeaders := true // Start off processing row headers
	row := 0

	// Sample attributes start at column 4 or greater. Remember the format is:
	// |sample|parent sample|<blank>|sample attr
	// That is there must be a blank column before sample attributes start. The
	// column that this starts can be greater than 4 if there are process attributes.
	// For example the following example would have startingSampleAttrsCol = 6
	//      1        2             3           4         5   6
	//   |sample|parent sample|process attr|process attr||sample attr
	// where 5 is our blank column
	startingSampleAttrsCol := 4
	for rows.Next() {
		row++
		if rowHeaders {
			column := 0
			inProcessAttrs := true
			for _, colCell := range rows.Columns() {
				column++
				if column < 3 {
					// column one is sample name
					// column two is parent sample
					continue
				}

				if colCell == "" && inProcessAttrs {
					inProcessAttrs = false
					startingSampleAttrsCol = column + 1
				} else if inProcessAttrs {
					attr := model.NewAttribute(colCell, "", column)
					process.AddAttribute(attr)
				} else {
					attr := model.NewAttribute(colCell, "", column)
					process.AddSampleAttr(attr)
				}
			}
			rowHeaders = false
		} else {
			column := 0
			inProcessAttrs := true
			var currentSample *model.Sample

			currentSample = nil
			for _, colCell := range rows.Columns() {
				column++
				if column == 1 {
					currentSample = model.NewSample(colCell, row)
					process.AddSample(currentSample)
					// Sample
				} else if column == 2 {
					// parent sample
					currentSample.Parent = colCell
				} else if colCell == "" && inProcessAttrs {
					inProcessAttrs = false
				} else if inProcessAttrs {
					// Not sure what to do here
				} else {
					// in sample attrs
					attr := process.SampleAttrs[column-startingSampleAttrsCol]
					sampleAttr := model.NewAttribute(attr.Name, attr.Unit, attr.Column)
					sampleAttr.Value = colCell
					currentSample.AddAttribute(sampleAttr)
				}
			}
		}
	}

	return process, nil
}
