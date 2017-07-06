package kafka_test

import (
	"os"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/altairsix/pkg/eventsourcex/kafka"
	"github.com/savaki/randx"
	"github.com/stretchr/testify/assert"
)

func TestEnvConfig(t *testing.T) {
	config := kafka.EnvConfig()
	assert.NotNil(t, config)
	assert.True(t, len(config.BrokerList) >= 1)
}

func TestProducerConsumer(t *testing.T) {
	config := kafka.EnvConfig()

	producer, err := kafka.Producer(config)
	assert.Nil(t, err)
	defer producer.Close()

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "test"
	}

	startedAt := time.Now()
	key := randx.AlphaN(12)
	value := randx.AlphaN(12)
	partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(value),
	})
	if !assert.Nil(t, err) {
		return
	}

	consumer, err := kafka.Consumer(config)
	if !assert.Nil(t, err) {
		return
	}
	defer consumer.Close()

	pc, err := consumer.ConsumePartition(topic, partition, offset)
	if !assert.Nil(t, err) {
		return
	}

	m := <-pc.Messages()
	assert.Equal(t, key, string(m.Key))
	assert.Equal(t, value, string(m.Value))
	assert.True(t, time.Now().Sub(startedAt) < time.Second*10, "expected round trip to kafka within 10s")
}
