package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/chaitin/WhaleHire/backend/internal/queue"
)

// RedisQueue Redis队列实现
type RedisQueue struct {
	client *redis.Client
	opts   *Options
}

// Options Redis队列选项
type Options struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewRedisQueue 创建Redis队列实例
func NewRedisQueue(opts *Options) *RedisQueue {
	if opts == nil {
		opts = &Options{
			Addr:         "localhost:6379",
			PoolSize:     10,
			MinIdleConns: 5,
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:         opts.Addr,
		Password:     opts.Password,
		DB:           opts.DB,
		PoolSize:     opts.PoolSize,
		MinIdleConns: opts.MinIdleConns,
		MaxRetries:   opts.MaxRetries,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
	})

	return &RedisQueue{
		client: client,
		opts:   opts,
	}
}

// Publish 发布消息到指定流
func (r *RedisQueue) Publish(ctx context.Context, stream string, data map[string]interface{}) error {
	return r.PublishWithID(ctx, stream, "*", data)
}

// PublishWithID 发布消息到指定流，指定消息ID
func (r *RedisQueue) PublishWithID(ctx context.Context, stream string, id string, data map[string]interface{}) error {
	// 将数据转换为Redis Stream格式
	values := make(map[string]interface{})
	for k, v := range data {
		if b, err := json.Marshal(v); err == nil {
			values[k] = string(b)
		} else {
			values[k] = fmt.Sprintf("%v", v)
		}
	}

	// 添加时间戳和追踪ID
	values["timestamp"] = time.Now().Unix()
	if traceID, ok := data["trace_id"]; ok {
		values["trace_id"] = traceID
	}

	_, err := r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		ID:     id,
		Values: values,
	}).Result()

	return err
}

