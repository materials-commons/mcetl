package mcapi

import (
	"testing"

	"github.com/materials-commons/gomcapi/pkg/tutils/assert"

	uuid "github.com/hashicorp/go-uuid"
)

const testURL = "http://mcdev.localhost/api"

func newTestClient() *Client {
	c := NewClient(testURL)
	c.APIKey = "totally-bogus"
	return c
}

func uniqueName(t *testing.T) string {
	id, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatalf("Failed creating uuid %s", err)
	}

	return id
}

func TestCreateProject(t *testing.T) {
	c := newTestClient()
	p, err := c.CreateProject(uniqueName(t), "projdesc")
	assert.Ok(t, err)
	assert.NotNil(t, p)

	err = c.DeleteProject(p.ID)
	assert.Ok(t, err)
}
