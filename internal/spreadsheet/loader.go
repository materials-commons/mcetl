package spreadsheet

import (
	"github.com/360EntSecGroup-Skylar/excelize"

	"github.com/materials-commons/mcetl/internal/spreadsheet/model"
)

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

func loadWorksheet(xlsx *excelize.File, worksheetName string, index int) (*model.Process, error) {
	process := &model.Process{
		Name:  worksheetName,
		Index: index,
	}
	rows, err := xlsx.Rows(process.Name)
	if err != nil {
		return process, err
	}

	rowHeaders := true
	row := 0
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
