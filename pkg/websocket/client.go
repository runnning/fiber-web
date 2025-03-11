package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
)

// NewClient 创建新的客户端实例
func NewClient(conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		ID:    fmt.Sprintf("%d", time.Now().UnixNano()),
		Conn:  conn,
		Hub:   hub,
		Send:  make(chan Message, hub.config.MessageBuffer),
		rooms: make(map[string]struct{}),
	}
}

// ReadPump 处理从客户端读取消息
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(c.Hub.config.MaxMessageSize)

	// 设置读取超时
	if c.Hub.config.ReadTimeout > 0 {
		c.Conn.SetReadDeadline(time.Now().Add(c.Hub.config.ReadTimeout))
	}

	for {
		messageType, data, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break
		}

		// 处理消息
		message := Message{
			Type: MessageType(messageType),
			Time: time.Now(),
			From: c.ID,
		}

		// 尝试解析JSON消息
		if err := json.Unmarshal(data, &message); err != nil {
			// 如果不是JSON，将原始数据作为内容
			message.Data = string(data)
		}

		// 处理心跳消息
		if message.Event == "ping" {
			c.handlePing()
			continue
		}

		// 如果配置了事件处理器，交给处理器处理
		if c.Hub.config.EventHandler != nil {
			if err := c.Hub.config.EventHandler.HandleEvent(c, message); err != nil {
				if c.Hub.config.ErrorHandler != nil {
					c.Hub.config.ErrorHandler.HandleError(c, err)
				}
				continue
			}
		}

		// 广播消息
		c.Hub.broadcast <- message
	}
}

// WritePump 处理发送消息到客户端
func (c *Client) WritePump() {
	ticker := time.NewTicker(c.Hub.config.PingInterval)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			// 设置写入超时
			if c.Hub.config.WriteTimeout > 0 {
				c.Conn.SetWriteDeadline(time.Now().Add(c.Hub.config.WriteTimeout))
			}

			if !ok {
				// 通道已关闭
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 发送消息
			if err := c.writeJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			if c.Hub.config.EnablePing {
				if err := c.ping(); err != nil {
					return
				}
			}
		}
	}
}

// writeJSON 发送JSON消息
func (c *Client) writeJSON(message Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.WriteJSON(message)
}

// ping 发送ping消息
func (c *Client) ping() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Conn.WriteMessage(websocket.PingMessage, nil)
}

// handlePing 处理ping消息
func (c *Client) handlePing() {
	c.Send <- Message{
		Type:  PongMessage,
		Event: "pong",
		Time:  time.Now(),
	}
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 发送关闭消息
	message := Message{
		Type:  CloseMessage,
		Event: "close",
		Time:  time.Now(),
		From:  c.ID,
	}

	if err := c.writeJSON(message); err != nil {
		return fmt.Errorf("error sending close message: %v", err)
	}

	return c.Conn.Close()
}

// SendMessage 发送消息
func (c *Client) SendMessage(message Message) error {
	select {
	case c.Send <- message:
		return nil
	default:
		return fmt.Errorf("send buffer full")
	}
}

// JoinRoom 加入房间
func (c *Client) JoinRoom(roomID string) error {
	return c.Hub.JoinRoom(c, roomID)
}

// LeaveRoom 离开房间
func (c *Client) LeaveRoom(roomID string) error {
	return c.Hub.LeaveRoom(c, roomID)
}

// GetRooms 获取已加入的房间列表
func (c *Client) GetRooms() []string {
	return c.Hub.GetClientRooms(c)
}

// SetProperty 设置客户端属性
func (c *Client) SetProperty(key string, value interface{}) {
	c.Properties.Store(key, value)
}

// GetProperty 获取客户端属性
func (c *Client) GetProperty(key string) (interface{}, bool) {
	return c.Properties.Load(key)
}

// DeleteProperty 删除客户端属性
func (c *Client) DeleteProperty(key string) {
	c.Properties.Delete(key)
}
