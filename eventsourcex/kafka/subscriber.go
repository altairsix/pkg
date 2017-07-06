package kafka

import (
	"context"
	"strings"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/altairsix/pkg/eventsourcex"
	"github.com/pkg/errors"
)

// Subscription provides a reference to a running stream
type Subscription struct {
	cancel func()
	done   <-chan struct{}
}

// Shutdown attempts to stop all the running goroutines
func (s *Subscription) Shutdown(ctx context.Context) error {
	s.cancel()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.done:
		return nil
	}
}

// SubscribeStream subscribes a message handler to the specified Kafka topic
func SubscribeStream(ctx context.Context, consumer sarama.Consumer, topic string, h eventsourcex.Handler) (*Subscription, error) {
	child, cancel := context.WithCancel(ctx)

	partitions, err := consumer.Partitions(topic)
	if err != nil {
		cancel()
		return nil, errors.Wrapf(err, "unable to find partitions for topic, %v", topic)
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(partitions))
	for _, partition := range partitions {
		c, err := consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			cancel()
			return nil, errors.Wrapf(err, "unable to consumer topic:partition, %v:%v", topic, partition)
		}

		go func() {
			defer c.Close()
			<-child.Done()
		}()

		go func() {
			defer wg.Done()
			for message := range c.Messages() {
				h.Handle(uint64(message.Offset), message.Value)
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
	}()

	return &Subscription{
		cancel: cancel,
		done:   done,
	}, nil
}

// MakeTopicName returns the topic name for the arguments provided; prefix is a convenience for the Heroku Kafka topic prefix
func MakeTopicName(prefix, env, boundedContext string, args ...string) string {
	segments := []string{prefix, env, boundedContext}
	segments = append(segments, args...)
	return strings.Join(segments, ".")
}
