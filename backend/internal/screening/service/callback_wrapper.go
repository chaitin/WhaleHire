package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
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
	collector *screening.AgentCallbackCollector,
	nodeRunRepo domain.ScreeningNodeRunRepo,
	screeningRepo domain.ScreeningRepo,
	screeningGraph *screening.ScreeningChatGraph,
	taskID, resumeID uuid.UUID,
	traceID string,
	logger *slog.Logger,
) *CallbackCollectorWrapper {
	if collector == nil {
		collector = screening.NewAgentCallbackCollector()
	}
	return &CallbackCollectorWrapper{
		AgentCallbackCollector: collector,
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
	// 获取原始采集器的回调选项，确保令牌用量等统计逻辑正常注册
	collectorOptions := w.AgentCallbackCollector.ComposeOptions()

	// 添加数据库存储的回调处理器
	nodeKeys := []string{
		domain.TaskMetaDataNode,
		domain.DispatcherNode,
		domain.BasicInfoAgent,
		domain.EducationAgent,
		domain.ExperienceAgent,
		domain.IndustryAgent,
		domain.ResponsibilityAgent,
		domain.SkillAgent,
		domain.AggregatorAgent,
	}

	dbOptions := make([]compose.Option, 0, len(nodeKeys))
	for _, nodeKey := range nodeKeys {
		dbOptions = append(dbOptions,
			compose.WithCallbacks(w.newDatabaseCallbackHandler(nodeKey)).DesignateNode(nodeKey),
		)
	}

	// 合并选项
	return append(collectorOptions, dbOptions...)
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
				var outputData any
				var tokenUsage *model.TokenUsage

				switch nodeKey {
				case domain.TaskMetaDataNode:
					if result, ok := output.(*domain.TaskMetaData); ok {
						outputData = result
					} else {
						w.logger.Warn("TaskMetaDataNode输出类型断言失败",
							slog.String("nodeKey", nodeKey),
							slog.String("actualType", fmt.Sprintf("%T", output)),
						)
					}
				case domain.DispatcherNode:
					if result, ok := output.(map[string]any); ok {
						outputData = result
					} else {
						w.logger.Warn("DispatcherNode输出类型断言失败",
							slog.String("nodeKey", nodeKey),
							slog.String("actualType", fmt.Sprintf("%T", output)),
						)
					}
				case domain.BasicInfoAgent:
					if result, ok := output.(*domain.BasicMatchDetail); ok {
						outputData = result
					} else {
						w.logger.Warn("BasicInfoAgent输出类型断言失败",
							slog.String("nodeKey", nodeKey),
							slog.String("actualType", fmt.Sprintf("%T", output)),
						)
					}
					if usage, ok := w.AgentCallbackCollector.BasicInfoTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.EducationAgent:
					if result, ok := output.(*domain.EducationMatchDetail); ok {
						outputData = result
					} else {
						w.logger.Warn("EducationAgent输出类型断言失败",
							slog.String("nodeKey", nodeKey),
							slog.String("actualType", fmt.Sprintf("%T", output)),
						)
					}
					if usage, ok := w.AgentCallbackCollector.EducationTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.ExperienceAgent:
					if result, ok := output.(*domain.ExperienceMatchDetail); ok {
						outputData = result
					} else {
						w.logger.Warn("ExperienceAgent输出类型断言失败",
							slog.String("nodeKey", nodeKey),
							slog.String("actualType", fmt.Sprintf("%T", output)),
						)
					}
					if usage, ok := w.AgentCallbackCollector.ExperienceTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.IndustryAgent:
					if result, ok := output.(*domain.IndustryMatchDetail); ok {
						outputData = result
					} else {
						w.logger.Warn("IndustryAgent输出类型断言失败",
							slog.String("nodeKey", nodeKey),
							slog.String("actualType", fmt.Sprintf("%T", output)),
						)
					}
					if usage, ok := w.AgentCallbackCollector.IndustryTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.ResponsibilityAgent:
					if result, ok := output.(*domain.ResponsibilityMatchDetail); ok {
						outputData = result
					} else {
						w.logger.Warn("ResponsibilityAgent输出类型断言失败",
							slog.String("nodeKey", nodeKey),
							slog.String("actualType", fmt.Sprintf("%T", output)),
						)
					}
					if usage, ok := w.AgentCallbackCollector.ResponsibilityTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.SkillAgent:
					if result, ok := output.(*domain.SkillMatchDetail); ok {
						outputData = result
					} else {
						w.logger.Warn("SkillAgent输出类型断言失败",
							slog.String("nodeKey", nodeKey),
							slog.String("actualType", fmt.Sprintf("%T", output)),
						)
					}
					if usage, ok := w.AgentCallbackCollector.SkillTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.AggregatorAgent:
					if result, ok := output.(*domain.JobResumeMatch); ok {
						outputData = result
					} else {
						w.logger.Warn("AggregatorAgent输出类型断言失败",
							slog.String("nodeKey", nodeKey),
							slog.String("actualType", fmt.Sprintf("%T", output)),
						)
					}
					if usage, ok := w.AgentCallbackCollector.AggregatorTokenUsage(); ok {
						tokenUsage = usage
					}
				}

				w.updateNodeRunInDB(ctx, nodeKey, outputData, nil, tokenUsage, consts.ScreeningNodeRunStatusCompleted)
			}
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			if info.Name == nodeKey {
				// 失败时记录错误信息和token使用情况，但不记录输入数据
				var tokenUsage *model.TokenUsage

				switch nodeKey {
				case domain.BasicInfoAgent:
					if usage, ok := w.AgentCallbackCollector.BasicInfoTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.EducationAgent:
					if usage, ok := w.AgentCallbackCollector.EducationTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.ExperienceAgent:
					if usage, ok := w.AgentCallbackCollector.ExperienceTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.IndustryAgent:
					if usage, ok := w.AgentCallbackCollector.IndustryTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.ResponsibilityAgent:
					if usage, ok := w.AgentCallbackCollector.ResponsibilityTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.SkillAgent:
					if usage, ok := w.AgentCallbackCollector.SkillTokenUsage(); ok {
						tokenUsage = usage
					}
				case domain.AggregatorAgent:
					if usage, ok := w.AgentCallbackCollector.AggregatorTokenUsage(); ok {
						tokenUsage = usage
					}
					// TaskMetaDataNode 和 DispatcherNode 通常不消耗token，无需处理
				}

				go w.updateNodeRunInDB(ctx, nodeKey, nil, err, tokenUsage, consts.ScreeningNodeRunStatusFailed)
			}
			return ctx
		}).Build()
}

