package spreadsheet

type ColumnAttributeType int

const (
	SampleAttributeColumn = iota + 1
	ProcessAttributeColumn
	FileAttributeColumn
	IgnoreAttributeColumn
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
	case IgnoreAttributeColumn:
		return "IgnoreAttributeColumn"
	default:
		return "UnknownAttributeColumn"
	}
}
