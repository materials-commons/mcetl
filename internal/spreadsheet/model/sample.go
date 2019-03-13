package model

type Sample struct {
	Name       string
	Parent     string
	Row        int
	Attributes []*Attribute
}

func (s *Sample) AddAttribute(attribute *Attribute) {
	s.Attributes = append(s.Attributes, attribute)
}

func NewSample(name string, row int) *Sample {
	return &Sample{
		Name: name,
		Row:  row,
	}
}
