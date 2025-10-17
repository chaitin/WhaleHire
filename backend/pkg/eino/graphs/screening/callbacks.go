package screening

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
)

// AgentCallbackCollector 负责收集各 Agent 节点的输出结果
type AgentCallbackCollector struct {
	mu sync.RWMutex

	taskMetaInput            *domain.TaskMetaData
	taskMetaOutput           *domain.TaskMetaData
	taskMetaErr              error
	dispatcherInput          *domain.MatchInput
	dispatcherOutput         map[string]any
	dispatcherErr            error
	basicInfoInput           *domain.BasicInfoData
	basicInfo                *domain.BasicMatchDetail
	basicInfoErr             error
	educationInput           *domain.EducationData
	education                *domain.EducationMatchDetail
	educationErr             error
	experienceInput          *domain.ExperienceData
	experience               *domain.ExperienceMatchDetail
	experienceErr            error
	industryInput            *domain.IndustryData
	industry                 *domain.IndustryMatchDetail
	industryErr              error
	responsibilityInput      *domain.ResponsibilityData
	responsibility           *domain.ResponsibilityMatchDetail
	responsibilityErr        error
	skillInput               *domain.SkillData
	skill                    *domain.SkillMatchDetail
	skillErr                 error
	basicInfoTokenUsage      *model.TokenUsage
	educationTokenUsage      *model.TokenUsage
	experienceTokenUsage     *model.TokenUsage
	industryTokenUsage       *model.TokenUsage
	responsibilityTokenUsage *model.TokenUsage
	skillTokenUsage          *model.TokenUsage
	aggregatorTokenUsage     *model.TokenUsage
	aggregatorInput          map[string]any
	aggregatedMatch          *domain.JobResumeMatch
	aggregatorErr            error

	raw         map[string]any               // 输出汇总
	rawInputs   map[string]any               // 输入汇总
	rawErrors   map[string]error             // 错误汇总
	tokenUsages map[string]*model.TokenUsage // 令牌使用汇总
	// enableDetailedCallbacks 标记是否注入详细的节点回调，默认关闭避免干扰主流程
	enableDetailedCallbacks bool
}

// NewAgentCallbackCollector 创建 Agent 输出收集器
func NewAgentCallbackCollector() *AgentCallbackCollector {
	return &AgentCallbackCollector{
		raw:         make(map[string]any),
		rawInputs:   make(map[string]any),
		rawErrors:   make(map[string]error),
		tokenUsages: make(map[string]*model.TokenUsage),
	}
}

// Reset 清空已收集的输出结果
func (c *AgentCallbackCollector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.taskMetaInput = nil
	c.taskMetaOutput = nil
	c.taskMetaErr = nil
	c.dispatcherInput = nil
	c.dispatcherOutput = nil
	c.dispatcherErr = nil
	c.basicInfoInput = nil
	c.basicInfo = nil
	c.basicInfoErr = nil
	c.educationInput = nil
	c.education = nil
	c.educationErr = nil
	c.experienceInput = nil
	c.experience = nil
	c.experienceErr = nil
	c.industryInput = nil
	c.industry = nil
	c.industryErr = nil
	c.responsibilityInput = nil
	c.responsibility = nil
	c.responsibilityErr = nil
	c.skillInput = nil
	c.skill = nil
	c.skillErr = nil
	c.aggregatorInput = nil
	c.aggregatedMatch = nil
	c.aggregatorErr = nil
	c.basicInfoTokenUsage = nil
	c.educationTokenUsage = nil
	c.experienceTokenUsage = nil
	c.industryTokenUsage = nil
	c.responsibilityTokenUsage = nil
	c.skillTokenUsage = nil
	c.aggregatorTokenUsage = nil
	c.raw = make(map[string]any)
	c.rawInputs = make(map[string]any)
	c.rawErrors = make(map[string]error)
	c.tokenUsages = make(map[string]*model.TokenUsage)
}

// EnableDetailedCallbacks 启用详细回调，仅在示例调试中使用
func (c *AgentCallbackCollector) EnableDetailedCallbacks() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enableDetailedCallbacks = true
}

// detailedCallbacksEnabled 返回当前是否开启详细回调
func (c *AgentCallbackCollector) detailedCallbacksEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.enableDetailedCallbacks
}

