package processor

//
//

import (
	"fmt"

	"github.com/materials-commons/gomcapi"
	"github.com/materials-commons/mcetl/internal/spreadsheet/model"
)

// Creater holds the state needed to create the workflow on the server.
type Creater struct {
	// The project we are adding to
	ProjectID string

	// Name of the experiment to create
	Name string

	// Description of the experiment to create
	Description string

	// The created experiment's ID. This and ProjectID are needed
	// for many of the mcapi REST calls.
	ExperimentID string

	// Does the second column represent the parent column that points to other worksheets. This allows the user
	// to construct a workflow graph.
	HasParent bool

	// Total number of API calls made
	Count int

	// Counts by API call
	ByCallCounts map[string]int

	client *mcapi.Client
}

func NewCreater(projectID, name, description string, client *mcapi.Client) *Creater {
	return &Creater{
		ProjectID:    projectID,
		Name:         name,
		Description:  description,
		client:       client,
		ByCallCounts: make(map[string]int),
	}
}

// Apply implements the Process interface. This version creates the workflow on the server.
func (c *Creater) Apply(worksheets []*model.Worksheet) error {
	// 1. Create the experiment on the server to load the workflow into.
	if err := c.createExperiment(); err != nil {
		return nil
	}

	// 2. Create the workflow from the worksheets
	wf := newWorkflow()
	wf.HasParent = c.HasParent

	wf.constructWorkflow(worksheets)

	// 3. Walk through the workflow creating each of the steps.
	for _, wp := range wf.root {
		if err := c.createWorkflowSteps(wp); err != nil {
			// Even though there were errors the experiment loading is no longer "in progress", so
			// adjust its status. Ignore errors as there is nothing we can do if this fails.
			var _ = c.client.UpdateExperimentProgressStatus(c.ProjectID, c.ExperimentID, false)
			return err
		}
	}

	fmt.Println("Total calls:", c.Count)
	fmt.Printf("%#v\n", c.ByCallCounts)

	// Ignore error - doesn't really matter if this succeeds
	var _ = c.client.UpdateExperimentProgressStatus(c.ProjectID, c.ExperimentID, false)
	return nil
}

// createWorkflowSteps walks the list of steps for a particular workflow item creating the
// samples and processes.
func (c *Creater) createWorkflowSteps(wp *WorkflowProcess) error {
	if wp.Worksheet == nil {
		// Creating the sample
		if sample, err := c.createSample(wp.Samples[0]); err != nil {
			return err
		} else {
			wp.Out = append(wp.Out, sample)
		}
	} else {
		// Create the process if it doesn't already exist
		// 1. Find the input sample
		// 2. Create the process with that input sample and attr
		if wp.Process == nil {
			// Create the process
			p, err := c.createProcessWithAttrs(wp.Worksheet, wp.Samples[0].ProcessAttrs)
			if err != nil {
				return err
			}

			wp.Process = p

			// Add the samples to the process
			inputSamples := c.getInputSamples(wp)

			for _, sample := range inputSamples {
				worksheetSample := c.findSampleInWorksheet(sample.Name, wp.Worksheet.Samples)
				if s, err := c.addSampleAndFilesToProcess(wp.Process.ID, sample, worksheetSample); err != nil {
					return err
				} else {
					wp.Out = append(wp.Out, s)

					// Add measurements
					if worksheetSample != nil {
						if err := c.addMeasurements(wp.Process.ID, s.ID, s.PropertySetID, worksheetSample); err != nil {
							return err
						}
					}
				}
			}
		}

	}

	// Now walk all the WorkflowProcess steps that it sends samples into
	// and create those workflow steps. Do this by recursively calling
	// ourselves (createWorkflowSteps).
	for _, next := range wp.To {
		if err := c.createWorkflowSteps(next); err != nil {
			return err
		}
	}

	return nil
}

func (c *Creater) AddCount(what string) {
	value := c.ByCallCounts[what]
	value++
	c.ByCallCounts[what] = value
}

// createExperiment will create a new experiment in the given project
func (c *Creater) createExperiment() error {
	c.Count++
	c.AddCount("createExperiment")
	experiment, err := c.client.CreateExperiment(c.ProjectID, c.Name, c.Description, true)
	if err != nil {
		return err
	}

	c.ExperimentID = experiment.ID
	return nil
}

// createProcessWithAttrs will create a new process with the given set of process attributes.
func (c *Creater) createProcessWithAttrs(process *model.Worksheet, attrs []*model.Attribute) (*mcapi.Process, error) {
	c.Count++
	c.AddCount("createProcessWithAttrs")
	//return &mcapi.Process{}, nil
	setup := mcapi.Setup{
		Name:      "Conditions",
		Attribute: "conditions",
	}

	for _, attr := range attrs {
		if attr.Value != nil {
			p := mcapi.SetupProperty{
				Name:      attr.Name,
				Attribute: attr.Name,
				OType:     "object",
				Unit:      attr.Unit,
				Value:     attr.Value["value"],
			}
			setup.Properties = append(setup.Properties, &p)
		}
	}

	return c.client.CreateProcess(c.ProjectID, c.ExperimentID, process.Name, []mcapi.Setup{setup})
}

