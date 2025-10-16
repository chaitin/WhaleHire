package service

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/domain"
	screening "github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
)

// CallbackCollectorWrapper 包装AgentCallbackCollector，添加数据库存储功能
type CallbackCollectorWrapper struct {
	*screening.AgentCallbackCollector
	nodeRunRepo    domain.ScreeningNodeRunRepo
	screeningRepo  domain.ScreeningRepo
	screeningGraph *screening.ScreeningChatGraph
	taskID         uuid.UUID
	resumeID       uuid.UUID
	traceID        string
	logger         *slog.Logger
}

// NewCallbackCollectorWrapper 创建带数据库存储功能的CallbackCollector包装器
func NewCallbackCollectorWrapper(
	nodeRunRepo domain.ScreeningNodeRunRepo,
	screeningRepo domain.ScreeningRepo,
	screeningGraph *screening.ScreeningChatGraph,
	taskID, resumeID uuid.UUID,
	traceID string,
	logger *slog.Logger,
) *CallbackCollectorWrapper {
	return &CallbackCollectorWrapper{
		AgentCallbackCollector: screening.NewAgentCallbackCollector(),
		nodeRunRepo:            nodeRunRepo,
		screeningRepo:          screeningRepo,
		screeningGraph:         screeningGraph,
		taskID:                 taskID,
		resumeID:               resumeID,
		traceID:                traceID,
		logger:                 logger,
	}
}

// ComposeOptions 返回包含数据库存储功能的compose选项
func (w *CallbackCollectorWrapper) ComposeOptions() []compose.Option {
	// 获取原始的compose选项
	originalOptions := w.AgentCallbackCollector.ComposeOptions()

	// 添加数据库存储的回调处理器
	dbOptions := []compose.Option{
		compose.WithCallbacks(
			w.newDatabaseCallbackHandler(domain.TaskMetaDataNode),
			w.newDatabaseCallbackHandler(domain.DispatcherNode),
			w.newDatabaseCallbackHandler(domain.BasicInfoAgent),
			w.newDatabaseCallbackHandler(domain.EducationAgent),
			w.newDatabaseCallbackHandler(domain.ExperienceAgent),
			w.newDatabaseCallbackHandler(domain.IndustryAgent),
			w.newDatabaseCallbackHandler(domain.ResponsibilityAgent),
			w.newDatabaseCallbackHandler(domain.SkillAgent),
			w.newDatabaseCallbackHandler(domain.AggregatorAgent),
		),
	}

	// 合并选项
	return append(originalOptions, dbOptions...)
}

