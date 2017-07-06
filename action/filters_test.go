package action_test

import (
	"context"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/altairsix/pkg/action"
	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	ctx := context.Background()
	retry := action.Retry(3, time.Second)

	t.Run("only 1 call on success", func(t *testing.T) {
		calls := int32(0)
		err := retry.AndThen(Run(&calls)).Do(ctx)
		assert.Nil(t, err)
		assert.Equal(t, int32(1), calls)
	})

	t.Run("returns err if never succeeds", func(t *testing.T) {
		calls := int32(0)
		a := func(ctx context.Context) error {
			atomic.AddInt32(&calls, 1)
			return io.ErrUnexpectedEOF
		}
		err := retry.AndThen(a).Do(ctx)
		assert.Equal(t, io.ErrUnexpectedEOF, err)
		assert.Equal(t, int32(4), calls)
	})
}

func TestForever(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*125)
	defer cancel()

	forever := action.Forever(time.Millisecond * 25)

	startedAt := time.Now()
	calls := int32(0)
	err := forever.AndThen(Run(&calls)).Do(ctx)
	assert.Nil(t, err)
	assert.True(t, calls > 3)
	assert.True(t, time.Now().Sub(startedAt) < time.Millisecond*200)
}
