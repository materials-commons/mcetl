package mcapi

import (
	"testing"

	"github.com/materials-commons/gomcapi/pkg/tutils/assert"
)

func TestCreateProcess(t *testing.T) {
	c := newTestClient()

	p, err := c.CreateProject(uniqueName(t), "projdesc")
	assert.Ok(t, err)
	assert.NotNil(t, p)

	var e *Experiment

	e, err = c.CreateExperiment(p.ID, uniqueName(t), "expdesc", false)
	assert.Ok(t, err)
	assert.NotNil(t, e)

	value := struct {
		A int `json:"a"`
		B int `json:"b"`
	}{
		A: 1,
		B: 2,
	}

	var proc *Process
	setup := Setup{
		Name:      "Test",
		Attribute: "test",
		Properties: []*SetupProperty{
			{Name: "Grain Size", Attribute: "grain_size", OType: "object", Unit: "mm", Value: value},
		},
	}

	proc, err = c.CreateProcess(p.ID, e.ID, uniqueName(t), []Setup{setup})
	assert.Ok(t, err)
	assert.NotNil(t, proc)

	err = c.DeleteProject(p.ID)
	assert.Ok(t, err)
}