// ComposeOptions 构建注入到 Compose 图中的回调选项
func (c *AgentCallbackCollector) ComposeOptions() []compose.Option {
	options := make([]compose.Option, 0, 16)

	if c.detailedCallbacksEnabled() {
		options = append(options,
			compose.WithCallbacks(c.newBasicInfoHandler()).DesignateNode(domain.BasicInfoAgent),
			compose.WithCallbacks(c.newEducationHandler()).DesignateNode(domain.EducationAgent),
			compose.WithCallbacks(c.newExperienceHandler()).DesignateNode(domain.ExperienceAgent),
			compose.WithCallbacks(c.newIndustryHandler()).DesignateNode(domain.IndustryAgent),
			compose.WithCallbacks(c.newResponsibilityHandler()).DesignateNode(domain.ResponsibilityAgent),
			compose.WithCallbacks(c.newSkillHandler()).DesignateNode(domain.SkillAgent),
			compose.WithCallbacks(c.newTaskMetaHandler()).DesignateNode(domain.TaskMetaDataNode),
			compose.WithCallbacks(c.newDispatcherHandler()).DesignateNode(domain.DispatcherNode),
			compose.WithCallbacks(c.newAggregatorHandler()).DesignateNode(domain.AggregatorAgent),
		)
	}

	options = append(options,
		// Model usage handlers
		compose.WithCallbacks(c.newModelUsageHandler(domain.BasicInfoAgent)).
			DesignateNodeWithPath(compose.NewNodePath(domain.BasicInfoAgent, "chat_model")),
		compose.WithCallbacks(c.newModelUsageHandler(domain.EducationAgent)).
			DesignateNodeWithPath(compose.NewNodePath(domain.EducationAgent, "chat_model")),
		compose.WithCallbacks(c.newModelUsageHandler(domain.ExperienceAgent)).
			DesignateNodeWithPath(compose.NewNodePath(domain.ExperienceAgent, "chat_model")),
		compose.WithCallbacks(c.newModelUsageHandler(domain.IndustryAgent)).
			DesignateNodeWithPath(compose.NewNodePath(domain.IndustryAgent, "chat_model")),
		compose.WithCallbacks(c.newModelUsageHandler(domain.ResponsibilityAgent)).
			DesignateNodeWithPath(compose.NewNodePath(domain.ResponsibilityAgent, "chat_model")),
		compose.WithCallbacks(c.newModelUsageHandler(domain.SkillAgent)).
			DesignateNodeWithPath(compose.NewNodePath(domain.SkillAgent, "chat_model")),
		compose.WithCallbacks(c.newModelUsageHandler(domain.AggregatorAgent)).
			DesignateNodeWithPath(compose.NewNodePath(domain.AggregatorAgent, "chat_model")),
	)

	return options
}

// BasicInfoOutput 返回基础信息匹配结果
func (c *AgentCallbackCollector) BasicInfoOutput() (*domain.BasicMatchDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.basicInfo == nil {
		return nil, false
	}
	return c.basicInfo, true
}

// BasicInfoInput 返回基础信息Agent输入
func (c *AgentCallbackCollector) BasicInfoInput() (*domain.BasicInfoData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.basicInfoInput == nil {
		return nil, false
	}
	return c.basicInfoInput, true
}

// BasicInfoError 返回基础信息Agent执行错误
func (c *AgentCallbackCollector) BasicInfoError() (error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.basicInfoErr == nil {
		return nil, false
	}
	return c.basicInfoErr, true
}

// EducationOutput 返回教育匹配结果
func (c *AgentCallbackCollector) EducationOutput() (*domain.EducationMatchDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.education == nil {
		return nil, false
	}
	return c.education, true
}

// EducationInput 返回教育Agent输入
func (c *AgentCallbackCollector) EducationInput() (*domain.EducationData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.educationInput == nil {
		return nil, false
	}
	return c.educationInput, true
}

// EducationError 返回教育Agent执行错误
func (c *AgentCallbackCollector) EducationError() (error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.educationErr == nil {
		return nil, false
	}
	return c.educationErr, true
}

// ExperienceOutput 返回经验匹配结果
func (c *AgentCallbackCollector) ExperienceOutput() (*domain.ExperienceMatchDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.experience == nil {
		return nil, false
	}
	return c.experience, true
}

