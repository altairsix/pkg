package k

import (
	"fmt"
	"time"

	"github.com/altairsix/pkg/types"
	"github.com/opentracing/opentracing-go/log"
)

// ID logs general id
func ID(id types.Key) log.Field {
	return log.String("id", id.String())
}

// OrgID logs org id
func OrgID(orgID types.Key) log.Field {
	return log.String("org", orgID.String())
}

// Elapsed time measured in seconds
func Elapsed(d time.Duration) log.Field {
	return log.Int("elapsed", int(d/time.Second))
}

// Delay measured in seconds
func Delay(d time.Duration) log.Field {
	return log.Int("delay", int(d/time.Second))
}

// Name provide a name field
func Name(name string) log.Field {
	return log.String("name", name)
}

// Method provides an http method
func Method(method string) log.Field {
	return log.String("method", method)
}

// StatusCode provides an http status code
func StatusCode(sc int) log.Field {
	return log.Int("status-code", sc)
}

// URL provides an http url
func URL(url string) log.Field {
	return log.String("url", url)
}

// Err provides an error
func Err(err error) log.Field {
	if err == nil {
		return log.Noop()
	}
	return log.Error(err)
}

// Key provides a standard key
func Key(key string) log.Field {
	return log.String("key", key)
}

// Duration measured in seconds
func Duration(d time.Duration) log.Field {
	return log.Int("duration", int(d/time.Second))
}

// Value allows the logging of arbitrary values
func Value(in interface{}) log.Field {
	switch v := in.(type) {
	case string:
		return log.String("value", v)
	case int:
		return log.Int("value", v)
	case int32:
		return log.Int32("value", v)
	case int64:
		return log.Int64("value", v)
	case float32:
		return log.Float32("value", v)
	case float64:
		return log.Float64("value", v)
	default:
		panic(fmt.Sprintf("unhandled type, %v", in))
	}
}
