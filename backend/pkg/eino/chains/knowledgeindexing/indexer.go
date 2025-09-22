package knowledgeindexing

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/pkg/store"
	"github.com/cloudwego/eino-ext/components/indexer/redis"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

// NewIndexer component initialization function of node 'RedisIndexer' in graph 'KnowledgeIndexing'
func NewIndexer(ctx context.Context, cfg *config.Config) (idr indexer.Indexer, err error) {
	// 获取配置
	// cfg, err := config.Init()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to load config: %w", err)
	// }

	// 初始化向量索引
	err = store.InitVectorIndex(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init redis index: %w", err)
	}

	redisClient := store.NewRedisCli(cfg)

	config := &redis.IndexerConfig{
		Client:    redisClient,
		KeyPrefix: store.RedisPrefix,
		BatchSize: 1,
		DocumentToHashes: func(ctx context.Context, doc *schema.Document) (*redis.Hashes, error) {
			if doc.ID == "" {
				doc.ID = uuid.New().String()
			}
			key := doc.ID

			var metadataBytes []byte
			metadataBytes, err = json.Marshal(doc.MetaData)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal metadata: %w", err)
			}

			return &redis.Hashes{
				Key: key,
				Field2Value: map[string]redis.FieldValue{
					store.DocumentIDField: {Value: strings.Split(doc.ID, "_")[0]},
					store.ContentField:    {Value: doc.Content, EmbedKey: store.VectorField},
					store.MetadataField:   {Value: string(metadataBytes)},
				},
			}, nil
		},
	}

	embeddingIns11, err := newEmbedding(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if embeddingIns11 != nil {
		// 如果有嵌入模型，设置到配置中
		config.Embedding = embeddingIns11
	}

	idr, err = redis.NewIndexer(ctx, config)
	if err != nil {
		return nil, err
	}
	return idr, nil
}
