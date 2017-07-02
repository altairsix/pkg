package gormx_test

import (
	"testing"

	"github.com/altairsix/eventsource/mysqlstore"
	"github.com/altairsix/pkg/dbase"
	"github.com/altairsix/pkg/dbase/gormx"
	"github.com/stretchr/testify/assert"
)

type Mock struct {
	dbase.Accessor
}

func TestImplements(t *testing.T) {
	target := &Mock{}
	var accessor interface{} = &gormx.Accessor{Target: target}
	_, ok := accessor.(mysqlstore.Accessor)
	assert.True(t, ok)
}
