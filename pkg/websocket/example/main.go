package main

import (
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

func (h *ChatHandler) OnConnect(client *websocket.Client) {
	log.Printf("Client connected: %s", client.ID)
	// 设置用户名
	client.SetProperty("username", "user_"+client.ID[:8])
}

func (h *ChatHandler) OnMessage(client *websocket.Client, message websocket.Message) {
	// 处理心跳响应消息
	if message.Type == int(websocket.PongMessage) {
		client.UpdatePing()
		return
	}

	// 解析消息
	var chatMsg ChatMessage
	if err := json.Unmarshal(message.Content, &chatMsg); err != nil {
		log.Printf("Error parsing message: %v", err)
		// 发送错误消息给客户端
		errMsg := ChatMessage{
			Type:    "error",
			Content: "Invalid message format",
			From:    "System",
		}
		if errBytes, err := json.Marshal(errMsg); err == nil {
			if err := client.Send(websocket.TextMessage, errBytes); err != nil {
				log.Printf("Error sending error message: %v", err)
			}
		}
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
	case "join_room":
		// 加入房间
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
		// 发送加入成功消息
		response := ChatMessage{
			Type:    "system",
			Content: username.(string) + " joined the room " + chatMsg.Room,
			Room:    chatMsg.Room,
			From:    "System",
		}
		responseBytes, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling join message: %v", err)
			return
		}
		if err := client.Pool.BroadcastToGroup(chatMsg.Room, websocket.TextMessage, responseBytes); err != nil {
			log.Printf("Error broadcasting join message: %v", err)
			return
		}

	case "leave_room":
		// 离开房间
		if err := client.Pool.LeaveGroup(chatMsg.Room, client.ID); err != nil {
			log.Printf("Error leaving room: %v", err)
			return
		}
		// 发送离开消息
		response := ChatMessage{
			Type:    "system",
			Content: username.(string) + " left the room " + chatMsg.Room,
			Room:    chatMsg.Room,
			From:    "System",
		}
		responseBytes, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling leave message: %v", err)
			return
		}
		if err := client.Pool.BroadcastToGroup(chatMsg.Room, websocket.TextMessage, responseBytes); err != nil {
			log.Printf("Error broadcasting leave message: %v", err)
			return
		}

	case "chat":
		// 发送聊天消息
		responseBytes, err := json.Marshal(chatMsg)
		if err != nil {
			log.Printf("Error marshaling chat message: %v", err)
			return
		}
		if chatMsg.Room != "" {
			if err := client.Pool.BroadcastToGroup(chatMsg.Room, websocket.TextMessage, responseBytes); err != nil {
				log.Printf("Error broadcasting to group: %v", err)
				return
			}
		} else {
			client.Pool.Broadcast(websocket.TextMessage, responseBytes)
		}

	case "set_username":
		// 设置用户名
		oldUsername := username.(string)
		client.SetProperty("username", chatMsg.Content)
		response := ChatMessage{
			Type:    "system",
			Content: oldUsername + " changed name to " + chatMsg.Content,
			From:    "System",
		}
		responseBytes, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling username change message: %v", err)
			return
		}
		client.Pool.Broadcast(websocket.TextMessage, responseBytes)

	default:
		log.Printf("Unknown message type: %s", chatMsg.Type)
		errMsg := ChatMessage{
			Type:    "error",
			Content: "Unknown message type",
			From:    "System",
		}
		if errBytes, err := json.Marshal(errMsg); err == nil {
			if err := client.Send(websocket.TextMessage, errBytes); err != nil {
				log.Printf("Error sending error message: %v", err)
			}
		}
	}
}

func (h *ChatHandler) OnClose(client *websocket.Client) {
	// 获取用户名
	username, ok := client.GetProperty("username")
	if !ok {
		log.Printf("Client disconnected: %s", client.ID)
		return
	}

	// 从所有房间中移除
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

// 定义中间件
func loggingMiddleware(client *websocket.Client, message websocket.Message) (websocket.Message, error) {
	log.Printf("Message from %s: %s", client.ID, string(message.Content))
	return message, nil
}

func main() {
	app := fiber.New()

	// 创建 WebSocket 配置
	config := websocket.Config{
		Handler:         &ChatHandler{},
		PingTimeout:     45 * time.Second, // 减少超时时间
		WriteTimeout:    10 * time.Second,
		ReadTimeout:     10 * time.Second,
		BufferSize:      1024,
		MessageBuffer:   256,
		EnableHeartbeat: true,
		HeartbeatPeriod: 15 * time.Second, // 减少心跳间隔
		Middlewares:     []websocket.MiddlewareFunc{loggingMiddleware},
	}

	// 设置静态文件服务
	app.Static("/", "./example")

	// 设置 WebSocket 路由
	app.Get("/ws", websocket.New(config))

	// 启动服务器
	log.Fatal(app.Listen(":3000"))
}
