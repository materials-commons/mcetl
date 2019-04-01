// Creater will take the loaded set of processes and create the workflow
// on the server. It steps through each process entry and then each
// of the samples for that process. For each sample associated with a top
// level process it will check to see if a new process should be created.
// To understand this layout look in the model to see how a process
// is laid out.

package processor

import (
	"fmt"
	"reflect"

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

	// previouslySeen is a map of process => list of samples. Whenever a sample
	// is created in a process instance it is added to this map. This allows
	// us to look up samples that were previously created and treat the new
	// instance of a sample as adding additional measurements to that sample
	// associated with that process.
	previouslySeen map[string][]createdSample

	// existingSamples is a map of known sample names => id. If the sample is not
	// in this map that means it's the first time it was encountered and the sample
	// needs to be created.
	existingSamples map[string]string
}

type createdSample struct {
	ID   string
	Name string
}

func NewCreater(projectID, name, description string, client *mcapi.Client) *Creater {
	return &Creater{
		ProjectID:      projectID,
		Name:           name,
		Description:    description,
		client:         client,
		previouslySeen: make(map[string][]createdSample),
	}
}

// Apply creates a new experiment then goes through the list of processes and creates the workflow from them.
func (c *Creater) Apply(worksheets []*model.Worksheet) error {
	if err := c.createExperiment(); err != nil {
		return err
	}

	for _, worksheet := range worksheets {
		if err := c.createWorkflowFromWorksheet(worksheet); err != nil {
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

// createWorkflowFromWorksheet will take a single worksheet entry that is composed
// of multiple samples. It will then attempt to create as many processes and samples as
// are needed from that particular worksheet entry. It creates a new process for that
// worksheet when it encounters a sample with a process attributes that are different
// than the last set of process attributes it saw.
func (c *Creater) createWorkflowFromWorksheet(process *model.Worksheet) error {
	var (
		processID               string
		p                       *mcapi.Process
		err                     error
		lastCreatedProcessAttrs []*model.Attribute
	)

	for _, sample := range process.Samples {
		switch {

		case processID == "":
			if p, err = c.createProcessWithAttrs(process, sample.ProcessAttrs); err != nil {
				return err
			}
			lastCreatedProcessAttrs = sample.ProcessAttrs
			processID = p.ID

		case needsNewProcess(sample, lastCreatedProcessAttrs):
			fmt.Println("Need to create new process for sample:", sample.Name)
			if p, err = c.createProcessWithAttrs(process, sample.ProcessAttrs); err != nil {
				return err
			}
			processID = p.ID
			lastCreatedProcessAttrs = sample.ProcessAttrs
		}

		if _, ok := c.existingSamples[sample.Name]; !ok {
			if s, err := c.createSample(sample); err != nil {
				return err
			} else {
				c.existingSamples[sample.Name] = s.ID
			}
		}

		if seenSample := c.findAlreadySeenSample(processID, sample); seenSample != nil {
			if err := c.addAdditionalMeasurements(processID, seenSample, sample); err != nil {
				return err
			}
		} else {
			if err := c.addSampleToProcess(processID, sample); err != nil {
				return err
			}
		}
	}

	return nil
}

// createProcessWithAttrs will create a new process with the given set of process attributes.
func (c *Creater) createProcessWithAttrs(process *model.Worksheet, attrs []*model.Attribute) (*mcapi.Process, error) {
	fmt.Printf("%sCreating Worksheet %s, in experiment %s with sample process attributes\n", spaces(4), process.Name, c.ExperimentID)
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
	proc, err := c.client.CreateProcess(c.ProjectID, c.ExperimentID, process.Name, []mcapi.Setup{setup})
	if err != nil {
		return nil, err
	}
	return proc, nil
}

// needsNewProcess will look through the process attributes associated with the sample and the
// last created processes process attributes. If they are different then it will return true
// meaning that a new process should be created.
func needsNewProcess(sample *model.Sample, lastSetOfAttrs []*model.Attribute) bool {
	for i := 0; i < len(lastSetOfAttrs); i++ {
		if !reflect.DeepEqual(sample.ProcessAttrs[i].Value, lastSetOfAttrs[i].Value) {
			return true
		}
	}

	return false
}

// createSample creates a new sample in the project.
func (c *Creater) createSample(sample *model.Sample) (*mcapi.Sample, error) {
	return &mcapi.Sample{}, nil
}

// findAlreadySeenSample looks for the sample in the list of samples associated with the given
// processID. Matches are made by sample name as sample names identify unique samples in a given
// process in the worksheet.
func (c *Creater) findAlreadySeenSample(processID string, sample *model.Sample) *createdSample {
	if samples, ok := c.previouslySeen[processID]; !ok {
		return nil
	} else {
		for _, seen := range samples {
			if sample.Name == seen.Name {
				return &seen
			}
		}
	}
	return nil
}

func (c *Creater) addAdditionalMeasurements(processID string, seenSample *createdSample, sample *model.Sample) error {
	fmt.Printf("%sAdd additional measurements for sample %s(%s) in process %s\n", spaces(6), sample.Name, seenSample.ID, processID)
	return nil
}

func (c *Creater) addSampleToProcess(processID string, sample *model.Sample) error {
	fmt.Printf("%sCreate Sample %s for process %s\n", spaces(6), sample.Name, processID)
	return nil
}
