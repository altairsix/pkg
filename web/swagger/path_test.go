package swagger_test

import (
	"testing"

	"github.com/altairsix/pkg/web/swagger"
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
			assert.Equal(t, tc.Expected, swagger.FixPath(tc.In))
		})
	}
}
