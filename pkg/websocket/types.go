package websocket

import (
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
)

// 消息类型常量
const (
	TextMessage MessageType = iota + 1
	BinaryMessage
	PingMessage
	PongMessage
	CloseMessage
)

// 系统事件常量
const (
	EventPing  = "ping"
	EventPong  = "pong"
	EventClose = "close"
	EventJoin  = "join_room"
	EventLeave = "leave_room"
)

// 默认配置常量
const (
	DefaultMessageBufferSize = 256
	DefaultMaxMessageSize    = 512 * 1024 // 512KB
)

type MessageType int

// Message 定义基础消息结构
type Message struct {
	Type   MessageType `json:"type"`
	Event  string      `json:"event"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
	Time   time.Time   `json:"time"`
	RoomID string      `json:"room_id,omitempty"`
	From   string      `json:"from,omitempty"`
	To     string      `json:"to,omitempty"`
}

// Client 表示一个WebSocket客户端连接
type Client struct {
	ID         string
	Conn       *websocket.Conn
	Hub        *Hub
	Send       chan Message
	Properties sync.Map
	mu         sync.RWMutex
	rooms      map[string]struct{}
	done       chan struct{} // 用于关闭信号
}

// Hub 管理所有WebSocket连接和房间
type Hub struct {
	clients    map[*Client]struct{}
	rooms      map[string]map[*Client]struct{}
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	stop       chan struct{} // 添加停止信号通道
	mu         sync.RWMutex
	config     *Config
}

// Config 定义WebSocket配置
type Config struct {
	WriteTimeout      time.Duration
	ReadTimeout       time.Duration
	PingInterval      time.Duration
	MaxMessageSize    int64
	MessageBuffer     int
	EnableCompression bool
	EnablePing        bool
	AuthHandler       AuthHandler
	ErrorHandler      ErrorHandler
	EventHandler      EventHandler
}

type AuthHandler interface {
	Authenticate(params map[string]interface{}) (string, error)
}

type ErrorHandler interface {
	HandleError(client *Client, err error)
}

type EventHandler interface {
	HandleEvent(client *Client, message Message) error
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		WriteTimeout:      10 * time.Second,
		ReadTimeout:       60 * time.Second,
		PingInterval:      30 * time.Second,
		MaxMessageSize:    DefaultMaxMessageSize,
		MessageBuffer:     DefaultMessageBufferSize,
		EnableCompression: true,
		EnablePing:        true,
	}
}
