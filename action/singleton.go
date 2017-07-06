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

type Heartbeat interface {
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

// Singleton takes an instance and ensure that only a single instance of it will
// run
func Singleton(heartbeat Heartbeat, opts ...SingletonOption) Filter {
	cfg := &singleton{
		interval:  time.Second * 3,
		elections: time.Second * 13,
		lease:     time.Minute * 13,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return func(a Action) Action {
		return func(ctx context.Context) error {
			id := strconv.FormatInt(r.Int63(), 36)
			startedAt := time.Now()

			segment, ctx := tracer.NewSegment(ctx, "action:singleton")
			segment.SetBaggageItem("singleton-id", id)
			defer segment.Finish()

			ch, err := heartbeat.Receive(ctx)
			if err != nil {
				return err
			}

			var finished chan error
			cancel := func() {}

			leader := false
			leases := make([]time.Time, 0, 12)

			t := time.NewTicker(jitter(cfg.interval))
			defer t.Stop()

			election := time.NewTicker(cfg.elections)
			defer election.Stop()

			run := func() {
				finished = make(chan error, 1) // done is scoped to our parent
				defer close(finished)

				var child context.Context
				child, cancel = context.WithCancel(ctx) // child and cancel are also both scoped to our parent
				defer cancel()

				childSegment, child := tracer.NewSegment(child, "singleton:run:finished")
				childSegment.Info("singleton:run:started")
				defer childSegment.Finish()

				finished <- a.Do(child)
			}

			for {
				select {
				case <-ctx.Done():
					segment.Info("singleton:canceled")
					return nil

				case v := <-ch:
					if v.ID != id && v.StartedAt.Before(startedAt) {
						leases = append(leases, time.Now().Add(cfg.lease))
					}

				case <-t.C:
					heartbeat.Publish(Tick{
						ID:        id,
						StartedAt: startedAt,
					})

					now := time.Now()
					for len(leases) > 0 && leases[0].Before(now) {
						leases = leases[1:]
					}
					if leader && len(leases) > 0 {
						segment.Info("singleton:lost_leadership")
						cancel()
					}

				case err := <-finished:
					if err != nil {
						segment.LogFields(log.Error(err))
					}
					return err

				case <-election.C:
					if leader {
						continue
					}

					if count := len(leases); count != 0 {
						cancel()
						return nil
					}

					segment.Info("singleton:elected_leader")
					go run()

					leader = true
				}
			}
		}
	}
}
