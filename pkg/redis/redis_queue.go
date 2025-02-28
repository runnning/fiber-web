package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrQueueClosed = errors.New("queue is closed")

	ErrCreateStream = errors.New("failed to create stream")

	ErrGroupExists = errors.New("consumer group already exists")
	ErrCreateGroup = errors.New("failed to create consumer group")
	ErrGetGroups   = errors.New("failed to get consumer groups")

	ErrMessageMarshal = errors.New("failed to marshal message")

	ErrMessageAck = errors.New("failed to acknowledge message")

	ErrWorkerTimeout = errors.New("timeout waiting for workers to finish")
)

// StreamMessage 表示队列中的消息
type StreamMessage struct {
	ID     string
	Values map[string]any
}

// MessageHandler 消息处理函数类型
type MessageHandler func(ctx context.Context, msg StreamMessage) error

// StreamOptions 队列配置选项
type StreamOptions struct {
	MaxLen         int64         // 流的最大长度，0 表示无限制
	ApproximateLen bool          // 是否使用 ~ 进行近似裁剪
	ReadTimeout    time.Duration // 读取超时时间
	WriteTimeout   time.Duration // 写入超时时间
	CloseTimeout   time.Duration // 关闭超时时间
}

// ConsumerOptions 消费者配置选项
type ConsumerOptions struct {
	BatchSize      int64         // 批量读取的大小
	BlockDuration  time.Duration // 阻塞读取的超时时间
	RetryDelay     time.Duration // 错误重试延迟
	MaxRetries     int           // 最大重试次数，-1 表示无限重试
	ConcurrentSize int           // 并发处理消息的数量
	MinIdleTime    time.Duration // 消息闲置多久后会被重新认领
}

// 默认配置
var (
	DefaultStreamOptions = &StreamOptions{
		MaxLen:         10000,
		ApproximateLen: true,
		ReadTimeout:    3 * time.Second,
		WriteTimeout:   3 * time.Second,
		CloseTimeout:   30 * time.Second,
	}

	DefaultConsumerOptions = &ConsumerOptions{
		BatchSize:      10,
		BlockDuration:  2 * time.Second,
		RetryDelay:     time.Second,
		MaxRetries:     3,
		ConcurrentSize: 1,
		MinIdleTime:    30 * time.Minute,
	}
)

// StreamQueue Redis Stream 队列实现
type StreamQueue struct {
	client  *Client
	stream  string
	options *StreamOptions
	wg      sync.WaitGroup
	closed  chan struct{}
	running bool
	mu      sync.RWMutex
}

// NewStreamQueue 创建新的 Stream 队列
func NewStreamQueue(client *Client, stream string, opts *StreamOptions) *StreamQueue {
	if opts == nil {
		opts = DefaultStreamOptions
	}
	return &StreamQueue{
		client:  client,
		stream:  stream,
		options: opts,
		closed:  make(chan struct{}),
		running: true,
	}
}

// isRunning 检查队列是否在运行
func (sq *StreamQueue) isRunning() bool {
	sq.mu.RLock()
	defer sq.mu.RUnlock()
	return sq.running
}

// marshalValue 序列化消息值
func (sq *StreamQueue) marshalValue(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrMessageMarshal, err)
	}
	return string(data), nil
}

// unmarshalStreamMessage 解析 Redis 消息为 StreamMessage
func (sq *StreamQueue) unmarshalStreamMessage(id string, values map[string]any) (StreamMessage, error) {
	msg := StreamMessage{
		ID:     id,
		Values: make(map[string]any, len(values)),
	}
	for k, v := range values {
		msg.Values[k] = v
	}
	return msg, nil
}

