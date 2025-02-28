package queue

import (
	"bytes"
	"compress/gzip"
	"context"
	"fiber_web/pkg/config"
	"fiber_web/pkg/logger"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/nsqio/go-nsq"
	"go.uber.org/zap"
)

// 监控指标
type metrics struct {
	published int64
	consumed  int64
	errors    int64
}

// Producer 封装 NSQ 生产者
type Producer struct {
	producer *nsq.Producer
	compress bool
	metrics  *metrics
}

// Consumer 封装 NSQ 消费者
type Consumer struct {
	consumer *nsq.Consumer
	topic    string
	channel  string
	metrics  *metrics
}

// Options 配置选项
type Options struct {
	MaxRetries        int
	RetryInterval     time.Duration
	WriteTimeout      time.Duration
	HeartbeatInterval time.Duration
	MaxInFlight       int
	ReadTimeout       time.Duration
	RequeueDelay      time.Duration
	MaxAttempts       uint16
	Compress          bool
}

// DefaultOptions 默认配置选项
var DefaultOptions = Options{
	MaxRetries:        3,
	RetryInterval:     time.Second,
	WriteTimeout:      time.Second * 5,
	HeartbeatInterval: time.Second * 30,
	MaxInFlight:       50,
	ReadTimeout:       time.Second * 60,
	RequeueDelay:      time.Second * 5,
	MaxAttempts:       5,
	Compress:          true,
}

// NewProducer 创建一个新的 NSQ 生产者
func NewProducer(cfg *config.NSQConfig, opts *Options) (*Producer, error) {
	if opts == nil {
		opts = &DefaultOptions
	}

	nsqConfig := nsq.NewConfig()
	nsqConfig.WriteTimeout = opts.WriteTimeout
	nsqConfig.HeartbeatInterval = opts.HeartbeatInterval

	addr := fmt.Sprintf("%s:%d", cfg.NSQD.Host, cfg.NSQD.Port)
	producer, err := nsq.NewProducer(addr, nsqConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 NSQ 生产者失败: %w", err)
	}
	producer.SetLoggerLevel(nsq.LogLevelError)

	if err := producer.Ping(); err != nil {
		return nil, fmt.Errorf("连接 NSQ 失败: %w", err)
	}

	return &Producer{
		producer: producer,
		compress: opts.Compress,
		metrics:  &metrics{},
	}, nil
}

// Publish 发布消息到 NSQ，支持重试和压缩
func (p *Producer) Publish(ctx context.Context, topic string, message []byte) error {
	if p.compress {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := gz.Write(message); err != nil {
			atomic.AddInt64(&p.metrics.errors, 1)
			return fmt.Errorf("压缩消息失败: %w", err)
		}
		_ = gz.Close()
		message = buf.Bytes()
	}

	var err error
	for i := 0; i < DefaultOptions.MaxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err = p.producer.Publish(topic, message); err == nil {
				atomic.AddInt64(&p.metrics.published, 1)
				return nil
			}
			atomic.AddInt64(&p.metrics.errors, 1)
			logger.Warn("发布消息失败，准备重试",
				zap.String("主题", topic),
				zap.Int("重试次数", i+1),
				zap.Error(err))
			time.Sleep(DefaultOptions.RetryInterval)
		}
	}
	return fmt.Errorf("发布消息失败，已重试 %d 次: %w", DefaultOptions.MaxRetries, err)
}

// NewConsumer 创建一个新的 NSQ 消费者
func NewConsumer(topic, channel string, cfg *config.NSQConfig, opts *Options) (*Consumer, error) {
	if opts == nil {
		opts = &DefaultOptions
	}

	nsqConfig := nsq.NewConfig()
	nsqConfig.MaxInFlight = opts.MaxInFlight
	nsqConfig.ReadTimeout = opts.ReadTimeout
	nsqConfig.HeartbeatInterval = opts.HeartbeatInterval
	nsqConfig.DefaultRequeueDelay = opts.RequeueDelay
	nsqConfig.MaxAttempts = opts.MaxAttempts

	consumer, err := nsq.NewConsumer(topic, channel, nsqConfig)
	if err != nil {
		logger.Error("创建 NSQ 消费者失败",
			zap.String("主题", topic),
			zap.String("通道", channel),
			zap.Error(err))
		return nil, err
	}

	// 设置日志级别
	consumer.SetLoggerLevel(nsq.LogLevelError)

	// 连接到 NSQ lookupd
	addr := fmt.Sprintf("%s:%d", cfg.Lookupd.Host, cfg.Lookupd.Port)
	if err := consumer.ConnectToNSQLookupd(addr); err != nil {
		logger.Error("连接 NSQ lookupd 失败",
			zap.String("地址", addr),
			zap.Error(err))
		return nil, err
	}

	logger.Info("成功连接 NSQ 消费者",
		zap.String("主题", topic),
		zap.String("通道", channel))

	return &Consumer{
		consumer: consumer,
		topic:    topic,
		channel:  channel,
		metrics:  &metrics{},
	}, nil
}

// Stop 停止 NSQ 生产者
func (p *Producer) Stop() {
	if p.producer != nil {
		p.producer.Stop()
	}
}

// Stop 停止 NSQ 消费者
func (c *Consumer) Stop() {
	if c.consumer != nil {
		c.consumer.Stop()
	}
}

// AddHandler 为消费者添加处理器，支持并发处理和消息解压缩
func (c *Consumer) AddHandler(handler nsq.Handler) {
	wrapper := nsq.HandlerFunc(func(msg *nsq.Message) error {
		reader, err := gzip.NewReader(bytes.NewReader(msg.Body))
		if err == nil {
			body, err := io.ReadAll(reader)
			_ = reader.Close()
			if err != nil {
				atomic.AddInt64(&c.metrics.errors, 1)
				return err
			}
			msg.Body = body
		}

		if err := handler.HandleMessage(msg); err != nil {
			atomic.AddInt64(&c.metrics.errors, 1)
			return err
		}

		atomic.AddInt64(&c.metrics.consumed, 1)
		return nil
	})

	c.consumer.AddConcurrentHandlers(wrapper, DefaultOptions.MaxInFlight)
}

// Stats 返回消费者统计信息
func (c *Consumer) Stats() *nsq.ConsumerStats {
	return c.consumer.Stats()
}

// MaxInFlight 设置最大处理中消息数
func (c *Consumer) MaxInFlight(max int) {
	c.consumer.ChangeMaxInFlight(max)
}

// GetMetrics 获取监控指标
func (p *Producer) GetMetrics() (published, errors int64) {
	return atomic.LoadInt64(&p.metrics.published),
		atomic.LoadInt64(&p.metrics.errors)
}

// GetMetrics 获取消费者监控指标
func (c *Consumer) GetMetrics() (consumed, errors int64) {
	return atomic.LoadInt64(&c.metrics.consumed),
		atomic.LoadInt64(&c.metrics.errors)
}
