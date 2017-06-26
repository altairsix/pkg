package action

import (
	"context"
	"strconv"
	"time"

	"github.com/altairsix/pkg/tracer"
	"github.com/opentracing/opentracing-go/log"
)

type Tick struct {
	ID        string
	StartedAt time.Time
}

type Ticker interface {
	Publish(tick Tick) error
	Receive(ctx context.Context) (<-chan Tick, error)
}

type singleton struct {
	interval  time.Duration
	elections time.Duration
	lease     time.Duration
}

type SingletonOption func(*singleton)

func WithInterval(d time.Duration) SingletonOption {
	return func(s *singleton) {
		s.interval = d
	}
}

func WithElections(d time.Duration) SingletonOption {
	return func(s *singleton) {
		s.elections = d
	}
}

func WithLease(d time.Duration) SingletonOption {
	return func(s *singleton) {
		s.lease = d
	}
}

func Singleton(ticker Ticker, fn func(ctx context.Context) error, opts ...SingletonOption) func(ctx context.Context) error {
	cfg := &singleton{
		interval:  time.Second * 3,
		elections: time.Second * 13,
		lease:     time.Minute * 13,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return func(ctx context.Context) error {
		id := strconv.FormatInt(r.Int63(), 36)
		startedAt := time.Now()

		segment, ctx := tracer.NewSegment(ctx, "action:singleton")
		segment.SetBaggageItem("singleton-id", id)
		defer segment.Finish()

		ch, err := ticker.Receive(ctx)
		if err != nil {
			return err
		}

		leader := false
		restarts := 0

		var child context.Context
		var done chan struct{}
		cancel := func() {}

		errs := make(chan error, 10)
		leases := make([]time.Time, 0, 12)

		t := time.NewTicker(jitter(cfg.interval))
		defer t.Stop()

		elect := time.NewTicker(cfg.elections)
		defer elect.Stop()

		start := func() {
			child, cancel = context.WithCancel(ctx)
			done = make(chan struct{})
			go func() {
				defer close(done)
				errs <- fn(child)
			}()
		}

		for {
			select {
			case <-ctx.Done():
				return nil

			case v := <-ch:
				if v.ID != id && v.StartedAt.Before(startedAt) {
					leases = append(leases, time.Now().Add(cfg.lease))
				}

			case <-t.C:
				ticker.Publish(Tick{
					ID:        id,
					StartedAt: startedAt,
				})

				now := time.Now()
				for len(leases) > 0 && leases[0].Before(now) {
					leases = leases[1:]
				}
				if leader && len(leases) > 0 {
					segment.Info("singleton:lost_leadership")
					leader = false
					restarts = 0
					cancel()
				}

			case <-done:
				if leader {
					restarts++
					delay := cfg.interval
					if restarts > 100 {
						delay = time.Minute * 15
					} else if restarts > 20 {
						delay = time.Minute * 3
					}
					delay = jitter(delay)
					segment.Info("singleton:restart",
						log.Int("restarts", restarts),
						log.Int64("delay-ms", int64(delay/time.Millisecond)),
					)

					select {
					case <-ctx.Done():
						return nil
					case <-time.After(delay):
					}

					start()
				}

			case err := <-errs:
				if err != nil {
					segment.LogFields(log.Error(err))
					return err
				}

			case <-elect.C:
				if count := len(leases); (leader && count == 0) || (!leader && count > 0) {
					continue
				}

				if !leader {
					segment.Info("singleton:elected_leader")
					leader = true
					start()
				}
			}
		}

		return nil
	}
}