package kafka

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/altairsix/eventsource"
	"github.com/altairsix/pkg/eventsourcex"
)

// NewPublisher creates a kafka publisher
func NewPublisher(ctx context.Context, producer sarama.SyncProducer, topic string) eventsourcex.PublisherFunc {
	return func(event eventsource.StreamRecord) error {
		_, _, err := producer.SendMessage(&sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(event.AggregateID),
			Value: sarama.ByteEncoder(event.Data),
		})
		return err
	}
}
