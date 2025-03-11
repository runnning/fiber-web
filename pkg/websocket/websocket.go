package websocket

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// MessageType 定义消息类型
type MessageType int32

const (
	TextMessage MessageType = iota + 1
	BinaryMessage
	CloseMessage
	PingMessage
	PongMessage
)

// Message 定义消息结构
type Message struct {
	Type      int            `json:"type"`
	Content   []byte         `json:"content"`
	From      string         `json:"from,omitempty"`
	To        string         `json:"to,omitempty"`
	Extra     map[string]any `json:"extra,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	MessageID string         `json:"message_id"`
}

// Client 表示一个WebSocket客户端连接
type Client struct {
	ID         string
	Conn       *websocket.Conn
	Pool       *Pool
	mu         sync.RWMutex
	closed     atomic.Bool
	lastPing   atomic.Value
	Properties sync.Map
	send       chan Message
	ctx        context.Context
	cancel     context.CancelFunc
}

// Pool 管理所有websocket连接
type Pool struct {
	clients     sync.Map
	register    chan *Client
	unregister  chan *Client
	broadcast   chan Message
	groups      sync.Map
	middlewares []MiddlewareFunc
	metrics     *Metrics
}

// Metrics 用于统计和监控
type Metrics struct {
	ConnectedClients atomic.Int64
	MessagesSent     atomic.Int64
	MessagesReceived atomic.Int64
	ErrorCount       atomic.Int64
	LastErrorTime    atomic.Value
	AverageLatency   atomic.Int64
}

// MiddlewareFunc 定义中间件函数类型
type MiddlewareFunc func(context.Context, *Client, Message) (Message, error)

// Handler 定义了处理WebSocket消息的接口
type Handler interface {
	OnConnect(client *Client)
	OnMessage(client *Client, message Message)
	OnClose(client *Client)
	OnError(client *Client, err error)
}

// Config WebSocket配置
type Config struct {
	Handler Handler
	// 基础配置
	PingTimeout   time.Duration
	WriteTimeout  time.Duration
	ReadTimeout   time.Duration
	BufferSize    int
	MessageBuffer int

	// 高级配置
	Compression     bool
	MaxMessageSize  int64
	EnableHeartbeat bool
	HeartbeatPeriod time.Duration

	// 自定义配置
	ClientIDGenerator func() string
	ErrorHandler      func(err error)
	Middlewares       []MiddlewareFunc

	// 新增配置
	EnableMetrics   bool
	EnableRateLimit bool
	RateLimit       int // 每秒最大消息数
	MaxClients      int // 最大客户端连接数
	EnableSSL       bool
	SSLCert         string
	SSLKey          string

	// 心跳重连配置
	EnableReconnect bool          // 是否启用自动重连
	MaxRetries      int           // 最大重试次数
	RetryInterval   time.Duration // 重试间隔
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		PingTimeout:     60 * time.Second,
		WriteTimeout:    10 * time.Second,
		ReadTimeout:     10 * time.Second,
		BufferSize:      1024,
		MessageBuffer:   256,
		Compression:     true,
		MaxMessageSize:  32 << 20,
		EnableHeartbeat: true,
		HeartbeatPeriod: 30 * time.Second,
		EnableMetrics:   true,
		EnableRateLimit: true,
		RateLimit:       100,
		MaxClients:      10000,
		EnableReconnect: true,
		MaxRetries:      3,
		RetryInterval:   5 * time.Second,
		ClientIDGenerator: func() string {
			return fmt.Sprintf("%d-%s", time.Now().UnixNano(), time.Now().Format("20060102150405"))
		},
	}
}

// New 创建新的WebSocket处理器
func New(config Config) fiber.Handler {
	if config.PingTimeout == 0 {
		config.PingTimeout = DefaultConfig().PingTimeout
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = DefaultConfig().WriteTimeout
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = DefaultConfig().ReadTimeout
	}
	if config.BufferSize == 0 {
		config.BufferSize = DefaultConfig().BufferSize
	}
	if config.MessageBuffer == 0 {
		config.MessageBuffer = DefaultConfig().MessageBuffer
	}
	if config.ClientIDGenerator == nil {
		config.ClientIDGenerator = DefaultConfig().ClientIDGenerator
	}

	pool := &Pool{
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan Message, config.MessageBuffer),
		middlewares: config.Middlewares,
		metrics:     &Metrics{},
	}

	if config.EnableMetrics {
		go pool.collectMetrics()
	}

	go pool.run()

	return websocket.New(func(c *websocket.Conn) {
		if config.MaxClients > 0 && pool.metrics.ConnectedClients.Load() >= int64(config.MaxClients) {
			_ = c.WriteMessage(int(CloseMessage), []byte("max clients reached"))
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		client := &Client{
			ID:     config.ClientIDGenerator(),
			Conn:   c,
			Pool:   pool,
			send:   make(chan Message, config.MessageBuffer),
			ctx:    ctx,
			cancel: cancel,
		}
		client.lastPing.Store(time.Now())

		// 启动客户端的读写协程
		go client.writePump(config)
		go client.readPump(config)

		pool.register <- client
		pool.metrics.ConnectedClients.Add(1)

		if config.Handler != nil {
			config.Handler.OnConnect(client)
		}

		<-ctx.Done()
		pool.unregister <- client
		pool.metrics.ConnectedClients.Add(-1)
		if config.Handler != nil {
			config.Handler.OnClose(client)
		}
	})
}

// readPump 处理消息接收
func (c *Client) readPump(config Config) {
	defer c.cancel()

	c.Conn.SetReadLimit(config.MaxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(config.PingTimeout))
	retries := 0

	if config.EnableHeartbeat {
		c.Conn.SetPongHandler(func(string) error {
			c.UpdatePing()
			_ = c.Conn.SetReadDeadline(time.Now().Add(config.PingTimeout))
			retries = 0 // 重置重试次数
			return nil
		})
	}

	var messageCounter int
	var rateLimitReset time.Time

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			messageType, data, err := c.Conn.ReadMessage()
			if err != nil {
				if config.EnableReconnect && retries < config.MaxRetries {
					retries++
					time.Sleep(config.RetryInterval)
					continue
				}
				if config.Handler != nil {
					config.Handler.OnError(c, err)
				}
				if config.ErrorHandler != nil {
					config.ErrorHandler(err)
				}
				c.Pool.metrics.ErrorCount.Add(1)
				c.Pool.metrics.LastErrorTime.Store(time.Now())
				return
			}
			retries = 0 // 重置重试次数

			// 速率限制
			if config.EnableRateLimit {
				now := time.Now()
				if now.After(rateLimitReset) {
					messageCounter = 0
					rateLimitReset = now.Add(time.Second)
				}
				messageCounter++
				if messageCounter > config.RateLimit {
					continue
				}
			}

			message := Message{
				Type:      messageType,
				Content:   data,
				From:      c.ID,
				Timestamp: time.Now(),
				MessageID: fmt.Sprintf("%s-%d", c.ID, time.Now().UnixNano()),
			}

			c.Pool.metrics.MessagesReceived.Add(1)

			// 应用中间件
			ctx := context.Background()
			for _, middleware := range c.Pool.middlewares {
				if message, err = middleware(ctx, c, message); err != nil {
					if config.ErrorHandler != nil {
						config.ErrorHandler(err)
					}
					continue
				}
			}

			if config.Handler != nil {
				config.Handler.OnMessage(c, message)
			}
		}
	}
}

// writePump 处理消息发送
func (c *Client) writePump(config Config) {
	ticker := time.NewTicker(config.HeartbeatPeriod)
	retries := 0
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		case message, ok := <-c.send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(config.WriteTimeout))
			if !ok {
				_ = c.Conn.WriteMessage(int(CloseMessage), []byte{})
				return
			}

			start := time.Now()
			if err := c.Conn.WriteMessage(message.Type, message.Content); err != nil {
				if config.EnableReconnect && retries < config.MaxRetries {
					retries++
					time.Sleep(config.RetryInterval)
					continue
				}
				return
			}
			retries = 0 // 重置重试次数
			c.Pool.metrics.MessagesSent.Add(1)

			// 更新平均延迟
			latency := time.Since(start).Microseconds()
			currentAvg := c.Pool.metrics.AverageLatency.Load()
			if currentAvg == 0 {
				c.Pool.metrics.AverageLatency.Store(latency)
			} else {
				newAvg := (currentAvg + latency) / 2
				c.Pool.metrics.AverageLatency.Store(newAvg)
			}

		case <-ticker.C:
			if config.EnableHeartbeat {
				_ = c.Conn.SetWriteDeadline(time.Now().Add(config.WriteTimeout))
				if err := c.Conn.WriteMessage(int(PingMessage), nil); err != nil {
					if config.EnableReconnect && retries < config.MaxRetries {
						retries++
						time.Sleep(config.RetryInterval)
						continue
					}
					return
				}
				retries = 0 // 重置重试次数
			}
		}
	}
}

// collectMetrics 收集指标
func (p *Pool) collectMetrics() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// 这里可以实现指标导出到监控系统
		// 例如 Prometheus 或其他监控系统
	}
}