// PublishBatch 批量发布消息
func (r *RedisQueue) PublishBatch(ctx context.Context, stream string, messages []map[string]interface{}) error {
	pipe := r.client.Pipeline()

	for _, data := range messages {
		values := make(map[string]interface{})
		for k, v := range data {
			if b, err := json.Marshal(v); err == nil {
				values[k] = string(b)
			} else {
				values[k] = fmt.Sprintf("%v", v)
			}
		}
		values["timestamp"] = time.Now().Unix()

		pipe.XAdd(ctx, &redis.XAddArgs{
			Stream: stream,
			ID:     "*",
			Values: values,
		})
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Subscribe 订阅指定流
func (r *RedisQueue) Subscribe(ctx context.Context, stream string, group string, consumer string) (<-chan queue.Message, error) {
	// 创建消费者组（如果不存在）
	r.client.XGroupCreateMkStream(ctx, stream, group, "0")

	msgChan := make(chan queue.Message, 100)

	go func() {
		defer close(msgChan)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				// 读取消息
				streams, err := r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
					Group:    group,
					Consumer: consumer,
					Streams:  []string{stream, ">"},
					Count:    10,
					Block:    time.Second,
				}).Result()

				if err != nil {
					if err == redis.Nil {
						continue
					}
					// 记录错误但继续运行
					continue
				}

				for _, stream := range streams {
					for _, message := range stream.Messages {
						msg := queue.Message{
							ID:     message.ID,
							Stream: stream.Stream,
							Data:   make(map[string]interface{}),
						}

						// 解析消息数据
						for k, v := range message.Values {
							if k == "trace_id" {
								msg.TraceID = fmt.Sprintf("%v", v)
								continue
							}

							// 尝试JSON解析
							var parsed interface{}
							if str, ok := v.(string); ok {
								if err := json.Unmarshal([]byte(str), &parsed); err == nil {
									msg.Data[k] = parsed
								} else {
									msg.Data[k] = str
								}
							} else {
								msg.Data[k] = v
							}
						}

						select {
						case msgChan <- msg:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}
	}()

	return msgChan, nil
}

// Ack 确认消息处理完成
func (r *RedisQueue) Ack(ctx context.Context, stream string, group string, messageID string) error {
	_, err := r.client.XAck(ctx, stream, group, messageID).Result()
	return err
}

// CreateStream 创建流
func (r *RedisQueue) CreateStream(ctx context.Context, stream string, options *queue.ProducerOptions) error {
	// Redis Stream会在第一次写入时自动创建
	// 这里我们可以设置一些初始配置
	if options != nil && options.MaxLen > 0 {
		// 设置流的最大长度
		_, err := r.client.XTrimMaxLen(ctx, stream, options.MaxLen).Result()
		return err
	}
	return nil
}

// DeleteStream 删除流
func (r *RedisQueue) DeleteStream(ctx context.Context, stream string) error {
	_, err := r.client.Del(ctx, stream).Result()
	return err
}

// StreamExists 检查流是否存在
func (r *RedisQueue) StreamExists(ctx context.Context, stream string) (bool, error) {
	exists, err := r.client.Exists(ctx, stream).Result()
	return exists > 0, err
}

// GetStreamInfo 获取流信息
func (r *RedisQueue) GetStreamInfo(ctx context.Context, stream string) (*queue.StreamInfo, error) {
	info, err := r.client.XInfoStream(ctx, stream).Result()
	if err != nil {
		return nil, err
	}

	return &queue.StreamInfo{
		Name:         stream,
		Length:       info.Length,
		LastID:       info.LastGeneratedID,
		FirstID:      info.FirstEntry.ID,
		MaxDeletedID: info.MaxDeletedEntryID,
		CreatedAt:    time.Now(), // Redis不提供创建时间，使用当前时间
	}, nil
}

// SendToDeadLetter 发送消息到死信队列
func (r *RedisQueue) SendToDeadLetter(ctx context.Context, originalStream string, message queue.Message, reason string) error {
	deadLetterStream := originalStream + ":deadletter"

	data := message.Data
	data["original_stream"] = originalStream
	data["original_id"] = message.ID
	data["dead_letter_reason"] = reason
	data["dead_letter_time"] = time.Now().Unix()

	return r.Publish(ctx, deadLetterStream, data)
}

// GetDeadLetterMessages 获取死信消息
func (r *RedisQueue) GetDeadLetterMessages(ctx context.Context, stream string, limit int64) ([]queue.Message, error) {
	deadLetterStream := stream + ":deadletter"

	streams, err := r.client.XRead(ctx, &redis.XReadArgs{
		Streams: []string{deadLetterStream, "0"},
		Count:   limit,
	}).Result()

	if err != nil {
		return nil, err
	}

	var messages []queue.Message
	for _, stream := range streams {
		for _, msg := range stream.Messages {
			message := queue.Message{
				ID:     msg.ID,
				Stream: stream.Stream,
				Data:   make(map[string]interface{}),
			}

			for k, v := range msg.Values {
				if k == "trace_id" {
					message.TraceID = fmt.Sprintf("%v", v)
					continue
				}

				var parsed interface{}
				if str, ok := v.(string); ok {
					if err := json.Unmarshal([]byte(str), &parsed); err == nil {
						message.Data[k] = parsed
					} else {
						message.Data[k] = str
					}
				} else {
					message.Data[k] = v
				}
			}

			messages = append(messages, message)
		}
	}

	return messages, nil
}

// ReprocessDeadLetter 重新处理死信消息
func (r *RedisQueue) ReprocessDeadLetter(ctx context.Context, deadLetterStream string, messageID string, targetStream string) error {
	// 获取死信消息
	streams, err := r.client.XRange(ctx, deadLetterStream, messageID, messageID).Result()
	if err != nil {
		return err
	}

	if len(streams) == 0 {
		return fmt.Errorf("message not found: %s", messageID)
	}

	message := streams[0]
	data := make(map[string]interface{})

	for k, v := range message.Values {
		// 跳过死信相关字段
		if k == "dead_letter_reason" || k == "dead_letter_time" || k == "original_stream" || k == "original_id" {
			continue
		}

		var parsed interface{}
		if str, ok := v.(string); ok {
			if err := json.Unmarshal([]byte(str), &parsed); err == nil {
				data[k] = parsed
			} else {
				data[k] = str
			}
		} else {
			data[k] = v
		}
	}

	// 重新发布到目标流
	if err := r.Publish(ctx, targetStream, data); err != nil {
		return err
	}

	// 删除死信消息
	_, err = r.client.XDel(ctx, deadLetterStream, messageID).Result()
	return err
}

// Close 关闭连接
func (r *RedisQueue) Close() error {
	return r.client.Close()
}

// Ping 检查连接
func (r *RedisQueue) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
