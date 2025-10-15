package screening

import (
	"context"
	"fmt"
	"sync"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
)

// AgentOutputCollector 负责收集各 Agent 节点的输出结果
type AgentOutputCollector struct {
	mu sync.RWMutex

	taskMetaInput            *domain.TaskMetaData
	taskMetaOutput           *domain.TaskMetaData
	taskMetaErr              error
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

	raw         map[string]any
	rawInputs   map[string]any
	rawErrors   map[string]error
	tokenUsages map[string]*model.TokenUsage
}

// NewAgentOutputCollector 创建 Agent 输出收集器
func NewAgentOutputCollector() *AgentOutputCollector {
	return &AgentOutputCollector{
		raw:         make(map[string]any),
		rawInputs:   make(map[string]any),
		rawErrors:   make(map[string]error),
		tokenUsages: make(map[string]*model.TokenUsage),
	}
}

// Reset 清空已收集的输出结果
func (c *AgentOutputCollector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.taskMetaInput = nil
	c.taskMetaOutput = nil
	c.taskMetaErr = nil
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

// ComposeOptions 构建注入到 Compose 图中的回调选项
func (c *AgentOutputCollector) ComposeOptions() []compose.Option {
	return []compose.Option{
		compose.WithCallbacks(c.newBasicInfoHandler()).DesignateNode(domain.BasicInfoAgent),
		compose.WithCallbacks(c.newEducationHandler()).DesignateNode(domain.EducationAgent),
		compose.WithCallbacks(c.newExperienceHandler()).DesignateNode(domain.ExperienceAgent),
		compose.WithCallbacks(c.newIndustryHandler()).DesignateNode(domain.IndustryAgent),
		compose.WithCallbacks(c.newResponsibilityHandler()).DesignateNode(domain.ResponsibilityAgent),
		compose.WithCallbacks(c.newSkillHandler()).DesignateNode(domain.SkillAgent),
		compose.WithCallbacks(c.newTaskMetaHandler()).DesignateNode(domain.TaskMetaDataNode),
		compose.WithCallbacks(c.newAggregatorHandler()).DesignateNode(domain.AggregatorAgent),

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
	}
}

// BasicInfo 返回基础信息匹配结果
func (c *AgentOutputCollector) BasicInfo() (*domain.BasicMatchDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.basicInfo == nil {
		return nil, false
	}
	return c.basicInfo, true
}

// BasicInfoInput 返回基础信息Agent输入
func (c *AgentOutputCollector) BasicInfoInput() (*domain.BasicInfoData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.basicInfoInput == nil {
		return nil, false
	}
	return c.basicInfoInput, true
}

// BasicInfoError 返回基础信息Agent执行错误
func (c *AgentOutputCollector) BasicInfoError() (error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.basicInfoErr == nil {
		return nil, false
	}
	return c.basicInfoErr, true
}

// Education 返回教育匹配结果
func (c *AgentOutputCollector) Education() (*domain.EducationMatchDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.education == nil {
		return nil, false
	}
	return c.education, true
}

// EducationInput 返回教育Agent输入
func (c *AgentOutputCollector) EducationInput() (*domain.EducationData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.educationInput == nil {
		return nil, false
	}
	return c.educationInput, true
}

// EducationError 返回教育Agent执行错误
func (c *AgentOutputCollector) EducationError() (error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.educationErr == nil {
		return nil, false
	}
	return c.educationErr, true
}

// Experience 返回经验匹配结果
func (c *AgentOutputCollector) Experience() (*domain.ExperienceMatchDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.experience == nil {
		return nil, false
	}
	return c.experience, true
}

// ExperienceInput 返回经验Agent输入
func (c *AgentOutputCollector) ExperienceInput() (*domain.ExperienceData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.experienceInput == nil {
		return nil, false
	}
	return c.experienceInput, true
}

// ExperienceError 返回经验Agent执行错误
func (c *AgentOutputCollector) ExperienceError() (error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.experienceErr == nil {
		return nil, false
	}
	return c.experienceErr, true
}

// Industry 返回行业匹配结果
func (c *AgentOutputCollector) Industry() (*domain.IndustryMatchDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.industry == nil {
		return nil, false
	}
	return c.industry, true
}

// IndustryInput 返回行业Agent输入
func (c *AgentOutputCollector) IndustryInput() (*domain.IndustryData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.industryInput == nil {
		return nil, false
	}
	return c.industryInput, true
}

// IndustryError 返回行业Agent执行错误
func (c *AgentOutputCollector) IndustryError() (error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.industryErr == nil {
		return nil, false
	}
	return c.industryErr, true
}

// Responsibility 返回职责匹配结果
func (c *AgentOutputCollector) Responsibility() (*domain.ResponsibilityMatchDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.responsibility == nil {
		return nil, false
	}
	return c.responsibility, true
}

// ResponsibilityInput 返回职责Agent输入
func (c *AgentOutputCollector) ResponsibilityInput() (*domain.ResponsibilityData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.responsibilityInput == nil {
		return nil, false
	}
	return c.responsibilityInput, true
}

// ResponsibilityError 返回职责Agent执行错误
func (c *AgentOutputCollector) ResponsibilityError() (error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.responsibilityErr == nil {
		return nil, false
	}
	return c.responsibilityErr, true
}

// Skill 返回技能匹配结果
func (c *AgentOutputCollector) Skill() (*domain.SkillMatchDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.skill == nil {
		return nil, false
	}
	return c.skill, true
}

// SkillInput 返回技能Agent输入
func (c *AgentOutputCollector) SkillInput() (*domain.SkillData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.skillInput == nil {
		return nil, false
	}
	return c.skillInput, true
}

// SkillError 返回技能Agent执行错误
func (c *AgentOutputCollector) SkillError() (error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.skillErr == nil {
		return nil, false
	}
	return c.skillErr, true
}

// BasicInfoTokenUsage 返回基础信息 Agent 对大模型的 Token 消耗
func (c *AgentOutputCollector) BasicInfoTokenUsage() (*model.TokenUsage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.basicInfoTokenUsage == nil {
		return nil, false
	}
	return cloneTokenUsage(c.basicInfoTokenUsage), true
}

// EducationTokenUsage 返回教育 Agent 对大模型的 Token 消耗
func (c *AgentOutputCollector) EducationTokenUsage() (*model.TokenUsage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.educationTokenUsage == nil {
		return nil, false
	}
	return cloneTokenUsage(c.educationTokenUsage), true
}

// ExperienceTokenUsage 返回经验 Agent 对大模型的 Token 消耗
func (c *AgentOutputCollector) ExperienceTokenUsage() (*model.TokenUsage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.experienceTokenUsage == nil {
		return nil, false
	}
	return cloneTokenUsage(c.experienceTokenUsage), true
}

// IndustryTokenUsage 返回行业 Agent 对大模型的 Token 消耗
func (c *AgentOutputCollector) IndustryTokenUsage() (*model.TokenUsage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.industryTokenUsage == nil {
		return nil, false
	}
	return cloneTokenUsage(c.industryTokenUsage), true
}

// ResponsibilityTokenUsage 返回职责 Agent 对大模型的 Token 消耗
func (c *AgentOutputCollector) ResponsibilityTokenUsage() (*model.TokenUsage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.responsibilityTokenUsage == nil {
		return nil, false
	}
	return cloneTokenUsage(c.responsibilityTokenUsage), true
}

// SkillTokenUsage 返回技能 Agent 对大模型的 Token 消耗
func (c *AgentOutputCollector) SkillTokenUsage() (*model.TokenUsage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.skillTokenUsage == nil {
		return nil, false
	}
	return cloneTokenUsage(c.skillTokenUsage), true
}

// AggregatorTokenUsage 返回聚合 Agent 对大模型的 Token 消耗
func (c *AgentOutputCollector) AggregatorTokenUsage() (*model.TokenUsage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.aggregatorTokenUsage == nil {
		return nil, false
	}
	return cloneTokenUsage(c.aggregatorTokenUsage), true
}

// AggregatedMatch 返回聚合后的最终匹配结果
func (c *AgentOutputCollector) AggregatedMatch() (*domain.JobResumeMatch, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.aggregatedMatch == nil {
		return nil, false
	}
	return c.aggregatedMatch, true
}

// AggregatorInput 返回聚合节点输入副本
func (c *AgentOutputCollector) AggregatorInput() (map[string]any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.aggregatorInput == nil {
		return nil, false
	}
	return cloneAnyMap(c.aggregatorInput), true
}

// AggregatorError 返回聚合节点执行错误
func (c *AgentOutputCollector) AggregatorError() (error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.aggregatorErr == nil {
		return nil, false
	}
	return c.aggregatorErr, true
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

// Inputs 获取所有 Agent 输入副本
func (c *AgentOutputCollector) Inputs() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	clone := make(map[string]any, len(c.rawInputs))
	for k, v := range c.rawInputs {
		clone[k] = v
	}
	return clone
}

// Errors 获取所有 Agent 错误副本
func (c *AgentOutputCollector) Errors() map[string]error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	clone := make(map[string]error, len(c.rawErrors))
	for k, v := range c.rawErrors {
		clone[k] = v
	}
	return clone
}

