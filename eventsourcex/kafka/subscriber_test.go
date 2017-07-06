package kafka_test

import (
	"context"
	"io"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/altairsix/pkg/eventsourcex"
	"github.com/altairsix/pkg/eventsourcex/kafka"
	"github.com/savaki/randx"
	"github.com/stretchr/testify/assert"
)

type MockConsumer struct {
	sarama.Consumer
	sarama.PartitionConsumer

	err      error
	messages <-chan *sarama.ConsumerMessage
}

func (m *MockConsumer) Partitions(topic string) ([]int32, error) {
	return []int32{0}, m.err
}

func (m *MockConsumer) Close() error {
	return m.err
}

func (m *MockConsumer) ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
	return m, m.err
}

func (m *MockConsumer) Messages() <-chan *sarama.ConsumerMessage {
	if m.messages == nil {
		panic("Use NewConsumer to instantiate MockConsumer instances")
	}
	return m.messages
}

func (m *MockConsumer) WithErr(err error) *MockConsumer {
	m.err = err
	return m
}

func NewConsumer(messages ...*sarama.ConsumerMessage) *MockConsumer {
	ch := make(chan *sarama.ConsumerMessage, len(messages))
	for _, m := range messages {
		ch <- m
	}
	close(ch)

	return &MockConsumer{
		messages: ch,
	}
}

func TestSubscriberFailsIfNoPartitions(t *testing.T) {
	consumer := NewConsumer().WithErr(io.ErrUnexpectedEOF)
	_, err := kafka.SubscribeStream(context.Background(), consumer, "topic", nil)
	assert.NotNil(t, err)
}

func TestSubscriber(t *testing.T) {
	receivedCount := 0
	expected := []byte("content")
	fn := eventsourcex.HandlerFunc(func(offset uint64, data []byte) {
		receivedCount++
		assert.Equal(t, expected, data)
	})

	topic := randx.AlphaN(12)
	consumer := NewConsumer(&sarama.ConsumerMessage{
		Key:   []byte("hello"),
		Value: expected,
	})

	ctx := context.Background()
	sub, err := kafka.SubscribeStream(ctx, consumer, topic, fn)
	if !assert.Nil(t, err) {
		return
	}
	sub.Shutdown(ctx)
	assert.Equal(t, 1, receivedCount)
}

func TestMakeTopicName(t *testing.T) {
	topic := kafka.MakeTopicName("a", "b", "c", "d")
	assert.Equal(t, "a.b.c.d", topic)
}
