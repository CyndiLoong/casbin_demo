// Package ws 提供 WebSocket 实时推送 Hub，支持多实例部署通过 Redis Pub/Sub 广播。
//
// 架构设计：
//
//  单实例：   Client ←→ Hub (直接通过 channel 广播)
//  多实例：   Client ←→ Hub ←→ Redis PubSub ←→ 其他实例 Hub ←→ Client
//            (Redis PubSub 负责跨实例消息分发，每个实例收到后推送本地连接)
//
// 连接管理：
//   - 用户通过 JWT 认证后升级为 WebSocket 连接
//   - 同一用户多端同时在线时，消息广播到所有连接
//   - 单用户最大连接数限制 MaxConnsPerUser=5，超出时关闭最旧连接
//   - 断开连接时自动清理注册信息
//
// 心跳机制：
//   - 服务端每 25s 发送 Ping（控制帧），客户端 60s 内无 Pong 则断开
//   - 客户端可发送文本 "ping"，服务端回复文本 "pong"
//
// 幂等去重：
//   - 基于 notification ID 的 LRU 缓存防止 MQ 重投递导致重复弹窗
package ws

import (
	"container/list"
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

const (
	PubSubChannel   = "ws:broadcast"
	AdminRoom       = "admin"
	WriteWait       = 10 * time.Second
	PongWait        = 60 * time.Second
	PingPeriod      = 25 * time.Second
	MaxMessageSize  = 1024
	MaxConnsPerUser = 5
	DedupWindow     = 5 * time.Minute
	DedupMaxEntries = 10000
)

// WsMessage WebSocket 推送消息结构。
type WsMessage struct {
	Type       string      `json:"type"`
	TargetType string      `json:"target_type"`
	TargetID   uint        `json:"target_id,omitempty"`
	Room       string      `json:"room,omitempty"`
	Data       interface{} `json:"data"`
	Timestamp  time.Time   `json:"timestamp"`
}

// IDProvider 可提取消息 ID 的接口，用于幂等去重。
type IDProvider interface {
	GetID() string
}

// extractMsgID 从 WsMessage 的 Data 中提取幂等去重用的 ID。
// 支持 map[string]interface{}（key="id"）、IDProvider 接口，以及通过 JSON 序列化兜底。
func extractMsgID(data interface{}) string {
	if data == nil {
		return ""
	}
	if ip, ok := data.(IDProvider); ok {
		return ip.GetID()
	}
	if m, ok := data.(map[string]interface{}); ok {
		if id, ok := m["id"].(string); ok {
			return id
		}
	}
	b, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	var tmp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return ""
	}
	return tmp.ID
}

// Client 单个 WebSocket 连接。
type Client struct {
	UserID uint
	Roles  []string
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
	mu     sync.Mutex
	closed bool
}

// dedupEntry 幂等去重缓存条目。
type dedupEntry struct {
	id        string
	expiresAt time.Time
}

// Hub 管理所有 WebSocket 连接。
type Hub struct {
	clients    map[*Client]bool
	userConns  map[uint]map[*Client]bool
	roomConns  map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	direct     chan *directMessage
	room       chan *roomMessage
	rdb        *redis.Client
	pubsub     *redis.PubSub
	stopCh     chan struct{}
	wg         sync.WaitGroup
	mu         sync.RWMutex

	dedupMu  sync.Mutex
	dedupLL  *list.List
	dedupMap map[string]*list.Element
}

type directMessage struct {
	UserID uint
	Data   []byte
}

type roomMessage struct {
	Room string
	Data []byte
}

// NewHub 创建 WebSocket Hub 实例。
func NewHub(rdb *redis.Client) *Hub {
	h := &Hub{
		clients:    make(map[*Client]bool),
		userConns:  make(map[uint]map[*Client]bool),
		roomConns:  make(map[string]map[*Client]bool),
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
		broadcast:  make(chan []byte, 256),
		direct:     make(chan *directMessage, 256),
		room:       make(chan *roomMessage, 256),
		rdb:        rdb,
		stopCh:     make(chan struct{}),
		dedupLL:    list.New(),
		dedupMap:   make(map[string]*list.Element),
	}
	return h
}

