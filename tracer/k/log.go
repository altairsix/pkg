package k

import (
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
