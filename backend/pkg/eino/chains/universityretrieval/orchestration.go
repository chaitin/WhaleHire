package universityretrieval

import (
	"context"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// UniversityRetrievalChain 高校标签召回处理链。
type UniversityRetrievalChain struct {
	chain *compose.Chain[string, []*schema.Document]
}

// GetChain 返回内部链实例，便于进一步组合。
func (c *UniversityRetrievalChain) GetChain() *compose.Chain[string, []*schema.Document] {
	return c.chain
}

// Compile 编译成可运行对象。
func (c *UniversityRetrievalChain) Compile(ctx context.Context) (compose.Runnable[string, []*schema.Document], error) {
	return c.chain.Compile(ctx)
}

// NewUniversityRetrievalChain 创建高校标签召回链。
func NewUniversityRetrievalChain(ctx context.Context, cfg *config.Config) (*UniversityRetrievalChain, error) {
	ret, err := NewRetriever(ctx, WithDataSource(cfg.Database.Master))
	if err != nil {
		return nil, err
	}

	normalizeLambda := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
		return normalizeQuery(input)
	})

	chain := compose.NewChain[string, []*schema.Document]()
	chain.AppendLambda(normalizeLambda, compose.WithNodeName("NormalizeUniversityQuery")).
		AppendRetriever(ret, compose.WithNodeName("PgvectorUniversityRetriever"))

	return &UniversityRetrievalChain{chain: chain}, nil
}
