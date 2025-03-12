package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2/utils"
	"log"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
)

// NewClient 创建新的客户端实例
func NewClient(conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		ID:    utils.UUID(),
		Conn:  conn,
		Hub:   hub,
		Send:  make(chan Message, hub.config.MessageBuffer),
		rooms: make(map[string]struct{}),
		done:  make(chan struct{}),
	}
}

// ReadPump 处理从客户端读取消息
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		close(c.done)
		_ = c.Conn.Close()
	}()

	c.Conn.SetReadLimit(c.Hub.config.MaxMessageSize)
	if c.Hub.config.ReadTimeout > 0 {
		_ = c.Conn.SetReadDeadline(time.Now().Add(c.Hub.config.ReadTimeout))
	}

	// 创建消息处理工作池
	const numWorkers = 3
	jobs := make(chan []byte, numWorkers*2) // 增加缓冲区大小
	results := make(chan *Message, numWorkers)

	// 启动工作池
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case data, ok := <-jobs:
					if !ok {
						return
					}
					if message := c.processMessage(websocket.TextMessage, data); message != nil {
						select {
						case results <- message:
						case <-ctx.Done():
							return
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// 启动结果处理协程
	go func() {
		for message := range results {
			select {
			case c.Hub.broadcast <- *message:
			case <-ctx.Done():
				return
			}
		}
	}()

	// 读取消息并发送到工作池
	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break
		}

		select {
		case jobs <- data:
		case <-c.done:
			return
		default:
			// 如果工作池已满，使用临时goroutine处理
			go func(d []byte) {
				if message := c.processMessage(websocket.TextMessage, d); message != nil {
					select {
					case c.Hub.broadcast <- *message:
					case <-c.done:
					}
				}
			}(data)
		}
	}

	cancel()
	close(jobs)
	close(results)
	wg.Wait()
}

// processMessage 处理接收到的消息
func (c *Client) processMessage(messageType int, data []byte) *Message {
	message := &Message{
		Type: MessageType(messageType),
		Time: time.Now(),
		From: c.ID,
	}

	if err := json.Unmarshal(data, message); err != nil {
		message.Data = string(data)
	}

	if message.Event == EventPing {
		c.handlePing()
		return nil
	}

	if c.Hub.config.EventHandler != nil {
		if err := c.Hub.config.EventHandler.HandleEvent(c, *message); err != nil {
			if c.Hub.config.ErrorHandler != nil {
				c.Hub.config.ErrorHandler.HandleError(c, err)
			}
			return nil
		}
	}

	return message
}

// WritePump 处理发送消息到客户端
func (c *Client) WritePump() {
	ticker := time.NewTicker(c.Hub.config.PingInterval)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.writeMessage(message); err != nil {
				return
			}

		case <-ticker.C:
			if c.Hub.config.EnablePing {
				if err := c.ping(); err != nil {
					return
				}
			}

		case <-c.done:
			return
		}
	}
}

// writeMessage 写入消息到连接
func (c *Client) writeMessage(message Message) error {
	if c.Hub.config.WriteTimeout > 0 {
		_ = c.Conn.SetWriteDeadline(time.Now().Add(c.Hub.config.WriteTimeout))
	}

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
		Event: EventPong,
		Time:  time.Now(),
	}
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	message := Message{
		Type:  CloseMessage,
		Event: EventClose,
		Time:  time.Now(),
		From:  c.ID,
	}

	if err := c.writeMessage(message); err != nil {
		return fmt.Errorf("error sending close message: %v", err)
	}

	close(c.done)
	return c.Conn.Close()
}

// SendMessage 发送消息（带超时和重试机制）
func (c *Client) SendMessage(message Message) error {
	const maxRetries = 3
	var err error

	for i := 0; i < maxRetries; i++ {
		timer := time.NewTimer(2 * time.Second) // 减少单次超时时间，但允许重试
		select {
		case c.Send <- message:
			timer.Stop()
			return nil
		case <-timer.C:
			err = fmt.Errorf("send timeout (attempt %d/%d)", i+1, maxRetries)
		case <-c.done:
			timer.Stop()
			return fmt.Errorf("client closed")
		default:
			timer.Stop()
			// 如果通道已满，等待一小段时间后重试
			time.Sleep(100 * time.Millisecond)
			continue
		}
		timer.Stop()
	}
	return fmt.Errorf("send failed after %d attempts: %v", maxRetries, err)
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
