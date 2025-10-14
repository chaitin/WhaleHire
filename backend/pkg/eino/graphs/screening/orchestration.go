package aggregator

import (
	"context"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/aggregator"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/basicinfo"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/dispatcher"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/education"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/experience"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/industry"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/responsibility"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/skill"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/taskmeta"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/types"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
)

// 智能简历匹配图结构
type ScreeningChatGraph struct {
	graph *compose.Graph[*types.MatchInput, *types.JobResumeMatch]
}

// GetGraph 获取图结构
func (r *ScreeningChatGraph) GetGraph() *compose.Graph[*types.MatchInput, *types.JobResumeMatch] {
	return r.graph
}

// Compile 编译链为可执行的Runnable
func (r *ScreeningChatGraph) Compile(ctx context.Context) (compose.Runnable[*types.MatchInput, *types.JobResumeMatch], error) {
	agent, err := r.graph.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile chain: %w", err)
	}
	return agent, nil
}

// NewScreeningChatGraph 使用配置创建智能简历匹配图
func NewScreeningChatGraph(ctx context.Context, chatModel model.ToolCallingChatModel, cfg *config.Config) (*ScreeningChatGraph, error) {

	g := compose.NewGraph[*types.MatchInput, *types.JobResumeMatch]()

	dispatcher := dispatcher.NewDispatcher()
	baseinfoAgent, err := basicinfo.NewBasicInfoAgent(ctx, chatModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create basic info agent: %w", err)
	}

	educationAgent, err := education.NewEducationAgent(ctx, chatModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create education agent: %w", err)
	}

	experienceAgent, err := experience.NewExperienceAgent(ctx, chatModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create experience agent: %w", err)
	}

	industryAgent, err := industry.NewIndustryAgent(ctx, chatModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create industry agent: %w", err)
	}

	responseIndustryAgent, err := responsibility.NewResponsibilityAgent(ctx, chatModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create industry agent: %w", err)
	}

	skillAgent, err := skill.NewSkillAgent(ctx, chatModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create skill agent: %w", err)
	}

	aggregatorAgent, err := aggregator.NewAggregatorAgent(ctx, chatModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create aggregator: %w", err)
	}

	// 添加调度节点
	_ = g.AddLambdaNode(types.DispatcherNode, compose.InvokableLambda(dispatcher.Process), compose.WithNodeName(types.DispatcherNode))

	// 批量添加各匹配代理节点
	_ = g.AddGraphNode(types.BasicInfoAgent, baseinfoAgent.GetChain(),
		compose.WithInputKey(types.BasicInfoAgent),
		compose.WithOutputKey(types.BasicInfoAgent),
		compose.WithNodeName(types.BasicInfoAgent))

	_ = g.AddGraphNode(types.EducationAgent, educationAgent.GetChain(),
		compose.WithInputKey(types.EducationAgent),
		compose.WithOutputKey(types.EducationAgent),
		compose.WithNodeName(types.EducationAgent))

	_ = g.AddGraphNode(types.ExperienceAgent, experienceAgent.GetChain(),
		compose.WithInputKey(types.ExperienceAgent),
		compose.WithOutputKey(types.ExperienceAgent),
		compose.WithNodeName(types.ExperienceAgent))

	_ = g.AddGraphNode(types.IndustryAgent, industryAgent.GetChain(),
		compose.WithInputKey(types.IndustryAgent),
		compose.WithOutputKey(types.IndustryAgent),
		compose.WithNodeName(types.IndustryAgent))

	_ = g.AddGraphNode(types.ResponsibilityAgent, responseIndustryAgent.GetChain(),
		compose.WithInputKey(types.ResponsibilityAgent),
		compose.WithOutputKey(types.ResponsibilityAgent),
		compose.WithNodeName(types.ResponsibilityAgent))

	_ = g.AddGraphNode(types.SkillAgent, skillAgent.GetChain(),
		compose.WithInputKey(types.SkillAgent),
		compose.WithOutputKey(types.SkillAgent),
		compose.WithNodeName(types.SkillAgent))

	_ = g.AddLambdaNode(types.TaskMetaDataNode, compose.InvokableLambda(taskmeta.NewLambdaTaskMeta),
		compose.WithInputKey(types.TaskMetaDataNode),
		compose.WithOutputKey(types.TaskMetaDataNode),
		compose.WithNodeName(types.TaskMetaDataNode))

	_ = g.AddGraphNode(types.AggregatorAgent, aggregatorAgent.GetChain(),
		compose.WithNodeName(types.AggregatorAgent))

	_ = g.AddEdge(compose.START, types.DispatcherNode)
	_ = g.AddEdge(types.DispatcherNode, types.BasicInfoAgent)
	_ = g.AddEdge(types.DispatcherNode, types.EducationAgent)
	_ = g.AddEdge(types.DispatcherNode, types.ExperienceAgent)
	_ = g.AddEdge(types.DispatcherNode, types.IndustryAgent)
	_ = g.AddEdge(types.DispatcherNode, types.ResponsibilityAgent)
	_ = g.AddEdge(types.DispatcherNode, types.SkillAgent)
	_ = g.AddEdge(types.DispatcherNode, types.TaskMetaDataNode)

	_ = g.AddEdge(types.BasicInfoAgent, types.AggregatorAgent)
	_ = g.AddEdge(types.EducationAgent, types.AggregatorAgent)
	_ = g.AddEdge(types.ExperienceAgent, types.AggregatorAgent)
	_ = g.AddEdge(types.IndustryAgent, types.AggregatorAgent)
	_ = g.AddEdge(types.ResponsibilityAgent, types.AggregatorAgent)
	_ = g.AddEdge(types.SkillAgent, types.AggregatorAgent)
	_ = g.AddEdge(types.TaskMetaDataNode, types.AggregatorAgent)

	_ = g.AddEdge(types.AggregatorAgent, compose.END)

	return &ScreeningChatGraph{
		graph: g,
	}, nil
}
