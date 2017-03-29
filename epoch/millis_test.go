package epoch_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/altairsix/pkg/epoch"
	"github.com/stretchr/testify/assert"
)

func TestNow(t *testing.T) {
	started := time.Now()
	assert.True(t, epoch.Now().Time().Sub(started) < time.Millisecond, "expected now to be the same as time.Now()")
}

func TestJSON(t *testing.T) {
	t.Run("obj", func(t *testing.T) {
		now := epoch.Now()

		data, err := json.Marshal(now)
		assert.Nil(t, err)

		var m epoch.Millis
		err = json.Unmarshal(data, &m)
		assert.Nil(t, err)

		assert.Equal(t, now, m)
	})

	t.Run("int64", func(t *testing.T) {
		now := epoch.Now()

		data, err := json.Marshal(now.Int64())
		assert.Nil(t, err)

		var m epoch.Millis
		err = json.Unmarshal(data, &m)
		assert.Nil(t, err)

		assert.Equal(t, now, m)
	})

	t.Run("string", func(t *testing.T) {
		now := epoch.Now()
		now -= now % 1000 // strip off the millis

		data, err := json.Marshal(now.Format(time.RFC3339))
		assert.Nil(t, err)

		var m epoch.Millis
		err = json.Unmarshal(data, &m)
		assert.Nil(t, err)

		assert.Equal(t, now, m)
	})
}