// Publish 发布消息到队列
func (sq *StreamQueue) Publish(ctx context.Context, values map[string]any) (string, error) {
	if !sq.isRunning() {
		return "", ErrQueueClosed
	}

	if sq.options.WriteTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, sq.options.WriteTimeout)
		defer cancel()
	}

	// 序列化消息内容
	args := make([]any, 0, len(values)*2)
	for k, v := range values {
		data, err := sq.marshalValue(v)
		if err != nil {
			return "", fmt.Errorf("key=%s: %w", k, err)
		}
		args = append(args, k, data)
	}

	// 使用 XADD 命令添加消息
	streamArgs := &redis.XAddArgs{
		Stream: sq.stream,
		Values: args,
	}

	if sq.options.MaxLen > 0 {
		streamArgs.MaxLen = sq.options.MaxLen
		streamArgs.Approx = sq.options.ApproximateLen
	}

	return sq.client.client.XAdd(ctx, streamArgs).Result()
}

// createConsumerGroup 创建消费者组
func (sq *StreamQueue) createConsumerGroup(ctx context.Context, groupName string) error {
	// 检查消费者组是否存在
	groups, err := sq.client.client.XInfoGroups(ctx, sq.stream).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// 流不存在，创建一个空消息
			_, err = sq.client.client.XAdd(ctx, &redis.XAddArgs{
				Stream: sq.stream,
				ID:     "0-0",
				Values: map[string]interface{}{"init": "init"},
			}).Result()
			if err != nil {
				return fmt.Errorf("%w: %v", ErrCreateStream, err)
			}
		} else {
			return fmt.Errorf("%w: %v", ErrGetGroups, err)
		}
	} else {
		for _, group := range groups {
			if group.Name == groupName {
				return nil
			}
		}
	}

	// 创建消费者组
	err = sq.client.client.XGroupCreate(ctx, sq.stream, groupName, "0").Err()
	if err != nil {
		var redisError interface{ RedisError() string }
		if errors.As(err, &redisError) && strings.Contains(redisError.RedisError(), "BUSYGROUP") {
			return ErrGroupExists
		}
		return fmt.Errorf("%w: %v", ErrCreateGroup, err)
	}
	return nil
}

// processMessage 处理单条消息
func (sq *StreamQueue) processMessage(ctx context.Context, groupName string, msg StreamMessage, handler MessageHandler, opts *ConsumerOptions) {
	retries := 0
	for {
		// 处理消息
		err := handler(ctx, msg)
		if err == nil {
			if ackErr := sq.client.client.XAck(ctx, sq.stream, groupName, msg.ID).Err(); ackErr != nil {
				fmt.Printf("%v: message_id=%s: %v\n", ErrMessageAck, msg.ID, ackErr)
			}
			return
		}

		// 处理重试
		retries++
		if opts.MaxRetries >= 0 && retries > opts.MaxRetries {
			fmt.Printf("Message %s exceeded max retries: %v\n", msg.ID, err)
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(opts.RetryDelay):
			continue
		}
	}
}

// handlePendingMessages 处理超时的待处理消息
func (sq *StreamQueue) handlePendingMessages(ctx context.Context, groupName, consumerName string, opts *ConsumerOptions, workChan chan<- StreamMessage) {
	ticker := time.NewTicker(opts.MinIdleTime / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 获取待处理消息
			pending, err := sq.client.client.XPending(ctx, sq.stream, groupName).Result()
			if err != nil || pending.Count == 0 {
				if errors.Is(err, redis.Nil) {
					continue
				}
				continue
			}

			// 获取详细的待处理消息信息
			entries, err := sq.client.client.XPendingExt(ctx, &redis.XPendingExtArgs{
				Stream: sq.stream,
				Group:  groupName,
				Start:  "-",
				End:    "+",
				Count:  pending.Count,
			}).Result()
			if err != nil {
				if errors.Is(err, redis.Nil) {
					continue
				}
				continue
			}

			// 处理每条超时消息
			for _, entry := range entries {
				if entry.Idle >= opts.MinIdleTime {
					messages, err := sq.client.client.XClaim(ctx, &redis.XClaimArgs{
						Stream:   sq.stream,
						Group:    groupName,
						Consumer: consumerName,
						MinIdle:  opts.MinIdleTime,
						Messages: []string{entry.ID},
					}).Result()
					if err != nil {
						if errors.Is(err, redis.Nil) {
							continue
						}
						continue
					}

					// 发送消息到工作通道
					for _, msg := range messages {
						streamMsg, err := sq.unmarshalStreamMessage(msg.ID, msg.Values)
						if err != nil {
							continue
						}

						select {
						case <-ctx.Done():
							return
						case workChan <- streamMsg:
						}
					}
				}
			}
		}
	}
}