// TokenUsages 获取所有 Agent Token 消耗数据的副本
func (c *AgentOutputCollector) TokenUsages() map[string]*model.TokenUsage {
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

// TaskMetaInput 返回任务元数据节点输入
func (c *AgentOutputCollector) TaskMetaInput() (*domain.TaskMetaData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.taskMetaInput == nil {
		return nil, false
	}
	return c.taskMetaInput, true
}

// TaskMetaOutput 返回任务元数据节点输出
func (c *AgentOutputCollector) TaskMetaOutput() (*domain.TaskMetaData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.taskMetaOutput == nil {
		return nil, false
	}
	return c.taskMetaOutput, true
}

// TaskMetaError 返回任务元数据节点错误
func (c *AgentOutputCollector) TaskMetaError() (error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.taskMetaErr == nil {
		return nil, false
	}
	return c.taskMetaErr, true
}

func (c *AgentOutputCollector) newBasicInfoHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if data, ok := input.(*domain.BasicInfoData); ok {
				c.recordInput(domain.BasicInfoAgent, data, func() { c.basicInfoInput = data })
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if detail, ok := output.(*domain.BasicMatchDetail); ok {
				c.record(domain.BasicInfoAgent, detail, func() { c.basicInfo = detail })
				fmt.Printf("[回调] 节点 %s 输出: %+v\n", info.Name, detail)
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

func (c *AgentOutputCollector) newEducationHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if data, ok := input.(*domain.EducationData); ok {
				c.recordInput(domain.EducationAgent, data, func() { c.educationInput = data })
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if detail, ok := output.(*domain.EducationMatchDetail); ok {
				c.record(domain.EducationAgent, detail, func() { c.education = detail })
				fmt.Printf("[回调] 节点 %s 输出: %+v\n", info.Name, detail)
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

func (c *AgentOutputCollector) newExperienceHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if data, ok := input.(*domain.ExperienceData); ok {
				c.recordInput(domain.ExperienceAgent, data, func() { c.experienceInput = data })
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if detail, ok := output.(*domain.ExperienceMatchDetail); ok {
				c.record(domain.ExperienceAgent, detail, func() { c.experience = detail })
				fmt.Printf("[回调] 节点 %s 输出: %+v\n", info.Name, detail)
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

func (c *AgentOutputCollector) newIndustryHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if data, ok := input.(*domain.IndustryData); ok {
				c.recordInput(domain.IndustryAgent, data, func() { c.industryInput = data })
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if detail, ok := output.(*domain.IndustryMatchDetail); ok {
				c.record(domain.IndustryAgent, detail, func() { c.industry = detail })
				fmt.Printf("[回调] 节点 %s 输出: %+v\n", info.Name, detail)
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

func (c *AgentOutputCollector) newResponsibilityHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if data, ok := input.(*domain.ResponsibilityData); ok {
				c.recordInput(domain.ResponsibilityAgent, data, func() { c.responsibilityInput = data })
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if detail, ok := output.(*domain.ResponsibilityMatchDetail); ok {
				c.record(domain.ResponsibilityAgent, detail, func() { c.responsibility = detail })
				fmt.Printf("[回调] 节点 %s 输出: %+v\n", info.Name, detail)
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

func (c *AgentOutputCollector) newSkillHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if data, ok := input.(*domain.SkillData); ok {
				c.recordInput(domain.SkillAgent, data, func() { c.skillInput = data })
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if detail, ok := output.(*domain.SkillMatchDetail); ok {
				c.record(domain.SkillAgent, detail, func() { c.skill = detail })
				fmt.Printf("[回调] 节点 %s 输出: %+v\n", info.Name, detail)
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

func (c *AgentOutputCollector) newAggregatorHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if data, ok := input.(map[string]any); ok {
				cloned := cloneAnyMap(data)
				c.recordInput(domain.AggregatorAgent, cloned, func() { c.aggregatorInput = cloned })
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if detail, ok := output.(*domain.JobResumeMatch); ok {
				c.record(domain.AggregatorAgent, detail, func() { c.aggregatedMatch = detail })
				fmt.Printf("[回调] 节点 %s 输出: %+v\n", info.Name, detail)
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

func (c *AgentOutputCollector) newTaskMetaHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if data, ok := input.(*domain.TaskMetaData); ok {
				c.recordInput(domain.TaskMetaDataNode, data, func() { c.taskMetaInput = data })
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if meta, ok := output.(*domain.TaskMetaData); ok {
				c.record(domain.TaskMetaDataNode, meta, func() { c.taskMetaOutput = meta })
				fmt.Printf("[回调] 节点 %s 输出: %+v\n", info.Name, meta)
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

func (c *AgentOutputCollector) newModelUsageHandler(agentKey string) callbacks.Handler {
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

func (c *AgentOutputCollector) record(key string, value any, assign func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	assign()
	c.raw[key] = value
}

func (c *AgentOutputCollector) recordInput(key string, value any, assign func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	assign()
	c.rawInputs[key] = value
}

func (c *AgentOutputCollector) recordError(key string, err error, assign func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	assign()
	c.rawErrors[key] = err
}

func (c *AgentOutputCollector) recordTokenUsage(key string, usage *model.TokenUsage, assign func()) {
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
