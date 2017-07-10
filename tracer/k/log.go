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

// String logs a key and value
func String(key, value string) log.Field {
	return log.String(key, value)
}

// Int logs a key and value
func Int(key string, value int) log.Field {
	return log.Int(key, value)
}

// Text logs a string with the key text; useful for messages which are
// not intended to be read except by humans
func Text(value string) log.Field {
	return log.String("text", value)
}

// Subject references the nats subject the message was sent to or
// received from
func Subject(value string) log.Field {
	return log.String("subject", value)
}

// Bool records primitive booleans
func Bool(key string, value bool) log.Field {
	return log.Bool(key, value)
}

// Int64 records primitive int64
func Int64(key string, value int64) log.Field {
	return log.Int64(key, value)
}

// Int32 records primitive int32
func Int32(key string, value int32) log.Field {
	return log.Int32(key, value)
}

// Float64 records primitive float64
func Float64(key string, value float64) log.Field {
	return log.Float64(key, value)
}

// Float32 records primitive int32
func Float32(key string, value float32) log.Field {
	return log.Float32(key, value)
}
