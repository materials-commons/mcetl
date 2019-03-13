package model

type Sample struct {
	Name         string
	Parent       string
	Row          int
	Attributes   []*Attribute
	ProcessAttrs []*Attribute
}

func (s *Sample) AddAttribute(attribute *Attribute) {
	s.Attributes = append(s.Attributes, attribute)
}

func (s *Sample) AddProcessAttribute(attribute *Attribute) {
	s.ProcessAttrs = append(s.ProcessAttrs, attribute)
}

func NewSample(name string, row int) *Sample {
	return &Sample{
		Name: name,
		Row:  row,
	}
}
