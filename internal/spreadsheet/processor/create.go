package processor

import (
	"fmt"

	"github.com/hashicorp/go-uuid"
	"github.com/materials-commons/gomcapi"
	"github.com/materials-commons/mcetl/internal/spreadsheet/model"
)

type Creater struct {
	ProjectID    string
	Name         string
	Description  string
	ExperimentID string
	client       *mcapi.Client
}

func NewCreater(projectID, name, description string, client *mcapi.Client) *Creater {
	return &Creater{
		ProjectID:   projectID,
		Name:        name,
		Description: description,
		client:      client,
	}
}

func (c *Creater) Apply(processes []*model.Process) error {
	if err := c.createExperiment(); err != nil {
		return err
	}

	for _, process := range processes {
		if err := c.addProcessToExperiment(process); err != nil {
			return err
		}
	}
	return nil
}

func (c *Creater) createExperiment() error {
	fmt.Printf("Creating Experiment: %s\n", c.Name)
	experiment, err := c.client.CreateExperiment(c.ProjectID, c.Name, c.Description)
	if err != nil {
		return err
	}

	c.ExperimentID = experiment.ID
	return nil
}

func (c *Creater) addProcessToExperiment(process *model.Process) error {
	var (
		processID string
		err       error
	)

	for _, sample := range process.Samples {
		switch {
		case processID == "":
			if processID, err = c.createProcessWithAttrs(process, sample.ProcessAttrs); err != nil {
				return err
			}
		case needsNewProcess(sample):
			fmt.Println("Need to create new process for sample:", sample.Name)
			if processID, err = c.createProcessWithAttrs(process, sample.ProcessAttrs); err != nil {
				return err
			}
		}

		if err := c.addSampleToProcess(processID, sample); err != nil {
			return err
		}
	}

	return nil
}

func (c *Creater) createProcessWithAttrs(process *model.Process, attrs []*model.Attribute) (string, error) {
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
	c.client.CreateProcess(c.ProjectID, c.ExperimentID, process.Name, []mcapi.Setup{setup})
	id, err := uuid.GenerateUUID()
	return id, err
}

// needsNewProcess will look through the process attributes associated with the sample. If all their values
// are blank then it will return false, if any of them have a value then it will return true signifying that
// a new process needs to be created.
func needsNewProcess(sample *model.Sample) bool {
	for _, attr := range sample.ProcessAttrs {
		if attr.Value != nil {
			return true
		}
	}

	return false
}

func (c *Creater) addSampleToProcess(processID string, sample *model.Sample) error {
	fmt.Printf("%sCreate Sample %s for process %s\n", spaces(6), sample.Name, processID)
	return nil
}