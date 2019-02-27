package spreadsheet

type Sample struct {
	Name       string
	Parent     string
	Attributes []Attribute
}

func (s *Sample) AddAttribute(attribute Attribute) {
	s.Attributes = append(s.Attributes, attribute)
}
