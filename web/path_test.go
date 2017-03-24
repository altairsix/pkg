package web_test

import (
	"testing"

	"github.com/altairsix/pkg/web"
	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {
	testCases := map[string]struct {
		In  string
		Out string
	}{
		"single": {
			In:  "/api/orgs/{org}",
			Out: "/api/orgs/:org",
		},
		"multiple": {
			In:  "/{a}/{b}/{c}",
			Out: "/:a/:b/:c",
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			assert.EqualValues(t, tc.Out, web.FixPath(tc.In))
		})
	}
}
