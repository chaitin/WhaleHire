package universityretrieval

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/chaitin/WhaleHire/backend/domain"
	pgretriever "github.com/chaitin/WhaleHire/backend/pkg/eino/components/retriever/pgvector"
	einoembedding "github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	pgvec "github.com/pgvector/pgvector-go"
)

const (
	searchUniversitySQL = `
SELECT
    id,
    name_cn,
    name_en,
    alias,
    country,
    is_double_first_class,
    is_project_985,
    is_project_211,
    is_qs_top100,
    rank_qs,
    overall_score,
    metadata,
    vector_content,
    created_at,
    updated_at,
    1 - (vector <=> $1) AS similarity
FROM university_profiles
WHERE deleted_at IS NULL
  AND vector IS NOT NULL
ORDER BY vector <=> $1
LIMIT $2
`

	searchUniversityWithThresholdSQL = `
SELECT
    id,
    name_cn,
    name_en,
    alias,
    country,
    is_double_first_class,
    is_project_985,
    is_project_211,
    is_qs_top100,
    rank_qs,
    overall_score,
    metadata,
    vector_content,
    created_at,
    updated_at,
    1 - (vector <=> $1) AS similarity
FROM university_profiles
WHERE deleted_at IS NULL
  AND vector IS NOT NULL
  AND (vector <=> $1) <= $3
ORDER BY vector <=> $1
LIMIT $2
`

	// 精确匹配查询SQL - 根据name_cn进行精确匹配
	searchUniversityByNameSQL = `
SELECT
    id,
    name_cn,
    name_en,
    alias,
    country,
    is_double_first_class,
    is_project_985,
    is_project_211,
    is_qs_top100,
    rank_qs,
    overall_score,
    metadata,
    vector_content,
    created_at,
    updated_at,
    1.0 AS similarity
FROM university_profiles
WHERE deleted_at IS NULL
  AND name_cn = $1
ORDER BY created_at DESC
LIMIT $2
`
)

// RetrieverConfig 包含构建 Retriever 所需的配置参数
type RetrieverConfig struct {
	DB                *sql.DB
	DataSource        string
	Embedder          einoembedding.Embedder
	Dimension         int
	TopK              int
	DistanceThreshold float64
	EnableExactMatch  bool // 是否启用精确匹配功能
}

// RetrieverOption 定义用于配置 Retriever 的选项函数类型
type RetrieverOption func(*RetrieverConfig)

// WithDB 设置数据库连接实例
func WithDB(db *sql.DB) RetrieverOption {
	return func(cfg *RetrieverConfig) {
		cfg.DB = db
	}
}

// WithDataSource 设置数据库连接字符串
func WithDataSource(dataSource string) RetrieverOption {
	return func(cfg *RetrieverConfig) {
		cfg.DataSource = dataSource
	}
}

// WithEmbedding 设置嵌入模型和维度
func WithEmbedding(embedder einoembedding.Embedder, dimension int) RetrieverOption {
	return func(cfg *RetrieverConfig) {
		cfg.Embedder = embedder
		cfg.Dimension = dimension
	}
}

// WithTopK 设置默认返回的文档数量
func WithTopK(topK int) RetrieverOption {
	return func(cfg *RetrieverConfig) {
		cfg.TopK = topK
	}
}

// WithDistanceThreshold 设置距离阈值
func WithDistanceThreshold(threshold float64) RetrieverOption {
	return func(cfg *RetrieverConfig) {
		cfg.DistanceThreshold = threshold
	}
}

// WithExactMatch 启用或禁用精确匹配功能
func WithExactMatch(enable bool) RetrieverOption {
	return func(cfg *RetrieverConfig) {
		cfg.EnableExactMatch = enable
	}
}

// NewRetriever 使用 Option 模式构建高校画像的通用向量召回器
func NewRetriever(ctx context.Context, options ...RetrieverOption) (retriever.Retriever, error) {
	// 设置默认配置
	cfg := &RetrieverConfig{
		TopK:              3,
		DistanceThreshold: math.NaN(),
	}

	// 应用所有选项
	for _, option := range options {
		option(cfg)
	}

	// 验证必需的配置
	if cfg.DB == nil && cfg.DataSource == "" {
		return nil, errors.New("university retriever: 未配置数据库连接或数据库实例")
	}

	if cfg.Embedder == nil {
		return nil, errors.New("university retriever: 未配置嵌入模型")
	}

	// 处理距离阈值
	var defaultThreshold = math.NaN()
	if cfg.DistanceThreshold > 0 && !math.IsNaN(cfg.DistanceThreshold) {
		if cfg.DistanceThreshold > 1 {
			cfg.DistanceThreshold = 1
		}
		defaultThreshold = cfg.DistanceThreshold
	}

	queryFunc := queryUniversityVectorRecords
	if cfg.EnableExactMatch {
		queryFunc = queryUniversityCombinedRecords
	}

	return pgretriever.NewRetriever(ctx, &pgretriever.RetrieverConfig{
		DB:                    cfg.DB,
		DataSource:            cfg.DataSource,
		Embedding:             cfg.Embedder,
		Dimension:             cfg.Dimension,
		DefaultTopK:           cfg.TopK,
		DefaultScoreThreshold: defaultThreshold,
		QueryFunc:             queryFunc,
	})
}

