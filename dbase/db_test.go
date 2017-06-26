package dbase_test

import (
	"testing"

	"github.com/altairsix/pkg/dbase"
	"github.com/stretchr/testify/assert"
)

func TestMock(t *testing.T) {
	var v interface{} = &dbase.Mock{}
	_, ok := v.(dbase.Accessor)
	assert.True(t, ok)
}
