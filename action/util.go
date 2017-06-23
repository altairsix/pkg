package action

import (
	"math/rand"
	"time"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func jitter(d time.Duration) time.Duration {
	fragment := d / 5
	return d - fragment + time.Duration(r.Int63n(2*int64(fragment)))
}
