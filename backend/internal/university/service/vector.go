package service

import (
	"context"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/components/embedding"
	"github.com/pgvector/pgvector-go"
)

// VectorService 向量生成服务接口
type VectorService interface {
	// GenerateVector 为文本内容生成向量
	GenerateVector(ctx context.Context, content string) (pgvector.Vector, error)
	// BatchGenerateVectors 批量为文本内容生成向量
	BatchGenerateVectors(ctx context.Context, contents []string) ([]pgvector.Vector, error)
}

type vectorService struct {
	config *config.Config
}

// NewVectorService 创建向量生成服务
func NewVectorService(cfg *config.Config) VectorService {
	return &vectorService{
		config: cfg,
	}
}

// GenerateVector 为文本内容生成向量
func (s *vectorService) GenerateVector(ctx context.Context, content string) (pgvector.Vector, error) {
	if content == "" {
		return pgvector.Vector{}, fmt.Errorf("content is required")
	}

	// 创建嵌入模型
	embedder, _, err := embedding.NewEmbedding(ctx, s.config)
	if err != nil {
		return pgvector.Vector{}, fmt.Errorf("创建嵌入模型失败: %w", err)
	}

	// 生成向量
	vectors, err := embedder.EmbedStrings(ctx, []string{content})
	if err != nil {
		return pgvector.Vector{}, fmt.Errorf("生成向量失败: %w", err)
	}

	if len(vectors) == 0 || len(vectors[0]) == 0 {
		return pgvector.Vector{}, fmt.Errorf("生成的向量为空")
	}

	// 转换为 pgvector.Vector
	vec := make([]float32, len(vectors[0]))
	for i, v := range vectors[0] {
		vec[i] = float32(v)
	}

	return pgvector.NewVector(vec), nil
}

// BatchGenerateVectors 批量为文本内容生成向量
func (s *vectorService) BatchGenerateVectors(ctx context.Context, contents []string) ([]pgvector.Vector, error) {
	if len(contents) == 0 {
		return nil, fmt.Errorf("contents list is empty")
	}

	// 创建嵌入模型
	embedder, _, err := embedding.NewEmbedding(ctx, s.config)
	if err != nil {
		return nil, fmt.Errorf("创建嵌入模型失败: %w", err)
	}

	// 批量生成向量
	vectors, err := embedder.EmbedStrings(ctx, contents)
	if err != nil {
		return nil, fmt.Errorf("批量生成向量失败: %w", err)
	}

	if len(vectors) != len(contents) {
		return nil, fmt.Errorf("生成的向量数量不匹配: 期望 %d，实际 %d", len(contents), len(vectors))
	}

	// 构建结果列表
	result := make([]pgvector.Vector, len(vectors))
	for i, vector := range vectors {
		if len(vector) == 0 {
			return nil, fmt.Errorf("生成的向量为空 (索引: %d)", i)
		}

		// 转换为 pgvector.Vector
		vec := make([]float32, len(vector))
		for j, v := range vector {
			vec[j] = float32(v)
		}

		result[i] = pgvector.NewVector(vec)
	}

	return result, nil
}
