package model

// Worksheet represents a single worksheet in excel. Each worksheet
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
// column 2 is sample parent, and that sample and process attributes have
// keywords so we can tell them apart. Look in loader.go to see the keywords.
// This worksheet has the name "Heat Treatment".
//     sample   parent   proc att      proc att        sample att      sample att
//   |Name   |Parent   |p:Time(s) |p:Temperature(c)|s:Grain Size(mm)|s:Composition(at%)|
//   |S1     |         | 300      |400             | 2              |mg 20             |
//   |S2     |         | 300      |400             | 1              |mg 19.8           |
//   |S3     |         | 500      |50              | 1              |al 30             |
//
// Here we have 3 samples. Samples S1 and S2 will share the same process because S1 will
// create a new process, and S2 will use that process because its process attributes
// (Time and Temperature) are the same as the previous process. S3 will create a new
// process with new Time and Temperature process attributes.
//
// After this is parsed the data structure will look as follows:
//     struct Worksheet {
//          Name: "Heat Treatment"
//          Index: 1 // Index of Worksheet
//          Attributes: [{Name: "Time", Unit: "s"},{Name:"Temperature", Unit: "c"}]
//          SampleAttrs: [{Name: "Grain Size", Unit: "mm"}, {Name: "Composition": Unit: "at%"}]
//          Samples: [
//              {
//                   Name: "S1",
//                   ProcessAttrs: [{Name: "Time", Value: 300, Unit: "s"}, {Name: "Temperature", Value: 400, Unit:"c"}],
//                   Attributes: [{Name: "Grain Size", Value: 2, Unit: "mm"}, {Name: "Composition", Value: "mg 20", Unit: "at%"}]
//              },
//                   Name: "S2",
//                   ProcessAttrs: [{Name: "Time", Value: 300, Unit: "s"}, {Name: "Temperature", Value: 400, Unit:"c"}],
//                   Attributes: [{Name: "Grain Size", Value: 1, Unit: "mm"}, {Name: "Composition", Value: "mg 19.8", Unit: "at%"}]
//              },
//              {
//                   Name: "S3",
//                   ProcessAttrs: [{Name: "Time", Value: 500, Unit: "s"}, {Name: "Temperature", Value: 50, Unit:"c"}],
//                   Attributes: [{Name: "Grain Size", Value: 1, Unit: "mm"}, {Name: "Composition", Value: "al 30", Unit: "at%"}]
//              },
//
// Because S2 Worksheet Attrs are the same as the last created process it will be associated with the process created from S1.
// However S3 has different values in ProcessAttrs so it will create a new process and the sample will be associated with it.
type Worksheet struct {
	Name         string
	Index        int
	ProcessAttrs []*Attribute
	Samples      []*Sample
	SampleAttrs  []*Attribute
}

func (w *Worksheet) AddSample(sample *Sample) {
	w.Samples = append(w.Samples, sample)
}

func (w *Worksheet) AddSampleAttr(attribute *Attribute) {
	w.SampleAttrs = append(w.SampleAttrs, attribute)
}

func (w *Worksheet) AddProcessAttr(attribute *Attribute) {
	w.ProcessAttrs = append(w.ProcessAttrs, attribute)
}

/////////////////////////////////////////////////////////////////

type Sample struct {
	Name         string
	Parent       string
	Row          int
	Attributes   []*Attribute
	ProcessAttrs []*Attribute
	Files        []File
}

type File struct {
	Path   string
	Column int
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

func (s *Sample) AddFile(path string, column int) {
	file := File{Path: path, Column: column}
	s.Files = append(s.Files, file)
}

/////////////////////////////////////////////////////////////////

type Attribute struct {
	Name   string
	Unit   string
	Column int
	Value  map[string]interface{}
}

func NewAttribute(name, unit string, column int) *Attribute {
	return &Attribute{
		Name:   name,
		Unit:   unit,
		Column: column,
	}
}
