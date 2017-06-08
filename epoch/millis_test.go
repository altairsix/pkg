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
		String   string
		Value    int64
		Unit     string
	}{
		"now": {
			Duration: 0,
			String:   "moments ago",
		},
		"1min": {
			Duration: epoch.Millis(60 * 1000),
			String:   "1 minute ago",
			Value:    1,
			Unit:     epoch.Minute,
		},
		"5min": {
			Duration: epoch.Millis(5 * 60 * 1000),
			String:   "5 minutes ago",
			Value:    5,
			Unit:     epoch.Minute,
		},
		"1h": {
			Duration: epoch.Millis(60 * 60 * 1000),
			String:   "1 hour ago",
			Value:    1,
			Unit:     epoch.Hour,
		},
		"3h": {
			Duration: epoch.Millis(3 * 60 * 60 * 1000),
			String:   "3 hours ago",
			Value:    3,
			Unit:     epoch.Hour,
		},
		"1d": {
			Duration: epoch.Millis(24 * 60 * 60 * 1000),
			String:   "1 day ago",
			Value:    1,
			Unit:     epoch.Day,
		},
		"3d": {
			Duration: epoch.Millis(3 * 24 * 60 * 60 * 1000),
			String:   "3 days ago",
			Value:    3,
			Unit:     epoch.Day,
		},
		"1mo": {
			Duration: epoch.Millis(30 * 24 * 60 * 60 * 1000),
			String:   "1 month ago",
			Value:    1,
			Unit:     epoch.Month,
		},
		"5mo": {
			Duration: epoch.Millis(5 * 30 * 24 * 60 * 60 * 1000),
			String:   "5 months ago",
			Value:    5,
			Unit:     epoch.Month,
		},
		"1y": {
			Duration: epoch.Millis(365 * 24 * 60 * 60 * 1000),
			String:   "1 year ago",
			Value:    1,
			Unit:     epoch.Year,
		},
		"5y": {
			Duration: epoch.Millis(5 * 365 * 24 * 60 * 60 * 1000),
			String:   "5 years ago",
			Value:    5,
			Unit:     epoch.Year,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			ago := tc.Duration.Ago()
			assert.Equal(t, tc.String, ago.String)
			assert.Equal(t, tc.Value, ago.Value)
			assert.Equal(t, tc.Unit, ago.Unit)
		})
	}
}
