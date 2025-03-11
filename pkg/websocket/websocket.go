package websocket

import (
	"context"
	"fmt"
	"log"
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
		BufferSize:      4096,
		MessageBuffer:   1024,
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

// writePump 处理消息发送
func (c *Client) writePump(config Config) {
	if c == nil || c.Conn == nil {
		return
	}

	ticker := time.NewTicker(config.HeartbeatPeriod)
	retries := 0
	defer func() {
		ticker.Stop()
		if r := recover(); r != nil {
			log.Printf("Recovered in writePump: %v", r)
		}
		if c != nil && c.Pool != nil && !c.closed.Load() {
			c.Pool.unregister <- c
		}
	}()

	for {
		select {
		case <-c.ctx.Done():
			c.mu.Lock()
			if c.Conn != nil && !c.closed.Load() {
				// 正常关闭连接
				_ = c.Conn.WriteControl(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, "connection closed"),
					time.Now().Add(time.Second),
				)
			}
			c.mu.Unlock()
			return
		case message, ok := <-c.send:
			if !ok {
				c.mu.Lock()
				if c.Conn != nil && !c.closed.Load() {
					// 发送关闭帧
					_ = c.Conn.WriteControl(
						websocket.CloseMessage,
						websocket.FormatCloseMessage(websocket.CloseNormalClosure, "channel closed"),
						time.Now().Add(time.Second),
					)
				}
				c.mu.Unlock()
				return
			}

			// 预先检查连接状态
			if c.closed.Load() {
				return
			}

			start := time.Now()
			c.mu.Lock()
			if c.Conn == nil {
				c.mu.Unlock()
				return
			}

			// 设置写入超时
			_ = c.Conn.SetWriteDeadline(time.Now().Add(config.WriteTimeout))

			// 快速解锁以减少锁持有时间
			conn := c.Conn
			c.mu.Unlock()

			err := conn.WriteMessage(message.Type, message.Content)

			if err != nil {
				if !websocket.IsCloseError(err,
					websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseNoStatusReceived) {
					log.Printf("Write error: %v", err)
				}
				if config.EnableReconnect && retries < config.MaxRetries &&
					!websocket.IsCloseError(err,
						websocket.CloseNormalClosure,
						websocket.CloseGoingAway) {
					retries++
					time.Sleep(config.RetryInterval)
					continue
				}
				return
			}
			retries = 0 // 重置重试次数

			// 更新指标
			if c.Pool != nil && c.Pool.metrics != nil {
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
			}

		case <-ticker.C:
			if config.EnableHeartbeat {
				c.mu.Lock()
				if c.Conn == nil || c.closed.Load() {
					c.mu.Unlock()
					return
				}

				_ = c.Conn.SetWriteDeadline(time.Now().Add(config.WriteTimeout))
				err := c.Conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(config.WriteTimeout))
				c.mu.Unlock()

				if err != nil {
					if !websocket.IsCloseError(err,
						websocket.CloseNormalClosure,
						websocket.CloseGoingAway,
						websocket.CloseNoStatusReceived) {
						log.Printf("Ping error: %v", err)
					}
					if config.EnableReconnect && retries < config.MaxRetries &&
						!websocket.IsCloseError(err,
							websocket.CloseNormalClosure,
							websocket.CloseGoingAway) {
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

// readPump 处理消息接收
func (c *Client) readPump(config Config) {
	if c == nil || c.Conn == nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered in readPump: %v", r)
		}
		if c != nil && c.Pool != nil && !c.closed.Load() {
			c.Pool.unregister <- c
		}
	}()

	c.mu.Lock()
	if c.Conn != nil && !c.closed.Load() {
		c.Conn.SetReadLimit(config.MaxMessageSize)
		_ = c.Conn.SetReadDeadline(time.Now().Add(config.PingTimeout))
	}
	c.mu.Unlock()

	retries := 0

	if config.EnableHeartbeat {
		c.Conn.SetPongHandler(func(string) error {
			c.UpdatePing()
			c.mu.Lock()
			if c.Conn != nil && !c.closed.Load() {
				_ = c.Conn.SetReadDeadline(time.Now().Add(config.PingTimeout))
			}
			c.mu.Unlock()
			retries = 0 // 重置重试次数
			return nil
		})

		// 设置关闭处理器
		c.Conn.SetCloseHandler(func(code int, text string) error {
			if code != websocket.CloseNoStatusReceived {
				log.Printf("Connection closed with code %d: %s", code, text)
			}
			c.mu.Lock()
			if c.Conn != nil && !c.closed.Load() {
				message := websocket.FormatCloseMessage(code, "")
				_ = c.Conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(time.Second))
			}
			c.mu.Unlock()
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
			c.mu.Lock()
			if c.Conn == nil || c.closed.Load() {
				c.mu.Unlock()
				return
			}
			messageType, data, err := c.Conn.ReadMessage()
			c.mu.Unlock()

			if err != nil {
				if !websocket.IsCloseError(err,
					websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseNoStatusReceived) {
					log.Printf("Read error: %v", err)
				}
				if config.EnableReconnect && retries < config.MaxRetries &&
					!websocket.IsCloseError(err,
						websocket.CloseNormalClosure,
						websocket.CloseGoingAway) {
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
				if c.Pool != nil && c.Pool.metrics != nil {
					c.Pool.metrics.ErrorCount.Add(1)
					c.Pool.metrics.LastErrorTime.Store(time.Now())
				}
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

			if c.Pool != nil && c.Pool.metrics != nil {
				c.Pool.metrics.MessagesReceived.Add(1)
			}

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

// collectMetrics 收集指标
func (p *Pool) collectMetrics() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// 这里可以实现指标导出到监控系统
		// 例如 Prometheus 或其他监控系统
	}
}
