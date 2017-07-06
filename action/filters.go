package action

import (
	"context"
	"time"

	"github.com/altairsix/pkg/tracer"
	"github.com/opentracing/opentracing-go/log"
)

// StopAfter stops an action after the specified period of time has elapsed
func StopAfter(d time.Duration) Filter {
	return func(a Action) Action {
		return func(ctx context.Context) error {
			child, cancel := context.WithTimeout(ctx, d)
			defer cancel()

			return a(child)
		}
	}
}

// Retry failed retries up to the specified number of times
func Retry(retries int, delay time.Duration) Filter {
	return func(a Action) Action {
		return func(ctx context.Context) (err error) {
			for attempt := 0; attempt <= retries; attempt++ {
				if err = a.Do(ctx); err == nil {
					return nil
				}

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(delay):
				}
			}
			return err
		}
	}
}

// Forever repeats the target action forever until the context is canceled
func Forever(delay time.Duration) Filter {
	return func(a Action) Action {
		return func(ctx context.Context) error {
			segment, child := tracer.NewSegment(ctx, "forever", log.Int64("delay-ms", int64(delay/time.Millisecond)))
			defer segment.Finish()

			for {
				if err := a.Do(child); err != nil {
					segment.Info("forever:err", log.Error(err))
				}

				select {
				case <-ctx.Done():
					return nil
				case <-time.After(jitter(delay)):
				}
			}
		}
	}
}
