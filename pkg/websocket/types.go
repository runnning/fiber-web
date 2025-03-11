package websocket

import (
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
)

// MessageType 定义消息类型
type MessageType int

const (
	TextMessage MessageType = iota + 1
	BinaryMessage
	PingMessage
	PongMessage
	CloseMessage
)

// Message 定义基础消息结构
type Message struct {
	Type   MessageType `json:"type"`
	Event  string      `json:"event"`             // 事件类型
	Data   interface{} `json:"data,omitempty"`    // 消息数据
	Error  string      `json:"error,omitempty"`   // 错误信息
	Time   time.Time   `json:"time"`              // 消息时间
	RoomID string      `json:"room_id,omitempty"` // 房间ID
	From   string      `json:"from,omitempty"`    // 发送者
	To     string      `json:"to,omitempty"`      // 接收者
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
}

// Hub 管理所有WebSocket连接和房间
type Hub struct {
	// 所有活跃的客户端
	clients map[*Client]struct{}

	// 房间管理，key是房间ID，value是房间中的客户端集合
	rooms map[string]map[*Client]struct{}

	// 广播消息通道
	broadcast chan Message

	// 注册新客户端
	register chan *Client

	// 注销客户端
	unregister chan *Client

	// 互斥锁保护并发访问
	mu sync.RWMutex

	// 配置选项
	config *Config
}

// Config 定义WebSocket配置
type Config struct {
	// 基础配置
	WriteTimeout   time.Duration // 写超时
	ReadTimeout    time.Duration // 读超时
	PingInterval   time.Duration // 心跳间隔
	MaxMessageSize int64         // 最大消息大小
	MessageBuffer  int           // 消息缓冲区大小

	// 高级配置
	EnableCompression bool         // 是否启用压缩
	EnablePing        bool         // 是否启用心跳
	AuthHandler       AuthHandler  // 认证处理器
	ErrorHandler      ErrorHandler // 错误处理器
	EventHandler      EventHandler // 事件处理器
}

// AuthHandler 定义认证处理接口
type AuthHandler interface {
	Authenticate(params map[string]interface{}) (string, error)
}

// ErrorHandler 定义错误处理接口
type ErrorHandler interface {
	HandleError(client *Client, err error)
}

// EventHandler 定义事件处理接口
type EventHandler interface {
	HandleEvent(client *Client, message Message) error
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		WriteTimeout:      10 * time.Second,
		ReadTimeout:       60 * time.Second,
		PingInterval:      30 * time.Second,
		MaxMessageSize:    512 * 1024, // 512KB
		MessageBuffer:     256,
		EnableCompression: true,
		EnablePing:        true,
	}
}
