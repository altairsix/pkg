package natsx

import (
	"bytes"
	"context"
	"strconv"
	"time"

	"github.com/altairsix/pkg/tracer"
	"github.com/nats-io/go-nats"
	"github.com/opentracing/opentracing-go/log"
)

// Func defines a function for use by Singleton
type Func func(ctx context.Context) error

// Singleton ensures that only one instance of a long running function will be running at a time
func Singleton(nc *nats.Conn, subject string, interval time.Duration, fn Func) (<-chan struct{}, Func) {
	// done signals once the singleton has completed
	done := make(chan struct{})

	// payload contains the start time for this singleton.  the current election rule is the singleton that
	// was started first wins
	payload := []byte(strconv.FormatInt(time.Now().UnixNano(), 10))

	return done, func(parent context.Context) error {
		ctx, cancel := context.WithCancel(parent)
		defer cancel()

		segment, _ := tracer.NewSegment(parent, "natz.singleton",
			log.String("subject", subject),
			log.String("payoad", string(payload)),
		)

		sub, err := nc.Subscribe(subject, func(msg *nats.Msg) {
			if v := msg.Data; v != nil && bytes.Compare(payload, v) > 0 {
				segment.Info("natz.singleton.yield", log.String("other-payload", string(v)))
				cancel()
			}
		})
		if err != nil {
			segment.Info("natz.singleton.err.subscribe_failed", log.Error(err))
			return err
		}

		// Broadcast ping at interval to prevent more than one action from running
		//
		go func() {
			t := time.NewTicker(interval)
			defer t.Stop()

			for {
				select {
				case <-t.C:
					nc.Publish(subject, payload)
				case <-ctx.Done():
					segment.Info("natz.singleton.stop_ticker")
					return
				}
			}
		}()

		// Unsubscribe once the action is completed
		//
		go func() {
			defer close(done)

			select {
			case <-ctx.Done():
				segment.Info("natz.singleton.cleanup")
				sub.Unsubscribe()
			}
		}()

		return fn(ctx)
	}
}
