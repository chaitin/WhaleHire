package aggregator

import (
	"context"
	"sync"

	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/types"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/compose"
)

// AgentOutputCollector 负责收集各 Agent 节点的输出结果
type AgentOutputCollector struct {
	mu sync.RWMutex

	basicInfo       *types.BasicMatchDetail
	education       *types.EducationMatchDetail
	experience      *types.ExperienceMatchDetail
	industry        *types.IndustryMatchDetail
	responsibility  *types.ResponsibilityMatchDetail
	skill           *types.SkillMatchDetail
	aggregatedMatch *types.JobResumeMatch

	raw map[string]any
}

// NewAgentOutputCollector 创建 Agent 输出收集器
func NewAgentOutputCollector() *AgentOutputCollector {
	return &AgentOutputCollector{
		raw: make(map[string]any),
	}
}

// Reset 清空已收集的输出结果
func (c *AgentOutputCollector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.basicInfo = nil
	c.education = nil
	c.experience = nil
	c.industry = nil
	c.responsibility = nil
	c.skill = nil
	c.aggregatedMatch = nil
	c.raw = make(map[string]any)
}

// ComposeOptions 构建注入到 Compose 图中的回调选项
func (c *AgentOutputCollector) ComposeOptions() []compose.Option {
	return []compose.Option{
		compose.WithCallbacks(c.newBasicInfoHandler()).DesignateNode(types.BasicInfoAgent),
		compose.WithCallbacks(c.newEducationHandler()).DesignateNode(types.EducationAgent),
		compose.WithCallbacks(c.newExperienceHandler()).DesignateNode(types.ExperienceAgent),
		compose.WithCallbacks(c.newIndustryHandler()).DesignateNode(types.IndustryAgent),
		compose.WithCallbacks(c.newResponsibilityHandler()).DesignateNode(types.ResponsibilityAgent),
		compose.WithCallbacks(c.newSkillHandler()).DesignateNode(types.SkillAgent),
		compose.WithCallbacks(c.newAggregatorHandler()).DesignateNode(types.AggregatorAgent),
	}
}

// BasicInfo 返回基础信息匹配结果
func (c *AgentOutputCollector) BasicInfo() (*types.BasicMatchDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.basicInfo == nil {
		return nil, false
	}
	return c.basicInfo, true
}

// Education 返回教育匹配结果
func (c *AgentOutputCollector) Education() (*types.EducationMatchDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.education == nil {
		return nil, false
	}
	return c.education, true
}

// Experience 返回经验匹配结果
func (c *AgentOutputCollector) Experience() (*types.ExperienceMatchDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.experience == nil {
		return nil, false
	}
	return c.experience, true
}

// Industry 返回行业匹配结果
func (c *AgentOutputCollector) Industry() (*types.IndustryMatchDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.industry == nil {
		return nil, false
	}
	return c.industry, true
}

// Responsibility 返回职责匹配结果
func (c *AgentOutputCollector) Responsibility() (*types.ResponsibilityMatchDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.responsibility == nil {
		return nil, false
	}
	return c.responsibility, true
}

// Skill 返回技能匹配结果
func (c *AgentOutputCollector) Skill() (*types.SkillMatchDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.skill == nil {
		return nil, false
	}
	return c.skill, true
}

// AggregatedMatch 返回聚合后的最终匹配结果
func (c *AgentOutputCollector) AggregatedMatch() (*types.JobResumeMatch, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.aggregatedMatch == nil {
		return nil, false
	}
	return c.aggregatedMatch, true
}

// Raw 获取所有原始输出的副本
func (c *AgentOutputCollector) Raw() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	clone := make(map[string]any, len(c.raw))
	for k, v := range c.raw {
		clone[k] = v
	}
	return clone
}

// InvokeWithCollector 结合输出收集器执行图，便于一次性注入回调选项
func InvokeWithCollector(
	ctx context.Context,
	runnable compose.Runnable[*types.MatchInput, *types.JobResumeMatch],
	input *types.MatchInput,
	collector *AgentOutputCollector,
	extra ...compose.Option,
) (*types.JobResumeMatch, error) {
	options := make([]compose.Option, 0, len(extra))
	options = append(options, extra...)
	if collector != nil {
		options = append(options, collector.ComposeOptions()...)
	}
	return runnable.Invoke(ctx, input, options...)
}

func (c *AgentOutputCollector) newBasicInfoHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if detail, ok := output.(*types.BasicMatchDetail); ok {
				c.record(types.BasicInfoAgent, detail, func() { c.basicInfo = detail })
			}
			return ctx
		}).Build()
}

func (c *AgentOutputCollector) newEducationHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if detail, ok := output.(*types.EducationMatchDetail); ok {
				c.record(types.EducationAgent, detail, func() { c.education = detail })
			}
			return ctx
		}).Build()
}

func (c *AgentOutputCollector) newExperienceHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if detail, ok := output.(*types.ExperienceMatchDetail); ok {
				c.record(types.ExperienceAgent, detail, func() { c.experience = detail })
			}
			return ctx
		}).Build()
}

func (c *AgentOutputCollector) newIndustryHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if detail, ok := output.(*types.IndustryMatchDetail); ok {
				c.record(types.IndustryAgent, detail, func() { c.industry = detail })
			}
			return ctx
		}).Build()
}

func (c *AgentOutputCollector) newResponsibilityHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if detail, ok := output.(*types.ResponsibilityMatchDetail); ok {
				c.record(types.ResponsibilityAgent, detail, func() { c.responsibility = detail })
			}
			return ctx
		}).Build()
}

func (c *AgentOutputCollector) newSkillHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if detail, ok := output.(*types.SkillMatchDetail); ok {
				c.record(types.SkillAgent, detail, func() { c.skill = detail })
			}
			return ctx
		}).Build()
}

func (c *AgentOutputCollector) newAggregatorHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if detail, ok := output.(*types.JobResumeMatch); ok {
				c.record(types.AggregatorAgent, detail, func() { c.aggregatedMatch = detail })
			}
			return ctx
		}).Build()
}

func (c *AgentOutputCollector) record(key string, value any, assign func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	assign()
	c.raw[key] = value
}
