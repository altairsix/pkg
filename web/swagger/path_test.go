package swagger

import (
	"testing"

	"github.com/altairsix/pkg/web"
	"github.com/savaki/swag/endpoint"
	"github.com/stretchr/testify/assert"
)

func TestFixPath(t *testing.T) {
	testCases := map[string]struct {
		In       string
		Expected string
	}{
		"simple": {
			In:       "/a/b",
			Expected: "/a/b",
		},
		"single": {
			In:       "/:wildcard",
			Expected: "/{wildcard}",
		},
		"multiple": {
			In:       "/:a/:b/:c",
			Expected: "/{a}/{b}/{c}",
		},
		"overlap": {
			In:       "/:a/:aa/:aaa",
			Expected: "/{a}/{aa}/{aaa}",
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			assert.Equal(t, tc.Expected, fixPath(tc.In))
		})
	}
}

func TestPathOptions(t *testing.T) {
	testCases := map[string]struct {
		Path     string
		Opts     []web.Option
		Expected []string // list of expected path routes (in order)
	}{
		"simple": {
			Path:     "/a/b",
			Expected: []string{},
		},
		"single": {
			Path:     "/:wildcard",
			Expected: []string{"wildcard"},
		},
		"multiple": {
			Path:     "/:a/:b/:c",
			Expected: []string{"a", "b", "c"},
		},
		"overlap": {
			Path:     "/:a/:aa/:aaa",
			Expected: []string{"a", "aa", "aaa"},
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			opts := pathOptions(tc.Path, tc.Opts...)
			e := endpoint.New("GET", tc.Path, "summary", opts...)

			assert.Len(t, e.Parameters, len(tc.Expected))
			for index, name := range tc.Expected {
				assert.Equal(t, "path", e.Parameters[index].In)
				assert.Equal(t, name, e.Parameters[index].Name)
				assert.NotZero(t, e.Parameters[index].Description)
				assert.True(t, e.Parameters[index].Required)
			}
		})
	}
}
