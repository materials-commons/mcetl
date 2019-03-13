package model

type Attribute struct {
	Name   string
	Unit   string
	Column int
	//Value  map[string]interface{}
	Value string
}

func NewAttribute(name, unit string, column int) *Attribute {
	return &Attribute{
		Name:   name,
		Unit:   unit,
		Column: column,
	}
}
