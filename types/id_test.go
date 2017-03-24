package types_test

import (
	"testing"

	"github.com/altairsix/pkg/types"

	"github.com/stretchr/testify/assert"
)

func TestID_IsZero(t *testing.T) {
	assert.True(t, types.ZeroID.IsZero())
}
