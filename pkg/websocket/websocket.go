package websocket

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// MessageType 定义消息类型
type MessageType int

const (
	TextMessage MessageType = iota + 1
	BinaryMessage
	CloseMessage
	PingMessage
	PongMessage
)

// Message 定义消息结构
type Message struct {
	Type    int            `json:"type"`
	Content []byte         `json:"content"`
	From    string         `json:"from,omitempty"`
	To      string         `json:"to,omitempty"`
	Extra   map[string]any `json:"extra,omitempty"`
}

// Client 表示一个WebSocket客户端连接
type Client struct {
	ID         string
	Conn       *websocket.Conn
	Pool       *Pool
	mu         sync.Mutex
	closed     bool
	lastPing   time.Time
	Properties map[string]any // 用于存储自定义属性
	send       chan Message   // 发送消息的通道
}

// Pool 管理所有websocket连接
type Pool struct {
	clients     map[string]*Client
	mu          sync.RWMutex
	register    chan *Client
	unregister  chan *Client
	broadcast   chan Message
	groups      map[string]map[string]*Client // 群组功能：group_id -> map[client_id]*Client
	groupsMu    sync.RWMutex
	middlewares []MiddlewareFunc
}

// MiddlewareFunc 定义中间件函数类型
type MiddlewareFunc func(client *Client, message Message) (Message, error)

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
	MessageBuffer int // 消息缓冲区大小

	// 高级配置
	Compression     bool          // 是否启用压缩
	MaxMessageSize  int64         // 最大消息大小
	EnableHeartbeat bool          // 是否启用心跳检测
	HeartbeatPeriod time.Duration // 心跳周期

	// 自定义配置
	ClientIDGenerator func() string    // 自定义客户端ID生成器
	ErrorHandler      func(err error)  // 自定义错误处理
	Middlewares       []MiddlewareFunc // 消息处理中间件
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
		MaxMessageSize:  32 << 20, // 32MB
		EnableHeartbeat: true,
		HeartbeatPeriod: 30 * time.Second,
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
		clients:     make(map[string]*Client),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan Message, config.MessageBuffer),
		groups:      make(map[string]map[string]*Client),
		middlewares: config.Middlewares,
	}

	go pool.run()

	return websocket.New(func(c *websocket.Conn) {
		client := &Client{
			ID:         config.ClientIDGenerator(),
			Conn:       c,
			Pool:       pool,
			lastPing:   time.Now(),
			Properties: make(map[string]interface{}),
			send:       make(chan Message, config.MessageBuffer),
		}

		// 启动客户端的写入协程
		go client.writePump(config)

		pool.register <- client

		if config.Handler != nil {
			config.Handler.OnConnect(client)
		}

		defer func() {
			pool.unregister <- client
			if config.Handler != nil {
				config.Handler.OnClose(client)
			}
		}()

		// 设置连接参数
		c.SetReadLimit(config.MaxMessageSize)
		_ = c.SetReadDeadline(time.Now().Add(config.PingTimeout))

		// 心跳处理
		if config.EnableHeartbeat {
			c.SetPongHandler(func(string) error {
				client.UpdatePing()
				_ = c.SetReadDeadline(time.Now().Add(config.PingTimeout))
				return nil
			})
		}

		for {
			messageType, data, err := c.ReadMessage()
			if err != nil {
				if config.Handler != nil {
					config.Handler.OnError(client, err)
				}
				if config.ErrorHandler != nil {
					config.ErrorHandler(err)
				}
				break
			}

			message := Message{
				Type:    messageType,
				Content: data,
				From:    client.ID,
			}

			// 应用中间件
			for _, middleware := range pool.middlewares {
				if processedMsg, err := middleware(client, message); err != nil {
					if config.ErrorHandler != nil {
						config.ErrorHandler(err)
					}
					continue
				} else {
					message = processedMsg
				}
			}

			if config.Handler != nil {
				config.Handler.OnMessage(client, message)
			}
		}
	})
}

// writePump 处理消息发送
func (c *Client) writePump(config Config) {
	ticker := time.NewTicker(config.HeartbeatPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(config.WriteTimeout))
			if !ok {
				_ = c.Conn.WriteMessage(int(CloseMessage), []byte{})
				return
			}

			if err := c.Conn.WriteMessage(message.Type, message.Content); err != nil {
				return
			}

		case <-ticker.C:
			if config.EnableHeartbeat {
				_ = c.Conn.SetWriteDeadline(time.Now().Add(config.WriteTimeout))
				if err := c.Conn.WriteMessage(int(PingMessage), nil); err != nil {
					return
				}
			}
		}
	}
}
