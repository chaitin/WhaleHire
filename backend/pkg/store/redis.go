package store

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"

	"github.com/chaitin/WhaleHire/backend/config"
)

const (
	RedisPrefix = "eino:doc:"
	IndexName   = "vector_index"

	DocumentIDField = "docid"
	ContentField    = "content"
	MetadataField   = "metadata"
	VectorField     = "content_vector"
	DistanceField   = "distance"
)

var initOnce sync.Once

func NewRedisCli(cfg *config.Config) *redis.Client {
	addr := net.JoinHostPort(cfg.Redis.Host, cfg.Redis.Port)
	rdb := redis.NewClient(&redis.Options{
		Addr:          addr,
		Password:      cfg.Redis.Pass,
		DB:            cfg.Redis.DB,
		MaxIdleConns:  3,
		UnstableResp3: true,
	})
	return rdb
}

type RedisResult struct {
	Key   string
	Value int64
}

func ScanRedis(ctx context.Context, rdb *redis.Client, pattern string, limit int, fn func(rs []*RedisResult) error) error {
	iter := rdb.Scan(ctx, 0, pattern, int64(limit)).Iterator()

	rs := make([]*RedisResult, 0)
	for iter.Next(ctx) {
		value, err := rdb.Get(ctx, iter.Val()).Int64()
		if err != nil {
			return err
		}
		rs = append(rs, &RedisResult{
			Key:   iter.Val(),
			Value: value,
		})

		if len(rs) >= limit {
			if err := fn(rs); err != nil {
				return err
			}
			rs = make([]*RedisResult, 0)
		}
	}

	if len(rs) > 0 {
		if err := fn(rs); err != nil {
			return err
		}
	}
	return nil
}

// VectorConfig 向量数据库配置
type VectorConfig struct {
	RedisAddr string

	Dimension int
}

// InitVectorIndex 初始化向量索引
func InitVectorIndex(ctx context.Context, cfg *config.Config) error {
	if !cfg.Redis.Vector.Enabled {
		return nil
	}

	var err error
	initOnce.Do(func() {
		err = initRedisIndex(ctx, cfg)
	})
	return err
}

// initRedisIndex 初始化 Redis 向量索引
func initRedisIndex(ctx context.Context, cfg *config.Config) error {
	if cfg.Redis.Vector.Dimension <= 0 {
		return fmt.Errorf("dimension must be positive")
	}

	client := NewRedisCli(cfg)

	defer func() {
		if err := recover(); err != nil {
			client.Close()
		}
	}()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	indexName := fmt.Sprintf("%s%s", RedisPrefix, IndexName)

	// 检查是否存在索引
	exists, err := client.Do(ctx, "FT.INFO", indexName).Result()
	if err != nil {
		if !strings.Contains(err.Error(), "Unknown index name") {
			return fmt.Errorf("failed to check if index exists: %w", err)
		}
		err = nil
	} else if exists != nil {
		return nil
	}

	// Create new index
	createIndexArgs := []interface{}{
		"FT.CREATE", indexName,
		"ON", "HASH",
		"PREFIX", "1", RedisPrefix,
		"SCHEMA",
		DocumentIDField, "TEXT",
		ContentField, "TEXT",
		MetadataField, "TEXT",
		VectorField, "VECTOR", "FLAT",
		"6",
		"TYPE", "FLOAT32",
		"DIM", strconv.Itoa(cfg.Redis.Vector.Dimension),
		"DISTANCE_METRIC", "COSINE",
	}

	fmt.Printf("initRedisIndex, indexName=%s, cfg.Redis.Vector.Dimension=%d\n", indexName, cfg.Redis.Vector.Dimension)
	if err = client.Do(ctx, createIndexArgs...).Err(); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	// 验证索引是否创建成功
	if _, err = client.Do(ctx, "FT.INFO", indexName).Result(); err != nil {
		return fmt.Errorf("failed to verify index creation: %w", err)
	}

	return nil
}
