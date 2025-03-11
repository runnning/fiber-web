package websocket

import (
	"fmt"
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

	// 发送欢迎消息
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

	// 从所有房间中移除
	for room := range client.rooms {
		h.removeFromRoom(client, room)
	}

	// 从客户端列表中移除
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.Send)
	}
}

// broadcastMessage 广播消息
func (h *Hub) broadcastMessage(message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// 如果指定了房间，只发送给房间内的客户端
	if message.RoomID != "" {
		if room, ok := h.rooms[message.RoomID]; ok {
			for client := range room {
				h.sendToClient(client, message)
			}
		}
		return
	}

	// 如果指定了接收者，只发送给特定客户端
	if message.To != "" {
		for client := range h.clients {
			if id, _ := client.Properties.Load("id"); id == message.To {
				h.sendToClient(client, message)
				break
			}
		}
		return
	}

	// 广播给所有客户端
	for client := range h.clients {
		h.sendToClient(client, message)
	}
}

// sendToClient 发送消息给特定客户端
func (h *Hub) sendToClient(client *Client, message Message) {
	select {
	case client.Send <- message:
	default:
		// 如果客户端的发送缓冲区已满，关闭连接
		h.unregisterClient(client)
	}
}

// JoinRoom 将客户端加入房间
func (h *Hub) JoinRoom(client *Client, roomID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 创建房间（如果不存在）
	if _, ok := h.rooms[roomID]; !ok {
		h.rooms[roomID] = make(map[*Client]struct{})
	}

	// 将客户端加入房间
	h.rooms[roomID][client] = struct{}{}

	// 更新客户端的房间列表
	if client.rooms == nil {
		client.rooms = make(map[string]struct{})
	}
	client.rooms[roomID] = struct{}{}

	// 通知房间其他成员
	message := Message{
		Type:   TextMessage,
		Event:  "join_room",
		RoomID: roomID,
		From:   client.ID,
		Time:   time.Now(),
	}

	for c := range h.rooms[roomID] {
		if c != client {
			h.sendToClient(c, message)
		}
	}

	return nil
}

// LeaveRoom 将客户端从房间中移除
func (h *Hub) LeaveRoom(client *Client, roomID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := h.removeFromRoom(client, roomID); err != nil {
		return fmt.Errorf("failed to leave room: %v", err)
	}

	// 通知房间其他成员
	message := Message{
		Type:   TextMessage,
		Event:  "leave_room",
		RoomID: roomID,
		From:   client.ID,
		Time:   time.Now(),
	}

	if room, ok := h.rooms[roomID]; ok {
		for c := range room {
			h.sendToClient(c, message)
		}
	}

	return nil
}

// removeFromRoom 从房间中移除客户端（内部方法）
func (h *Hub) removeFromRoom(client *Client, roomID string) error {
	if room, ok := h.rooms[roomID]; ok {
		delete(room, client)
		delete(client.rooms, roomID)

		// 如果房间为空，删除房间
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
	defer h.mu.RUnlock()

	if _, ok := h.rooms[roomID]; !ok {
		return fmt.Errorf("room %s not found", roomID)
	}

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

	return map[string]interface{}{
		"total_clients": len(h.clients),
		"total_rooms":   len(h.rooms),
		"rooms_stats":   h.getRoomsStats(),
	}
}

// getRoomsStats 获取所有房间的统计信息
func (h *Hub) getRoomsStats() map[string]interface{} {
	stats := make(map[string]interface{})
	for roomID, room := range h.rooms {
		stats[roomID] = len(room)
	}
	return stats
}
