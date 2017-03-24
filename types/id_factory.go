package types

import (
	"time"

	"github.com/pkg/errors"
	"github.com/savaki/snowflake"
)

// -- IDFactory ------------------------------------------

type IDFactory func() ID

func (fn IDFactory) NewID() ID {
	return fn()
}

func (fn IDFactory) NewKey() Key {
	return fn().Key()
}

func MockIDFactory() func() ID {
	return func() ID {
		return ID(time.Now().UnixNano())
	}
}

func NewIDFactory(hosts ...string) (IDFactory, error) {
	c, err := snowflake.NewClient(snowflake.WithHosts(hosts...))
	if err != nil {
		return nil, errors.Wrap(err, "types:new_id_factory:err")
	}

	bc := snowflake.NewBufferedClient(c)
	idFactory := IDFactory(func() ID {
		return ID(bc.Id())
	})

	return idFactory, nil
}