// ExperienceInput 返回经验Agent输入
func (c *AgentCallbackCollector) ExperienceInput() (*domain.ExperienceData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.experienceInput == nil {
		return nil, false
	}
	return c.experienceInput, true
}

// ExperienceError 返回经验Agent执行错误
func (c *AgentCallbackCollector) ExperienceError() (error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.experienceErr == nil {
		return nil, false
	}
	return c.experienceErr, true
}

// IndustryOutput 返回行业匹配结果
func (c *AgentCallbackCollector) IndustryOutput() (*domain.IndustryMatchDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.industry == nil {
		return nil, false
	}
	return c.industry, true
}

// IndustryInput 返回行业Agent输入
func (c *AgentCallbackCollector) IndustryInput() (*domain.IndustryData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.industryInput == nil {
		return nil, false
	}
	return c.industryInput, true
}

// IndustryError 返回行业Agent执行错误
func (c *AgentCallbackCollector) IndustryError() (error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.industryErr == nil {
		return nil, false
	}
	return c.industryErr, true
}

// ResponsibilityOutput 返回职责匹配结果
func (c *AgentCallbackCollector) ResponsibilityOutput() (*domain.ResponsibilityMatchDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.responsibility == nil {
		return nil, false
	}
	return c.responsibility, true
}

// ResponsibilityInput 返回职责Agent输入
func (c *AgentCallbackCollector) ResponsibilityInput() (*domain.ResponsibilityData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.responsibilityInput == nil {
		return nil, false
	}
	return c.responsibilityInput, true
}

// ResponsibilityError 返回职责Agent执行错误
func (c *AgentCallbackCollector) ResponsibilityError() (error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.responsibilityErr == nil {
		return nil, false
	}
	return c.responsibilityErr, true
}

// SkillOutput 返回技能匹配结果
func (c *AgentCallbackCollector) SkillOutput() (*domain.SkillMatchDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.skill == nil {
		return nil, false
	}
	return c.skill, true
}

// SkillInput 返回技能Agent输入
func (c *AgentCallbackCollector) SkillInput() (*domain.SkillData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.skillInput == nil {
		return nil, false
	}
	return c.skillInput, true
}

// SkillError 返回技能Agent执行错误
func (c *AgentCallbackCollector) SkillError() (error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.skillErr == nil {
		return nil, false
	}
	return c.skillErr, true
}

// BasicInfoTokenUsage 返回基础信息 Agent 对大模型的 Token 消耗
func (c *AgentCallbackCollector) BasicInfoTokenUsage() (*model.TokenUsage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.basicInfoTokenUsage == nil {
		return nil, false
	}
	return cloneTokenUsage(c.basicInfoTokenUsage), true
}

// EducationTokenUsage 返回教育 Agent 对大模型的 Token 消耗
func (c *AgentCallbackCollector) EducationTokenUsage() (*model.TokenUsage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.educationTokenUsage == nil {
		return nil, false
	}
	return cloneTokenUsage(c.educationTokenUsage), true
}

// ExperienceTokenUsage 返回经验 Agent 对大模型的 Token 消耗
func (c *AgentCallbackCollector) ExperienceTokenUsage() (*model.TokenUsage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.experienceTokenUsage == nil {
		return nil, false
	}
	return cloneTokenUsage(c.experienceTokenUsage), true
}

// IndustryTokenUsage 返回行业 Agent 对大模型的 Token 消耗
func (c *AgentCallbackCollector) IndustryTokenUsage() (*model.TokenUsage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.industryTokenUsage == nil {
		return nil, false
	}
	return cloneTokenUsage(c.industryTokenUsage), true
}

// ResponsibilityTokenUsage 返回职责 Agent 对大模型的 Token 消耗
func (c *AgentCallbackCollector) ResponsibilityTokenUsage() (*model.TokenUsage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.responsibilityTokenUsage == nil {
		return nil, false
	}
	return cloneTokenUsage(c.responsibilityTokenUsage), true
}

// SkillTokenUsage 返回技能 Agent 对大模型的 Token 消耗
func (c *AgentCallbackCollector) SkillTokenUsage() (*model.TokenUsage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.skillTokenUsage == nil {
		return nil, false
	}
	return cloneTokenUsage(c.skillTokenUsage), true
}

