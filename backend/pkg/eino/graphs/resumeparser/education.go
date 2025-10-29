package resumeparsergraph

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/chains/universityretrieval"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

// EducationEnrichmentInput 教育增强输入。
type EducationEnrichmentInput struct {
	Educations []*domain.ParsedEducation
}

// EducationEnrichmentResult 教育增强输出。
type EducationEnrichmentResult struct {
	Educations []*domain.ParsedEducation
}

type educationNode struct {
	retriever retriever.Retriever
	logger    *slog.Logger
}

const (
	metaKeyDoubleFirstClass = "is_double_first_class"
	metaKeyProject985       = "is_project_985"
	metaKeyProject211       = "is_project_211"
	metaKeyQsTop100         = "is_qs_top100"
)

var metaTagMappings = []struct {
	key string
	tag consts.UniversityType
}{
	{key: metaKeyDoubleFirstClass, tag: consts.UniversityTypeDoubleFirstClass},
	{key: metaKeyProject985, tag: consts.UniversityType985},
	{key: metaKeyProject211, tag: consts.UniversityType211},
	{key: metaKeyQsTop100, tag: consts.UniversityTypeQSTop100},
}

func newEducationNode(ctx context.Context, db *sql.DB, cfg *config.Config, logger *slog.Logger) (*educationNode, error) {
	// 创建嵌入模型
	embedder, _, err := embedding.NewEmbedding(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("education enrichment: 创建嵌入模型失败: %w", err)
	}

	// 创建向量检索器，传入数据库连接和嵌入模型
	ret, err := universityretrieval.NewRetriever(ctx,
		universityretrieval.WithDB(db),
		universityretrieval.WithEmbedding(embedder, cfg.Embedding.Dimension),
		universityretrieval.WithTopK(cfg.Retriever.TopK),
		universityretrieval.WithDistanceThreshold(cfg.Retriever.DistanceThreshold),
		universityretrieval.WithExactMatch(true),
	)
	if err != nil {
		return nil, fmt.Errorf("education enrichment: 创建高校检索器失败: %w", err)
	}

	nodeLogger := logger
	if nodeLogger == nil {
		nodeLogger = slog.Default()
	}

	return &educationNode{
		retriever: ret,
		logger:    nodeLogger.With("component", "resumeparser.education"),
	}, nil
}

func (n *educationNode) close() {
	if closer, ok := n.retriever.(interface{ Close() error }); ok {
		_ = closer.Close()
	}
}

func (n *educationNode) enrich(ctx context.Context, data *EducationEnrichmentInput) (*EducationEnrichmentResult, error) {
	if data == nil || len(data.Educations) == 0 {
		return &EducationEnrichmentResult{Educations: []*domain.ParsedEducation{}}, nil
	}

	if n.retriever == nil {
		return nil, errors.New("education enrichment: 检索器未初始化")
	}

	result := make([]*domain.ParsedEducation, 0, len(data.Educations))
	if n.logger != nil {
		n.logger.Debug("开始执行教育信息增强", slog.Int("count", len(data.Educations)))
	}
	cache := newQueryCache()

	for _, item := range data.Educations {
		if item == nil {
			continue
		}

		enriched := cloneEducation(item)
		query := strings.TrimSpace(enriched.School)
		if query == "" {
			applyManualFallback(enriched)
			if n.logger != nil {
				n.logger.Debug("学校名称为空，已跳过检索并回退默认标签")
			}
			result = append(result, enriched)
			continue
		}

		normalized, err := universityretrieval.NormalizeQuery(query)
		if err != nil {
			if n.logger != nil {
				n.logger.Warn("高校名称归一化失败，已回退默认标签", slog.String("school", query), slog.Any("error", err))
			}
			applyManualFallback(enriched)
			result = append(result, enriched)
			continue
		}

		docs, err := n.fetchDocsWithCache(ctx, cache, normalized)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil, err
			}
			if n.logger != nil {
				n.logger.Error("高校检索失败，已回退默认标签", slog.String("school", normalized), slog.Any("error", err))
			}
			applyManualFallback(enriched)
			result = append(result, enriched)
			continue
		}

		if len(docs) == 0 || docs[0] == nil {
			if n.logger != nil {
				n.logger.Info("高校检索未找到匹配结果，已回退默认标签", slog.String("school", normalized))
			}
			applyManualFallback(enriched)
			result = append(result, enriched)
			continue
		}

		fillEducationWithDoc(enriched, docs[0])
		if n.logger != nil {
			n.logger.Debug("高校匹配成功", slog.String("school", normalized), slog.Float64("score", docs[0].Score()))
		}
		result = append(result, enriched)
	}

	if n.logger != nil {
		n.logger.Debug("教育信息增强完成", slog.Int("result_count", len(result)))
	}
	return &EducationEnrichmentResult{
		Educations: result,
	}, nil
}

