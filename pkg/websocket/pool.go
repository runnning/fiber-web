package websocket

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrClientNotFound     = errors.New("client not found")
	ErrGroupNotFound      = errors.New("group not found")
	ErrGroupExists        = errors.New("group already exists")
	ErrClientInGroup      = errors.New("client already in group")
	ErrSendBufferFull     = errors.New("send buffer full")
	ErrClientDisconnected = errors.New("client disconnected")
)

// run 启动连接池的管理循环
func (p *Pool) run() {
	for {
		select {
		case client := <-p.register:
			p.mu.Lock()
			p.clients[client.ID] = client
			p.mu.Unlock()
		case client := <-p.unregister:
			p.mu.Lock()
			if _, ok := p.clients[client.ID]; ok {
				delete(p.clients, client.ID)
				close(client.send)
				client.Close()
			}
			p.mu.Unlock()

			// 从所有群组中移除
			p.groupsMu.Lock()
			for groupID, group := range p.groups {
				if _, ok := group[client.ID]; ok {
					delete(group, client.ID)
					// 如果群组为空，删除群组
					if len(group) == 0 {
						delete(p.groups, groupID)
					}
				}
			}
			p.groupsMu.Unlock()
		case message := <-p.broadcast:
			p.broadcastMessage(message)
		}
	}
}

// broadcastMessage 广播消息的具体实现
func (p *Pool) broadcastMessage(message Message) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 如果指定了目标群组
	if message.To != "" {
		p.groupsMu.RLock()
		group, ok := p.groups[message.To]
		if !ok {
			p.groupsMu.RUnlock()
			return
		}

		failedClients := make([]string, 0)
		for clientID, client := range group {
			select {
			case client.send <- message:
			default:
				failedClients = append(failedClients, clientID)
			}
		}
		p.groupsMu.RUnlock()

		// 处理发送失败的客户端
		if len(failedClients) > 0 {
			for _, clientID := range failedClients {
				if client, ok := p.clients[clientID]; ok {
					p.unregister <- client
				}
			}
		}
		return
	}

	// 广播给所有客户端
	failedClients := make([]string, 0)
	for clientID, client := range p.clients {
		select {
		case client.send <- message:
		default:
			failedClients = append(failedClients, clientID)
		}
	}

	// 处理发送失败的客户端
	if len(failedClients) > 0 {
		for _, clientID := range failedClients {
			if client, ok := p.clients[clientID]; ok {
				p.unregister <- client
			}
		}
	}
}

// CreateGroup 创建一个新的群组
func (p *Pool) CreateGroup(groupID string) error {
	p.groupsMu.Lock()
	defer p.groupsMu.Unlock()

	if _, exists := p.groups[groupID]; exists {
		return fmt.Errorf("%w: %s", ErrGroupExists, groupID)
	}

	p.groups[groupID] = make(map[string]*Client)
	return nil
}

// JoinGroup 将客户端加入群组
func (p *Pool) JoinGroup(groupID string, clientID string) error {
	p.groupsMu.Lock()
	defer p.groupsMu.Unlock()

	group, ok := p.groups[groupID]
	if !ok {
		return fmt.Errorf("%w: %s", ErrGroupNotFound, groupID)
	}

	if _, exists := group[clientID]; exists {
		return fmt.Errorf("%w: client %s in group %s", ErrClientInGroup, clientID, groupID)
	}

	p.mu.RLock()
	client, ok := p.clients[clientID]
	p.mu.RUnlock()

	if !ok {
		return fmt.Errorf("%w: %s", ErrClientNotFound, clientID)
	}

	group[clientID] = client
	return nil
}

// LeaveGroup 将客户端从群组中移除
func (p *Pool) LeaveGroup(groupID string, clientID string) error {
	p.groupsMu.Lock()
	defer p.groupsMu.Unlock()

	group, ok := p.groups[groupID]
	if !ok {
		return fmt.Errorf("%w: %s", ErrGroupNotFound, groupID)
	}

	if _, exists := group[clientID]; !exists {
		return fmt.Errorf("%w: %s", ErrClientNotFound, clientID)
	}

	delete(group, clientID)

	// 如果群组为空，删除群组
	if len(group) == 0 {
		delete(p.groups, groupID)
	}

	return nil
}

// DeleteGroup 删除一个群组
func (p *Pool) DeleteGroup(groupID string) error {
	p.groupsMu.Lock()
	defer p.groupsMu.Unlock()

	if _, exists := p.groups[groupID]; !exists {
		return fmt.Errorf("%w: %s", ErrGroupNotFound, groupID)
	}

	delete(p.groups, groupID)
	return nil
}

// GetGroupMembers 获取群组成员
func (p *Pool) GetGroupMembers(groupID string) ([]*Client, error) {
	p.groupsMu.RLock()
	defer p.groupsMu.RUnlock()

	group, ok := p.groups[groupID]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrGroupNotFound, groupID)
	}

	members := make([]*Client, 0, len(group))
	for _, client := range group {
		members = append(members, client)
	}
	return members, nil
}

// Broadcast 向所有连接的客户端广播消息
func (p *Pool) Broadcast(messageType MessageType, content []byte) {
	message := Message{
		Type:    int(messageType),
		Content: content,
	}
	p.broadcast <- message
}

// BroadcastToGroup 向指定群组广播消息
func (p *Pool) BroadcastToGroup(groupID string, messageType MessageType, content []byte) error {
	p.groupsMu.RLock()
	_, ok := p.groups[groupID]
	p.groupsMu.RUnlock()

	if !ok {
		return fmt.Errorf("%w: %s", ErrGroupNotFound, groupID)
	}

	message := Message{
		Type:    int(messageType),
		Content: content,
		To:      groupID,
	}
	p.broadcast <- message
	return nil
}

// GetClient 根据ID获取客户端
func (p *Pool) GetClient(id string) (*Client, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	client, ok := p.clients[id]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrClientNotFound, id)
	}
	return client, nil
}

// Count 返回当前连接的客户端数量
func (p *Pool) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.clients)
}

// GroupCount 返回群组数量
func (p *Pool) GroupCount() int {
	p.groupsMu.RLock()
	defer p.groupsMu.RUnlock()
	return len(p.groups)
}

// Send 发送消息到客户端
func (c *Client) Send(messageType MessageType, content []byte) error {
	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()

	if closed {
		return ErrClientDisconnected
	}

	message := Message{
		Type:    int(messageType),
		Content: content,
		From:    c.ID,
	}

	select {
	case c.send <- message:
		return nil
	default:
		return ErrSendBufferFull
	}
}

// Close 关闭客户端连接
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		c.closed = true
		_ = c.Conn.Close()
	}
}

// UpdatePing 更新最后一次心跳时间
func (c *Client) UpdatePing() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastPing = time.Now()
}

// IsAlive 检查客户端是否存活
func (c *Client) IsAlive(timeout time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return time.Since(c.lastPing) < timeout
}

// SetProperty 设置客户端属性
func (c *Client) SetProperty(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Properties[key] = value
}

// GetProperty 获取客户端属性
func (c *Client) GetProperty(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.Properties[key]
	return value, ok
}

// GetClientGroups 获取客户端所在的所有群组
func (p *Pool) GetClientGroups(clientID string) ([]string, error) {
	p.groupsMu.RLock()
	defer p.groupsMu.RUnlock()

	groups := make([]string, 0)
	for groupID, group := range p.groups {
		if _, exists := group[clientID]; exists {
			groups = append(groups, groupID)
		}
	}
	return groups, nil
}
