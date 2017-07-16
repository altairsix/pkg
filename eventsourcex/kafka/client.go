package kafka

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"strings"

	"github.com/Shopify/sarama"
)

// Config contains the configuration parameters for the kafka producer
type Config struct {
	CertPEM    []byte
	KeyPEM     []byte
	CaPEM      []byte
	BrokerList []string
	VerifyTLS  bool
}

func (c *Config) tlsEnabled() bool {
	return c.CertPEM != nil && c.KeyPEM != nil && c.CaPEM != nil
}

// Build uses the config to construct a new *sarama.Config.  For now the *Config
// only provides support for TLS sarama connections
func (c *Config) Build() (*sarama.Config, error) {
	config := sarama.NewConfig()

	if c.tlsEnabled() {
		tlsConfig, err := createTLSConfiguration(c)
		if err != nil {
			return nil, err
		}
		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Enable = true
	}

	return config, nil
}

func getBytes(name string) []byte {
	v := os.Getenv(name)
	if v == "" {
		return nil
	}

	return []byte(v)
}

func getArrayOrElse(name string, defaultValue []string) []string {
	v := os.Getenv(name)
	if v == "" {
		return defaultValue
	}

	segments := strings.Split(v, ",")
	arr := make([]string, 0, len(segments))
	for _, item := range segments {
		arr = append(arr, strings.TrimSpace(item))
	}
	return arr
}

// EnvConfig returns a new Config instance populated with values from the environment.
// Expected keys are KAFKA_CERT, KAFKA_KEY, KAFKA_CA, KAFKA_BROKERS
// If KAFKA_BROKERS is not set, defaults to localhost:9092
func EnvConfig() *Config {
	return &Config{
		CertPEM:    getBytes("KAFKA_CERT"),
		KeyPEM:     getBytes("KAFKA_KEY"),
		CaPEM:      getBytes("KAFKA_CA"),
		BrokerList: getArrayOrElse("KAFKA_BROKERS", []string{"localhost:9092"}),
	}
}

func createTLSConfiguration(cfg *Config) (*tls.Config, error) {
	cert, err := tls.X509KeyPair(cfg.CertPEM, cfg.KeyPEM)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cfg.CaPEM)

	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: true,
	}, nil
}

// Option provides functional operators for Sarama
type Option func(*sarama.Config)

// Producer creates a new kafka sync producer
func Producer(cfg *Config, opts ...Option) (sarama.SyncProducer, error) {

	// For the data collector, we are looking for strong consistency semantics.
	// Because we don't change the flush settings, sarama will try to produce messages
	// as fast as possible to keep latency low.
	config, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	config.Producer.Return.Successes = true

	for _, opt := range opts {
		opt(config)
	}

	// On the broker side, you may want to change the following settings to get
	// stronger consistency guarantees:
	// - For your broker, set `unclean.leader.election.enable` to false
	// - For the topic, you could increase `min.insync.replicas`.

	producer, err := sarama.NewSyncProducer(cfg.BrokerList, config)
	if err != nil {
		return nil, err
	}

	return producer, nil
}

// Consumer creates a new kafka sync producer
func Consumer(cfg *Config, opts ...Option) (sarama.Consumer, error) {
	config, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(config)
	}

	return sarama.NewConsumer(cfg.BrokerList, config)
}
