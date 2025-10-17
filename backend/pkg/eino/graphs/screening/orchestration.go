package screening

import (
	"context"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/aggregator"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/basicinfo"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/dispatcher"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/education"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/experience"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/industry"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/responsibility"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/skill"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/taskmeta"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
)

// 智能简历匹配图结构
type ScreeningChatGraph struct {
	graph               *compose.Graph[*domain.MatchInput, *domain.JobResumeMatch]
	version             string
	basicInfoAgent      *basicinfo.BasicInfoAgent
	educationAgent      *education.EducationAgent
	experienceAgent     *experience.ExperienceAgent
	industryAgent       *industry.IndustryAgent
	responsibilityAgent *responsibility.ResponsibilityAgent
	skillAgent          *skill.SkillAgent
	aggregatorAgent     *aggregator.AggregatorAgent
	taskMetaProcessor   *taskmeta.TaskMetaProcessor
	dispatcher          *dispatcher.Dispatcher
}

// GetGraph 获取图结构
func (r *ScreeningChatGraph) GetGraph() *compose.Graph[*domain.MatchInput, *domain.JobResumeMatch] {
	return r.graph
}

// Compile 编译链为可执行的Runnable
func (r *ScreeningChatGraph) Compile(ctx context.Context) (compose.Runnable[*domain.MatchInput, *domain.JobResumeMatch], error) {
	agent, err := r.graph.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile chain: %w", err)
	}
	return agent, nil
}

// GetVersion 返回Agent版本
func (r *ScreeningChatGraph) GetVersion() string {
	return r.version
}

// GetSubAgentVersions 返回各个子Agent的版本信息
func (r *ScreeningChatGraph) GetSubAgentVersions() map[string]string {
	versions := make(map[string]string)

	if r.basicInfoAgent != nil {
		versions[domain.BasicInfoAgent] = r.basicInfoAgent.GetVersion()
	}
	if r.educationAgent != nil {
		versions[domain.EducationAgent] = r.educationAgent.GetVersion()
	}
	if r.experienceAgent != nil {
		versions[domain.ExperienceAgent] = r.experienceAgent.GetVersion()
	}
	if r.industryAgent != nil {
		versions[domain.IndustryAgent] = r.industryAgent.GetVersion()
	}
	if r.responsibilityAgent != nil {
		versions[domain.ResponsibilityAgent] = r.responsibilityAgent.GetVersion()
	}
	if r.skillAgent != nil {
		versions[domain.SkillAgent] = r.skillAgent.GetVersion()
	}
	if r.aggregatorAgent != nil {
		versions[domain.AggregatorAgent] = r.aggregatorAgent.GetVersion()
	}
	if r.taskMetaProcessor != nil {
		versions[domain.TaskMetaDataNode] = r.taskMetaProcessor.GetVersion()
	}
	if r.dispatcher != nil {
		versions[domain.DispatcherNode] = r.dispatcher.GetVersion()
	}

	return versions
}

// GetAgentVersion 根据节点键返回对应Agent的版本
func (r *ScreeningChatGraph) GetAgentVersion(nodeKey string) string {
	switch nodeKey {
	case domain.BasicInfoAgent:
		if r.basicInfoAgent != nil {
			return r.basicInfoAgent.GetVersion()
		}
	case domain.EducationAgent:
		if r.educationAgent != nil {
			return r.educationAgent.GetVersion()
		}
	case domain.ExperienceAgent:
		if r.experienceAgent != nil {
			return r.experienceAgent.GetVersion()
		}
	case domain.IndustryAgent:
		if r.industryAgent != nil {
			return r.industryAgent.GetVersion()
		}
	case domain.ResponsibilityAgent:
		if r.responsibilityAgent != nil {
			return r.responsibilityAgent.GetVersion()
		}
	case domain.SkillAgent:
		if r.skillAgent != nil {
			return r.skillAgent.GetVersion()
		}
	case domain.AggregatorAgent:
		if r.aggregatorAgent != nil {
			return r.aggregatorAgent.GetVersion()
		}
	case domain.TaskMetaDataNode:
		if r.taskMetaProcessor != nil {
			return r.taskMetaProcessor.GetVersion()
		}
	case domain.DispatcherNode:
		if r.dispatcher != nil {
			return r.dispatcher.GetVersion()
		}
	}
	return ""
}

