package kafka

import (
	"context"
	"strings"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/altairsix/pkg/eventsourcex"
	"github.com/altairsix/pkg/tracer"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
)

// Subscription provides a reference to a running stream
type Subscription struct {
	cancel func()
	done   <-chan struct{}
}

// Done waits for the subscription to be finished
func (s *Subscription) Done() <-chan struct{} {
	return s.done
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
	segment, child := tracer.NewSegment(child, "kafka:subscription", log.String("topic", topic))

	partitions, err := consumer.Partitions(topic)
	if err != nil {
		cancel()
		segment.LogFields(log.Error(err))
		segment.Finish()
		return nil, errors.Wrapf(err, "unable to find partitions for topic, %v", topic)
	}
	segment.Info("kafka:partitions", log.Int("partition-count", len(partitions)))

	wg := &sync.WaitGroup{}
	wg.Add(len(partitions))
	for _, partition := range partitions {
		segment.Info("kafka:consuming_partition", log.Int32("partition", partition))
		c, err := consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			cancel()
			segment.LogFields(log.Error(err))
			segment.Finish()
			return nil, errors.Wrapf(err, "unable to consumer topic:partition, %v:%v", topic, partition)
		}

		go func() {
			defer c.Close()
			<-child.Done()
		}()

		go func() {
			defer wg.Done()
			for message := range c.Messages() {
				h.Receive(uint64(message.Offset), message.Value)
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer segment.Finish()
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
	for len(segments) > 0 && segments[0] == "" {
		segments = segments[1:]
	}
	return strings.Join(segments, ".")
}
