package webdb_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/altairsix/pkg/dbase"
	"github.com/altairsix/pkg/web"
	"github.com/altairsix/pkg/web/webdb"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestOpen(t *testing.T) {
	db := &gorm.DB{Value: "tracer"}
	accessor := &dbase.Mock{DB: db}
	tracer := errors.New("tracer")

	filter := webdb.Filter(accessor)
	fn := func(c *web.Context) error {
		v, err := webdb.Open(c)
		assert.Nil(t, err)
		assert.Equal(t, db, v)
		return tracer
	}
	fn = filter.Apply(fn)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost", nil)
	err := fn(&web.Context{Request: req, Response: w})
	assert.Equal(t, tracer, err)
	assert.Equal(t, 1, accessor.OpenCount)
	assert.Equal(t, 1, accessor.RollbackCount)
}
