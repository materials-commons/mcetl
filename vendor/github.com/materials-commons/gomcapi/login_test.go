package mcapi

import (
	"testing"

	"github.com/materials-commons/gomcapi/pkg/tutils/assert"
)

func TestLogin(t *testing.T) {
	c := NewClient("http://mcdev.localhost/api")
	err := c.Login("test@test.mc", "test")
	assert.Okf(t, err, "Login failed with err :%s", err)
	assert.Equals(t, c.APIKey, "totally-bogus")
}
