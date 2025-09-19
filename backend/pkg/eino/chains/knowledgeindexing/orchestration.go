package knowledgeindexing

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/ptonlix/whalehire/backend/config"
)

// KnowledgeIndexingChain 知识索引处理链
type KnowledgeIndexingChain struct {
	chain *compose.Chain[[]*schema.Document, []string]
}

// GetChain 获取处理链
func (kic *KnowledgeIndexingChain) GetChain() *compose.Chain[[]*schema.Document, []string] {
	return kic.chain
}

// Compile 编译链为可执行的Runnable
func (kic *KnowledgeIndexingChain) Compile(ctx context.Context) (compose.Runnable[[]*schema.Document, []string], error) {
	runnable, err := kic.chain.Compile(ctx)
	if err != nil {
		return nil, err
	}
	return runnable, nil
}

func newtest(ctx context.Context, input []*schema.Document, opts ...any) (output []*schema.Document, err error) {
	for _, doc := range input {
		fmt.Println(doc.ID, doc.Content)
		fmt.Println("-----------")
	}
	return input, nil
}

// NewKnowledgeIndexingChain 创建知识索引处理链
func NewKnowledgeIndexingChain(ctx context.Context, cfg *config.Config) (*KnowledgeIndexingChain, error) {

	transformer, err := NewDocumentTransformer(ctx)
	if err != nil {
		return nil, err
	}
	indexer, err := NewIndexer(ctx, cfg)
	if err != nil {
		return nil, err
	}
	// 创建处理链
	chain := compose.NewChain[[]*schema.Document, []string]()
	chain.AppendDocumentTransformer(transformer).
		AppendLambda(compose.InvokableLambdaWithOption(newtest), compose.WithNodeName("test")).
		AppendIndexer(indexer)

	return &KnowledgeIndexingChain{
		chain: chain,
	}, nil
}
