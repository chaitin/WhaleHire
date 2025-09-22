package knowledgeindexing

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/ptonlix/whalehire/backend/config"
)

func newEmbedding(ctx context.Context, cfg *config.Config) (eb embedding.Embedder, err error) {
	fmt.Printf("newEmbedding, cfg.Embedding.Dimension=%d\n", cfg.Embedding.Dimension)
	config := &openai.EmbeddingConfig{
		BaseURL:    cfg.Embedding.APIEndpoint,
		APIKey:     cfg.Embedding.APIKey,
		Model:      cfg.Embedding.ModelName,
		Dimensions: &cfg.Embedding.Dimension,
	}
	eb, err = openai.NewEmbedder(ctx, config)
	if err != nil {
		return nil, err
	}
	return eb, nil
}
