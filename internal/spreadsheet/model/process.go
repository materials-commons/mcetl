package model

type Process struct {
	Name        string
	Index       int
	Attributes  []*Attribute
	Samples     []*Sample
	SampleAttrs []*Attribute
}

func (p *Process) AddSample(sample *Sample) {
	p.Samples = append(p.Samples, sample)
}

func (p *Process) AddSampleAttr(attribute *Attribute) {
	p.SampleAttrs = append(p.SampleAttrs, attribute)
}

func (p *Process) AddAttribute(attribute *Attribute) {
	p.Attributes = append(p.Attributes, attribute)
}
