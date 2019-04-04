// Creater will take the loaded set of processes and create the workflow
// on the server. It steps through each process entry and then each
// of the samples for that process. For each sample associated with a top
// level process it will check to see if a new process should be created.
// To understand this layout look in the model to see how a process
// is laid out.

package processor

import (
	"fmt"

	"github.com/materials-commons/gomcapi"
	"github.com/materials-commons/mcetl/internal/spreadsheet/model"
)

type Creater struct {
	// The project we are adding to
	ProjectID string

	// Name of the experiment to create
	Name string

	// Description of the experiment to create
	Description string

	// The created experiments ID, this and ProjectID are needed
	// for many of the mcapi REST calls.
	ExperimentID string

	client *mcapi.Client
}

type createdSample struct {
	ID   string
	Name string
}

func NewCreater(projectID, name, description string, client *mcapi.Client) *Creater {
	return &Creater{
		ProjectID:   projectID,
		Name:        name,
		Description: description,
		client:      client,
	}
}

func (c *Creater) Apply(worksheets []*model.Worksheet) error {
	if err := c.createExperiment(); err != nil {
		return nil
	}

	wf := newWorkflow()
	wf.constructWorkflow(worksheets)
	for _, wp := range wf.root {
		if err := c.createWorkflowSteps(wp); err != nil {
			return err
		}
	}
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
			fmt.Printf("   Created sample %#v\n", sample)
			wp.Out = append(wp.Out, sample)
		}
	} else {
		// Create the process
		// 1. Find the input sample
		// 2. Create the process with that input sample and attr
		//     == or if process already exists ==
		// 1. Add additional measurements for that process/sample
		//
		// Going to need to keep track of the samples so we know what the inputs are
		inputSamples := c.getInputSamples(wp)
		if wp.Process == nil {
			p, err := c.createProcessWithAttrs(wp.Worksheet, wp.Worksheet.ProcessAttrs)
			if err != nil {
				return err
			}

			fmt.Printf("Created Process %s %#v\n", wp.Worksheet.Name, p)
			wp.Process = p
		}

		// TODO: Add measurements to process

		for _, sample := range inputSamples {
			if s, err := c.addSampleToProcess(wp.Process.ID, sample); err != nil {
				return err
			} else {
				fmt.Printf("  added Sample To Process %s %#v\n", wp.Worksheet.Name, s)
				wp.Out = append(wp.Out, s)
			}
		}
	}

	for _, next := range wp.To {
		if err := c.createWorkflowSteps(next); err != nil {
			return err
		}
	}

	return nil
}

// createExperiment will create a new experiment in the given project
func (c *Creater) createExperiment() error {
	fmt.Printf("Creating Experiment: %s\n", c.Name)
	experiment, err := c.client.CreateExperiment(c.ProjectID, c.Name, c.Description)
	if err != nil {
		return err
	}

	c.ExperimentID = experiment.ID
	return nil
}

// createProcessWithAttrs will create a new process with the given set of process attributes.
func (c *Creater) createProcessWithAttrs(process *model.Worksheet, attrs []*model.Attribute) (*mcapi.Process, error) {
	fmt.Printf("%sCreating Process %s, in experiment %s with sample process attributes\n", spaces(4), process.Name, c.ExperimentID)
	setup := mcapi.Setup{
		Name:      "Test",
		Attribute: "test",
	}
	for _, attr := range attrs {
		if attr.Value != nil {
			p := mcapi.SetupProperty{
				Name:      attr.Name,
				Attribute: attr.Name,
				OType:     "object",
				Unit:      attr.Unit,
				Value:     attr.Value,
			}
			setup.Properties = append(setup.Properties, &p)
		}
	}

	return c.client.CreateProcess(c.ProjectID, c.ExperimentID, process.Name, []mcapi.Setup{setup})
}

// createSample creates a new sample in the project.
func (c *Creater) createSample(sample *model.Sample) (*mcapi.Sample, error) {
	fmt.Printf("%sCreating Sample %s", spaces(4), sample.Name)
	var attrs []mcapi.Property
	for _, attr := range sample.Attributes {
		property := mcapi.Property{
			Name: attr.Name,
		}
		//fmt.Printf("   attr.Value = %#v: %#v\n", attr.Value, attr.Value["value"])
		attrs = append(attrs, property)
		m := mcapi.Measurement{
			Unit:  attr.Unit,
			Value: attr.Value,
			OType: "object",
		}
		property.Measurements = append(property.Measurements, m)
	}

	return c.client.CreateSample(c.ProjectID, c.ExperimentID, sample.Name, attrs)
}

func (c *Creater) addAdditionalMeasurements(processID string, seenSample *createdSample, sample *model.Sample) error {
	fmt.Printf("%sAdd additional measurements for sample %s(%s) in process %s\n", spaces(6), sample.Name, seenSample.ID, processID)
	return nil
}

func (c *Creater) addSampleToProcess(processID string, sample *mcapi.Sample) (*mcapi.Sample, error) {
	fmt.Printf("%sAdd Sample %s to process %s %#v\n", spaces(6), sample.Name, processID, sample)
	connect := mcapi.ConnectSampleToProcess{
		ProcessID:     processID,
		SampleID:      sample.ID,
		PropertySetID: sample.PropertySetID,
		Transform:     true,
	}
	s, err := c.client.AddSampleToProcess(c.ProjectID, c.ExperimentID, connect)
	return s, err
}

func (c *Creater) getInputSamples(wp *WorkflowProcess) []*mcapi.Sample {
	var samples []*mcapi.Sample
	// retrieve all samples from parent steps
	for _, parentWorkflow := range wp.From {
		samples = append(samples, parentWorkflow.Out...)
	}
	return samples
}
