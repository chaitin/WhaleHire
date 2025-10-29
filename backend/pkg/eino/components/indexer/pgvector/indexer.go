package pgvector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
	_ "github.com/lib/pq"
)

// Row 表示一条待写入 pgvector 的记录，包含写库所需的向量与自定义载荷。
type Row struct {
	// Document 指向原始 schema.Document。若为空，索引器会在运行时补齐。
	Document *schema.Document
	// Vector 是最终写入数据库的向量；当缺失时索引器会尝试调用嵌入模型填充。
	Vector []float64
	// EmbeddingText 指定缺少向量时用于嵌入的文本。若为空则回退至 Document.Content。
	EmbeddingText string
	// Payload 为调用方自定义的附加数据，通常用于 InsertFunc 中构造 SQL 参数。
	Payload any
}

// DocumentToRow 将 schema.Document 转换为 Row，负责抽取业务字段与写库所需载荷。
type DocumentToRow func(ctx context.Context, doc *schema.Document) (*Row, error)

// InsertFunc 执行具体的写库逻辑，接收填充完向量的 Row 并返回最终写入的主键/标识。
type InsertFunc func(ctx context.Context, tx *sql.Tx, row *Row) (string, error)

// IndexerConfig 定义 pgvector 索引器的通用配置。
type IndexerConfig struct {
	// DB 允许直接注入已有连接池。
	DB *sql.DB
	// DataSource 在未提供 DB 时使用，用于初始化新的连接池。
	DataSource string
	// Embedding 嵌入组件；当 Row.Vector 为空时用于补全向量。
	Embedding embedding.Embedder
	// Dimension 可选向量维度校验；为 0 时跳过校验。
	Dimension int
	// DocumentToRow 将文档转换为 Row，必须提供。
	DocumentToRow DocumentToRow
	// InsertFunc 执行最终插入/更新操作，必须提供。
	InsertFunc InsertFunc
}

// Indexer 基于 pgvector 的通用索引器实现。
type Indexer struct {
	db        *sql.DB
	embedder  embedding.Embedder
	dimension int
	docToRow  DocumentToRow
	insertFn  InsertFunc
	ownsConn  bool
	closed    bool
}

// NewIndexer 根据配置初始化索引器并验证数据库连通性。
func NewIndexer(ctx context.Context, cfg *IndexerConfig) (*Indexer, error) {
	if cfg == nil {
		return nil, errors.New("pgvector indexer: 缺少配置")
	}
	if cfg.DocumentToRow == nil {
		return nil, errors.New("pgvector indexer: 未提供 DocumentToRow")
	}
	if cfg.InsertFunc == nil {
		return nil, errors.New("pgvector indexer: 未提供 InsertFunc")
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
			return nil, fmt.Errorf("pgvector indexer: 打开数据库失败: %w", err)
		}
		owns = true
	default:
		return nil, errors.New("pgvector indexer: 既未提供 DB 也未提供 DataSource")
	}

	if err = db.PingContext(ctx); err != nil {
		if owns {
			_ = db.Close()
		}
		return nil, fmt.Errorf("pgvector indexer: 数据库连通性检查失败: %w", err)
	}

	return &Indexer{
		db:        db,
		embedder:  cfg.Embedding,
		dimension: cfg.Dimension,
		docToRow:  cfg.DocumentToRow,
		insertFn:  cfg.InsertFunc,
		ownsConn:  owns,
	}, nil
}

// Close 当索引器内部创建数据库实例时需要显式关闭。
func (i *Indexer) Close() error {
	if !i.ownsConn || i.closed {
		return nil
	}
	i.closed = true
	return i.db.Close()
}

