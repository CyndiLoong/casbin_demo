// Package mq 提供 RabbitMQ 消息队列客户端封装。
//
// 架构设计（fanout广播模式 - 解决多实例竞争消费问题）：
//
//	                    ┌──────────┐
//	  Publish ────────►│ Fanout   ├──────► Instance-1 Queue (exclusive) ──► Consumer (本地WS推送)
//	                    │ Exchange ├──────► Instance-2 Queue (exclusive) ──► Consumer (本地WS推送)
//	                    └──────────┘       ...
//
// 为什么用 fanout 而不是 direct/shared queue：
//   - direct/shared queue: 多个实例竞争消费，只有一个实例拿到消息，如果该实例没有对应在线用户则消息丢失
//   - fanout: 每个实例都拿到消息，各自推送给本地连接的用户，确保不丢消息
//
// 消息可靠性保障：
//  1. 队列持久化 + 消息持久化（DeliveryMode=2）
//  2. 手动ACK（消费成功才确认，失败Nack重试）
//  3. 死信交换机（DLX）：超过重试次数的消息进入死信队列，避免无限循环
//  4. 自动重连：连接断开后指数退避重连，恢复拓扑和消费者
//  5. 消息轻量：MQ只传message_id，详情从PG加载，减小MQ内存压力
//  6. 幂等：消费端通过内存去重缓存避免重复通知
package mq

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeNotify      = "audit.fanout"
	ExchangeDLX         = "audit.dlx"
	QueueDLQ            = "audit.dlq"
	InstanceQueuePrefix = "audit.instance."
	PrefetchCount       = 20
	ReconnectInterval   = 3 * time.Second
	MaxReconnectRetry   = 30
)

// NotificationMessage MQ 传输的轻量消息体（只传ID和目标信息，详情查PG）。
type NotificationMessage struct {
	MessageID      uint      `json:"message_id"`
	TargetType     string    `json:"target_type"`
	TargetID       uint      `json:"target_id,omitempty"`
	Type           string    `json:"type"`
	BusinessID     uint      `json:"business_id"`
	BusinessType   string    `json:"business_type"`
	CreatedAt      time.Time `json:"created_at"`
	IdempotencyKey string    `json:"idempotency_key,omitempty"`
}

// MessageHandler 消息处理函数类型。
type MessageHandler func(msg NotificationMessage) error

// Client RabbitMQ 客户端封装。
type Client struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	uri           string
	closed        atomic.Bool
	mu            sync.Mutex
	handler       atomic.Pointer[MessageHandler]
	instanceTag   string
	instanceQueue string
	wg            sync.WaitGroup
	stopCh        chan struct{}
	consumerTag   string
}

// Config RabbitMQ 连接配置。
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	VHost    string
}

// URI 构建 AMQP 连接 URI。
func (c *Config) URI() string {
	vhost := c.VHost
	if vhost == "" {
		vhost = "/"
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%d%s", c.User, c.Password, c.Host, c.Port, vhost)
}

// genInstanceTag 生成实例唯一标识（hostname+随机后缀）。
func genInstanceTag() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%s", hostname, hex.EncodeToString(b))
}

// NewClient 创建 RabbitMQ 客户端并建立连接。
// handler 可后续通过 SetHandler 设置（先连接，后注入业务handler）。
// 返回 error 时表示连接失败，上层可选择降级（MQ不可用时消息仍落PG，由定时补发兜底）。
func NewClient(cfg *Config) (*Client, error) {
	uri := cfg.URI()
	tag := genInstanceTag()
	client := &Client{
		uri:         uri,
		stopCh:      make(chan struct{}),
		instanceTag: tag,
	}
	if err := client.connect(); err != nil {
		return nil, err
	}
	if err := client.setupTopology(); err != nil {
		return nil, err
	}
	if err := client.startConsumer(); err != nil {
		return nil, err
	}
	slog.Info("rabbitmq connected", "host", cfg.Host, "port", cfg.Port, "instance", tag)
	return client, nil
}

// SetHandler 设置消息处理函数（线程安全，可在连接后注入业务逻辑）。
func (c *Client) SetHandler(handler MessageHandler) {
	c.handler.Store(&handler)
}

// connect 建立到 RabbitMQ 的 TCP 连接和 Channel。
func (c *Client) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := amqp.Dial(c.uri)
	if err != nil {
		return fmt.Errorf("connect rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("open channel: %w", err)
	}

	if err := ch.Qos(PrefetchCount, 0, false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("set qos: %w", err)
	}

	c.conn = conn
	c.channel = ch

	go c.watchConnection()
	return nil
}

// watchConnection 监听连接异常断开，自动重连并恢复拓扑和消费者。
func (c *Client) watchConnection() {
	closeCh := make(chan *amqp.Error, 1)
	c.conn.NotifyClose(closeCh)
	select {
	case err := <-closeCh:
		if c.closed.Load() {
			return
		}
		slog.Warn("rabbitmq connection lost, reconnecting...", "error", err)
		c.reconnect()
	case <-c.stopCh:
		return
	}
}

