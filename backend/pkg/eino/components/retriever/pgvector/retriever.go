package pgvector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

// Options 定义 pgvector 召回器的实现特有选项。
type Options struct {
	// Vector 允许直接传入查询向量，当不为空时将跳过内部的嵌入调用。
	Vector []float64
}

// WithVector 指定检索使用的向量。
func WithVector(vector []float64) retriever.Option {
	copied := make([]float64, len(vector))
	copy(copied, vector)
	return retriever.WrapImplSpecificOptFn(func(opts *Options) {
		opts.Vector = copied
	})
}

// QueryInput 描述执行检索时可用的全部上下文信息。
type QueryInput struct {
	// Query 原始查询文本。
	Query string
	// Vector 检索所用向量，若由外部提供则不会再触发嵌入。
	Vector []float64
	// TopK 本次检索的返回上限。
	TopK int
	// ScoreThreshold 相似度阈值，仅当 HasScoreThreshold 为 true 时有效。
	ScoreThreshold float64
	// HasScoreThreshold 标记是否启用阈值过滤。
	HasScoreThreshold bool
	// CommonOptions 为 Eino 通用选项，包含 index/subIndex/Embedding 等信息。
	CommonOptions *retriever.Options
	// ImplOptions 为 pgvector 自定义选项，目前仅包含外部传入的向量。
	ImplOptions *Options
}

// QueryFunc 负责执行实际的数据库检索逻辑。
type QueryFunc func(ctx context.Context, db *sql.DB, input *QueryInput) ([]*schema.Document, error)

// RetrieverConfig 定义 pgvector 召回器的配置。
type RetrieverConfig struct {
	// DB 允许直接注入已有连接。
	DB *sql.DB
	// DataSource 在未提供 DB 时使用，用于初始化新的连接池。
	DataSource string
	// Embedding 当未传入向量时用于生成查询向量。
	Embedding embedding.Embedder
	// Dimension 可选向量维度校验，<=0 表示跳过。
	Dimension int
	// DefaultTopK TopK 默认值，<=0 时内部会回退为 3。
	DefaultTopK int
	// DefaultScoreThreshold 默认相似度阈值；若未启用请置为 math.NaN() 或调用者显式设置。
	DefaultScoreThreshold float64
	// QueryFunc 执行最终检索，必须提供。
	QueryFunc QueryFunc
}

// Retriever 基于 pgvector 的通用召回器实现。
type Retriever struct {
	db                  *sql.DB
	embedder            embedding.Embedder
	dimension           int
	defaultTopK         int
	defaultThreshold    float64
	hasDefaultThreshold bool
	queryFn             QueryFunc
	ownsConn            bool
	closed              bool
}

// NewRetriever 根据配置初始化召回器。
func NewRetriever(ctx context.Context, cfg *RetrieverConfig) (*Retriever, error) {
	if cfg == nil {
		return nil, errors.New("pgvector retriever: 缺少配置")
	}
	if cfg.QueryFunc == nil {
		return nil, errors.New("pgvector retriever: 未提供 QueryFunc")
	}

	var (
		db   *sql.DB
		err  error
		owns bool
	)

	switch {
	case cfg.DB != nil:
		db = cfg.DB
	case cfg.DataSource != "":
		db, err = sql.Open("postgres", cfg.DataSource)
		if err != nil {
			return nil, fmt.Errorf("pgvector retriever: 打开数据库失败: %w", err)
		}
		owns = true
	default:
		return nil, errors.New("pgvector retriever: 既未提供 DB 也未提供 DataSource")
	}

	if err = db.PingContext(ctx); err != nil {
		if owns {
			_ = db.Close()
		}
		return nil, fmt.Errorf("pgvector retriever: 数据库连通性检查失败: %w", err)
	}

	topK := cfg.DefaultTopK
	if topK <= 0 {
		topK = 3
	}

	ret := &Retriever{
		db:          db,
		embedder:    cfg.Embedding,
		dimension:   cfg.Dimension,
		defaultTopK: topK,
		queryFn:     cfg.QueryFunc,
		ownsConn:    owns,
	}

	if !math.IsNaN(cfg.DefaultScoreThreshold) {
		ret.defaultThreshold = cfg.DefaultScoreThreshold
		ret.hasDefaultThreshold = true
	}

	return ret, nil
}

