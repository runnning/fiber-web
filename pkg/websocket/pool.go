package websocket

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrClientNotFound     = errors.New("client not found")
	ErrGroupNotFound      = errors.New("group not found")
	ErrGroupExists        = errors.New("group already exists")
	ErrClientInGroup      = errors.New("client already in group")
	ErrSendBufferFull     = errors.New("send buffer full")
	ErrClientDisconnected = errors.New("client disconnected")
	ErrInvalidGroupID     = errors.New("invalid group id")
	ErrInvalidClientID    = errors.New("invalid client id")
)

// Group 表示一个WebSocket群组
type Group struct {
	ID      string
	clients sync.Map
	pool    *Pool
}

// run 启动连接池的管理循环
func (p *Pool) run() {
	for {
		select {
		case client := <-p.register:
			p.clients.Store(client.ID, client)
		case client := <-p.unregister:
			if _, ok := p.clients.LoadAndDelete(client.ID); ok {
				close(client.send)
				client.Close()

				// 从所有群组中移除
				p.groups.Range(func(key, value interface{}) bool {
					if group, ok := value.(*Group); ok {
						group.clients.Delete(client.ID)
					}
					return true
				})
			}
		case message := <-p.broadcast:
			p.broadcastMessage(message)
		}
	}
}

// broadcastMessage 广播消息的具体实现
func (p *Pool) broadcastMessage(message Message) {
	start := time.Now()

	// 如果指定了目标群组
	if message.To != "" {
		if value, ok := p.groups.Load(message.To); ok {
			group := value.(*Group)
			group.clients.Range(func(key, value interface{}) bool {
				if client, ok := value.(*Client); ok {
					select {
					case client.send <- message:
					default:
						p.unregister <- client
					}
				}
				return true
			})
		}
		return
	}

	// 广播给所有客户端
	p.clients.Range(func(key, value interface{}) bool {
		if client, ok := value.(*Client); ok {
			select {
			case client.send <- message:
			default:
				p.unregister <- client
			}
		}
		return true
	})

	// 更新广播延迟指标
	if p.metrics != nil {
		latency := time.Since(start).Microseconds()
		currentAvg := p.metrics.AverageLatency.Load()
		if currentAvg == 0 {
			p.metrics.AverageLatency.Store(latency)
		} else {
			newAvg := (currentAvg + latency) / 2
			p.metrics.AverageLatency.Store(newAvg)
		}
	}
}

// CreateGroup 创建一个新的群组
func (p *Pool) CreateGroup(groupID string) error {
	if groupID == "" {
		return ErrInvalidGroupID
	}

	group := &Group{
		ID:   groupID,
		pool: p,
	}

	if _, loaded := p.groups.LoadOrStore(groupID, group); loaded {
		return fmt.Errorf("%w: %s", ErrGroupExists, groupID)
	}

	return nil
}

// JoinGroup 将客户端加入群组
func (p *Pool) JoinGroup(groupID string, clientID string) error {
	if groupID == "" {
		return ErrInvalidGroupID
	}
	if clientID == "" {
		return ErrInvalidClientID
	}

	value, ok := p.groups.Load(groupID)
	if !ok {
		return fmt.Errorf("%w: %s", ErrGroupNotFound, groupID)
	}
	group := value.(*Group)

	clientValue, ok := p.clients.Load(clientID)
	if !ok {
		return fmt.Errorf("%w: %s", ErrClientNotFound, clientID)
	}
	client := clientValue.(*Client)

	if _, loaded := group.clients.LoadOrStore(clientID, client); loaded {
		return fmt.Errorf("%w: client %s in group %s", ErrClientInGroup, clientID, groupID)
	}

	return nil
}

// LeaveGroup 将客户端从群组中移除
func (p *Pool) LeaveGroup(groupID string, clientID string) error {
	if groupID == "" {
		return ErrInvalidGroupID
	}
	if clientID == "" {
		return ErrInvalidClientID
	}

	value, ok := p.groups.Load(groupID)
	if !ok {
		return fmt.Errorf("%w: %s", ErrGroupNotFound, groupID)
	}
	group := value.(*Group)

	if _, ok := group.clients.LoadAndDelete(clientID); !ok {
		return fmt.Errorf("%w: %s", ErrClientNotFound, clientID)
	}

	// 检查群组是否为空
	empty := true
	group.clients.Range(func(key, value interface{}) bool {
		empty = false
		return false
	})

	if empty {
		p.groups.Delete(groupID)
	}

	return nil
}

// DeleteGroup 删除一个群组
func (p *Pool) DeleteGroup(groupID string) error {
	if groupID == "" {
		return ErrInvalidGroupID
	}

	if _, ok := p.groups.LoadAndDelete(groupID); !ok {
		return fmt.Errorf("%w: %s", ErrGroupNotFound, groupID)
	}

	return nil
}

