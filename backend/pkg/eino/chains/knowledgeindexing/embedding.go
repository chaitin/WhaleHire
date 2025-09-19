/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package knowledgeindexing

import (
	"context"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/ptonlix/whalehire/backend/config"
)

func newEmbedding(ctx context.Context, cfg *config.Config) (eb embedding.Embedder, err error) {
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