func queryUniversityVectorRecords(ctx context.Context, db *sql.DB, input *pgretriever.QueryInput) ([]*schema.Document, error) {
	if len(input.Vector) == 0 {
		return nil, errors.New("university retriever: 缺少查询向量")
	}
	topK := input.TopK
	if topK <= 0 {
		topK = 1
	}

	vector, err := toPGVector(input.Vector)
	if err != nil {
		return nil, err
	}

	var (
		rows *sql.Rows
	)

	if input.HasScoreThreshold {
		distanceLimit := 1 - input.ScoreThreshold
		if distanceLimit < 0 {
			distanceLimit = 0
		}
		rows, err = db.QueryContext(ctx, searchUniversityWithThresholdSQL, vector, topK, distanceLimit)
	} else {
		rows, err = db.QueryContext(ctx, searchUniversitySQL, vector, topK)
	}
	if err != nil {
		return nil, fmt.Errorf("university retriever: 执行检索 SQL 失败: %w", err)
	}
	defer rows.Close()

	results := make([]*schema.Document, 0, topK)
	for rows.Next() {
		doc, scanErr := scanUniversityRow(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		if doc.MetaData == nil {
			doc.MetaData = make(map[string]any)
		}
		doc.MetaData["retriever_query"] = input.Query
		if input.HasScoreThreshold && doc.Score() < input.ScoreThreshold {
			continue
		}
		results = append(results, doc)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("university retriever: 读取检索结果失败: %w", err)
	}

	for _, doc := range results {
		ensureMatchType(doc, "vector")
	}

	return results, nil
}

func queryUniversityCombinedRecords(ctx context.Context, db *sql.DB, input *pgretriever.QueryInput) ([]*schema.Document, error) {
	topK := input.TopK
	if topK <= 0 {
		topK = 1
	}

	results := make([]*schema.Document, 0, topK)
	seen := make(map[string]struct{})

	query := strings.TrimSpace(input.Query)
	if query != "" {
		exactDocs, err := QueryUniversityByName(ctx, db, query, topK)
		if err != nil {
			return nil, err
		}

		for _, doc := range exactDocs {
			if doc == nil {
				continue
			}
			ensureMatchType(doc, "exact")
			if doc.ID != "" {
				if _, ok := seen[doc.ID]; ok {
					continue
				}
				seen[doc.ID] = struct{}{}
			}
			results = append(results, doc)
			if len(results) >= topK {
				return results[:topK], nil
			}
		}
	}

	remaining := topK - len(results)
	if remaining <= 0 {
		return results, nil
	}

	vectorInput := *input
	vectorInput.TopK = remaining

	vectorDocs, err := queryUniversityVectorRecords(ctx, db, &vectorInput)
	if err != nil {
		return nil, err
	}

	for _, doc := range vectorDocs {
		if doc == nil {
			continue
		}
		ensureMatchType(doc, "vector")
		if doc.ID != "" {
			if _, ok := seen[doc.ID]; ok {
				continue
			}
			seen[doc.ID] = struct{}{}
		}
		results = append(results, doc)
		if len(results) >= topK {
			break
		}
	}

	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

func ensureMatchType(doc *schema.Document, matchType string) {
	if doc == nil {
		return
	}
	if doc.MetaData == nil {
		doc.MetaData = make(map[string]any)
	}
	if matchType != "" {
		doc.MetaData["match_type"] = matchType
	}
}

func scanUniversityRow(rows *sql.Rows) (*schema.Document, error) {
	var (
		id                 string
		nameCN             string
		nameEN             sql.NullString
		alias              sql.NullString
		country            sql.NullString
		isDoubleFirstClass bool
		isProject985       bool
		isProject211       bool
		isQSTop100         bool
		rankQS             sql.NullInt64
		overallScore       sql.NullFloat64
		rawMetadata        []byte
		vectorContent      sql.NullString
		createdAt          time.Time
		updatedAt          time.Time
		similarity         float64
	)

	if err := rows.Scan(
		&id,
		&nameCN,
		&nameEN,
		&alias,
		&country,
		&isDoubleFirstClass,
		&isProject985,
		&isProject211,
		&isQSTop100,
		&rankQS,
		&overallScore,
		&rawMetadata,
		&vectorContent,
		&createdAt,
		&updatedAt,
		&similarity,
	); err != nil {
		return nil, fmt.Errorf("university retriever: 解析结果行失败: %w", err)
	}

	meta := make(map[string]any)
	if len(rawMetadata) > 0 {
		if err := json.Unmarshal(rawMetadata, &meta); err != nil {
			return nil, fmt.Errorf("university retriever: 解析元数据失败: %w", err)
		}
	}
	if meta == nil {
		meta = make(map[string]any)
	}

	university := &domain.University{
		ID:                 id,
		NameCn:             nameCN,
		IsDoubleFirstClass: isDoubleFirstClass,
		IsProject985:       isProject985,
		IsProject211:       isProject211,
		IsQsTop100:         isQSTop100,
		VectorContent:      strings.TrimSpace(vectorContent.String),
		Metadata:           meta,
		CreatedAt:          createdAt.Unix(),
		UpdatedAt:          updatedAt.Unix(),
	}

	if nameEN.Valid {
		nameENVal := nameEN.String
		university.NameEn = &nameENVal
	}
	if alias.Valid {
		aliasVal := alias.String
		university.Alias = &aliasVal
	}
	if country.Valid {
		countryVal := country.String
		university.Country = &countryVal
	}
	if rankQS.Valid {
		rankQSVal := int(rankQS.Int64)
		university.RankQs = &rankQSVal
	}
	if overallScore.Valid {
		overallScoreVal := overallScore.Float64
		university.OverallScore = &overallScoreVal
	}

	doc, err := university.ToDocument()
	if err != nil {
		return nil, fmt.Errorf("university retriever: 生成文档失败: %w", err)
	}

	if strings.TrimSpace(university.VectorContent) != "" {
		doc.Content = university.VectorContent
	}

	if doc.MetaData == nil {
		doc.MetaData = make(map[string]any)
	}

	doc.MetaData["id"] = id
	doc.MetaData["similarity"] = similarity
	distance := 1 - similarity
	if distance < 0 {
		distance = 0
	}
	doc.MetaData["distance"] = distance
	doc.MetaData["created_at"] = createdAt.Unix()
	doc.MetaData["updated_at"] = updatedAt.Unix()
	if alias.Valid {
		doc.MetaData["alias"] = alias.String
	}
	if strings.TrimSpace(university.VectorContent) != "" {
		doc.MetaData["vector_content"] = university.VectorContent
	}

	doc.WithScore(similarity)
	return doc, nil
}

func toPGVector(values []float64) (pgvec.Vector, error) {
	if len(values) == 0 {
		return pgvec.Vector{}, errors.New("university retriever: 查询向量为空")
	}
	vec := make([]float32, len(values))
	for idx, val := range values {
		if math.IsNaN(val) || math.IsInf(val, 0) {
			return pgvec.Vector{}, fmt.Errorf("university retriever: 查询向量包含非法值: %f", val)
		}
		vec[idx] = float32(val)
	}
	return pgvec.NewVector(vec), nil
}

func normalizeQuery(input string) (string, error) {
	query := strings.TrimSpace(input)
	if query == "" {
		return "", errors.New("university retriever: 高校名称不能为空")
	}
	return query, nil
}

// NormalizeQuery 对外暴露的高校名称标准化工具，便于 CLI 或测试调用。
func NormalizeQuery(input string) (string, error) {
	return normalizeQuery(input)
}

// QueryUniversityByName 根据高校中文名称进行精确匹配查询
func QueryUniversityByName(ctx context.Context, db *sql.DB, nameCN string, topK int) ([]*schema.Document, error) {
	query, err := normalizeQuery(nameCN)
	if err != nil {
		return nil, err
	}

	if topK <= 0 {
		topK = 1
	}

	rows, err := db.QueryContext(ctx, searchUniversityByNameSQL, query, topK)
	if err != nil {
		return nil, fmt.Errorf("university retriever: 执行精确匹配查询失败: %w", err)
	}
	defer rows.Close()

	results := make([]*schema.Document, 0, topK)
	for rows.Next() {
		doc, scanErr := scanUniversityRow(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		if doc.MetaData == nil {
			doc.MetaData = make(map[string]any)
		}
		doc.MetaData["retriever_query"] = query
		doc.MetaData["query_type"] = "exact_match"
		doc.MetaData["match_type"] = "exact"
		results = append(results, doc)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("university retriever: 读取精确匹配结果失败: %w", err)
	}
	return results, nil
}
