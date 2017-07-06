package timeofday_test

import (
	"testing"
	"time"

	"github.com/altairsix/pkg/timeofday"
	"github.com/stretchr/testify/assert"
)

func TestClock(t *testing.T) {
	now := time.Now()
	assert.Equal(t, now.Format(timeofday.DateFormat), timeofday.Time(now).String())
}

func TestRelativeTimes(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Minute)
	clock := timeofday.Time(now)

	assert.True(t, clock.LT(later))
	assert.True(t, clock.LTE(later))
	assert.True(t, clock.EQ(now))
	assert.True(t, clock.GTE(now))
	assert.False(t, clock.GT(now))
}
