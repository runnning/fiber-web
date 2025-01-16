package queue

import (
	"fiber_web/pkg/config"
	"fiber_web/pkg/logger"
	"fmt"
	"time"

	"github.com/nsqio/go-nsq"
	"go.uber.org/zap"
)

// Producer 封装 NSQ 生产者
type Producer struct {
	producer *nsq.Producer
}

// Consumer 封装 NSQ 消费者
type Consumer struct {
	consumer *nsq.Consumer
}

// NewProducer 创建一个新的 NSQ 生产者
func NewProducer(cfg *config.NSQConfig) (*Producer, error) {
	nsqConfig := nsq.NewConfig()
	nsqConfig.DefaultRequeueDelay = time.Second * 5

	addr := fmt.Sprintf("%s:%d", cfg.NSQD.Host, cfg.NSQD.Port)
	prod, err := nsq.NewProducer(addr, nsqConfig)
	if err != nil {
		logger.Error("创建 NSQ 生产者失败", zap.Error(err))
		return nil, err
	}

	// 测试连接
	if err := prod.Ping(); err != nil {
		logger.Error("连接 NSQ 失败", zap.Error(err))
		return nil, err
	}

	logger.Info("成功连接到 NSQ")
	return &Producer{producer: prod}, nil
}

// NewConsumer 创建一个新的 NSQ 消费者
func NewConsumer(topic, channel string, cfg *config.Config) (*Consumer, error) {
	nsqConfig := nsq.NewConfig()
	nsqConfig.DefaultRequeueDelay = time.Second * 5
	nsqConfig.MaxInFlight = 10

	consumer, err := nsq.NewConsumer(topic, channel, nsqConfig)
	if err != nil {
		logger.Error("创建 NSQ 消费者失败",
			zap.String("主题", topic),
			zap.String("通道", channel),
			zap.Error(err))
		return nil, err
	}

	// 连接到 NSQ lookupd
	addr := fmt.Sprintf("%s:%d", cfg.NSQ.Lookupd.Host, cfg.NSQ.Lookupd.Port)
	if err := consumer.ConnectToNSQLookupd(addr); err != nil {
		logger.Error("连接 NSQ lookupd 失败",
			zap.String("地址", addr),
			zap.Error(err))
		return nil, err
	}

	logger.Info("成功连接 NSQ 消费者",
		zap.String("主题", topic),
		zap.String("通道", channel))

	return &Consumer{consumer: consumer}, nil
}

// Publish 发布消息到 NSQ
func (p *Producer) Publish(topic string, message []byte) error {
	if p.producer == nil {
		return fmt.Errorf("NSQ 生产者未初始化")
	}
	return p.producer.Publish(topic, message)
}

// PublishDeferred 发布延迟消息到 NSQ
func (p *Producer) PublishDeferred(topic string, delay time.Duration, message []byte) error {
	if p.producer == nil {
		return fmt.Errorf("NSQ 生产者未初始化")
	}
	return p.producer.DeferredPublish(topic, delay, message)
}

// Stop 停止 NSQ 生产者
func (p *Producer) Stop() error {
	if p.producer != nil {
		p.producer.Stop()
	}
	return nil
}

// Stop 停止 NSQ 消费者
func (c *Consumer) Stop() error {
	if c.consumer != nil {
		c.consumer.Stop()
	}
	return nil
}

// AddHandler 为消费者添加处理器
func (c *Consumer) AddHandler(handler nsq.Handler) {
	c.consumer.AddHandler(handler)
}

// Stats 返回消费者统计信息
func (c *Consumer) Stats() *nsq.ConsumerStats {
	return c.consumer.Stats()
}

// MaxInFlight 设置最大处理中消息数
func (c *Consumer) MaxInFlight(max int) {
	c.consumer.ChangeMaxInFlight(max)
}
