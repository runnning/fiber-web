package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"fiber_web/pkg/websocket"

	"github.com/gofiber/fiber/v2"
)

// ChatMessage 定义聊天消息结构
type ChatMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Room    string `json:"room,omitempty"`
	From    string `json:"from"`
}

type ChatHandler struct{}

// 发送错误消息
func sendErrorMessage(client *websocket.Client, content string) {
	errMsg := ChatMessage{
		Type:    "error",
		Content: content,
		From:    "System",
	}
	if errBytes, err := json.Marshal(errMsg); err == nil {
		client.Send(websocket.TextMessage, errBytes)
	}
}

// 处理聊天消息
func handleChatMessage(client *websocket.Client, chatMsg ChatMessage) {
	responseBytes, err := json.Marshal(chatMsg)
	if err != nil {
		log.Printf("Error marshaling chat message: %v", err)
		return
	}
	if chatMsg.Room != "" {
		client.Pool.BroadcastToGroup(chatMsg.Room, websocket.TextMessage, responseBytes)
	} else {
		client.Pool.Broadcast(websocket.TextMessage, responseBytes)
	}
}

// 处理加入房间
func handleJoinRoom(client *websocket.Client, chatMsg ChatMessage, username string) {
	if err := client.Pool.JoinGroup(chatMsg.Room, client.ID); err != nil {
		if errors.Is(err, websocket.ErrGroupNotFound) {
			if err := client.Pool.CreateGroup(chatMsg.Room); err != nil {
				log.Printf("Error CreateGroup: %v", err)
				return
			}
			if err := client.Pool.JoinGroup(chatMsg.Room, client.ID); err != nil {
				log.Printf("Error joining room after creation: %v", err)
				return
			}
		} else {
			log.Printf("Error joining room: %v", err)
			return
		}
	}

	response := ChatMessage{
		Type:    "system",
		Content: username + " joined the room " + chatMsg.Room,
		Room:    chatMsg.Room,
		From:    "System",
	}
	if responseBytes, err := json.Marshal(response); err == nil {
		client.Pool.BroadcastToGroup(chatMsg.Room, websocket.TextMessage, responseBytes)
	}
}

// 处理离开房间
func handleLeaveRoom(client *websocket.Client, chatMsg ChatMessage, username string) {
	if err := client.Pool.LeaveGroup(chatMsg.Room, client.ID); err != nil {
		log.Printf("Error leaving room: %v", err)
		return
	}

	response := ChatMessage{
		Type:    "system",
		Content: username + " left the room " + chatMsg.Room,
		Room:    chatMsg.Room,
		From:    "System",
	}
	if responseBytes, err := json.Marshal(response); err == nil {
		client.Pool.BroadcastToGroup(chatMsg.Room, websocket.TextMessage, responseBytes)
	}
}

// 处理设置用户名
func handleSetUsername(client *websocket.Client, chatMsg ChatMessage, oldUsername string) {
	client.SetProperty("username", chatMsg.Content)
	response := ChatMessage{
		Type:    "system",
		Content: oldUsername + " changed name to " + chatMsg.Content,
		From:    "System",
	}
	if responseBytes, err := json.Marshal(response); err == nil {
		client.Pool.Broadcast(websocket.TextMessage, responseBytes)
	}
}

func (h *ChatHandler) OnConnect(client *websocket.Client) {
	log.Printf("Client connected: %s", client.ID)
	client.SetProperty("username", "user_"+client.ID[:8])
}

func (h *ChatHandler) OnMessage(client *websocket.Client, message websocket.Message) {
	// 使用对象池来减少内存分配
	var chatMsg ChatMessage
	if message.Type == int(websocket.PongMessage) {
		client.UpdatePing()
		return
	}

	// 快速处理心跳响应
	if len(message.Content) == 0 {
		return
	}

	// 解析消息
	if err := json.Unmarshal(message.Content, &chatMsg); err != nil {
		log.Printf("Error parsing message: %v", err)
		sendErrorMessage(client, "Invalid message format")
		return
	}

	// 处理心跳响应
	if chatMsg.Type == "pong" {
		client.UpdatePing()
		return
	}

	username, ok := client.GetProperty("username")
	if !ok {
		log.Printf("Username not found for client: %s", client.ID)
		return
	}
	chatMsg.From = username.(string)

	// 根据消息类型处理
	switch chatMsg.Type {
	case "chat":
		// 优先处理聊天消息
		handleChatMessage(client, chatMsg)
	case "join_room":
		handleJoinRoom(client, chatMsg, username.(string))
	case "leave_room":
		handleLeaveRoom(client, chatMsg, username.(string))
	case "set_username":
		handleSetUsername(client, chatMsg, username.(string))
	default:
		sendErrorMessage(client, "Unknown message type")
	}
}

func (h *ChatHandler) OnClose(client *websocket.Client) {
	username, ok := client.GetProperty("username")
	if !ok {
		log.Printf("Client disconnected: %s", client.ID)
		return
	}

	groups, err := client.Pool.GetClientGroups(client.ID)
	if err != nil {
		log.Printf("Error getting client groups: %v", err)
	} else {
		for _, group := range groups {
			if err := client.Pool.LeaveGroup(group, client.ID); err != nil {
				log.Printf("Error leaving group %s: %v", group, err)
			}
		}
	}

	log.Printf("Client disconnected: %s (%s)", client.ID, username)
}

func (h *ChatHandler) OnError(client *websocket.Client, err error) {
	username, ok := client.GetProperty("username")
	if !ok {
		log.Printf("Error from client %s: %v", client.ID, err)
		return
	}
	log.Printf("Error from client %s (%s): %v", client.ID, username, err)
}

func loggingMiddleware(ctx context.Context, client *websocket.Client, message websocket.Message) (websocket.Message, error) {
	log.Printf("Message from %s: %s", client.ID, string(message.Content))
	return message, nil
}

func main() {
	app := fiber.New(fiber.Config{
		ReadTimeout:     60 * time.Second,
		WriteTimeout:    60 * time.Second,
		IdleTimeout:     75 * time.Second,
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		//DisableStartupMessage: true,
	})

	// 创建 WebSocket 配置
	config := websocket.Config{
		Handler:         &ChatHandler{},
		PingTimeout:     60 * time.Second,
		WriteTimeout:    10 * time.Second,
		ReadTimeout:     10 * time.Second,
		BufferSize:      4096,
		MessageBuffer:   1024,
		EnableHeartbeat: true,
		HeartbeatPeriod: 25 * time.Second,
		EnableReconnect: true,
		MaxRetries:      3,
		RetryInterval:   5 * time.Second,
		Compression:     true,
		MaxMessageSize:  32 << 20,
		EnableMetrics:   true,
		EnableRateLimit: true,
		RateLimit:       200, // 增加速率限制
		Middlewares:     []websocket.MiddlewareFunc{loggingMiddleware},
	}

	// 设置 WebSocket 路由
	app.Get("/ws", websocket.New(config))

	// 启动服务器
	log.Printf("Server starting on :3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
