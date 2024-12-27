package queue

import (
	"fiber_web/pkg/config"
	"fiber_web/pkg/logger"
	"fmt"
	"time"

	"github.com/nsqio/go-nsq"
	"go.uber.org/zap"
)

// Producer wraps NSQ producer
type Producer struct {
	producer *nsq.Producer
}

// Consumer wraps NSQ consumer
type Consumer struct {
	consumer *nsq.Consumer
}

// NewProducer creates a new NSQ producer
func NewProducer(cfg *config.NSQConfig) (*Producer, error) {
	nsqConfig := nsq.NewConfig()
	nsqConfig.DefaultRequeueDelay = time.Second * 5

	addr := fmt.Sprintf("%s:%d", cfg.NSQD.Host, cfg.NSQD.Port)
	prod, err := nsq.NewProducer(addr, nsqConfig)
	if err != nil {
		logger.Error("Failed to create NSQ producer", zap.Error(err))
		return nil, err
	}

	// Test the connection
	if err := prod.Ping(); err != nil {
		logger.Error("Failed to connect to NSQ", zap.Error(err))
		return nil, err
	}

	logger.Info("Successfully connected to NSQ")
	return &Producer{producer: prod}, nil
}

// NewConsumer creates a new NSQ consumer
func NewConsumer(topic, channel string, cfg *config.Config) (*Consumer, error) {
	nsqConfig := nsq.NewConfig()
	nsqConfig.DefaultRequeueDelay = time.Second * 5
	nsqConfig.MaxInFlight = 10

	consumer, err := nsq.NewConsumer(topic, channel, nsqConfig)
	if err != nil {
		logger.Error("Failed to create NSQ consumer",
			zap.String("topic", topic),
			zap.String("channel", channel),
			zap.Error(err))
		return nil, err
	}

	// Connect to NSQ lookupd
	addr := fmt.Sprintf("%s:%d", cfg.NSQ.Lookupd.Host, cfg.NSQ.Lookupd.Port)
	if err := consumer.ConnectToNSQLookupd(addr); err != nil {
		logger.Error("Failed to connect to NSQ lookupd",
			zap.String("address", addr),
			zap.Error(err))
		return nil, err
	}

	logger.Info("Successfully connected NSQ consumer",
		zap.String("topic", topic),
		zap.String("channel", channel))

	return &Consumer{consumer: consumer}, nil
}

// Publish publishes a message to NSQ
func (p *Producer) Publish(topic string, message []byte) error {
	if p.producer == nil {
		return fmt.Errorf("NSQ producer not initialized")
	}
	return p.producer.Publish(topic, message)
}

// PublishDeferred publishes a delayed message to NSQ
func (p *Producer) PublishDeferred(topic string, delay time.Duration, message []byte) error {
	if p.producer == nil {
		return fmt.Errorf("NSQ producer not initialized")
	}
	return p.producer.DeferredPublish(topic, delay, message)
}

// Stop stops the NSQ producer
func (p *Producer) Stop() error {
	if p.producer != nil {
		p.producer.Stop()
	}
	return nil
}

// Stop stops the NSQ consumer
func (c *Consumer) Stop() error {
	if c.consumer != nil {
		c.consumer.Stop()
	}
	return nil
}

// AddHandler adds a handler for the consumer
func (c *Consumer) AddHandler(handler nsq.Handler) {
	c.consumer.AddHandler(handler)
}

// Stats returns consumer statistics
func (c *Consumer) Stats() *nsq.ConsumerStats {
	return c.consumer.Stats()
}

// MaxInFlight sets the maximum number of messages in flight
func (c *Consumer) MaxInFlight(max int) {
	c.consumer.ChangeMaxInFlight(max)
}