// Close 关闭由组件内部创建的连接池。
func (r *Retriever) Close() error {
	if !r.ownsConn || r.closed {
		return nil
	}
	r.closed = true
	return r.db.Close()
}

// Retrieve 实现 retriever.Retriever 接口，返回与查询向量最相似的文档。
func (r *Retriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	query = strings.TrimSpace(query)

	commonBase := &retriever.Options{
		Embedding: r.embedder,
	}
	if r.defaultTopK > 0 {
		commonBase.TopK = &r.defaultTopK
	}
	if r.hasDefaultThreshold {
		commonBase.ScoreThreshold = &r.defaultThreshold
	}

	commonOpts := retriever.GetCommonOptions(commonBase, opts...)
	implOpts := retriever.GetImplSpecificOptions(&Options{}, opts...)

	topK := r.defaultTopK
	if commonOpts.TopK != nil && *commonOpts.TopK > 0 {
		topK = *commonOpts.TopK
	}
	if topK <= 0 {
		topK = 1
	}

	var (
		scoreThreshold float64
		useThreshold   bool
	)
	if commonOpts.ScoreThreshold != nil {
		scoreThreshold = *commonOpts.ScoreThreshold
		useThreshold = true
	}

	callbackInput := &retriever.CallbackInput{
		Query: query,
		TopK:  topK,
	}
	if useThreshold {
		callbackInput.ScoreThreshold = &scoreThreshold
	}
	ctx = callbacks.OnStart(ctx, callbackInput)

	vector := implOpts.Vector
	if len(vector) == 0 {
		if query == "" {
			err := errors.New("pgvector retriever: 查询文本与向量不可同时为空")
			callbacks.OnError(ctx, err)
			return nil, err
		}

		embedder := commonOpts.Embedding
		if embedder == nil {
			err := errors.New("pgvector retriever: 缺少嵌入组件，无法生成查询向量")
			callbacks.OnError(ctx, err)
			return nil, err
		}

		res, err := embedder.EmbedStrings(ctx, []string{query})
		if err != nil {
			wrapped := fmt.Errorf("pgvector retriever: 嵌入调用失败: %w", err)
			callbacks.OnError(ctx, wrapped)
			return nil, wrapped
		}
		if len(res) != 1 {
			err := fmt.Errorf("pgvector retriever: 嵌入结果数量不一致，期望 1 实际 %d", len(res))
			callbacks.OnError(ctx, err)
			return nil, err
		}
		vector = res[0]
	}

	if err := r.checkDimension(len(vector)); err != nil {
		callbacks.OnError(ctx, err)
		return nil, err
	}

	input := &QueryInput{
		Query:             query,
		Vector:            vector,
		TopK:              topK,
		ScoreThreshold:    scoreThreshold,
		HasScoreThreshold: useThreshold,
		CommonOptions:     commonOpts,
		ImplOptions:       implOpts,
	}

	docs, err := r.queryFn(ctx, r.db, input)
	if err != nil {
		wrapped := fmt.Errorf("pgvector retriever: 检索失败: %w", err)
		callbacks.OnError(ctx, wrapped)
		return nil, wrapped
	}

	callbacks.OnEnd(ctx, &retriever.CallbackOutput{Docs: docs})
	return docs, nil
}

// GetType 返回组件类型标识。
func (r *Retriever) GetType() string {
	return "PgvectorRetriever"
}

func (r *Retriever) checkDimension(length int) error {
	if r.dimension > 0 && length != r.dimension {
		return fmt.Errorf("pgvector retriever: 向量维度不匹配，期望 %d 实际 %d", r.dimension, length)
	}
	return nil
}
