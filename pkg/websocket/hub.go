package websocket

import (
	"fmt"
	"sync"
	"time"
)

// NewHub 创建一个新的Hub实例
func NewHub(config *Config) *Hub {
	if config == nil {
		config = DefaultConfig()
	}

	return &Hub{
		clients:    make(map[*Client]struct{}),
		rooms:      make(map[string]map[*Client]struct{}),
		broadcast:  make(chan Message, config.MessageBuffer),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		config:     config,
	}
}

// Run 启动Hub的主循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient 注册新客户端
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()

	client.Send <- Message{
		Type:  TextMessage,
		Event: "system",
		Data:  "Welcome!",
		Time:  time.Now(),
	}
}

// unregisterClient 注销客户端
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for room := range client.rooms {
		_ = h.removeFromRoom(client, room)
	}

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.Send)
	}
}

// broadcastMessage 广播消息
func (h *Hub) broadcastMessage(message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if message.RoomID != "" {
		h.broadcastToRoom(message.RoomID, message)
		return
	}

	if message.To != "" {
		h.sendToSpecificClient(message)
		return
	}

	// 使用工作池处理广播
	h.broadcastToAll(message)
}

// broadcastToRoom 向特定房间广播消息
func (h *Hub) broadcastToRoom(roomID string, message Message) {
	if room, ok := h.rooms[roomID]; ok {
		clients := make([]*Client, 0, len(room))
		for client := range room {
			clients = append(clients, client)
		}

		// 并发发送消息给房间内的客户端
		var wg sync.WaitGroup
		for _, client := range clients {
			wg.Add(1)
			go func(c *Client) {
				defer wg.Done()
				h.sendToClient(c, message)
			}(client)
		}
		wg.Wait()
	}
}

// sendToSpecificClient 发送消息给特定客户端
func (h *Hub) sendToSpecificClient(message Message) {
	for client := range h.clients {
		if id, _ := client.Properties.Load("id"); id == message.To {
			h.sendToClient(client, message)
			break
		}
	}
}

// broadcastToAll 广播消息给所有客户端
func (h *Hub) broadcastToAll(message Message) {
	clients := make([]*Client, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client)
	}

	// 使用goroutine并发发送消息
	var wg sync.WaitGroup
	for _, client := range clients {
		wg.Add(1)
		go func(c *Client) {
			defer wg.Done()
			h.sendToClient(c, message)
		}(client)
	}
	wg.Wait()
}

// sendToClient 发送消息给特定客户端
func (h *Hub) sendToClient(client *Client, message Message) {
	select {
	case client.Send <- message:
	default:
		h.unregisterClient(client)
	}
}

// JoinRoom 将客户端加入房间
func (h *Hub) JoinRoom(client *Client, roomID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[roomID]; !ok {
		h.rooms[roomID] = make(map[*Client]struct{})
	}

	h.rooms[roomID][client] = struct{}{}
	client.rooms[roomID] = struct{}{}

	h.notifyRoomEvent(client, roomID, EventJoin)
	return nil
}

// LeaveRoom 将客户端从房间中移除
func (h *Hub) LeaveRoom(client *Client, roomID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := h.removeFromRoom(client, roomID); err != nil {
		return fmt.Errorf("failed to leave room: %v", err)
	}

	h.notifyRoomEvent(client, roomID, EventLeave)
	return nil
}

// notifyRoomEvent 通知房间事件
func (h *Hub) notifyRoomEvent(client *Client, roomID, event string) {
	message := Message{
		Type:   TextMessage,
		Event:  event,
		RoomID: roomID,
		From:   client.ID,
		Time:   time.Now(),
	}

	if room, ok := h.rooms[roomID]; ok {
		for c := range room {
			if c != client {
				h.sendToClient(c, message)
			}
		}
	}
}

// removeFromRoom 从房间中移除客户端（内部方法）
func (h *Hub) removeFromRoom(client *Client, roomID string) error {
	if room, ok := h.rooms[roomID]; ok {
		delete(room, client)
		delete(client.rooms, roomID)

		if len(room) == 0 {
			delete(h.rooms, roomID)
		}
		return nil
	}
	return fmt.Errorf("room %s not found", roomID)
}

// GetRoomClients 获取房间中的所有客户端
func (h *Hub) GetRoomClients(roomID string) ([]*Client, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if room, ok := h.rooms[roomID]; ok {
		clients := make([]*Client, 0, len(room))
		for client := range room {
			clients = append(clients, client)
		}
		return clients, nil
	}
	return nil, fmt.Errorf("room %s not found", roomID)
}

// GetClientRooms 获取客户端加入的所有房间
func (h *Hub) GetClientRooms(client *Client) []string {
	client.mu.RLock()
	defer client.mu.RUnlock()

	rooms := make([]string, 0, len(client.rooms))
	for room := range client.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// Broadcast 向所有客户端广播消息
func (h *Hub) Broadcast(message Message) {
	h.broadcast <- message
}

// BroadcastToRoom 向指定房间广播消息
func (h *Hub) BroadcastToRoom(roomID string, message Message) error {
	h.mu.RLock()
	if _, ok := h.rooms[roomID]; !ok {
		h.mu.RUnlock()
		return fmt.Errorf("room %s not found", roomID)
	}
	h.mu.RUnlock()

	message.RoomID = roomID
	h.broadcast <- message
	return nil
}

// SendToClient 发送消息给指定客户端
func (h *Hub) SendToClient(clientID string, message Message) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if id, _ := client.Properties.Load("id"); id == clientID {
			message.To = clientID
			h.sendToClient(client, message)
			return nil
		}
	}
	return fmt.Errorf("client %s not found", clientID)
}

// Stats 返回Hub的统计信息
func (h *Hub) Stats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := make(map[string]interface{})
	roomStats := make(map[string]int, len(h.rooms))

	// 并发统计房间信息
	var wg sync.WaitGroup
	var statsLock sync.Mutex

	for roomID, room := range h.rooms {
		wg.Add(1)
		go func(id string, r map[*Client]struct{}) {
			defer wg.Done()
			statsLock.Lock()
			roomStats[id] = len(r)
			statsLock.Unlock()
		}(roomID, room)
	}
	wg.Wait()

	stats["total_clients"] = len(h.clients)
	stats["total_rooms"] = len(h.rooms)
	stats["rooms_stats"] = roomStats

	return stats
}
