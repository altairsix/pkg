package domainx

import (
	"time"

	"github.com/nats-io/go-nats"
)

// AggregateSubject returns the subject for a specific bounded context
func AggregateSubject(env, boundedContext string) string {
	return env + ".aggregate." + boundedContext
}

func subscribeForUpdates(nc *nats.Conn, subject string, timeout time.Duration) <-chan struct{} {
	updated := make(chan struct{}, 1)

	sub, err := nc.Subscribe(subject, func(m *nats.Msg) {
		select {
		case updated <- struct{}{}:
		default:
		}
	})
	if err != nil {
		close(updated)
		return updated
	}

	go func() {
		defer close(updated)
		defer sub.Unsubscribe()

		select {
		case <-time.After(timeout):
		case <-updated:
		}
	}()

	return updated
}
