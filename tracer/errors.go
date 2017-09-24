package tracer

import "fmt"

type Causer interface {
	Cause() error
}

type errType struct {
	cause   error
	message string
}

func (e errType) Cause() error {
	return e.cause
}

func (e errType) Error() string {
	if e.cause == nil {
		return e.message
	}

	return e.message + ": " + e.cause.Error()
}

func Errorf(cause error, message string, args ...interface{}) error {
	return errType{
		cause:   cause,
		message: fmt.Sprintf(message, args...),
	}
}

func HasErr(err error, fn func(err error) bool) bool {
	if err == nil {
		return false

	} else if ok := fn(err); ok {
		return true

	} else if v, ok := err.(Causer); ok {
		if cause := v.Cause(); cause != nil {
			return HasErr(cause, fn)
		}
	}

	return false
}