// AggregatorTokenUsage 返回聚合 Agent 对大模型的 Token 消耗
func (c *AgentCallbackCollector) AggregatorTokenUsage() (*model.TokenUsage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.aggregatorTokenUsage == nil {
		return nil, false
	}
	return cloneTokenUsage(c.aggregatorTokenUsage), true
}

// AggregatedMatchOutput 返回聚合后的最终匹配结果
func (c *AgentCallbackCollector) AggregatedMatchOutput() (*domain.JobResumeMatch, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.aggregatedMatch == nil {
		return nil, false
	}
	return c.aggregatedMatch, true
}

// AggregatorInput 返回聚合节点输入副本
func (c *AgentCallbackCollector) AggregatorInput() (map[string]any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.aggregatorInput == nil {
		return nil, false
	}
	return cloneAnyMap(c.aggregatorInput), true
}

// AggregatorError 返回聚合节点执行错误
func (c *AgentCallbackCollector) AggregatorError() (error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.aggregatorErr == nil {
		return nil, false
	}
	return c.aggregatorErr, true
}

// Raw 获取所有原始输出的副本
func (c *AgentCallbackCollector) Raw() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	clone := make(map[string]any, len(c.raw))
	for k, v := range c.raw {
		clone[k] = v
	}
	return clone
}

// Inputs 获取所有 Agent 输入副本
func (c *AgentCallbackCollector) Inputs() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	clone := make(map[string]any, len(c.rawInputs))
	for k, v := range c.rawInputs {
		clone[k] = v
	}
	return clone
}

// Errors 获取所有 Agent 错误副本
func (c *AgentCallbackCollector) Errors() map[string]error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	clone := make(map[string]error, len(c.rawErrors))
	for k, v := range c.rawErrors {
		clone[k] = v
	}
	return clone
}

// TokenUsages 获取所有 Agent Token 消耗数据的副本
func (c *AgentCallbackCollector) TokenUsages() map[string]*model.TokenUsage {
	c.mu.RLock()
	defer c.mu.RUnlock()
	clone := make(map[string]*model.TokenUsage, len(c.tokenUsages))
	for k, v := range c.tokenUsages {
		if v == nil {
			clone[k] = nil
			continue
		}
		clone[k] = cloneTokenUsage(v)
	}
	return clone
}

// SerializeTokenUsages 序列化Token使用情况并计算总计统计
func (c *AgentCallbackCollector) SerializeTokenUsages() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.tokenUsages) == 0 {
		return nil
	}

	result := make(map[string]any)
	var totalPromptTokens, totalCompletionTokens, totalTokens int

	for key, usage := range c.tokenUsages {
		if usage == nil {
			continue
		}
		result[key] = map[string]any{
			"prompt_tokens":     usage.PromptTokens,
			"completion_tokens": usage.CompletionTokens,
			"total_tokens":      usage.TotalTokens,
		}

		// 累加各个Agent的Token使用量
		totalPromptTokens += usage.PromptTokens
		totalCompletionTokens += usage.CompletionTokens
		totalTokens += usage.TotalTokens
	}

	// 添加总计统计
	result["total"] = map[string]any{
		"prompt_tokens":     totalPromptTokens,
		"completion_tokens": totalCompletionTokens,
		"total_tokens":      totalTokens,
	}

	return result
}

// TaskMetaInput 返回任务元数据节点输入
func (c *AgentCallbackCollector) TaskMetaInput() (*domain.TaskMetaData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.taskMetaInput == nil {
		return nil, false
	}
	return c.taskMetaInput, true
}

// TaskMetaOutput 返回任务元数据节点输出
func (c *AgentCallbackCollector) TaskMetaOutput() (*domain.TaskMetaData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.taskMetaOutput == nil {
		return nil, false
	}
	return c.taskMetaOutput, true
}

// TaskMetaError 返回任务元数据节点错误
func (c *AgentCallbackCollector) TaskMetaError() (error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.taskMetaErr == nil {
		return nil, false
	}
	return c.taskMetaErr, true
}

// DispatcherInput 返回分发器节点输入
func (c *AgentCallbackCollector) DispatcherInput() (*domain.MatchInput, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.dispatcherInput == nil {
		return nil, false
	}
	return c.dispatcherInput, true
}

