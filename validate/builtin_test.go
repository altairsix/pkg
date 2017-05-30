package validate_test

import (
	"testing"

	"github.com/altairsix/pkg/types"
	"github.com/altairsix/pkg/validate"
	"github.com/stretchr/testify/assert"
)

func TestIsRequired(t *testing.T) {
	testCases := map[string]struct {
		In interface{}
		Ok bool
	}{
		"string": {
			In: "hello",
			Ok: true,
		},
		"types.StringSet": {
			In: types.StringSet{"a": struct{}{}},
			Ok: true,
		},
		"[]types.ID": {
			In: []types.ID{1, 2, 3},
			Ok: true,
		},
		"[]types.ID empty": {
			In: []types.ID{},
			Ok: false,
		},
		"[]types.Key": {
			In: []types.Key{"a", "b", "c"},
			Ok: true,
		},
		"[]types.Key empty": {
			In: []types.Key{},
			Ok: false,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			assert.Equal(t, tc.Ok, validate.IsRequired(tc.In))
		})
	}
}
