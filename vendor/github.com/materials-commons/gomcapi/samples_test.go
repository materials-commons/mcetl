package mcapi

import (
	"fmt"
	"testing"

	"github.com/materials-commons/gomcapi/pkg/tutils/assert"
)

func TestCreateSample(t *testing.T) {
	c := newTestClient()
	p, e := createTestProjAndExperiment(t, c)

	measurements := []Measurement{
		{OType: "object", Unit: "mm", Value: 1},
	}

	attrs := []Property{
		{Name: "attr1", Measurements: measurements},
	}

	s, err := c.CreateSample(p.ID, e.ID, "s1", attrs)
	assert.Ok(t, err)
	assert.NotNil(t, s)

	fmt.Printf("%#v\n", s)

	_ = c.DeleteProject(p.ID)
}

func TestAddSampleToProcess(t *testing.T) {
	c := newTestClient()
	p, e := createTestProjAndExperiment(t, c)
	s, err := c.CreateSample(p.ID, e.ID, "s1", nil)
	assert.Ok(t, err)
	assert.NotNil(t, s)

	proc, err := c.CreateProcess(p.ID, e.ID, uniqueName(t), nil)
	assert.Ok(t, err)
	assert.NotNil(t, proc)

	connect := ConnectSampleToProcess{
		ProcessID:     proc.ID,
		SampleID:      s.ID,
		PropertySetID: s.PropertySetID,
		Transform:     true,
	}
	s, err = c.AddSampleToProcess(p.ID, e.ID, true, connect)
	assert.Ok(t, err)
	assert.NotNil(t, s)
	fmt.Printf("%#v\n", s)
	//_ = c.DeleteProject(p.ID)
}

func createTestProjAndExperiment(t *testing.T, c *Client) (*Project, *Experiment) {
	var (
		p   *Project
		e   *Experiment
		err error
	)
	p, err = c.CreateProject(uniqueName(t), "test project")
	assert.Ok(t, err)
	assert.NotNil(t, p)

	e, err = c.CreateExperiment(p.ID, uniqueName(t), "test experiment", false)
	assert.Ok(t, err)
	assert.NotNil(t, e)

	return p, e
}
