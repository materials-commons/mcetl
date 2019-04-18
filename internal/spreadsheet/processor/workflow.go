package processor

/*
 * The workflow module turns the set of processed worksheets into a workflow. This workflow includes
 *
 * the initial "Create Samples" processes to create a sample. This module goes through 3 steps to
 * construct the workflow.
 *
 * The first step is to go through all the worksheets identifying the unique sample names. Each of these
 * is turned into a create samples process that sits at the root of each of the workflows.
 *
 * The second step involves creating all the processes that are in the worksheets. This involves going
 * through each worksheet and finding each worksheets sample/process attribute unique combinations. This
 * forms the set of processes that need to be created.
 *
 * For example imagine a worksheet that has the following (sample name, Process Attr):
 *  Worksheet: Heat Treatment
 *    S1 Temp:400
 *    S2 Temp:400
 *    S3 Temp:500
 *
 * This worksheet would create 2 processes because there are two sets of unique Process Parameters, one set at Temp:400
 * and one set at Temp:500. So the second step goes through each of the worksheet identifying all these processes
 * to be created.
 *
 * The 3rd and final step is to wire all the processes together. In the second step a map of all the unique processes
 * was created. For the 3rd step we again walk through the worksheets looking at the parent (2nd) column and use that
 * plus the properties to connect each of the processes together. A parent means that the referenced process/sample is
 * an input into the process with the parent attribute on it.
 * For example imagine the following (Sample name, Parent, Process Attr)
 *  Worksheet: Heat Treatment
 *    S1  "" Temp:400
 *
 *  Worksheet: SEM
 *    S1 "Heat Treatment" Grain Size:400
 *
 * Here SEM has a parent of Heat Treatment. This means that the sample S1 from Heat Treatment is an input into SEM, ie
 *   Heat Treatment --S1--> SEM
 *
 */

import (
	"crypto/sha256"
	"fmt"

	"github.com/materials-commons/gomcapi"
	"github.com/materials-commons/mcetl/internal/spreadsheet/model"
)

// Workflow describes the entire workflow for the set of worksheets being processed
type Workflow struct {
	// root is the starting point of each of the individual workflows. Each of these top level WorkflowProcess entries
	// will be a Create Samples entry for creating each of the samples.
	root []*WorkflowProcess

	// existingSamples is used to track each of the unique sample instances that need to be created.
	existingSamples map[string]*model.Sample

	// uniqueProcessInstances contains each of the unique processes (except for the Create Sample processes that
	// start in the root). These are all the actual processes that are directly identified in the worksheets. Since
	// Create Samples are implicitly defined, they aren't directly in the worksheets (that is there is no worksheet
	// that contains the create samples.
	uniqueProcessInstances map[string]*WorkflowProcess

	HasParent bool
}

