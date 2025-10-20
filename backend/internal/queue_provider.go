package internal

import (
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/internal/queue"
	queueredis "github.com/chaitin/WhaleHire/backend/internal/queue/redis"
)

// NewQueueProducer 创建队列生产者
func NewQueueProducer(redisClient *redis.Client, cfg *config.Config) queue.Producer {
	opts := &queueredis.Options{
		Addr:     redisClient.Options().Addr,
		Password: redisClient.Options().Password,
		DB:       redisClient.Options().DB,
	}
	return queueredis.NewRedisQueue(opts)
}

// NewQueueConsumer 创建队列消费者
func NewQueueConsumer(redisClient *redis.Client, cfg *config.Config) queue.Consumer {
	opts := &queueredis.Options{
		Addr:     redisClient.Options().Addr,
		Password: redisClient.Options().Password,
		DB:       redisClient.Options().DB,
	}
	return queueredis.NewRedisQueue(opts)
}

var QueueProvider = wire.NewSet(
	NewQueueProducer,
	NewQueueConsumer,
)
