package processor

//
//

import (
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

	client *mcapi.Client
}

func NewCreater(projectID, name, description string, client *mcapi.Client) *Creater {
	return &Creater{
		ProjectID:   projectID,
		Name:        name,
		Description: description,
		client:      client,
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
	wf.constructWorkflow(worksheets)

	// 3. Walk through the workflow creating each of the steps.
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
				if s, err := c.addSampleToProcess(wp.Process.ID, sample); err != nil {
					return err
				} else {
					wp.Out = append(wp.Out, s)

					// Add measurements
					worksheetSample := c.findSample(s, wp.Worksheet.Samples)
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

// createExperiment will create a new experiment in the given project
func (c *Creater) createExperiment() error {
	experiment, err := c.client.CreateExperiment(c.ProjectID, c.Name, c.Description)
	if err != nil {
		return err
	}

	c.ExperimentID = experiment.ID
	return nil
}

// createProcessWithAttrs will create a new process with the given set of process attributes.
func (c *Creater) createProcessWithAttrs(process *model.Worksheet, attrs []*model.Attribute) (*mcapi.Process, error) {
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
	var attrs []mcapi.Property
	for _, attr := range sample.Attributes {
		property := mcapi.Property{
			Name: attr.Name,
		}
		attrs = append(attrs, property)
		m := mcapi.Measurement{
			Unit:  attr.Unit,
			Value: attr.Value["value"],
			OType: "object",
		}
		property.Measurements = append(property.Measurements, m)
	}

	return c.client.CreateSample(c.ProjectID, c.ExperimentID, sample.Name, attrs)
}

// addMeasurements adds measurements from the model.Sample to the server side process and sample/property set.
// In the workflow a model.Sample contains all the measurements for a sample reference in the spreadsheet.
func (c *Creater) addMeasurements(processID string, sampleID, propertySetID string, sample *model.Sample) error {
	attrs := c.createAttributeMeasurements(sample.Attributes)

	sm := mcapi.SampleMeasurements{
		SampleID:      sampleID,
		PropertySetID: propertySetID,
		Attributes:    attrs,
	}

	_, err := c.client.AddMeasurementsToSampleInProcess(c.ProjectID, c.ExperimentID, processID, sm)
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
func (c *Creater) findSample(createdSample *mcapi.Sample, samples []*model.Sample) *model.Sample {
	for _, sample := range samples {
		if sample.Name == createdSample.Name {
			return sample
		}
	}

	return nil
}

// addSampleToProcess will add the sample to the process on the server. It hides the details of constructing
// the go-mcapi call.
func (c *Creater) addSampleToProcess(processID string, sample *mcapi.Sample) (*mcapi.Sample, error) {
	connect := mcapi.ConnectSampleToProcess{
		ProcessID:     processID,
		SampleID:      sample.ID,
		PropertySetID: sample.PropertySetID,
		Transform:     true,
	}
	s, err := c.client.AddSampleToProcess(c.ProjectID, c.ExperimentID, connect)
	return s, err
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