func (n *educationNode) Process(ctx context.Context, dispatchData map[string]any) (map[string]any, error) {
	if dispatchData == nil {
		return nil, fmt.Errorf("education enrichment: 分发数据为空")
	}

	raw, ok := dispatchData[dispatchKeyEducation]
	if !ok || raw == nil {
		return dispatchData, nil
	}

	input, ok := raw.(*EducationEnrichmentInput)
	if !ok {
		return dispatchData, nil
	}

	result, err := n.enrich(ctx, input)
	if err != nil {
		return nil, err
	}

	dispatchData[dispatchKeyEducation] = result
	return dispatchData, nil
}

func cloneEducation(src *domain.ParsedEducation) *domain.ParsedEducation {
	if src == nil {
		return nil
	}
	clone := *src
	if clone.UniversityMatchScore != nil {
		score := *clone.UniversityMatchScore
		clone.UniversityMatchScore = &score
	}
	if len(src.UniversityTags) > 0 {
		clone.UniversityTags = append([]consts.UniversityType(nil), src.UniversityTags...)
	} else {
		clone.UniversityTags = nil
	}
	return &clone
}

func fillEducationWithDoc(edu *domain.ParsedEducation, doc *schema.Document) {
	if edu == nil || doc == nil {
		return
	}

	meta := doc.MetaData

	tags := make([]consts.UniversityType, 0, len(metaTagMappings))
	for _, mapping := range metaTagMappings {
		if boolFromMeta(meta, mapping.key) {
			tags = append(tags, mapping.tag)
		}
	}
	if len(tags) == 0 {
		tags = []consts.UniversityType{consts.UniversityTypeOrdinary}
	}
	edu.UniversityTags = tags

	if score := doc.Score(); score > 0 {
		s := score
		edu.UniversityMatchScore = &s
	} else {
		edu.UniversityMatchScore = nil
	}

	// 根据匹配类型设置匹配来源
	if matchType, ok := meta["match_type"].(string); ok && matchType == "exact" {
		edu.UniversityMatchSource = consts.UniversityMatchSourceExact
	} else {
		edu.UniversityMatchSource = consts.UniversityMatchSourceVector
	}
}

func (n *educationNode) fetchDocsWithCache(ctx context.Context, cache *queryCache, normalized string) ([]*schema.Document, error) {
	if docs, found := cache.get(normalized); found {
		if n.logger != nil {
			n.logger.Debug("高校检索命中缓存", slog.String("school", normalized), slog.Int("doc_count", len(docs)))
		}
		return docs, nil
	}

	if n.logger != nil {
		n.logger.Debug("开始高校检索", slog.String("school", normalized))
	}

	docs, err := n.retriever.Retrieve(ctx, normalized)
	if err != nil {
		return nil, err
	}

	for _, doc := range docs {
		if doc == nil {
			continue
		}
		if doc.MetaData == nil {
			doc.MetaData = make(map[string]any)
		}
		if _, ok := doc.MetaData["match_type"]; !ok {
			doc.MetaData["match_type"] = "vector"
		}
	}

	cache.put(normalized, docs)

	if n.logger != nil {
		n.logger.Debug("高校检索完成", slog.String("school", normalized), slog.Int("doc_count", len(docs)))
	}

	return docs, nil
}

type queryCache struct {
	data map[string][]*schema.Document
	mu   sync.RWMutex
}

func newQueryCache() *queryCache {
	return &queryCache{
		data: make(map[string][]*schema.Document),
	}
}

func (c *queryCache) get(key string) ([]*schema.Document, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

func (c *queryCache) put(key string, docs []*schema.Document) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if docs == nil {
		return
	}
	cached := make([]*schema.Document, len(docs))
	copy(cached, docs)
	c.data[key] = cached
}

func applyManualFallback(edu *domain.ParsedEducation) {
	if edu == nil {
		return
	}
	edu.UniversityTags = []consts.UniversityType{consts.UniversityTypeOrdinary}
	edu.UniversityMatchSource = consts.UniversityMatchSourceManual
	edu.UniversityMatchScore = nil
}
