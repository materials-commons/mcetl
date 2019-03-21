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

// Apply creates a new experiment then goes through the list of processes and creates the workflow from them.
func (c *Creater) Apply(worksheets []*model.Worksheet) error {
	if err := c.createExperiment(); err != nil {
		return err
	}

	for _, worksheet := range worksheets {
		if err := c.createWorkfowFromWorksheet(worksheet); err != nil {
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

// createWorkfowFromWorksheet will take a single worksheet entry that is composed
// of multiple samples. It will then attempt to create as many processes and samples as
// are needed from that particular worksheet entry.
func (c *Creater) createWorkfowFromWorksheet(process *model.Worksheet) error {
	var (
		processID               string
		err                     error
		lastCreatedProcessAttrs []*model.Attribute
	)

	for _, sample := range process.Samples {
		switch {
		case processID == "":
			if processID, err = c.createProcessWithAttrs(process, sample.ProcessAttrs); err != nil {
				return err
			}
			lastCreatedProcessAttrs = sample.ProcessAttrs
		case needsNewProcess(sample, lastCreatedProcessAttrs):
			fmt.Println("Need to create new process for sample:", sample.Name)
			if processID, err = c.createProcessWithAttrs(process, sample.ProcessAttrs); err != nil {
				return err
			}
			lastCreatedProcessAttrs = sample.ProcessAttrs
		}

		if err := c.addSampleToProcess(processID, sample); err != nil {
			return err
		}
	}

	return nil
}

func (c *Creater) createProcessWithAttrs(process *model.Worksheet, attrs []*model.Attribute) (string, error) {
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
		return "", err
	}
	return proc.ID, nil
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

func (c *Creater) addSampleToProcess(processID string, sample *model.Sample) error {
	fmt.Printf("%sCreate Sample %s for process %s\n", spaces(6), sample.Name, processID)
	return nil
}
