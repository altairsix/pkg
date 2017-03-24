package epoch_test

import (
	"testing"
	"time"

	"github.com/altairsix/pkg/epoch"
	"github.com/stretchr/testify/assert"
)

func TestNow(t *testing.T) {
	started := time.Now()
	assert.True(t, epoch.Now().Time().Sub(started) < time.Millisecond, "expected now to be the same as time.Now()")
}
