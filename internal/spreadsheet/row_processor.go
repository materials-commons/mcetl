package spreadsheet

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/materials-commons/mcetl/internal/spreadsheet/model"
	"github.com/pkg/errors"
)

// rowProcessor handles processing of each row of a worksheet
type rowProcessor struct {
	// worksheet is the worksheet to load the excel worksheet into
	worksheet *model.Worksheet

	// Is column 2 the parent column?
	HasParent bool

	// columnType is built while processing the header row. It maps each
	// column to its column type (process attribute, sample attribute or file)
	columnType map[int]ColumnAttributeType

	// converter is used to convert sample or process attribute cells that
	// aren't blank into their relevant type (float, object, int, etc...)
	converter *cellConverter
}

func newRowProcessor(worksheetName string, hasParent bool, index int) *rowProcessor {
	return &rowProcessor{
		worksheet: &model.Worksheet{
			Name:  worksheetName,
			Index: index,
		},
		HasParent:  hasParent,
		converter:  newCellConverter(),
		columnType: make(map[int]ColumnAttributeType),
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
		// Check for columns to skip. Column 1 is sample name and column 2
		// could be the parent column. Always skip column 1, and optionally
		// skip column 2 if HasParent is true (indicating that column 2 is
		// used to construct the workflow).
		if column < 3 && r.HasParent {
			// column one is sample name
			// column two is parent sample
			continue
		} else if column < 2 {
			// column one is sample name
			// there is nothing special about column 2
			continue
		}

		if colCell == "" {
			// blank cell so nothing to process
			continue
		}

		// If you add a new type of keyword then don't forget to modify processSampleRow() case statement to handle
		// that keyword.

		switch columnAttributeTypeFromKeyword(colCell) {
		case ProcessAttributeColumn:
			name, unit := cell2NameAndUnit(colCell)
			attr := model.NewAttribute(name, unit, column)
			r.columnType[column] = ProcessAttributeColumn
			r.worksheet.AddProcessAttr(attr)
		case SampleAttributeColumn:
			name, unit := cell2NameAndUnit(colCell)
			attr := model.NewAttribute(name, unit, column)
			r.columnType[column] = SampleAttributeColumn
			r.worksheet.AddSampleAttr(attr)
		case FileAttributeColumn:
			fileHeader := createFileHeader(colCell, column)
			r.worksheet.AddFileHeader(fileHeader)
			r.columnType[column] = FileAttributeColumn
		case IgnoreAttributeColumn:
			r.columnType[column] = IgnoreAttributeColumn
		default:
			fmt.Printf("Warning: Worksheet %s heading column %d with value '%s' has unknown keyword to identify its type\n", r.worksheet.Name, column, colCell)
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
func (r *rowProcessor) processSampleRow(row *excelize.Rows, rowIndex int) error {
	column := 0
	var currentSample *model.Sample = nil

	for _, colCell := range row.Columns() {
		colCell = strings.TrimSpace(colCell)
		column++

		// Go through each column and capture its value. Column 1 is special and column
		// 2 may be special. If HasParent flag is true then column two is treated as a special column.
		// Empty cell handling is column specific. If column 1 cell is blank then we
		// skip the entire row. Column 2 is allowed to have blank columns. The rest
		// of the columns can have blank cells, but if they are blank we just skip
		// processing them. Since the rest of the columns represent various types of
		// attributes (process, sample or file) skipping blank cells prevents empty
		// attributes from being created on the server.

		if column == 1 {
			// Sample

			if colCell == "" {
				// No sample is listed in this column. Just skip the entire row.
				return nil
			}
			currentSample = model.NewSample(colCell, rowIndex)
			r.worksheet.AddSample(currentSample)
		} else if column == 2 && r.HasParent {
			// parent worksheet
			currentSample.Parent = colCell
		} else {
			// Process, Sample or File attribute

			// Column 1 is sample, column 2 is parent worksheet if HasParent is true, otherwise it is an attribute
			// column. All the other columns are attributes.
			//
			// At this point the header row has been processed (in processHeaderRow()). The rowProcessor identified
			// each of the header columns by their type (process, sample or file attribute). As we walk through the
			// columns that make up a row we refer back to the rowProcessor columnType which will tell us which type
			// of attribute we are looking at.
			colType, ok := r.columnType[column]

			if isBlank(colCell) {
				// This column cell is blank so skip processing. This way empty attributes
				// are not tracked and loaded onto the server.
				continue
			}

			switch {
			case !ok:
				// Couldn't find column type. This means the spreadsheet contains header columns with unknown keywords.
				continue

			case colType == SampleAttributeColumn:
				// This column is a sample attribute. Look up the given header information (stored in the
				// worksheet.SampleAttrs) so that we know which attribute we are looking at for this cell.
				// Ignore cells that are blank.
				attr := findAttr(r.worksheet.SampleAttrs, column)
				sampleAttr := model.NewAttribute(attr.Name, attr.Unit, attr.Column)

				if val, err := r.converter.cellToJSONMap(colCell); err != nil {
					errDesc := fmt.Sprintf("Error converting cell in worksheet %s: row: %d, column: %d with value %s",
						r.worksheet.Name, rowIndex, column, colCell)
					return errors.Wrapf(err, errDesc)
				} else {
					sampleAttr.Value = val
				}

				currentSample.AddAttribute(sampleAttr)

			case colType == ProcessAttributeColumn:
				// This column is a process attribute. As above look up the header so we know the attribute
				// associated with this cell. Ignore cells that are blank.
				attr := findAttr(r.worksheet.ProcessAttrs, column)
				processAttr := model.NewAttribute(attr.Name, attr.Unit, attr.Column)

				if val, err := r.converter.cellToJSONMap(colCell); err != nil {
					errDesc := fmt.Sprintf("Error converting cell in worksheet %s: row: %d, column: %d with value '%s'",
						r.worksheet.Name, rowIndex, column, colCell)
					return errors.Wrapf(err, errDesc)
				} else {
					processAttr.Value = val
				}

				currentSample.AddProcessAttribute(processAttr)

			case colType == FileAttributeColumn:
				fileHeader := findFileHeader(r.worksheet.FileHeaders, column)
				currentSample.AddFile(cell2Filepath(colCell, fileHeader), column)

			case colType == IgnoreAttributeColumn:
				// Ignore all values in this column
				continue

			default:
				// If we are here then what happened is that a new column type was created and added
				// into processHeaderRow(), but this case statement wasn't extended to handle that
				// column type.
				fmt.Printf("Bug: processHeaderRow() contains a new column type that isn't in processSampleRow. Cell with unknown header type %s, column %d\n", colCell, column)
			}
		}
	}

	return nil
}

// findAttr will look up the attribute in the given list of attributes. These attributes were built
// during the header processing. Each attribute has a column it is associated with and we can use that
// to find the given attribute in the header.
func findAttr(attributes []*model.Attribute, column int) *model.Attribute {
	for _, attr := range attributes {
		if attr.Column == column {
			return attr
		}
	}

	return nil
}

// findFileHeader will look up the file header in the given list of file headers. The file headers
// were built during the header processing stage. Each file header has a column associated with it
// and this method matches on the column to find the given file header.
func findFileHeader(fileHeaders []*model.FileHeader, column int) *model.FileHeader {
	for _, fileHeader := range fileHeaders {
		if fileHeader.Column == column {
			return fileHeader
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

	cell = strings.TrimSpace(cell)
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
		return strings.TrimSpace(name), strings.TrimSpace(unit)
	default:
		// indexClosingParen == -1, which means we have a string like: abc(c
		// that has no closing paren
		name = cell[:indexOpeningParen]
		unit = cell[indexOpeningParen+1:]
		return strings.TrimSpace(name), strings.TrimSpace(unit)
	}
}

// createFileHeader parses the cell for a file header. The format of a cell
// is keyword:description:path, keyword:path or keyword:.
func createFileHeader(cell string, column int) *model.FileHeader {
	// Example of parsing:
	//
	// fullCell := "file:abc:path/"
	// partialCell := "file:path/"
	//
	// firstColon := strings.Index(fullCell, ":")
	// secondColon := strings.LastIndex(fullCell, ":")
	// fmt.Printf("Description: '%s', Path: '%s'\n", fullCell[firstColon+1:secondColon], fullCell[secondColon+1:])
	//    => Description: 'abc', Path: 'path/'
	//
	// firstColon = strings.Index(partialCell, ":")
	// secondColon = strings.LastIndex(partialCell, ":")
	// fmt.Printf("Description: '%s', Path: '%s'\n", "", partialCell[firstColon+1:])
	//    => Description: '', Path: 'path/'
	//

	firstColon := strings.Index(cell, ":")
	secondColon := strings.LastIndex(cell, ":")
	if firstColon != secondColon {
		// if firstColon != secondColon then there is a description and a path
		// ie, the format is:  FILE:My description:directory-path/to/file/in/cell/in/materials-commons
		return model.NewFileHeader(cell[firstColon+1:secondColon], strings.TrimSpace(cell[secondColon+1:]), column)
	}

	// If we are here then firstColon == secondColon, which means the format is:
	// FILE:directory-path/to/file/in/cell/in/materials-commons
	return model.NewFileHeader("", strings.TrimSpace(cell[firstColon+1:]), column)
}

// cell2Filepath converts a given cell into a file path. It does this by first checking
// if the cell contains a '/', if it does then the cell is assumed to be a full path. If
// it doesn't then the cell references the file name and the path is derived from the
// fileHeader. If fileHeader is nil then it is ignored.
func cell2Filepath(cell string, fileHeader *model.FileHeader) string {
	i := strings.Index(cell, "/")
	if i != -1 {
		// The cell contains a '/' character so it is path, just
		// return the cell
		return cell
	}

	// If we are here then there was no '/' in the cell, so the cell just contains a
	// filename. Join the fileHeader Path to the file name if fileHeader isn't nil,
	// otherwise just return the cell.
	if fileHeader != nil {
		return filepath.Join(fileHeader.Path, cell)

	}

	return cell
}
