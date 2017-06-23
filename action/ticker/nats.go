package ticker

import (
	"context"
	"encoding/json"
	"sync/atomic"

	"github.com/altairsix/pkg/action"
	"github.com/nats-io/go-nats"
)

type natsTicker struct {
	nc      *nats.Conn
	subject string
}

func (n *natsTicker) Publish(tick action.Tick) error {
	data, err := json.Marshal(tick)
	if err != nil {
		return err
	}

	return n.nc.Publish(n.subject, data)
}

func (n *natsTicker) Receive(ctx context.Context) (<-chan action.Tick, error) {
	ch := make(chan action.Tick, 16)
	open := int32(1)

	sub, err := n.nc.Subscribe(n.subject, func(msg *nats.Msg) {
		tick := action.Tick{}
		if err := json.Unmarshal(msg.Data, &tick); err == nil {
			if atomic.LoadInt32(&open) == 1 {
				ch <- tick
			}
		}
	})
	if err != nil {
		return nil, err
	}

	go func() {
		select {
		case <-ctx.Done():
			atomic.StoreInt32(&open, 1)
			close(ch)
			sub.Unsubscribe()
		}
	}()

	return ch, nil
}

func Nats(nc *nats.Conn, subject string) action.Ticker {
	return &natsTicker{
		nc:      nc,
		subject: subject,
	}
}
