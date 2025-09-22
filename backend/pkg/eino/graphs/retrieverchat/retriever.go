package retrieverchat

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cloudwego/eino/schema"
	redisCli "github.com/redis/go-redis/v9"

	"github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/ptonlix/whalehire/backend/config"
	"github.com/ptonlix/whalehire/backend/pkg/store"
)

func NewRetriever(ctx context.Context, cfg *config.Config) (rtr retriever.Retriever, err error) {
	// 使用配置系统获取Redis客户端
	redisClient := store.NewRedisCli(cfg)
	fmt.Println("NewRetriever", cfg.Retriever.TopK)
	fmt.Println("NewRetriever", cfg.Retriever.DistanceThreshold)

	config := &redis.RetrieverConfig{
		Client:       redisClient,
		Index:        store.RedisPrefix + store.IndexName,
		Dialect:      2,
		ReturnFields: []string{store.DocumentIDField, store.ContentField, store.MetadataField, store.DistanceField},
		TopK:         cfg.Retriever.TopK,
		VectorField:  store.VectorField,
		DocumentConverter: func(ctx context.Context, doc redisCli.Document) (*schema.Document, error) {
			fmt.Println("DocumentConverter", doc.ID)
			resp := &schema.Document{
				ID:       doc.ID,
				Content:  "",
				MetaData: map[string]any{},
			}
			for field, val := range doc.Fields {
				if field == store.DocumentIDField {
					resp.ID = val
				} else if field == store.ContentField {
					resp.Content = val
				} else if field == store.MetadataField {
					resp.MetaData[field] = val
				} else if field == store.DistanceField {
					distance, _ := strconv.ParseFloat(val, 64)
					if err != nil {
						continue
					}
					resp.WithScore(1 - distance)
				}
			}

			return resp, nil
		},
	}

	embeddingIns11, err := NewEmbedding(ctx, cfg)
	if err != nil {
		return nil, err
	}
	config.Embedding = embeddingIns11

	rtr, err = redis.NewRetriever(ctx, config)
	if err != nil {
		return nil, err
	}
	return rtr, nil
}
