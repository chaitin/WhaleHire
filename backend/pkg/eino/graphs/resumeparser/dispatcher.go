package resumeparsergraph

import (
	"context"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/domain"
	chainresume "github.com/chaitin/WhaleHire/backend/pkg/eino/chains/resumeparser"
)

// Dispatcher 负责将简历基础解析结果拆分为各增强模块所需的输入。
type Dispatcher struct{}

// NewDispatcher 创建分发器。
func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

// Process 根据基础解析结果拆分输出。
func (d *Dispatcher) Process(ctx context.Context, result *chainresume.ResumeParseResult) (map[string]any, error) {
	if result == nil {
		return nil, fmt.Errorf("resume dispatcher: 基础解析结果为空")
	}

	data := make(map[string]any, 2)
	data[dispatchKeyBase] = result

	if len(result.Educations) > 0 {
		items := make([]*domain.ParsedEducation, 0, len(result.Educations))
		for _, edu := range result.Educations {
			if edu == nil {
				continue
			}

			clone := *edu
			clone.UniversityTags = nil
			clone.UniversityMatchScore = nil
			clone.UniversityMatchSource = ""
			items = append(items, &clone)
		}

		data[dispatchKeyEducation] = &EducationEnrichmentInput{
			Educations: items,
		}
	}

	return data, nil
}