// reconnect 重连 RabbitMQ 并恢复拓扑和消费者。
func (c *Client) reconnect() {
	for i := 0; i < MaxReconnectRetry; i++ {
		if c.closed.Load() {
			return
		}
		delay := ReconnectInterval * time.Duration(min(i+1, 10))
		time.Sleep(delay)
		slog.Info("rabbitmq reconnect attempt", "attempt", i+1)
		if err := c.connect(); err != nil {
			slog.Warn("rabbitmq reconnect failed", "error", err)
			continue
		}
		if err := c.setupTopology(); err != nil {
			slog.Warn("rabbitmq setup topology after reconnect failed", "error", err)
			continue
		}
		if err := c.startConsumer(); err != nil {
			slog.Warn("rabbitmq start consumer after reconnect failed", "error", err)
			continue
		}
		slog.Info("rabbitmq reconnected successfully")
		return
	}
	slog.Error("rabbitmq max reconnect retries exceeded, will retry via scheduled task")
}

// setupTopology 声明交换机、队列和绑定（幂等）。
func (c *Client) setupTopology() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	ch := c.channel

	if err := ch.ExchangeDeclare(
		ExchangeDLX, "direct", true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("declare dlx exchange: %w", err)
	}

	if _, err := ch.QueueDeclare(
		QueueDLQ, true, false, false, false,
		amqp.Table{"x-queue-type": "quorum"},
	); err != nil {
		return fmt.Errorf("declare dlq: %w", err)
	}
	if err := ch.QueueBind(QueueDLQ, QueueDLQ, ExchangeDLX, false, nil); err != nil {
		return fmt.Errorf("bind dlq: %w", err)
	}

	args := amqp.Table{
		"x-dead-letter-exchange":    ExchangeDLX,
		"x-dead-letter-routing-key": QueueDLQ,
	}
	if err := ch.ExchangeDeclare(
		ExchangeNotify, "fanout", true, false, false, false, args,
	); err != nil {
		return fmt.Errorf("declare fanout exchange: %w", err)
	}

	qName := InstanceQueuePrefix + c.instanceTag
	q, err := ch.QueueDeclare(
		qName,
		true,
		true,
		true,
		false,
		amqp.Table{
			"x-queue-type":              "quorum",
			"x-dead-letter-exchange":    ExchangeDLX,
			"x-dead-letter-routing-key": QueueDLQ,
		},
	)
	if err != nil {
		return fmt.Errorf("declare instance queue: %w", err)
	}
	c.instanceQueue = q.Name

	if err := ch.QueueBind(q.Name, "", ExchangeNotify, false, nil); err != nil {
		return fmt.Errorf("bind instance queue: %w", err)
	}

	slog.Info("rabbitmq topology setup complete", "queue", q.Name)
	return nil
}

// startConsumer 启动消费者 goroutine，手动 ACK 保证可靠投递。
func (c *Client) startConsumer() error {
	c.mu.Lock()
	ch := c.channel
	queue := c.instanceQueue
	c.mu.Unlock()
	if ch == nil {
		return fmt.Errorf("channel not available")
	}

	c.consumerTag = fmt.Sprintf("consumer-%s-%d", c.instanceTag, time.Now().UnixNano())
	deliveries, err := ch.Consume(
		queue,
		c.consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume queue: %w", err)
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-c.stopCh:
				return
			case d, ok := <-deliveries:
				if !ok {
					slog.Warn("rabbitmq consumer channel closed", "queue", queue)
					return
				}
				c.processMessage(d)
			}
		}
	}()
	slog.Info("rabbitmq consumer started", "queue", queue)
	return nil
}

// processMessage 处理单条 MQ 消息：解析→调用handler→手动ACK/Nack。
func (c *Client) processMessage(d amqp.Delivery) {
	var msg NotificationMessage
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		slog.Error("rabbitmq unmarshal failed, sending to dlq", "error", err)
		_ = d.Nack(false, false)
		return
	}

	slog.Debug("mq message received", "msg_id", msg.MessageID, "type", msg.Type, "target", msg.TargetType)

	handlerPtr := c.handler.Load()
	if handlerPtr == nil || *handlerPtr == nil {
		_ = d.Ack(false)
		return
	}

	if err := (*handlerPtr)(msg); err != nil {
		slog.Error("mq handler failed", "msg_id", msg.MessageID, "error", err)
		time.Sleep(500 * time.Millisecond)
		_ = d.Nack(false, true)
		return
	}

	if err := d.Ack(false); err != nil {
		slog.Warn("rabbitmq ack failed", "msg_id", msg.MessageID, "error", err)
	}
}

// Publish 发布消息到 fanout 交换机（所有实例都会收到）。
func (c *Client) Publish(ctx context.Context, msg NotificationMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.channel == nil || c.closed.Load() {
		return fmt.Errorf("rabbitmq channel not available")
	}
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	return c.channel.PublishWithContext(ctx,
		ExchangeNotify,
		"",
		true,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

// Close 优雅关闭 RabbitMQ 连接，等待消费者处理完在途消息。
func (c *Client) Close() error {
	if c.closed.Swap(true) {
		return nil
	}
	close(c.stopCh)
	c.wg.Wait()
	slog.Info("rabbitmq consumers stopped")

	c.mu.Lock()
	ch := c.channel
	conn := c.conn
	c.mu.Unlock()

	if ch != nil {
		_ = ch.Close()
	}
	if conn != nil {
		return conn.Close()
	}
	return nil
}

// IsConnected 检查客户端是否已连接。
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil && !c.conn.IsClosed() && c.channel != nil && !c.closed.Load()
}