// WorkflowProcess is a unique process step. Each process step contains all the samples associated with that
// step. And pointers to/from downstream/upstream processes. For ease of use in later modules this data structure
// also has placeholders for tracking the actual process and samples that are created on the server.
type WorkflowProcess struct {
	// The worksheet that this process came from
	Worksheet *model.Worksheet

	// Unique Process Key
	Key string

	// Sample that this process was computed for
	SampleName string

	// All the samples in the worksheet that involve this process. Remember that there are process attributes
	// attached to these samples. All the samples in this list will have the same process attributes.
	Samples []*model.Sample

	// The server side samples that are output by this process. For Create Samples this is new samples that
	// are created. For other processes this is updated samples with a new PropertySetID. This is not used
	// in this module but instead used in the creater when the workflow is turned into actual server side entities.
	Out []*mcapi.Sample

	// The server side process that this represents. This is not used in this module but instead used in the creater
	// when the workflow is turned into actual server side entities.
	Process *mcapi.Process

	// Workflow processes that send samples into this process. Essentially forward links for a linked list.
	To []*WorkflowProcess

	// Workflow processes that use samples from this process. Essentially backward links for a linked list.
	From []*WorkflowProcess
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

// constructWorkflow creates the workflow as described in the module following the 3 outlined steps.
func (w *Workflow) constructWorkflow(worksheets []*model.Worksheet) {
	// 1. Top level processes are all create sample processes
	w.createSampleProcesses(worksheets)

	// 2. Create a map containing all the unique processes
	w.createUniqueProcessesMap(worksheets)

	// 3. Connect processes by going through the worksheet and looking at the parent attribute.
	//    The parent will point to a sample on a worksheet, which means, for our purposes,
	//    that is the process that is sending that sample instance into this process.
	w.wireupWorkflow(worksheets)
}

// createSampleProcesses goes through all the worksheets and identifies all the
// samples that need to be created. It then adds them to the root field in the workflow.
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
			// Create a unique key for this process. This key is constructed based on the worksheet
			// name and the process attributes. This allows us to track all the unique process instances.
			key := w.makeSampleInstanceKey(sample, worksheet.Name)
			if wp, ok := w.uniqueProcessInstances[key]; !ok {
				// There is no instance for this process so create it and insert it into uniqueProcessInstances
				wp := newWorkflowProcess()
				wp.SampleName = sample.Name
				wp.Key = key
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
	var parentProcess *WorkflowProcess

	for _, worksheet := range worksheets {
		for _, sample := range worksheet.Samples {

			// First get the process from the worksheet that we are sending the sample to
			uniqueProcessFromWorksheet := w.findProcessFromSampleInWorksheet(sample, worksheet.Name)
			if uniqueProcessFromWorksheet == nil {
				// If this happens then we have a bug in the code for creating all the unique process instances
				// because this means we've found a process that isn't in that map.
				fmt.Printf("Can't find matching process to wire up %s %#v\n", worksheet.Name, sample)
				continue
			}

			// If Parent is blank then the input sample is from the original list of created samples
			if sample.Parent == "" {
				// Find the create sample process that is going to feed the sample into this process.
				parentProcess = w.findMatchingCreateSampleProcess(sample.Name)
			} else {
				// If we are here then sample.Parent in the worksheet is not blank. So we need to find the
				// process that Parent points to.
				parentProcess = w.findMatchingEntry(sample.Name, sample.Parent, worksheets)
			}

			if parentProcess == nil {
				// Should never happen
				fmt.Println("Can't find matching create sample process for ", sample.Name)
				continue
			}

			w.wireProcessesTogetherFromTo(parentProcess, uniqueProcessFromWorksheet)
		}
	}
}

// wireProcessesTogetherFromTo wires the processes together point correctly setting up the links
// in both directions.
func (w *Workflow) wireProcessesTogetherFromTo(fromProcess, toProcess *WorkflowProcess) {
	toProcess.From = append(toProcess.From, fromProcess)
	fromProcess.To = append(fromProcess.To, toProcess)
}

// findProcessFromSampleInWorksheet creates the unique name to look up a process process in uniqueProcessInstances.
func (w *Workflow) findProcessFromSampleInWorksheet(sample *model.Sample, worksheetName string) *WorkflowProcess {
	key := w.makeSampleInstanceKey(sample, worksheetName)
	if instance, ok := w.uniqueProcessInstances[key]; !ok {
		fmt.Printf("Can't find matching process to wire up %s %#v\n", worksheetName, sample)
		return nil
	} else {
		return instance
	}
}

// findMatchingCreateSampleProcess looks at the root workflow processes. All these processes are create sample
// processes. Then for each of these it looks at the samples for that process until it finds a sample that
// matches the sampleName passed in. Once there is a match we've found the Create Samples process that the named
// sample comes from.
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

// findMatchingEntry finds the workflow process that matches the given sample in a worksheet. It first goes
// through all the worksheets finding the worksheet (by name) then it goes through the samples in that worksheet
// and for each sample that matches the sampleName it creates the unique key to look up the process in the
// uniqueProcessInstances map. This should always find a match.
func (w *Workflow) findMatchingEntry(sampleName, worksheetName string, worksheets []*model.Worksheet) *WorkflowProcess {
	for _, worksheet := range worksheets {
		if worksheet.Name == worksheetName {
			for _, sample := range worksheet.Samples {
				if sample.Name == sampleName {
					key := w.makeSampleInstanceKey(sample, worksheetName)
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

// makeSampleInstanceKey creates the unique key for a sample and its process attributes, this key
// is used to store the unique processes. A key is constructed from the sample name and all its
// process attributes. We then run sha256 on it and get the hex key to create the unique key for
// that combination.
func (w *Workflow) makeSampleInstanceKey(sample *model.Sample, starting string) string {
	key := starting
	for _, attr := range sample.ProcessAttrs {
		key = fmt.Sprintf("%s%s%#v", key, attr.Unit, attr.Value)
	}

	if !w.HasParent {
		for _, attr := range sample.Attributes {
			key = fmt.Sprintf("%s%s%#v", key, attr.Unit, attr.Value)
		}
	}

	key = fmt.Sprintf("%s%s", sample.Name, key)

	return fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
}