// newDatabaseCallbackHandler 创建数据库存储的回调处理器
func (w *CallbackCollectorWrapper) newDatabaseCallbackHandler(nodeKey string) callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if info.Name == nodeKey {
				// 根据节点类型进行具体的数据类型检查，只保存正确的数据结构
				var inputData any
				switch nodeKey {
				case domain.TaskMetaDataNode:
					if data, ok := input.(*domain.TaskMetaData); ok {
						inputData = data
					}
				case domain.DispatcherNode:
					if data, ok := input.(*domain.MatchInput); ok {
						inputData = data
					}
				case domain.BasicInfoAgent:
					if data, ok := input.(*domain.BasicInfoData); ok {
						inputData = data
					}
				case domain.EducationAgent:
					if data, ok := input.(*domain.EducationData); ok {
						inputData = data
					}
				case domain.ExperienceAgent:
					if data, ok := input.(*domain.ExperienceData); ok {
						inputData = data
					}
				case domain.IndustryAgent:
					if data, ok := input.(*domain.IndustryData); ok {
						inputData = data
					}
				case domain.ResponsibilityAgent:
					if data, ok := input.(*domain.ResponsibilityData); ok {
						inputData = data
					}
				case domain.SkillAgent:
					if data, ok := input.(*domain.SkillData); ok {
						inputData = data
					}
				case domain.AggregatorAgent:
					if data, ok := input.(map[string]any); ok {
						inputData = data
					}
				}

				// 只有当输入数据类型正确时才保存到数据库
				if inputData != nil {
					w.saveNodeRunToDB(ctx, nodeKey, inputData, nil, nil, nil, consts.ScreeningNodeRunStatusRunning)
				}
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if info.Name == nodeKey {
				// 获取对应的输出数据和token使用情况
				var outputData any
				var tokenUsage *model.TokenUsage

				switch nodeKey {
				case domain.TaskMetaDataNode:
					// TaskMetaDataNode: 任务元数据处理节点
					// 输出: domain.TaskMetaData 结构，包含任务ID、简历ID、匹配任务ID和维度权重配置
					// 用途: 为后续Agent节点提供任务上下文信息和权重配置
					if result, ok := w.TaskMetaOutput(); ok {
						outputData = result
					}
					// TaskMetaDataNode 通常不消耗token，因为它是数据处理节点
				case domain.DispatcherNode:
					// DispatcherNode: 数据分发节点
					// 输出: map[string]any 分发数据，将输入数据按Agent类型分发到对应的处理节点
					// 用途: 为各个专业Agent提供结构化的输入数据
					if result, ok := w.DispatcherOutput(); ok {
						outputData = result
					}
				case domain.BasicInfoAgent:
					// BasicInfoAgent: 基本信息匹配Agent
					// 输出: domain.BasicMatchDetail 基本信息匹配详情
					// 包含: 地点匹配、薪资匹配、部门匹配等基础信息的匹配分析结果
					// 权重: 3% (DefaultDimensionWeights.Basic)
					if result, ok := w.BasicInfoOutput(); ok {
						outputData = result
					}
					if usage, ok := w.BasicInfoTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.EducationAgent:
					// EducationAgent: 教育背景匹配Agent
					// 输出: domain.EducationMatchDetail 教育背景匹配详情
					// 包含: 学历匹配、专业匹配、学校匹配等教育背景的匹配分析结果
					// 权重: 15% (DefaultDimensionWeights.Education)
					if result, ok := w.EducationOutput(); ok {
						outputData = result
					}
					if usage, ok := w.EducationTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.ExperienceAgent:
					// ExperienceAgent: 工作经验匹配Agent
					// 输出: domain.ExperienceMatchDetail 工作经验匹配详情
					// 包含: 工作年限匹配、职位匹配、行业经验匹配等工作经验的匹配分析结果
					// 权重: 20% (DefaultDimensionWeights.Experience)
					if result, ok := w.ExperienceOutput(); ok {
						outputData = result
					}
					if usage, ok := w.ExperienceTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.IndustryAgent:
					// IndustryAgent: 行业背景匹配Agent
					// 输出: domain.IndustryMatchDetail 行业背景匹配详情
					// 包含: 行业匹配、公司匹配等行业背景的匹配分析结果
					// 权重: 7% (DefaultDimensionWeights.Industry)
					if result, ok := w.IndustryOutput(); ok {
						outputData = result
					}
					if usage, ok := w.IndustryTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.ResponsibilityAgent:
					// ResponsibilityAgent: 职责匹配Agent
					// 输出: domain.ResponsibilityMatchDetail 职责匹配详情
					// 包含: 岗位职责与工作经历的匹配分析、相关经验识别等
					// 权重: 20% (DefaultDimensionWeights.Responsibility)
					if result, ok := w.ResponsibilityOutput(); ok {
						outputData = result
					}
					if usage, ok := w.ResponsibilityTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.SkillAgent:
					// SkillAgent: 技能匹配Agent
					// 输出: domain.SkillMatchDetail 技能匹配详情
					// 包含: 匹配技能列表、缺失技能列表、额外技能、LLM分析结果等
					// 权重: 35% (DefaultDimensionWeights.Skill) - 权重最高的匹配维度
					if result, ok := w.SkillOutput(); ok {
						outputData = result
					}
					if usage, ok := w.SkillTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.AggregatorAgent:
					// AggregatorAgent: 结果聚合Agent
					// 输出: domain.JobResumeMatch 综合匹配结果
					// 包含: 各维度匹配详情、综合评分、匹配建议等最终匹配结果
					// 用途: 将各个专业Agent的输出结果进行加权聚合，生成最终的匹配报告
					if result, ok := w.AggregatedMatchOutput(); ok {
						outputData = result
					}
					if usage, ok := w.AggregatorTokenUsage(); ok {
						tokenUsage = usage
					}
				}

				w.saveNodeRunToDB(ctx, nodeKey, nil, outputData, nil, tokenUsage, consts.ScreeningNodeRunStatusCompleted)
			}
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			if info.Name == nodeKey {
				// 失败时记录错误信息和token使用情况，但不记录输入数据
				var tokenUsage *model.TokenUsage

				switch nodeKey {
				case domain.BasicInfoAgent:
					if usage, ok := w.BasicInfoTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.EducationAgent:
					if usage, ok := w.EducationTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.ExperienceAgent:
					if usage, ok := w.ExperienceTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.IndustryAgent:
					if usage, ok := w.IndustryTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.ResponsibilityAgent:
					if usage, ok := w.ResponsibilityTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.SkillAgent:
					if usage, ok := w.SkillTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.AggregatorAgent:
					if usage, ok := w.AggregatorTokenUsage(); ok {
						tokenUsage = usage
					}
					// TaskMetaDataNode 和 DispatcherNode 通常不消耗token，无需处理
				}

				w.saveNodeRunToDB(ctx, nodeKey, nil, nil, err, tokenUsage, consts.ScreeningNodeRunStatusFailed)
			}
			return ctx
		}).Build()
}