// NewScreeningChatGraph 使用配置创建智能简历匹配图
func NewScreeningChatGraph(ctx context.Context, chatModel model.ToolCallingChatModel, cfg *config.Config) (*ScreeningChatGraph, error) {

	g := compose.NewGraph[*domain.MatchInput, *domain.JobResumeMatch]()

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
	taskMetaProcessor := taskmeta.NewTaskMetaProcessor()
	if taskMetaProcessor == nil {
		return nil, fmt.Errorf("failed to create task meta processor")
	}

	// 添加调度节点
	_ = g.AddLambdaNode(domain.DispatcherNode, compose.InvokableLambda(dispatcher.Process), compose.WithNodeName(domain.DispatcherNode))

	// 批量添加各匹配代理节点
	_ = g.AddGraphNode(domain.BasicInfoAgent, baseinfoAgent.GetChain(),
		compose.WithInputKey(domain.BasicInfoAgent),
		compose.WithOutputKey(domain.BasicInfoAgent),
		compose.WithNodeName(domain.BasicInfoAgent))

	_ = g.AddGraphNode(domain.EducationAgent, educationAgent.GetChain(),
		compose.WithInputKey(domain.EducationAgent),
		compose.WithOutputKey(domain.EducationAgent),
		compose.WithNodeName(domain.EducationAgent))

	_ = g.AddGraphNode(domain.ExperienceAgent, experienceAgent.GetChain(),
		compose.WithInputKey(domain.ExperienceAgent),
		compose.WithOutputKey(domain.ExperienceAgent),
		compose.WithNodeName(domain.ExperienceAgent))

	_ = g.AddGraphNode(domain.IndustryAgent, industryAgent.GetChain(),
		compose.WithInputKey(domain.IndustryAgent),
		compose.WithOutputKey(domain.IndustryAgent),
		compose.WithNodeName(domain.IndustryAgent))

	_ = g.AddGraphNode(domain.ResponsibilityAgent, responseIndustryAgent.GetChain(),
		compose.WithInputKey(domain.ResponsibilityAgent),
		compose.WithOutputKey(domain.ResponsibilityAgent),
		compose.WithNodeName(domain.ResponsibilityAgent))

	_ = g.AddGraphNode(domain.SkillAgent, skillAgent.GetChain(),
		compose.WithInputKey(domain.SkillAgent),
		compose.WithOutputKey(domain.SkillAgent),
		compose.WithNodeName(domain.SkillAgent))

	_ = g.AddLambdaNode(domain.TaskMetaDataNode, compose.InvokableLambda(taskMetaProcessor.ProcessTaskMeta),
		compose.WithInputKey(domain.TaskMetaDataNode),
		compose.WithOutputKey(domain.TaskMetaDataNode),
		compose.WithNodeName(domain.TaskMetaDataNode))

	_ = g.AddGraphNode(domain.AggregatorAgent, aggregatorAgent.GetChain(),
		compose.WithNodeName(domain.AggregatorAgent))

	_ = g.AddEdge(compose.START, domain.DispatcherNode)
	_ = g.AddEdge(domain.DispatcherNode, domain.BasicInfoAgent)
	_ = g.AddEdge(domain.DispatcherNode, domain.EducationAgent)
	_ = g.AddEdge(domain.DispatcherNode, domain.ExperienceAgent)
	_ = g.AddEdge(domain.DispatcherNode, domain.IndustryAgent)
	_ = g.AddEdge(domain.DispatcherNode, domain.ResponsibilityAgent)
	_ = g.AddEdge(domain.DispatcherNode, domain.SkillAgent)
	_ = g.AddEdge(domain.DispatcherNode, domain.TaskMetaDataNode)

	_ = g.AddEdge(domain.BasicInfoAgent, domain.AggregatorAgent)
	_ = g.AddEdge(domain.EducationAgent, domain.AggregatorAgent)
	_ = g.AddEdge(domain.ExperienceAgent, domain.AggregatorAgent)
	_ = g.AddEdge(domain.IndustryAgent, domain.AggregatorAgent)
	_ = g.AddEdge(domain.ResponsibilityAgent, domain.AggregatorAgent)
	_ = g.AddEdge(domain.SkillAgent, domain.AggregatorAgent)
	_ = g.AddEdge(domain.TaskMetaDataNode, domain.AggregatorAgent)

	_ = g.AddEdge(domain.AggregatorAgent, compose.END)

	return &ScreeningChatGraph{
		graph:               g,
		version:             "1.0.0",
		basicInfoAgent:      baseinfoAgent,
		educationAgent:      educationAgent,
		experienceAgent:     experienceAgent,
		industryAgent:       industryAgent,
		responsibilityAgent: responseIndustryAgent,
		skillAgent:          skillAgent,
		aggregatorAgent:     aggregatorAgent,
		taskMetaProcessor:   taskMetaProcessor,
		dispatcher:          dispatcher,
	}, nil
}
