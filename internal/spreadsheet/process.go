package spreadsheet

type Process struct {
	Name       string
	Index      int
	Attributes []Attribute
	Samples    []Sample
}

func (p *Process) AddSample(sample Sample) {
	p.Samples = append(p.Samples, sample)
}

func (p *Process) AddAttribute(attribute Attribute) {
	p.Attributes = append(p.Attributes, attribute)
}
