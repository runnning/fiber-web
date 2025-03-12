package websocket

import (
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// 连接限制常量
const (
	maxConcurrentConnections = 1000
	connectionTimeout        = 10 * time.Second
)

var (
	connectionSemaphore = make(chan struct{}, maxConcurrentConnections)
	activeConnections   sync.Map
)

// Handler 创建一个新的 WebSocket 处理器
func Handler(config *Config) fiber.Handler {
	hub := NewHub(config)
	go hub.Run()

	return websocket.New(handleWebSocket(hub, config), getWebSocketConfig(config))
}

// handleWebSocket 处理WebSocket连接
func handleWebSocket(hub *Hub, config *Config) func(*websocket.Conn) {
	return func(c *websocket.Conn) {
		// 尝试获取连接许可
		select {
		case connectionSemaphore <- struct{}{}:
			defer func() { <-connectionSemaphore }()
		case <-time.After(connectionTimeout):
			_ = c.WriteJSON(Message{
				Type:  TextMessage,
				Event: "error",
				Error: "connection limit reached",
				Time:  time.Now(),
			})
			_ = c.Close()
			return
		}

		// 验证客户端
		if !authenticateClient(c, config) {
			return
		}

		client := NewClient(c, hub)
		activeConnections.Store(client.ID, client)
		defer activeConnections.Delete(client.ID)

		// 注册客户端
		hub.register <- client

		// 启动读写协程
		done := make(chan struct{})
		go func() {
			client.WritePump()
			close(done)
		}()

		client.ReadPump()
		<-done // 等待写入协程完成
	}
}

// authenticateClient 认证客户端
func authenticateClient(c *websocket.Conn, config *Config) bool {
	if config.AuthHandler == nil {
		return true
	}

	authChan := make(chan bool, 1)
	timeout := time.After(5 * time.Second)

	go func() {
		params := make(map[string]interface{})
		c.Locals("params", params)

		_, err := config.AuthHandler.Authenticate(params)
		authChan <- err == nil
	}()

	select {
	case isAuthenticated := <-authChan:
		if !isAuthenticated {
			_ = c.WriteJSON(Message{
				Type:  TextMessage,
				Event: "error",
				Error: "unauthorized",
				Time:  time.Now(),
			})
			_ = c.Close()
		}
		return isAuthenticated
	case <-timeout:
		_ = c.WriteJSON(Message{
			Type:  TextMessage,
			Event: "error",
			Error: "authentication timeout",
			Time:  time.Now(),
		})
		_ = c.Close()
		return false
	}
}

// GetActiveConnections 获取当前活跃连接数
func GetActiveConnections() int {
	count := 0
	activeConnections.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// getWebSocketConfig 获取WebSocket配置
func getWebSocketConfig(config *Config) websocket.Config {
	return websocket.Config{
		HandshakeTimeout:  config.ReadTimeout,
		ReadBufferSize:    int(config.MaxMessageSize),
		WriteBufferSize:   int(config.MaxMessageSize),
		EnableCompression: config.EnableCompression,
	}
}

// UpgradeMiddleware 创建一个 WebSocket 升级中间件
func UpgradeMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			// 检查是否达到最大连接数
			if GetActiveConnections() >= maxConcurrentConnections {
				return fiber.NewError(fiber.StatusServiceUnavailable, "connection limit reached")
			}
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}
