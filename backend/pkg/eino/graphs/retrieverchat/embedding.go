package retrieverchat

import (
	"context"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/ptonlix/whalehire/backend/config"
)

func NewEmbedding(ctx context.Context, cfg *config.Config) (eb embedding.Embedder, err error) {
	config := &openai.EmbeddingConfig{
		BaseURL: cfg.Embedding.APIEndpoint,
		APIKey:  cfg.Embedding.APIKey,
		Model:   cfg.Embedding.ModelName,
	}
	eb, err = openai.NewEmbedder(ctx, config)
	if err != nil {
		return nil, err
	}
	return eb, nil
}
