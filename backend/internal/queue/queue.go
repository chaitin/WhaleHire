package queue

import (
	"context"
	"time"
)

// Message 队列消息
type Message struct {
	ID      string                 `json:"id"`
	Stream  string                 `json:"stream"`
	Data    map[string]interface{} `json:"data"`
	TraceID string                 `json:"trace_id"`
}

// Producer 消息生产者接口
type Producer interface {
	// Publish 发布消息到指定流
	Publish(ctx context.Context, stream string, data map[string]interface{}) error
	// PublishWithID 发布消息到指定流，指定消息ID
	PublishWithID(ctx context.Context, stream string, id string, data map[string]interface{}) error
	// PublishBatch 批量发布消息
	PublishBatch(ctx context.Context, stream string, messages []map[string]interface{}) error
	// Close 关闭生产者
	Close() error
}

// Consumer 消息消费者接口
type Consumer interface {
	// Subscribe 订阅指定流
	Subscribe(ctx context.Context, stream string, group string, consumer string) (<-chan Message, error)
	// Ack 确认消息处理完成
	Ack(ctx context.Context, stream string, group string, messageID string) error
	// Close 关闭消费者
	Close() error
}

// ConsumerOptions 消费者选项
type ConsumerOptions struct {
	Group       string        // 消费者组名
	Consumer    string        // 消费者名称
	Block       time.Duration // 阻塞时间
	Count       int64         // 每次读取消息数量
	StartFromID string        // 开始读取的消息ID，">" 表示从最新消息开始
}

// ProducerOptions 生产者选项
type ProducerOptions struct {
	MaxLen      int64 // 流的最大长度
	Approximate bool  // 是否使用近似长度限制
}

// QueueManager 队列管理器接口
type QueueManager interface {
	Producer
	Consumer
	// CreateStream 创建流
	CreateStream(ctx context.Context, stream string, options *ProducerOptions) error
	// DeleteStream 删除流
	DeleteStream(ctx context.Context, stream string) error
	// StreamExists 检查流是否存在
	StreamExists(ctx context.Context, stream string) (bool, error)
	// GetStreamInfo 获取流信息
	GetStreamInfo(ctx context.Context, stream string) (*StreamInfo, error)
}

// StreamInfo 流信息
type StreamInfo struct {
	Name         string    `json:"name"`
	Length       int64     `json:"length"`
	LastID       string    `json:"last_id"`
	FirstID      string    `json:"first_id"`
	MaxDeletedID string    `json:"max_deleted_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// DeadLetterQueue 死信队列接口
type DeadLetterQueue interface {
	// SendToDeadLetter 发送消息到死信队列
	SendToDeadLetter(ctx context.Context, originalStream string, message Message, reason string) error
	// GetDeadLetterMessages 获取死信消息
	GetDeadLetterMessages(ctx context.Context, stream string, limit int64) ([]Message, error)
	// ReprocessDeadLetter 重新处理死信消息
	ReprocessDeadLetter(ctx context.Context, deadLetterStream string, messageID string, targetStream string) error
}