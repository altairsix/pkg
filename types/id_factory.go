package types

// -- IDFactory ------------------------------------------

type IDFactory func() ID

func (fn IDFactory) NewID() ID {
	return fn()
}

func (fn IDFactory) NewKey() Key {
	return fn().Key()
}