// createSample creates a new sample in the project on the server.
func (c *Creater) createSample(sample *model.Sample) (*mcapi.Sample, error) {
	c.Count++
	c.AddCount("createSample")
	return c.client.CreateSample(c.ProjectID, c.ExperimentID, sample.Name, nil)
}

// addMeasurements adds measurements from the model.Sample to the server side process and sample/property set.
// In the workflow a model.Sample contains all the measurements for a sample reference in the spreadsheet.
func (c *Creater) addMeasurements(processID string, sampleID, propertySetID string, sample *model.Sample) error {
	c.Count++
	c.AddCount("addMeasurements")
	//return nil
	attrs := c.createAttributeMeasurements(sample.Attributes)

	sm := mcapi.SampleMeasurements{
		SampleID:      sampleID,
		PropertySetID: propertySetID,
		Attributes:    attrs,
	}

	_, err := c.client.AddMeasurementsToSampleInProcess(c.ProjectID, c.ExperimentID, processID, false, sm)
	return err
}

// createAttributeMeasurements iterates over the list of sample attributes creating a single
// SampleProperty for each attribute and merging the other attributes that match that name
// as separate measurements of that attribute.
func (c *Creater) createAttributeMeasurements(attrs []*model.Attribute) []mcapi.SampleProperty {
	samplePropertiesMap := make(map[string]*mcapi.SampleProperty)
	for _, attr := range attrs {
		sp, ok := samplePropertiesMap[attr.Name]
		if !ok {
			sp = &mcapi.SampleProperty{Name: attr.Name}
			samplePropertiesMap[attr.Name] = sp
		}

		m := mcapi.Measurement{
			Unit:  attr.Unit,
			Value: attr.Value["value"],
			OType: "object",
		}

		sp.Measurements = append(sp.Measurements, m)
	}

	var sampleProperties []mcapi.SampleProperty

	for key := range samplePropertiesMap {
		sampleProperties = append(sampleProperties, *samplePropertiesMap[key])
	}

	return sampleProperties
}

// findSample finds the model.Sample that corresponds to the server side sample. Matching is based
// on name as each sample in the worksheets will have a unique name.
func (c *Creater) findSampleInWorksheet(sampleName string, samples []*model.Sample) *model.Sample {
	for _, sample := range samples {
		if sample.Name == sampleName {
			return sample
		}
	}

	return nil
}

func (c *Creater) findSampleFromServer(sampleName string, samples []*mcapi.Sample) *mcapi.Sample {
	for _, sample := range samples {
		if sample.Name == sampleName {
			return sample
		}
	}

	return nil
}

// addSampleAndFilesToProcess will add the sample and associated files to the process on the server. It hides the details
// of constructing the go-mcapi call.
func (c *Creater) addSampleAndFilesToProcess(processID string, sample *mcapi.Sample, worksheetSample *model.Sample) (*mcapi.Sample, error) {
	c.Count++
	c.AddCount("addSampleAndFilesToProcess")
	//return &mcapi.Sample{}, nil
	connect := mcapi.ConnectSampleAndFilesToProcess{
		ProcessID:     processID,
		SampleID:      sample.ID,
		PropertySetID: sample.PropertySetID,
		Transform:     true,
	}

	if worksheetSample != nil {
		for _, file := range worksheetSample.Files {
			f := mcapi.FileAndDirection{
				Path:      file.Path,
				Direction: "in",
			}
			connect.FilesByName = append(connect.FilesByName, f)
		}
	}
	s, err := c.client.AddSampleAndFilesToProcess(c.ProjectID, c.ExperimentID, false, connect)
	return s, err
}

func (c *Creater) addSamplesToProcess(processID string, samples []*mcapi.Sample) ([]*mcapi.Sample, error) {
	c.Count++
	c.AddCount("addSamplesToProcess")
	connect := mcapi.ConnectSamplesToProcess{
		ProcessID: processID,
		Transform: true,
	}

	for _, sample := range samples {
		s := mcapi.SampleToConnect{
			SampleID:      sample.ID,
			PropertySetID: sample.PropertySetID,
			Name:          sample.Name,
		}
		connect.Samples = append(connect.Samples, s)
	}

	updatedSamples, err := c.client.AddSamplesToProcess(c.ProjectID, c.ExperimentID, connect)
	if err != nil {
		return nil, err
	}

	// API call returns []mcapi.Sample, we need to return []*mcapi.Sample
	var transformUpdatedSamples []*mcapi.Sample
	for _, sample := range updatedSamples {
		transformUpdatedSamples = append(transformUpdatedSamples, &sample)
	}

	return transformUpdatedSamples, nil
}

// getInputSamples goes to the parent workflow processes and constructs the list
// of samples that are input into the workflow process (in this case the wp
// parameter).
func (c *Creater) getInputSamples(wp *WorkflowProcess) []*mcapi.Sample {
	var samples []*mcapi.Sample
	// A WorkflowProcess contains a pointer to its parent workflow processes, this allows
	// it to retrieve all samples from the parent workflow process steps.
	for _, parentWorkflow := range wp.From {
		samples = append(samples, parentWorkflow.Out...)
	}
	return samples
}
