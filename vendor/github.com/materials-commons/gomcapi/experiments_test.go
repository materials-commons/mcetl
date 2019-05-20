package mcapi

import (
	"testing"

	"github.com/materials-commons/gomcapi/pkg/tutils/assert"
)

func TestCreateExperiment(t *testing.T) {
	c := newTestClient()

	p, err := c.CreateProject(uniqueName(t), "projdesc")
	assert.Ok(t, err)
	assert.NotNil(t, p)

	var e *Experiment

	e, err = c.CreateExperiment(p.ID, uniqueName(t), "expdesc", false)
	assert.Ok(t, err)
	assert.NotNil(t, e)

	err = c.DeleteProject(p.ID)
	assert.Ok(t, err)
}
