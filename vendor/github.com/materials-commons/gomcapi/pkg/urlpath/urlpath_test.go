package urlpath_test

import (
	"testing"

	"github.com/materials-commons/gomcapi/pkg/tutils/assert"
	"github.com/materials-commons/gomcapi/pkg/urlpath"
)

func TestURLPaths(t *testing.T) {
	tests := []struct {
		name     string
		site     string
		paths    []string
		expected string
	}{
		{
			name:     "",
			site:     "http://www.this.com",
			paths:    []string{"def"},
			expected: "http://www.this.com/def",
		},
	}

	for _, test := range tests {
		site, err := urlpath.JoinE(test.site, test.paths...)
		assert.Ok(t, err)
		assert.Equals(t, site, test.expected)
	}
}
