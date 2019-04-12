package spreadsheet

type ColumnAttributeType int

const (
	SampleAttributeColumn = iota + 1
	ProcessAttributeColumn
	FileAttributeColumn
	UnknownAttributeColumn
)

func (c ColumnAttributeType) String() string {
	switch c {
	case SampleAttributeColumn:
		return "SampleAttributeColumn"
	case ProcessAttributeColumn:
		return "ProcessAttributeColumn"
	case FileAttributeColumn:
		return "FileAttributeColumn"
	default:
		return "UnknownAttributeColumn"
	}
}
