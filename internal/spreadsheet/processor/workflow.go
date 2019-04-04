package processor

import (
	"crypto/sha256"
	"fmt"

	"github.com/materials-commons/gomcapi"
	"github.com/materials-commons/mcetl/internal/spreadsheet/model"
)

type Workflow struct {
	root                   []*WorkflowProcess
	existingSamples        map[string]*model.Sample
	uniqueProcessInstances map[string]*WorkflowProcess
}

type WorkflowProcess struct {
	Worksheet *model.Worksheet
	Samples   []*model.Sample
	Out       []*mcapi.Sample
	Process   *mcapi.Process
	To        []*WorkflowProcess
	From      []*WorkflowProcess
}

func newWorkflowProcess() *WorkflowProcess {
	return &WorkflowProcess{}
}

func newWorkflow() *Workflow {
	return &Workflow{
		existingSamples:        make(map[string]*model.Sample),
		uniqueProcessInstances: make(map[string]*WorkflowProcess),
	}
}

func (w *Workflow) constructWorkflow(worksheets []*model.Worksheet) {
	// 1. Top level processes are all create sample processes
	w.createSampleProcesses(worksheets)

	// 2. Create a map containing all the unique processes
	w.createUniqueProcessesMap(worksheets)

	// 3. Connect processes by going through the worksheet and looking at parent attributes.
	//    The parent will point to a sample on a worksheet, which means, for our purposes,
	//    that is the process that is sending that sample instance into this process.
	w.wireupWorkflow(worksheets)
}

// createSampleProcesses goes through all the worksheets and identifies all the
// samples that need to be created.
func (w *Workflow) createSampleProcesses(worksheets []*model.Worksheet) {
	// Build up a list of unique samples that need to be created
	for _, worksheet := range worksheets {
		for _, sample := range worksheet.Samples {
			if _, ok := w.existingSamples[sample.Name]; !ok {
				w.existingSamples[sample.Name] = sample
			}
		}
	}

	// Now add all those as top level nodes in the root. These are all "out" samples.
	for sampleName := range w.existingSamples {
		node := newWorkflowProcess()
		node.Samples = append(node.Samples, w.existingSamples[sampleName])
		w.root = append(w.root, node)
	}
}

// createUniqueProcessesMap goes through the worksheet and identifies all the unique process
// instances that need to be created. For example, in a worksheet a process will be created
// whenever the process attributes in that worksheet are uniquely specified. So a particular
// worksheet can result in multiple processes being created, even though each process will
// be of the same "type".
func (w *Workflow) createUniqueProcessesMap(worksheets []*model.Worksheet) {
	for _, worksheet := range worksheets {
		for _, sample := range worksheet.Samples {
			key := makeSampleInstanceKey(sample, worksheet.Name)
			if wp, ok := w.uniqueProcessInstances[key]; !ok {
				// There is no instance for this process so create it
				wp := newWorkflowProcess()
				wp.Worksheet = worksheet
				wp.Samples = append(wp.Samples, sample)
				w.uniqueProcessInstances[key] = wp
			} else {
				// There is an existing process instance, that means we've encountered this
				// a second time sample/worksheet combination before. When this happens
				// additional matches don't mean a new sample/process but rather that we
				// are going to add additional measures to the existing sample/process.
				wp.Samples = append(wp.Samples, sample)
			}
		}
	}
}

// wireupWorkflow walks through the worksheets and the unique list of processes looking for the parent
// attribute in the worksheets. The parent attribute is used to wire two processes together. For example
// given:
//   Worksheet: CT
//    Sample1 Parent: SEM
//   Worksheet: SEM
//     Sample1
// This will create a workflow that looks as follows SEM->CT. Where -> is S1 from SEM going into the CT process.
// There is one special case. Samples need to be created. Each Created sample belongs to a "Create Samples"
// process. The "Created Samples" process is not actually in the spreadsheet. Instead it is implicitly in there and
// is identified by unique sample names.
//
// Before wireupWorkflow is run the method "createSampleProcesses" runs and creates these process nodes and puts
// them in the root. Thus any sample that doesn't have an actual parent in the spreadsheet implicitly has a parent
// that is pointing to a "Create Samples" process. In the code you can see this where we check for sample.Parent == "".
func (w *Workflow) wireupWorkflow(worksheets []*model.Worksheet) {
	for _, worksheet := range worksheets {
		for _, sample := range worksheet.Samples {
			// Parent is blank then the input sample is from the original list of created samples
			if sample.Parent == "" {
				wp := w.findMatchingCreateSampleProcess(sample.Name)
				if wp == nil {
					// Should never happen
					fmt.Println("Can't find matching create sample process for ", sample.Name)
				} else {
					key := makeSampleInstanceKey(sample, worksheet.Name)
					if instance, ok := w.uniqueProcessInstances[key]; !ok {
						fmt.Printf("Can't find matching process to wire up %s %#v\n", worksheet.Name, sample)
					} else {
						instance.From = append(instance.From, wp)
						wp.To = append(wp.To, instance)
					}
				}
			} else {
				// pointer to parent so we need to find that item
				wp := w.findMatchingEntry(sample.Name, sample.Parent, worksheets)
				if wp == nil {
					// Should never happen
					fmt.Println("Can't find matching create sample process for ", sample.Name)
				} else {
					key := makeSampleInstanceKey(sample, worksheet.Name)
					if instance, ok := w.uniqueProcessInstances[key]; !ok {
						fmt.Printf("Can't find matching process to wire up %s %#v\n", worksheet.Name, sample)
					} else {
						instance.From = append(instance.From, wp)
						wp.To = append(wp.To, instance)
					}
				}
			}
		}
	}
}

func (w *Workflow) findMatchingCreateSampleProcess(sampleName string) *WorkflowProcess {
	for _, wp := range w.root {
		for _, sample := range wp.Samples {
			if sample.Name == sampleName {
				return wp
			}
		}
	}

	return nil
}

func (w *Workflow) findMatchingEntry(sampleName, worksheetName string, worksheets []*model.Worksheet) *WorkflowProcess {
	for _, worksheet := range worksheets {
		if worksheet.Name == worksheetName {
			for _, sample := range worksheet.Samples {
				if sample.Name == sampleName {
					key := makeSampleInstanceKey(sample, worksheetName)
					if instance, ok := w.uniqueProcessInstances[key]; !ok {
						return nil
					} else {
						return instance
					}
				}
			}
		}
	}

	return nil
}

func makeSampleInstanceKey(sample *model.Sample, starting string) string {
	key := starting
	for _, attr := range sample.ProcessAttrs {
		key = fmt.Sprintf("%s%s%#v", key, attr.Unit, attr.Value)
	}

	return fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
}
