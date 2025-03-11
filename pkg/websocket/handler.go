package websocket

import (
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// Handler 创建一个新的 WebSocket 处理器
func Handler(config *Config) fiber.Handler {
	// 创建一个新的 Hub
	hub := NewHub(config)
	go hub.Run()

	// 返回 WebSocket 升级处理器
	return websocket.New(func(c *websocket.Conn) {
		// 如果配置了认证处理器
		if config.AuthHandler != nil {
			params := make(map[string]interface{})
			// 从查询参数中获取认证信息
			c.Locals("params", params)
			if _, err := config.AuthHandler.Authenticate(params); err != nil {
				c.WriteJSON(Message{
					Type:  TextMessage,
					Event: "error",
					Error: "unauthorized",
					Time:  time.Now(),
				})
				c.Close()
				return
			}
		}

		// 创建新的客户端
		client := NewClient(c, hub)

		// 注册客户端
		hub.register <- client

		// 启动读写协程
		go client.WritePump()
		client.ReadPump()
	}, websocket.Config{
		HandshakeTimeout:  config.ReadTimeout,
		ReadBufferSize:    int(config.MaxMessageSize),
		WriteBufferSize:   int(config.MaxMessageSize),
		EnableCompression: config.EnableCompression,
	})
}

// UpgradeMiddleware 创建一个 WebSocket 升级中间件
func UpgradeMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}
