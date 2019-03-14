package model

// Process represents a single worksheet in excel. Each worksheet
// specifies a process template and the samples. Since the worksheet
// is a model for the process that means that multiple processes
// can come from that single worksheet. Each "process" from the
// worksheet is the same nominal type but with potentially different
// process attributes.
//
// This structure captures the process attributes in the header
// in the Attributes array. This array has no values associated
// with, it simply contains the name and units for each of the
// attributes.
//
// The same is true for the SampleAttrs array. Since a spreadsheet
// has a header row that specifies the process and sample attributes
// the attributes for all the samples in the spreadsheet are represented
// in the SampleAttrs array. Like the Attributes array it only contains
// the attribute name and unit.
//
// The Samples array contains all the information about the samples
// in the worksheet. Because this is row oriented it means that samples
// in a row will also have process attributes associated with them. These
// get stored in the sample and help to determine whether or not to create
// a new process for a sample. For example imagine the following worksheet
// where the first row is the headers. Remember that column 1 is sample,
// column 2 is sample parent, and that sample attributes come after the
// first blank header column. This worksheet has the name "Heat Treatment".
//     sample   parent   proc att    proc att        sample att      sample att
//   |Name   |Parent   |Time(s)   |Temperature(c)| |Grain Size(mm)|Composition(at%)|
//   |S1     |         | 300      |400           | | 2mm          |mg 20           |
//   |S2     |         |          |              | | 1mm          |mg 19.8         |
//   |S3     |         | 500      |50            | | 1mm          |al 30           |
//
// Here we have 3 samples. Samples S1 and S2 will share the same process because S1 will
// create a new process, and S2 will use that process because its process attributes
// (Time and Temperature) are blank, meaning use the last processes attributes. S3 will
// create a new process with new Time and Temperature process attributes.
//
// After this is parsed the data structure will look as follows:
//
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