// updateNodeRunInDB 更新现有的节点运行记录
func (w *CallbackCollectorWrapper) updateNodeRunInDB(
	ctx context.Context,
	nodeKey string,
	output any,
	err error,
	tokenUsage *model.TokenUsage,
	status consts.ScreeningNodeRunStatus,
) {
	taskResumeID, getErr := w.getTaskResumeID(ctx)
	if getErr != nil {
		w.logger.Error("获取任务简历ID失败",
			slog.String("nodeKey", nodeKey),
			slog.String("traceID", w.traceID),
			slog.Any("error", getErr),
		)
		return
	}

	// 查找现有记录
	existingRecord, findErr := w.nodeRunRepo.GetByTaskResumeAndNode(ctx, taskResumeID, nodeKey, 1)
	if findErr != nil {
		w.logger.Error("查找现有节点运行记录失败",
			slog.String("nodeKey", nodeKey),
			slog.String("taskResumeID", taskResumeID.String()),
			slog.String("traceID", w.traceID),
			slog.Any("error", findErr),
		)
		return
	}

	// 准备更新数据
	var outputPayload map[string]interface{}
	if output != nil {
		outputBytes, marshalErr := json.Marshal(output)
		if marshalErr != nil {
			w.logger.Warn("序列化输出数据失败",
				slog.String("nodeKey", nodeKey),
				slog.String("traceID", w.traceID),
				slog.Any("error", marshalErr),
			)
		} else {
			if unmarshalErr := json.Unmarshal(outputBytes, &outputPayload); unmarshalErr != nil {
				w.logger.Warn("反序列化输出数据失败",
					slog.String("nodeKey", nodeKey),
					slog.String("traceID", w.traceID),
					slog.Any("error", unmarshalErr),
				)
			}
		}
	}

	var errorMessage *string
	if err != nil {
		errMsg := err.Error()
		errorMessage = &errMsg
	}

	// 获取 Agent 版本信息
	var agentVersion *string
	if w.screeningGraph != nil {
		version := w.screeningGraph.GetAgentVersion(nodeKey)
		if version != "" {
			agentVersion = &version
		}
	}

	// 使用 nodeRunRepo 的 Update 方法更新记录
	_, updateErr := w.nodeRunRepo.Update(ctx, existingRecord.ID, func(tx *db.Tx, current *db.ScreeningNodeRun, updater *db.ScreeningNodeRunUpdateOne) error {
		updater.SetStatus(string(status))

		if outputPayload != nil {
			updater.SetOutputPayload(outputPayload)
		}

		if errorMessage != nil {
			updater.SetErrorMessage(*errorMessage)
		}

		if agentVersion != nil {
			updater.SetAgentVersion(*agentVersion)
		}

		if tokenUsage != nil {
			if tokenUsage.PromptTokens > 0 {
				updater.SetTokensInput(int64(tokenUsage.PromptTokens))
			}
			if tokenUsage.CompletionTokens > 0 {
				updater.SetTokensOutput(int64(tokenUsage.CompletionTokens))
			}
		}

		return nil
	})

	if updateErr != nil {
		w.logger.Error("更新节点运行数据失败",
			slog.String("nodeKey", nodeKey),
			slog.String("taskResumeID", taskResumeID.String()),
			slog.String("traceID", w.traceID),
			slog.Any("error", updateErr),
		)
	}
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
		data, jsonErr := json.Marshal(input)
		if jsonErr != nil {
			w.logger.Debug("序列化输入数据失败",
				slog.String("nodeKey", nodeKey),
				slog.String("traceID", w.traceID),
				slog.Any("error", jsonErr),
			)
		} else {
			if unmarshalErr := json.Unmarshal(data, &inputPayload); unmarshalErr != nil {
				w.logger.Debug("反序列化输入数据失败",
					slog.String("nodeKey", nodeKey),
					slog.String("traceID", w.traceID),
					slog.Any("error", unmarshalErr),
				)
			}
		}
	}
	if output != nil {
		data, jsonErr := json.Marshal(output)
		if jsonErr != nil {
			w.logger.Debug("序列化输出数据失败",
				slog.String("nodeKey", nodeKey),
				slog.String("traceID", w.traceID),
				slog.Any("error", jsonErr),
			)
		} else {
			if unmarshalErr := json.Unmarshal(data, &outputPayload); unmarshalErr != nil {
				w.logger.Debug("反序列化输出数据失败",
					slog.String("nodeKey", nodeKey),
					slog.String("traceID", w.traceID),
					slog.Any("error", unmarshalErr),
				)
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
			slog.String("traceID", w.traceID),
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

	// 同步保存到数据库
	if _, saveErr := w.nodeRunRepo.Create(ctx, req); saveErr != nil {
		w.logger.Error("保存节点运行数据失败",
			slog.String("nodeKey", nodeKey),
			slog.String("taskResumeID", taskResumeID.String()),
			slog.String("traceID", w.traceID),
			slog.Any("error", saveErr),
		)
	} else {
		w.logger.Debug("节点运行数据保存成功",
			slog.String("nodeKey", nodeKey),
			slog.String("status", string(status)),
			slog.String("taskResumeID", taskResumeID.String()),
		)
	}
}

// getTaskResumeID 根据TaskID和ResumeID获取TaskResumeID
func (w *CallbackCollectorWrapper) getTaskResumeID(ctx context.Context) (uuid.UUID, error) {
	taskResume, err := w.screeningRepo.GetScreeningTaskResume(ctx, w.taskID, w.resumeID)
	if err != nil {
		return uuid.Nil, err
	}
	return taskResume.ID, nil
}