// DispatcherOutput 返回分发器节点输出
func (c *AgentCallbackCollector) DispatcherOutput() (map[string]any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.dispatcherOutput == nil {
		return nil, false
	}
	return cloneAnyMap(c.dispatcherOutput), true
}

// DispatcherError 返回分发器节点错误
func (c *AgentCallbackCollector) DispatcherError() (error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.dispatcherErr == nil {
		return nil, false
	}
	return c.dispatcherErr, true
}

func (c *AgentCallbackCollector) newBasicInfoHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if data, ok := input.(*domain.BasicInfoData); ok {
				c.recordInput(domain.BasicInfoAgent, data, func() { c.basicInfoInput = data })
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if info.Name != domain.BasicInfoAgent {
				return ctx
			}
			if detail, ok := output.(*domain.BasicMatchDetail); ok {
				c.record(domain.BasicInfoAgent, detail, func() { c.basicInfo = detail })
				printCallbackOutput(info.Name, detail)
			}
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			if err != nil {
				c.recordError(domain.BasicInfoAgent, err, func() { c.basicInfoErr = err })
			}
			return ctx
		}).Build()
}

func (c *AgentCallbackCollector) newEducationHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if data, ok := input.(*domain.EducationData); ok {
				c.recordInput(domain.EducationAgent, data, func() { c.educationInput = data })
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if info.Name != domain.EducationAgent {
				return ctx
			}
			if detail, ok := output.(*domain.EducationMatchDetail); ok {
				c.record(domain.EducationAgent, detail, func() { c.education = detail })
				printCallbackOutput(info.Name, detail)
			}
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			if err != nil {
				c.recordError(domain.EducationAgent, err, func() { c.educationErr = err })
			}
			return ctx
		}).Build()
}

func (c *AgentCallbackCollector) newExperienceHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if data, ok := input.(*domain.ExperienceData); ok {
				c.recordInput(domain.ExperienceAgent, data, func() { c.experienceInput = data })
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if info.Name != domain.ExperienceAgent {
				return ctx
			}
			if detail, ok := output.(*domain.ExperienceMatchDetail); ok {
				c.record(domain.ExperienceAgent, detail, func() { c.experience = detail })
				printCallbackOutput(info.Name, detail)
			}
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			if err != nil {
				c.recordError(domain.ExperienceAgent, err, func() { c.experienceErr = err })
			}
			return ctx
		}).Build()
}

func (c *AgentCallbackCollector) newIndustryHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if data, ok := input.(*domain.IndustryData); ok {
				c.recordInput(domain.IndustryAgent, data, func() { c.industryInput = data })
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if info.Name != domain.IndustryAgent {
				return ctx
			}
			if detail, ok := output.(*domain.IndustryMatchDetail); ok {
				c.record(domain.IndustryAgent, detail, func() { c.industry = detail })
				printCallbackOutput(info.Name, detail)
			}
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			if err != nil {
				c.recordError(domain.IndustryAgent, err, func() { c.industryErr = err })
			}
			return ctx
		}).Build()
}

func (c *AgentCallbackCollector) newResponsibilityHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if data, ok := input.(*domain.ResponsibilityData); ok {
				c.recordInput(domain.ResponsibilityAgent, data, func() { c.responsibilityInput = data })
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if info.Name != domain.ResponsibilityAgent {
				return ctx
			}
			if detail, ok := output.(*domain.ResponsibilityMatchDetail); ok {
				c.record(domain.ResponsibilityAgent, detail, func() { c.responsibility = detail })
				printCallbackOutput(info.Name, detail)
			}
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			if err != nil {
				c.recordError(domain.ResponsibilityAgent, err, func() { c.responsibilityErr = err })
			}
			return ctx
		}).Build()
}

func (c *AgentCallbackCollector) newSkillHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if data, ok := input.(*domain.SkillData); ok {
				c.recordInput(domain.SkillAgent, data, func() { c.skillInput = data })
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if info.Name != domain.SkillAgent {
				return ctx
			}
			if detail, ok := output.(*domain.SkillMatchDetail); ok {
				c.record(domain.SkillAgent, detail, func() { c.skill = detail })
				printCallbackOutput(info.Name, detail)
			}
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			if err != nil {
				c.recordError(domain.SkillAgent, err, func() { c.skillErr = err })
			}
			return ctx
		}).Build()
}

