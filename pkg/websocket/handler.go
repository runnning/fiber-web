package websocket

import (
	"fmt"
	"log"
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
	if config == nil {
		config = DefaultConfig()
	}

	hub := NewHub(config)
	go hub.Run()

	// 添加连接统计
	go monitorConnections(hub)

	return websocket.New(handleWebSocket(hub, config), getWebSocketConfig(config))
}

// monitorConnections 监控连接数量
func monitorConnections(hub *Hub) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		stats := hub.Stats()
		log.Printf("WebSocket Stats - Total Clients: %d, Total Rooms: %d",
			stats["total_clients"], stats["total_rooms"])
	}
}

// handleWebSocket 处理WebSocket连接
func handleWebSocket(hub *Hub, config *Config) func(*websocket.Conn) {
	return func(c *websocket.Conn) {
		// 使用带超时的信号量获取
		timeout := time.NewTimer(connectionTimeout)
		defer timeout.Stop()

		select {
		case connectionSemaphore <- struct{}{}:
			defer func() { <-connectionSemaphore }()
		case <-timeout.C:
			_ = c.WriteJSON(Message{
				Type:  TextMessage,
				Event: "error",
				Error: "connection limit reached",
				Time:  time.Now(),
			})
			_ = c.Close()
			if config.ErrorHandler != nil {
				config.ErrorHandler.HandleError(nil, fmt.Errorf("connection limit reached"))
			}
			return
		}

		// 验证客户端
		if !authenticateClient(c, config) {
			return
		}

		client := NewClient(c, hub)
		activeConnections.Store(client.ID, client)

		// 确保在函数返回时清理资源
		defer func() {
			activeConnections.Delete(client.ID)
			if r := recover(); r != nil {
				log.Printf("WebSocket handler panic recovered: %v", r)
				if config.ErrorHandler != nil {
					config.ErrorHandler.HandleError(client, fmt.Errorf("handler panic: %v", r))
				}
			}
		}()

		// 注册客户端
		hub.register <- client

		// 启动读写协程
		done := make(chan struct{})
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("WritePump panic recovered: %v", r)
				}
				close(done)
			}()
			client.WritePump()
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