// Run 启动 Hub 事件循环和 Redis PubSub 订阅。
func (h *Hub) Run() {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		h.eventLoop()
	}()
	go h.runDedupCleanup()
	if h.rdb != nil {
		go h.subscribeRedis()
	}
	slog.Info("websocket hub started", "redis_pubsub", h.rdb != nil)
}

// subscribeRedis 订阅 Redis Pub/Sub 频道，接收其他实例的广播消息。
func (h *Hub) subscribeRedis() {
	h.mu.Lock()
	h.pubsub = h.rdb.Subscribe(context.Background(), PubSubChannel)
	h.mu.Unlock()

	ch := h.pubsub.Channel()
	for {
		select {
		case <-h.stopCh:
			return
		case msg, ok := <-ch:
			if !ok {
				slog.Warn("redis pubsub channel closed, will reconnect")
				time.Sleep(3 * time.Second)
				h.resubscribeRedis()
				return
			}
			if msg == nil {
				continue
			}
			h.handleRedisMessage([]byte(msg.Payload))
		}
	}
}

// resubscribeRedis Redis 断连后重新订阅。
func (h *Hub) resubscribeRedis() {
	for i := 0; i < 10; i++ {
		select {
		case <-h.stopCh:
			return
		default:
		}
		time.Sleep(time.Duration(i+1) * 3 * time.Second)
		h.mu.Lock()
		if h.pubsub != nil {
			_ = h.pubsub.Close()
		}
		h.mu.Unlock()
		if h.rdb == nil {
			return
		}
		pubsub := h.rdb.Subscribe(context.Background(), PubSubChannel)
		_, err := pubsub.Receive(context.Background())
		if err != nil {
			slog.Warn("redis pubsub resubscribe failed", "attempt", i+1, "error", err)
			_ = pubsub.Close()
			continue
		}
		h.mu.Lock()
		h.pubsub = pubsub
		h.mu.Unlock()
		slog.Info("redis pubsub resubscribed")
		go h.subscribeRedis()
		return
	}
	slog.Error("redis pubsub max resubscribe retries exceeded")
}

// handleRedisMessage 处理从 Redis Pub/Sub 收到的其他实例广播消息。
// 注意：不重复推送给本地发送者（通过幂等去重实现）。
func (h *Hub) handleRedisMessage(data []byte) {
	var msg WsMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		slog.Warn("redis pubsub unmarshal failed", "error", err)
		return
	}

	if id := extractMsgID(msg.Data); id != "" && h.isDuplicate(id) {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	switch msg.TargetType {
	case "room":
		if msg.Room == AdminRoom {
			for client := range h.roomConns[AdminRoom] {
				client.sendSafe(data)
			}
		}
	case "user":
		for client := range h.userConns[msg.TargetID] {
			client.sendSafe(data)
		}
	default:
		for client := range h.clients {
			client.sendSafe(data)
		}
	}
}

// eventLoop Hub 主事件循环。
func (h *Hub) eventLoop() {
	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)
		case client := <-h.unregister:
			h.removeClient(client)
		case message := <-h.broadcast:
			h.localBroadcastAll(message)
		case dm := <-h.direct:
			h.sendToUserLocal(dm.UserID, dm.Data)
		case rm := <-h.room:
			h.sendToRoomLocal(rm.Room, rm.Data)
		case <-h.stopCh:
			return
		}
	}
}

// handleRegister 处理客户端注册，实施连接数限制。
func (h *Hub) handleRegister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	existingConns := h.userConns[client.UserID]
	for len(existingConns) >= MaxConnsPerUser {
		var oldest *Client
		for c := range existingConns {
			oldest = c
			break
		}
		if oldest != nil {
			slog.Warn("max connections per user reached, closing oldest", "user_id", client.UserID)
			go oldest.close()
			delete(existingConns, oldest)
			delete(h.clients, oldest)
			h.removeFromRoomsLocked(oldest)
		}
	}

	h.clients[client] = true
	if _, ok := h.userConns[client.UserID]; !ok {
		h.userConns[client.UserID] = make(map[*Client]bool)
	}
	h.userConns[client.UserID][client] = true

	for _, role := range client.Roles {
		if role == "admin" {
			if _, ok := h.roomConns[AdminRoom]; !ok {
				h.roomConns[AdminRoom] = make(map[*Client]bool)
			}
			h.roomConns[AdminRoom][client] = true
		}
	}

	slog.Debug("ws client registered", "user_id", client.UserID, "roles", client.Roles)
}