func (c *AgentCallbackCollector) newAggregatorHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if data, ok := input.(map[string]any); ok {
				cloned := cloneAnyMap(data)
				c.recordInput(domain.AggregatorAgent, cloned, func() { c.aggregatorInput = cloned })
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if info.Name != domain.AggregatorAgent {
				return ctx
			}
			if detail, ok := output.(*domain.JobResumeMatch); ok {
				c.record(domain.AggregatorAgent, detail, func() { c.aggregatedMatch = detail })
				printCallbackOutput(info.Name, detail)
			}
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			if err != nil {
				c.recordError(domain.AggregatorAgent, err, func() { c.aggregatorErr = err })
			}
			return ctx
		}).Build()
}

func (c *AgentCallbackCollector) newTaskMetaHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if data, ok := input.(*domain.TaskMetaData); ok {
				c.recordInput(domain.TaskMetaDataNode, data, func() { c.taskMetaInput = data })
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if info.Name != domain.TaskMetaDataNode {
				return ctx
			}
			if data, ok := output.(*domain.TaskMetaData); ok {
				c.record(domain.TaskMetaDataNode, data, func() { c.taskMetaOutput = data })
			}
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			if err != nil {
				c.recordError(domain.TaskMetaDataNode, err, func() { c.taskMetaErr = err })
			}
			return ctx
		}).Build()
}

func (c *AgentCallbackCollector) newDispatcherHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if data, ok := input.(*domain.MatchInput); ok {
				c.recordInput(domain.DispatcherNode, data, func() { c.dispatcherInput = data })
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if info.Name != domain.DispatcherNode {
				return ctx
			}
			if data, ok := output.(map[string]any); ok {
				c.record(domain.DispatcherNode, data, func() { c.dispatcherOutput = data })
				printCallbackOutput(info.Name, data)
			}
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			if err != nil {
				c.recordError(domain.DispatcherNode, err, func() { c.dispatcherErr = err })
			}
			return ctx
		}).Build()
}

func (c *AgentCallbackCollector) newModelUsageHandler(agentKey string) callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			callbackOutput := model.ConvCallbackOutput(output)
			if callbackOutput == nil || callbackOutput.TokenUsage == nil {
				return ctx
			}
			usage := cloneTokenUsage(callbackOutput.TokenUsage)
			var assign func()
			switch agentKey {
			case domain.BasicInfoAgent:
				assign = func() { c.basicInfoTokenUsage = usage }
			case domain.EducationAgent:
				assign = func() { c.educationTokenUsage = usage }
			case domain.ExperienceAgent:
				assign = func() { c.experienceTokenUsage = usage }
			case domain.IndustryAgent:
				assign = func() { c.industryTokenUsage = usage }
			case domain.ResponsibilityAgent:
				assign = func() { c.responsibilityTokenUsage = usage }
			case domain.SkillAgent:
				assign = func() { c.skillTokenUsage = usage }
			case domain.AggregatorAgent:
				assign = func() { c.aggregatorTokenUsage = usage }
			default:
				assign = func() {}
			}
			c.recordTokenUsage(agentKey, usage, assign)
			return ctx
		}).Build()
}

func (c *AgentCallbackCollector) record(key string, value any, assign func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	assign()
	c.raw[key] = value
}

func (c *AgentCallbackCollector) recordInput(key string, value any, assign func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	assign()
	c.rawInputs[key] = value
}

func (c *AgentCallbackCollector) recordError(key string, err error, assign func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	assign()
	c.rawErrors[key] = err
}

func (c *AgentCallbackCollector) recordTokenUsage(key string, usage *model.TokenUsage, assign func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	assign()
	c.tokenUsages[key] = usage
}

func cloneAnyMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneTokenUsage(src *model.TokenUsage) *model.TokenUsage {
	if src == nil {
		return nil
	}
	clone := *src
	return &clone
}

func printCallbackOutput(node string, value any) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fmt.Printf("[回调] 节点 %s 输出序列化失败，使用默认格式: %+v\n", node, value)
		return
	}
	fmt.Printf("[回调] 节点 %s 输出:\n%s\n", node, string(data))
}
