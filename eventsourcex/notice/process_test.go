package notice_test

import (
	"context"
	"testing"

	"github.com/altairsix/pkg/eventsourcex/notice"
	"github.com/stretchr/testify/assert"
)

type Message struct {
	ID string
}

func (n Message) AggregateID() string {
	return n.ID
}

func (n Message) Close() error {
	return nil
}

func TestProcessFunc(t *testing.T) {
	message := Message{ID: "abc"}
	ch := make(chan notice.MessageCloser, 1)
	ch <- message
	close(ch)

	callCount := 0
	ctx := context.Background()
	notice.ProcessFunc(ctx, ch, func(ctx context.Context, notice notice.Message) {
		assert.Equal(t, notice, message)
		callCount++
	})

	assert.Equal(t, 1, callCount)
}
