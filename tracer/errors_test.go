package tracer_test

import (
	"io"
	"testing"

	"github.com/altairsix/pkg/tracer"
	"github.com/stretchr/testify/assert"
)

func TestHasErr(t *testing.T) {
	fn := func(err error) bool { return err == io.EOF }

	testCases := map[string]struct {
		Error error
		Match bool
	}{
		"nil": {
			Error: nil,
			Match: false,
		},
		"simple": {
			Error: io.EOF,
			Match: true,
		},
		"wrapped": {
			Error: tracer.Errorf(io.EOF, "wrapped message"),
			Match: true,
		},
		"wrapped fail": {
			Error: tracer.Errorf(io.ErrClosedPipe, "wrapped message"),
			Match: false,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			assert.Equal(t, tc.Match, tracer.HasErr(tc.Error, fn))
		})
	}
}
