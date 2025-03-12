package main

import (
	"log"
	"time"

	"fiber_web/pkg/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// ChatEventHandler 处理聊天事件
type ChatEventHandler struct{}

func (h *ChatEventHandler) HandleEvent(client *websocket.Client, message websocket.Message) error {
	switch message.Event {
	case "join_room":
		if roomID, ok := message.Data.(string); ok {
			return client.JoinRoom(roomID)
		}
	case "leave_room":
		if roomID, ok := message.Data.(string); ok {
			return client.LeaveRoom(roomID)
		}
	case "chat":
		// 直接广播消息
		client.Hub.Broadcast(message)
	}
	return nil
}

// ChatErrorHandler 处理错误
type ChatErrorHandler struct{}

func (h *ChatErrorHandler) HandleError(client *websocket.Client, err error) {
	log.Printf("Error from client %s: %v", client.ID, err)
	_ = client.SendMessage(websocket.Message{
		Type:  websocket.TextMessage,
		Event: "error",
		Error: err.Error(),
		Time:  time.Now(),
	})
}

func main() {
	// 创建 Fiber 应用
	app := fiber.New(fiber.Config{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	// 添加中间件
	app.Use(logger.New())
	app.Use(cors.New())

	// 设置静态文件服务
	app.Static("/", "./pkg/websocket/example")

	// 添加根路由处理
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./pkg/websocket/example/index.html")
	})

	// WebSocket 配置
	wsConfig := &websocket.Config{
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      10 * time.Second,
		PingInterval:      30 * time.Second,
		MaxMessageSize:    512 * 1024, // 512KB
		MessageBuffer:     256,
		EnableCompression: true,
		EnablePing:        true,
		EventHandler:      &ChatEventHandler{},
		ErrorHandler:      &ChatErrorHandler{},
	}

	// 设置 WebSocket 路由
	app.Use("/ws", websocket.UpgradeMiddleware())
	app.Get("/ws", websocket.Handler(wsConfig))

	// 启动服务器
	log.Printf("Server starting on :3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