// saveNodeRunToDB 保存节点运行数据到数据库
func (w *CallbackCollectorWrapper) saveNodeRunToDB(
	ctx context.Context,
	nodeKey string,
	input, output any,
	err error,
	tokenUsage *model.TokenUsage,
	status consts.ScreeningNodeRunStatus,
) {
	if w.nodeRunRepo == nil {
		return // 如果没有配置数据库存储，则跳过
	}

	// 序列化输入和输出数据
	var inputPayload, outputPayload map[string]interface{}
	if input != nil {
		if data, jsonErr := json.Marshal(input); jsonErr == nil {
			if unmarshalErr := json.Unmarshal(data, &inputPayload); unmarshalErr != nil {
				w.logger.Warn("Failed to unmarshal input payload", "error", unmarshalErr)
			}
		}
	}
	if output != nil {
		if data, jsonErr := json.Marshal(output); jsonErr == nil {
			if unmarshalErr := json.Unmarshal(data, &outputPayload); unmarshalErr != nil {
				w.logger.Warn("Failed to unmarshal output payload", "error", unmarshalErr)
			}
		}
	}

	// 获取对应Agent的版本号
	var agentVersion *string
	if w.screeningGraph != nil {
		subAgentVersions := w.screeningGraph.GetSubAgentVersions()
		if version, exists := subAgentVersions[nodeKey]; exists && version != "" {
			agentVersion = &version
		}
	}

	// 获取TaskResumeID
	taskResumeID, getTaskErr := w.getTaskResumeID(ctx)
	if getTaskErr != nil {
		w.logger.Error("获取TaskResumeID失败",
			slog.String("nodeKey", nodeKey),
			slog.String("taskID", w.taskID.String()),
			slog.String("resumeID", w.resumeID.String()),
			slog.Any("error", getTaskErr),
		)
		return
	}

	// 构建创建请求 - 使用正确的TaskResumeID
	req := &domain.CreateNodeRunRepoReq{
		TaskID:       w.taskID,
		TaskResumeID: taskResumeID,
		NodeKey:      nodeKey,
		Status:       status,
		AttemptNo:    1, // 默认尝试次数为1
		TraceID:      &w.traceID,
		AgentVersion: agentVersion,
	}

	if inputPayload != nil {
		req.InputPayload = inputPayload
	}
	if outputPayload != nil {
		req.OutputPayload = outputPayload
	}

	// 异步保存到数据库
	go func() {
		if _, saveErr := w.nodeRunRepo.Create(context.Background(), req); saveErr != nil {
			w.logger.Error("保存节点运行数据失败",
				slog.String("nodeKey", nodeKey),
				slog.String("taskID", w.taskID.String()),
				slog.String("resumeID", w.resumeID.String()),
				slog.String("taskResumeID", taskResumeID.String()),
				slog.String("traceID", w.traceID),
				slog.Any("agentVersion", agentVersion),
				slog.Any("error", saveErr),
			)
		} else {
			w.logger.Debug("节点运行数据保存成功",
				slog.String("nodeKey", nodeKey),
				slog.String("status", string(status)),
				slog.String("taskID", w.taskID.String()),
				slog.String("resumeID", w.resumeID.String()),
				slog.String("taskResumeID", taskResumeID.String()),
				slog.Any("agentVersion", agentVersion),
			)
		}
	}()
}

// getTaskResumeID 根据TaskID和ResumeID获取TaskResumeID
func (w *CallbackCollectorWrapper) getTaskResumeID(ctx context.Context) (uuid.UUID, error) {
	taskResume, err := w.screeningRepo.GetScreeningTaskResume(ctx, w.taskID, w.resumeID)
	if err != nil {
		return uuid.Nil, err
	}
	return taskResume.ID, nil
}
