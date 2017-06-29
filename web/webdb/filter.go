package webdb

import (
	"context"
	"sync"

	"github.com/altairsix/pkg/dbase"
	"github.com/altairsix/pkg/web"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

const (
	refKey = "webdb.ref"
)

type reference struct {
	sync.Mutex

	db       *gorm.DB
	accessor dbase.Accessor
}

func OpenContext(ctx context.Context) (*gorm.DB, error) {
	v := ctx.Value(refKey)
	if v == nil {
		return nil, errors.New("OpenContext called with no database reference.  Was webdb.Filter set up?")
	}

	ref := v.(*reference)

	ref.Lock()
	defer ref.Unlock()

	if ref.db == nil {
		db, err := ref.accessor.Open()
		if err != nil {
			return nil, err
		}
		ref.db = db
	}

	return ref.db, nil
}

func Open(c web.Context) (*gorm.DB, error) {
	return OpenContext(c.Request().Context())
}

func Filter(accessor dbase.Accessor) web.Filter {
	return func(h web.HandlerFunc) web.HandlerFunc {
		return func(c web.Context) error {
			ref := &reference{
				accessor: accessor,
			}
			ctx := context.WithValue(c.Request().Context(), refKey, ref)
			c = c.WithRequest(c.Request().WithContext(ctx))

			err := h(c)

			if db := ref.db; db != nil {
				defer accessor.Close(db)

				if err != nil {
					accessor.Rollback(db)
				} else {
					accessor.Commit(db)
				}
			}

			return err
		}
	}
}
