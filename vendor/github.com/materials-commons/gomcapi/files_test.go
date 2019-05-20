package mcapi

import (
	"testing"

	"github.com/materials-commons/gomcapi/pkg/tutils/assert"
)

func TestClient_GetFileByPathInProject(t *testing.T) {
	c := newTestClient()
	// hard coded for now - will fix in a bit
	f, err := c.GetFileByPathInProject("P1/D1/hello.txt", "f1425e92-76c1-402f-b1f4-fc85aaf1a387")
	assert.Ok(t, err)
	assert.NotNil(t, f)
}
