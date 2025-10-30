package resumeparsergraph

import (
	"context"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/domain"
	chainresume "github.com/chaitin/WhaleHire/backend/pkg/eino/chains/resumeparser"
)

// Aggregator 将各字段增强结果合并回最终的解析数据。
type Aggregator struct{}

// NewAggregator 创建聚合器。
func NewAggregator() *Aggregator {
	return &Aggregator{}
}

// Process 聚合所有字段输出。
func (a *Aggregator) Process(ctx context.Context, data map[string]any) (*chainresume.ResumeParseResult, error) {
	if data == nil {
		return nil, fmt.Errorf("resume aggregator: 输入数据为空")
	}

	baseAny, ok := data[dispatchKeyBase]
	if !ok {
		return nil, fmt.Errorf("resume aggregator: 缺少基础解析结果")
	}

	base, ok := baseAny.(*domain.ParsedResumeData)
	if !ok || base == nil {
		return nil, fmt.Errorf("resume aggregator: 基础解析结果类型不匹配")
	}

	if eduAny, ok := data[dispatchKeyEducation]; ok && eduAny != nil {
		if eduRes, ok := eduAny.(*EducationEnrichmentResult); ok && eduRes != nil && len(eduRes.Educations) > 0 {
			base.Educations = eduRes.Educations
		}
	}

	return base, nil
}
