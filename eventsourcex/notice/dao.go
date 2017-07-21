package notice

import (
	"context"

	"github.com/altairsix/eventsource"
	"github.com/altairsix/pkg/dbase"
	"github.com/jinzhu/gorm"
)

// DAO provides the shape of the required db access layer
type DAO interface {
	// Version returns the current version number of the aggregate; returns 0, nil if the aggregate wasn't found
	Version(ctx context.Context, db *gorm.DB, aggregateID string) (int, error)

	// HandleEvent updates the read model with the specified event
	HandleEvent(ctx context.Context, db *gorm.DB, event eventsource.Event) error
}

// UnmarshalFunc reads a []byte and returns an event
type UnmarshalFunc func([]byte) (eventsource.Event, error)

// NewDBHandler generates a new Handler for database read models
func NewDBHandler(accessor dbase.Accessor, dao DAO, store eventsource.Store, unmarshal UnmarshalFunc) HandlerFunc {
	return func(ctx context.Context, notice Message) {
		db, err := accessor.Open()
		if err != nil {
			return
		}
		defer db.Close()

		id := notice.AggregateID()
		version, err := dao.Version(ctx, db, id)
		if err != nil {
			return
		}

		history, err := store.Load(ctx, id, version, 0)
		if err != nil {
			return
		}

		for _, record := range history {
			event, err := unmarshal(record.Data)
			if err != nil {
				return
			}

			if err := dao.HandleEvent(ctx, db, event); err != nil {
				return
			}
		}
	}
}