// GetGroupMembers 获取群组成员
func (p *Pool) GetGroupMembers(groupID string) ([]*Client, error) {
	if groupID == "" {
		return nil, ErrInvalidGroupID
	}

	value, ok := p.groups.Load(groupID)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrGroupNotFound, groupID)
	}
	group := value.(*Group)

	var members []*Client
	group.clients.Range(func(key, value interface{}) bool {
		if client, ok := value.(*Client); ok {
			members = append(members, client)
		}
		return true
	})

	return members, nil
}

// Broadcast 向所有连接的客户端广播消息
func (p *Pool) Broadcast(messageType MessageType, content []byte) {
	message := Message{
		Type:      int(messageType),
		Content:   content,
		Timestamp: time.Now(),
		MessageID: fmt.Sprintf("broadcast-%d", time.Now().UnixNano()),
	}
	p.broadcast <- message
}

// BroadcastToGroup 向指定群组广播消息
func (p *Pool) BroadcastToGroup(groupID string, messageType MessageType, content []byte) error {
	if groupID == "" {
		return ErrInvalidGroupID
	}

	if _, ok := p.groups.Load(groupID); !ok {
		return fmt.Errorf("%w: %s", ErrGroupNotFound, groupID)
	}

	message := Message{
		Type:      int(messageType),
		Content:   content,
		To:        groupID,
		Timestamp: time.Now(),
		MessageID: fmt.Sprintf("group-%s-%d", groupID, time.Now().UnixNano()),
	}
	p.broadcast <- message
	return nil
}

// GetClient 根据ID获取客户端
func (p *Pool) GetClient(id string) (*Client, error) {
	if id == "" {
		return nil, ErrInvalidClientID
	}

	if value, ok := p.clients.Load(id); ok {
		return value.(*Client), nil
	}
	return nil, fmt.Errorf("%w: %s", ErrClientNotFound, id)
}

// Count 返回当前连接的客户端数量
func (p *Pool) Count() int {
	var count int
	p.clients.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// GroupCount 返回群组数量
func (p *Pool) GroupCount() int {
	var count int
	p.groups.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// GetClientGroups 获取客户端所在的所有群组
func (p *Pool) GetClientGroups(clientID string) ([]string, error) {
	if clientID == "" {
		return nil, ErrInvalidClientID
	}

	var groups []string
	p.groups.Range(func(key, value interface{}) bool {
		if group, ok := value.(*Group); ok {
			if _, exists := group.clients.Load(clientID); exists {
				groups = append(groups, group.ID)
			}
		}
		return true
	})

	return groups, nil
}

// Send 发送消息到客户端
func (c *Client) Send(messageType MessageType, content []byte) error {
	if c.closed.Load() {
		return ErrClientDisconnected
	}

	message := Message{
		Type:      int(messageType),
		Content:   content,
		From:      c.ID,
		Timestamp: time.Now(),
		MessageID: fmt.Sprintf("%s-%d", c.ID, time.Now().UnixNano()),
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
	if !c.closed.CompareAndSwap(false, true) {
		return
	}

	c.cancel()
	c.mu.Lock()
	if c.Conn != nil {
		_ = c.Conn.Close()
		c.Conn = nil
	}
	c.mu.Unlock()
}

// UpdatePing 更新最后一次ping时间
func (c *Client) UpdatePing() {
	c.lastPing.Store(time.Now())
}

// IsAlive 检查客户端是否存活
func (c *Client) IsAlive(timeout time.Duration) bool {
	lastPing, _ := c.lastPing.Load().(time.Time)
	return time.Since(lastPing) < timeout
}

// SetProperty 设置客户端属性
func (c *Client) SetProperty(key string, value interface{}) {
	c.Properties.Store(key, value)
}

// GetProperty 获取客户端属性
func (c *Client) GetProperty(key string) (interface{}, bool) {
	return c.Properties.Load(key)
}

// SendToClient 发送消息到指定客户端
func (p *Pool) SendToClient(targetID string, messageType MessageType, content []byte) error {
	if targetID == "" {
		return ErrInvalidClientID
	}

	value, ok := p.clients.Load(targetID)
	if !ok {
		return fmt.Errorf("%w: %s", ErrClientNotFound, targetID)
	}
	client := value.(*Client)

	return client.Send(messageType, content)
}

// BroadcastFiltered 向满足条件的客户端广播消息
func (p *Pool) BroadcastFiltered(messageType MessageType, content []byte, filter func(*Client) bool) {
	message := Message{
		Type:      int(messageType),
		Content:   content,
		Timestamp: time.Now(),
		MessageID: fmt.Sprintf("filtered-%d", time.Now().UnixNano()),
	}

	p.clients.Range(func(key, value interface{}) bool {
		if client, ok := value.(*Client); ok && filter(client) {
			select {
			case client.send <- message:
			default:
				p.unregister <- client
			}
		}
		return true
	})
}