// removeFromRoomsLocked 在锁内将客户端从所有 room 中移除。
func (h *Hub) removeFromRoomsLocked(client *Client) {
	for _, role := range client.Roles {
		if role == "admin" {
			if conns, ok := h.roomConns[AdminRoom]; ok {
				delete(conns, client)
				if len(conns) == 0 {
					delete(h.roomConns, AdminRoom)
				}
			}
		}
	}
}

// removeClient 从 Hub 中移除客户端。
func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[client]; !ok {
		return
	}
	client.close()
	delete(h.clients, client)
	if conns, ok := h.userConns[client.UserID]; ok {
		delete(conns, client)
		if len(conns) == 0 {
			delete(h.userConns, client.UserID)
		}
	}
	h.removeFromRoomsLocked(client)
	slog.Debug("ws client unregistered", "user_id", client.UserID)
}

// isDuplicate 检查消息是否已处理过（幂等）。
func (h *Hub) isDuplicate(id string) bool {
	if id == "" {
		return false
	}
	h.dedupMu.Lock()
	defer h.dedupMu.Unlock()
	if el, ok := h.dedupMap[id]; ok {
		entry := el.Value.(*dedupEntry)
		if time.Now().Before(entry.expiresAt) {
			h.dedupLL.MoveToFront(el)
			return true
		}
		h.dedupLL.Remove(el)
		delete(h.dedupMap, id)
	}
	entry := &dedupEntry{id: id, expiresAt: time.Now().Add(DedupWindow)}
	el := h.dedupLL.PushFront(entry)
	h.dedupMap[id] = el

	for h.dedupLL.Len() > DedupMaxEntries {
		tail := h.dedupLL.Back()
		if tail != nil {
			h.dedupLL.Remove(tail)
			if e, ok := tail.Value.(*dedupEntry); ok {
				delete(h.dedupMap, e.id)
			}
		}
	}
	return false
}

// runDedupCleanup 定期清理过期的去重条目。
func (h *Hub) runDedupCleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			h.cleanupDedup()
		case <-h.stopCh:
			return
		}
	}
}

func (h *Hub) cleanupDedup() {
	h.dedupMu.Lock()
	defer h.dedupMu.Unlock()
	now := time.Now()
	for el := h.dedupLL.Back(); el != nil; {
		prev := el.Prev()
		entry := el.Value.(*dedupEntry)
		if now.After(entry.expiresAt) {
			h.dedupLL.Remove(el)
			delete(h.dedupMap, entry.id)
		} else {
			break
		}
		el = prev
	}
}

// localBroadcastAll 将消息广播给本地所有连接。
func (h *Hub) localBroadcastAll(data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		client.sendSafe(data)
	}
}

// sendToUserLocal 向指定用户的本地连接推送消息（不跨实例）。
func (h *Hub) sendToUserLocal(userID uint, data []byte) {
	h.mu.RLock()
	conns := h.userConns[userID]
	h.mu.RUnlock()
	for client := range conns {
		client.sendSafe(data)
	}
}

// sendToAdminsLocal 向本地所有管理员推送消息。
func (h *Hub) sendToAdminsLocal(data []byte) {
	h.mu.RLock()
	conns := h.roomConns[AdminRoom]
	h.mu.RUnlock()
	for client := range conns {
		client.sendSafe(data)
	}
}

// sendToRoomLocal 向指定房间的本地连接推送消息。
func (h *Hub) sendToRoomLocal(room string, data []byte) {
	h.mu.RLock()
	conns := h.roomConns[room]
	h.mu.RUnlock()
	for client := range conns {
		client.sendSafe(data)
	}
}

// SendToUserLocal 对外暴露的本地用户直推方法（MQ消费端使用，不经过Redis PubSub避免循环）。
func (h *Hub) SendToUserLocal(userID uint, msg WsMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("ws marshal message failed", "error", err)
		return
	}
	h.direct <- &directMessage{UserID: userID, Data: data}
}

