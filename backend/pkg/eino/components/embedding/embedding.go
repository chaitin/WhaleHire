package embedding

import (
	"context"
	"fmt"
	"strings"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
)

// EmbeddingService 嵌入服务接口
type EmbeddingService interface {
	// GetEmbedder 获取嵌入模型实例
	GetEmbedder() embedding.Embedder
	// GetDimension 获取向量维度
	GetDimension() int
}

// embeddingService 嵌入服务实现
type embeddingService struct {
	embedder  embedding.Embedder
	dimension int
}

// NewEmbeddingService 创建嵌入服务实例
func NewEmbeddingService(ctx context.Context, cfg *config.Config) (EmbeddingService, error) {
	embedder, dimension, err := NewEmbedding(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &embeddingService{
		embedder:  embedder,
		dimension: dimension,
	}, nil
}

// GetEmbedder 获取嵌入模型实例
func (s *embeddingService) GetEmbedder() embedding.Embedder {
	return s.embedder
}

// GetDimension 获取向量维度
func (s *embeddingService) GetDimension() int {
	return s.dimension
}

// NewEmbedding 构建嵌入模型实例，同时返回有效向量维度（0 表示未知/由模型默认决定）
func NewEmbedding(ctx context.Context, cfg *config.Config) (embedding.Embedder, int, error) {
	if cfg.Embedding.APIEndpoint == "" || cfg.Embedding.APIKey == "" {
		return nil, 0, fmt.Errorf("embedding service: 嵌入服务配置缺失")
	}

	var (
		dimensionPtr *int
		dimension    = cfg.Embedding.Dimension
	)
	if supportDimensionOverride(cfg.Embedding.ModelName) && dimension > 0 {
		dimensionPtr = &dimension
	}

	config := &openai.EmbeddingConfig{
		BaseURL:    cfg.Embedding.APIEndpoint,
		APIKey:     cfg.Embedding.APIKey,
		Model:      cfg.Embedding.ModelName,
		Dimensions: dimensionPtr,
	}

	embedder, err := openai.NewEmbedder(ctx, config)
	if err != nil {
		return nil, 0, err
	}

	if dimensionPtr != nil {
		return embedder, dimension, nil
	}

	return embedder, 0, nil
}

// supportDimensionOverride 检查模型是否支持维度覆盖
func supportDimensionOverride(model string) bool {
	model = strings.ToLower(model)
	if model == "" {
		return false
	}
	// OpenAI text-embedding-3 系列支持指定维度；其他模型默认使用内置维度。
	return strings.Contains(model, "text-embedding-3")
}