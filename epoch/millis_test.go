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

func TestAgo(t *testing.T) {
	testCases := map[string]struct {
		Duration epoch.Millis
		Expected string
	}{
		"now": {
			Duration: 0,
			Expected: "just now",
		},
		"1min": {
			Duration: epoch.Millis(60 * 1000),
			Expected: "1 minute ago",
		},
		"5min": {
			Duration: epoch.Millis(5 * 60 * 1000),
			Expected: "5 minutes ago",
		},
		"1h": {
			Duration: epoch.Millis(60 * 60 * 1000),
			Expected: "1 hour ago",
		},
		"3h": {
			Duration: epoch.Millis(3 * 60 * 60 * 1000),
			Expected: "3 hours ago",
		},
		"1d": {
			Duration: epoch.Millis(24 * 60 * 60 * 1000),
			Expected: "1 day ago",
		},
		"3d": {
			Duration: epoch.Millis(3 * 24 * 60 * 60 * 1000),
			Expected: "3 days ago",
		},
		"1m": {
			Duration: epoch.Millis(30 * 24 * 60 * 60 * 1000),
			Expected: "1 month ago",
		},
		"5m": {
			Duration: epoch.Millis(5 * 30 * 24 * 60 * 60 * 1000),
			Expected: "5 months ago",
		},
		"1y": {
			Duration: epoch.Millis(365 * 24 * 60 * 60 * 1000),
			Expected: "1 year ago",
		},
		"5y": {
			Duration: epoch.Millis(5 * 365 * 24 * 60 * 60 * 1000),
			Expected: "5 years ago",
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			assert.Equal(t, tc.Expected, tc.Duration.Ago())
		})
	}
}