// startWorkers 启动工作协程
func (sq *StreamQueue) startWorkers(ctx context.Context, groupName string, handler MessageHandler, opts *ConsumerOptions, workChan <-chan StreamMessage) {
	for i := 0; i < opts.ConcurrentSize; i++ {
		sq.wg.Add(1)
		go func() {
			defer sq.wg.Done()
			for msg := range workChan {
				select {
				case <-sq.closed:
					return
				default:
					sq.processMessage(ctx, groupName, msg, handler, opts)
				}
			}
		}()
	}
}

// readMessages 读取消息的主循环
func (sq *StreamQueue) readMessages(ctx context.Context, groupName, consumerName string, opts *ConsumerOptions, workChan chan<- StreamMessage) {
	defer close(workChan)

	for {
		select {
		case <-ctx.Done():
			return
		case <-sq.closed:
			return
		default:
			if !sq.isRunning() {
				return
			}

			// 读取新消息
			streams, err := sq.client.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    groupName,
				Consumer: consumerName,
				Streams:  []string{sq.stream, ">"},
				Count:    opts.BatchSize,
				Block:    opts.BlockDuration,
			}).Result()

			if err != nil {
				if !errors.Is(err, redis.Nil) {
					time.Sleep(opts.RetryDelay)
				}
				continue
			}

			// 处理消息
			for _, stream := range streams {
				for _, message := range stream.Messages {
					streamMsg, err := sq.unmarshalStreamMessage(message.ID, message.Values)
					if err != nil {
						continue
					}

					select {
					case <-ctx.Done():
						return
					case <-sq.closed:
						return
					case workChan <- streamMsg:
					}
				}
			}
		}
	}
}

// Consume 使用回调函数消费消息
func (sq *StreamQueue) Consume(ctx context.Context, groupName, consumerName string, handler MessageHandler, opts *ConsumerOptions) error {
	if !sq.isRunning() {
		return ErrQueueClosed
	}

	if opts == nil {
		opts = DefaultConsumerOptions
	}

	// 创建或检查消费者组
	if err := sq.createConsumerGroup(ctx, groupName); err != nil {
		return err
	}

	// 创建工作通道
	workChan := make(chan StreamMessage, opts.BatchSize)
	consumeCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 启动工作协程
	sq.startWorkers(consumeCtx, groupName, handler, opts, workChan)

	// 启动超时消息处理协程
	sq.wg.Add(1)
	go func() {
		defer sq.wg.Done()
		sq.handlePendingMessages(consumeCtx, groupName, consumerName, opts, workChan)
	}()

	// 启动消息读取协程
	go sq.readMessages(consumeCtx, groupName, consumerName, opts, workChan)

	// 等待上下文取消或队列关闭
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-sq.closed:
		return ErrQueueClosed
	}
}

// Close 优雅关闭队列
func (sq *StreamQueue) Close() error {
	sq.mu.Lock()
	if !sq.running {
		sq.mu.Unlock()
		return nil
	}
	sq.running = false
	close(sq.closed)
	sq.mu.Unlock()

	// 等待所有工作协程完成
	done := make(chan struct{})
	go func() {
		sq.wg.Wait()
		close(done)
	}()

	// 设置最大等待时间
	select {
	case <-done:
		return nil
	case <-time.After(sq.options.CloseTimeout):
		return ErrWorkerTimeout
	}
}