// SendToAdminsLocal 对外暴露的本地管理员广播方法（MQ消费端使用，只推本地管理员房间，不经过Redis PubSub避免循环）。
func (h *Hub) SendToAdminsLocal(msg WsMessage) {
	if id := extractMsgID(msg.Data); id != "" && h.isDuplicate(id) {
		slog.Debug("ws duplicate admin message suppressed", "id", id)
		return
	}
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("ws marshal admin broadcast failed", "error", err)
		return
	}
	h.room <- &roomMessage{Room: AdminRoom, Data: data}
}

// SendToUser 对外暴露的用户直推方法（先本地推送，再Redis PubSub跨实例）。
func (h *Hub) SendToUser(userID uint, msg WsMessage) {
	if id := extractMsgID(msg.Data); id != "" && h.isDuplicate(id) {
		return
	}
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("ws marshal message failed", "error", err)
		return
	}
	h.direct <- &directMessage{UserID: userID, Data: data}
	h.publishToRedis(data)
}

// BroadcastToAdmins 向所有管理员广播消息（跨实例 via Redis PubSub）。
func (h *Hub) BroadcastToAdmins(msg WsMessage) {
	msg.TargetType = "room"
	msg.Room = AdminRoom
	if id := extractMsgID(msg.Data); id != "" && h.isDuplicate(id) {
		return
	}
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("ws marshal admin broadcast failed", "error", err)
		return
	}
	h.mu.RLock()
	if conns, ok := h.roomConns[AdminRoom]; ok {
		for client := range conns {
			client.sendSafe(data)
		}
	}
	h.mu.RUnlock()
	h.publishToRedis(data)
}

// publishToRedis 发布消息到 Redis Pub/Sub 供其他实例消费。
func (h *Hub) publishToRedis(data []byte) {
	if h.rdb == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := h.rdb.Publish(ctx, PubSubChannel, data).Err(); err != nil {
		slog.Warn("redis publish failed", "error", err)
	}
}

// sendSafe 线程安全地向客户端发送消息。
func (c *Client) sendSafe(data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	select {
	case c.Send <- data:
	default:
		slog.Warn("ws client send buffer full", "user_id", c.UserID)
	}
}

// close 关闭客户端连接和send channel。
func (c *Client) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.closed = true
	close(c.Send)
	_ = c.Conn.Close()
}

// ReadPump 从 WebSocket 连接读取消息（处理 ping/pong 心跳）。
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		_ = c.Conn.Close()
	}()
	c.Conn.SetReadLimit(MaxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(PongWait))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})
	for {
		msgType, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Warn("ws read error", "user_id", c.UserID, "error", err)
			}
			break
		}
		if msgType == websocket.TextMessage {
			text := string(message)
			if text == "ping" {
				c.mu.Lock()
				if !c.closed {
					_ = c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
					_ = c.Conn.WriteMessage(websocket.TextMessage, []byte("pong"))
				}
				c.mu.Unlock()
			}
		}
	}
}

// WritePump 向 WebSocket 连接写入消息（处理心跳 ping 和发送队列）。
func (c *Client) WritePump() {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)
			n := len(c.Send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-c.Send)
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.mu.Lock()
			if c.closed {
				c.mu.Unlock()
				return
			}
			_ = c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			err := c.Conn.WriteMessage(websocket.PingMessage, nil)
			c.mu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

// OnlineCount 返回当前在线连接数。
func (h *Hub) OnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Register 将客户端注册到 Hub。
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Stop 优雅停止 Hub。
func (h *Hub) Stop() {
	close(h.stopCh)
	h.mu.Lock()
	for client := range h.clients {
		client.close()
	}
	h.clients = make(map[*Client]bool)
	h.userConns = make(map[uint]map[*Client]bool)
	h.roomConns = make(map[string]map[*Client]bool)
	if h.pubsub != nil {
		_ = h.pubsub.Close()
		h.pubsub = nil
	}
	h.mu.Unlock()
	h.wg.Wait()
	slog.Info("websocket hub stopped")
}
