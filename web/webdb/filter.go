package webdb

import (
	"github.com/altairsix/pkg/dbase"
	"github.com/altairsix/pkg/web"
	"github.com/jinzhu/gorm"
)

const (
	accessorKey = "accessor"
	refKey      = "ref"
)

type reference struct {
	db *gorm.DB
}

func fetch(c web.Context) (*gorm.DB, bool) {
	if v := c.Get(refKey); v != nil {
		if ref, ok := v.(*reference); ok {
			return ref.db, true
		}
	}

	return nil, false
}

func Open(c web.Context) (*gorm.DB, error) {
	db, ok := fetch(c)
	if ok {
		return db, nil
	}

	accessor := c.Get(accessorKey).(dbase.Accessor)
	db, err := accessor.Open()
	if err != nil {
		return nil, err
	}

	ref := &reference{db: db}
	c.Set(refKey, ref)

	return db, nil
}

func Filter(accessor dbase.Accessor) web.Filter {
	return func(h web.HandlerFunc) web.HandlerFunc {
		return func(c web.Context) error {
			c.Set(accessorKey, accessor)
			err := h(c)

			if db, ok := fetch(c); ok {
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