// Store 实现 indexer.Indexer 接口，将文档转换为 Row 并委托 InsertFunc 写库。
func (i *Indexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) (ids []string, err error) {
	if len(docs) == 0 {
		return nil, nil
	}

	options := indexer.GetCommonOptions(&indexer.Options{
		Embedding: i.embedder,
	}, opts...)

	ctx = callbacks.EnsureRunInfo(ctx, i.GetType(), components.ComponentOfIndexer)
	ctx = callbacks.OnStart(ctx, &indexer.CallbackInput{Docs: docs})
	defer func() {
		if err != nil {
			callbacks.OnError(ctx, err)
		}
	}()

	rows := make([]*Row, len(docs))
	var (
		embedTexts []string
		embedIndex []int
	)

	for idx := range docs {
		doc := docs[idx]
		row, convErr := i.docToRow(ctx, doc)
		if convErr != nil {
			err = convErr
			return nil, err
		}
		if row == nil {
			err = errors.New("pgvector indexer: DocumentToRow 返回空 Row")
			return nil, err
		}

		if row.Document == nil {
			row.Document = doc
		}

		if len(row.Vector) == 0 {
			if vector := doc.DenseVector(); len(vector) > 0 {
				row.Vector = vector
			}
		}

		if len(row.Vector) == 0 {
			text := strings.TrimSpace(row.EmbeddingText)
			if text == "" {
				text = strings.TrimSpace(row.Document.Content)
			}
			if text == "" {
				err = fmt.Errorf("pgvector indexer: 文档 %d 缺少向量与嵌入文本", idx)
				return nil, err
			}
			embedTexts = append(embedTexts, text)
			embedIndex = append(embedIndex, idx)
		}

		rows[idx] = row
	}

	if len(embedIndex) > 0 {
		embedder := options.Embedding
		if embedder == nil {
			err = errors.New("pgvector indexer: 缺少嵌入组件，无法生成向量")
			return nil, err
		}

		vectors, embedErr := embedder.EmbedStrings(ctx, embedTexts)
		if embedErr != nil {
			err = fmt.Errorf("pgvector indexer: 嵌入调用失败: %w", embedErr)
			return nil, err
		}
		if len(vectors) != len(embedIndex) {
			err = fmt.Errorf("pgvector indexer: 嵌入结果数量不一致，期望 %d 得到 %d", len(embedIndex), len(vectors))
			return nil, err
		}

		for idx, target := range embedIndex {
			vector := vectors[idx]
			if err = i.checkDimension(len(vector)); err != nil {
				return nil, err
			}
			rows[target].Vector = vector
		}
	}

	for _, row := range rows {
		if len(row.Vector) == 0 {
			err = errors.New("pgvector indexer: 仍有记录缺少向量")
			return nil, err
		}
		if err = i.checkDimension(len(row.Vector)); err != nil {
			return nil, err
		}
	}

	tx, txErr := i.db.BeginTx(ctx, nil)
	if txErr != nil {
		err = fmt.Errorf("pgvector indexer: 开启事务失败: %w", txErr)
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	ids = make([]string, 0, len(rows))
	for idx, row := range rows {
		insertedID, execErr := i.insertFn(ctx, tx, row)
		if execErr != nil {
			err = fmt.Errorf("pgvector indexer: 写入第 %d 条记录失败: %w", idx, execErr)
			return nil, err
		}

		if insertedID != "" {
			row.Document.ID = insertedID
			docs[idx].ID = insertedID
			ids = append(ids, insertedID)
			continue
		}

		switch {
		case row.Document.ID != "":
			ids = append(ids, row.Document.ID)
		case docs[idx].ID != "":
			ids = append(ids, docs[idx].ID)
		default:
			ids = append(ids, "")
		}
	}

	if commitErr := tx.Commit(); commitErr != nil {
		err = fmt.Errorf("pgvector indexer: 提交事务失败: %w", commitErr)
		return nil, err
	}
	committed = true

	callbacks.OnEnd(ctx, &indexer.CallbackOutput{IDs: ids})
	return ids, nil
}

// GetType 返回组件类型标识。
func (i *Indexer) GetType() string {
	return "PgvectorIndexer"
}

func (i *Indexer) checkDimension(length int) error {
	if i.dimension > 0 && length != i.dimension {
		return fmt.Errorf("pgvector indexer: 向量维度不匹配，期望 %d 实际 %d", i.dimension, length)
	}
	return nil
}
