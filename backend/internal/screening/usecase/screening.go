package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/errcode"
	"github.com/chaitin/WhaleHire/backend/internal/screening/service"
)

// ScreeningUsecase 筛选业务实现
type ScreeningUsecase struct {
	repo          domain.ScreeningRepo
	nodeRunRepo   domain.ScreeningNodeRunRepo
	jobUsecase    domain.JobProfileUsecase
	resumeUsecase domain.ResumeUsecase
	matcher       service.MatchingService
	logger        *slog.Logger
	// 任务上下文管理器
	taskContexts sync.Map // map[uuid.UUID]context.CancelFunc
}

// processScreeningTaskAsync 异步处理筛选任务
func (u *ScreeningUsecase) processScreeningTaskAsync(task *db.ScreeningTask, taskResumes []*db.ScreeningTaskResume, jobDetail *domain.JobProfileDetail) {
	// 为异步处理创建可取消的context
	processCtx, cancel := context.WithCancel(context.Background())

	// 保存取消函数到任务上下文管理器
	u.taskContexts.Store(task.ID, cancel)

	// 确保任务完成后清理上下文
	defer func() {
		u.taskContexts.Delete(task.ID)
	}()

	// 使用并发处理简历，提升处理效率
	collector := NewResultCollector()

	// 直接调用并发处理方法
	u.processConcurrentResumes(processCtx, task, taskResumes, jobDetail, collector)

	// 获取最终统计结果
	processed, succeeded, failed, scoreSum, scoreCnt, histogram, tokenInput, tokenOutput := collector.GetStats()

	u.logger.Info("screening task processing completed",
		"processed", processed,
		"succeeded", succeeded,
		"failed", failed,
		"scoreSum", scoreSum,
		"scoreCnt", scoreCnt,
		"histogram", histogram,
		"tokenInput", tokenInput,
		"tokenOutput", tokenOutput,
	)

	// 如果有成功处理的简历，进行排名更新
	if succeeded > 0 {
		if err := u.updateResumeRankings(processCtx, task.ID); err != nil {
			u.logger.Warn("更新简历排名失败", slog.Any("task_id", task.ID), slog.Any("err", err))
		}
	}

	finalStatus := consts.ScreeningTaskStatusCompleted
	if failed > 0 && succeeded == 0 {
		finalStatus = consts.ScreeningTaskStatusFailed
	} else if failed > 0 {
		finalStatus = consts.ScreeningTaskStatusFailed
	}

	if err := u.repo.UpdateScreeningTask(processCtx, task.ID, map[string]any{
		"status":           finalStatus,
		"finished_at":      time.Now(),
		"resume_processed": processed,
		"resume_succeeded": succeeded,
		"resume_failed":    failed,
		"agent_version":    u.matcher.Version(),
	}); err != nil {
		u.logger.Warn("更新任务完成状态失败", slog.Any("task_id", task.ID), slog.Any("err", err))
	}

	metric := &db.ScreeningRunMetric{
		TaskID:       task.ID,
		AvgScore:     safeAvg(scoreSum, scoreCnt),
		Histogram:    convertHistogramToMap(histogram),
		TokensInput:  tokenInput,
		TokensOutput: tokenOutput,
		TotalCost:    0,
	}
	if _, err := u.repo.CreateScreeningRunMetric(processCtx, metric); err != nil {
		u.logger.Warn("保存运行指标失败", slog.Any("task_id", task.ID), slog.Any("err", err))
	}
}

// NewScreeningUsecase 实例化用例
func NewScreeningUsecase(
	repo domain.ScreeningRepo,
	nodeRunRepo domain.ScreeningNodeRunRepo,
	jobUsecase domain.JobProfileUsecase,
	resumeUsecase domain.ResumeUsecase,
	matcher service.MatchingService,
	logger *slog.Logger,
) domain.ScreeningUsecase {
	if logger == nil {
		logger = slog.Default()
	}
	return &ScreeningUsecase{
		repo:          repo,
		nodeRunRepo:   nodeRunRepo,
		jobUsecase:    jobUsecase,
		resumeUsecase: resumeUsecase,
		matcher:       matcher,
		logger:        logger.With("module", "screening_usecase"),
	}
}

// CreateScreeningTask 创建任务
func (u *ScreeningUsecase) CreateScreeningTask(ctx context.Context, req *domain.CreateScreeningTaskReq) (*domain.CreateScreeningTaskResp, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}
	if req.JobPositionID == uuid.Nil {
		return nil, fmt.Errorf("岗位ID不能为空")
	}
	if req.CreatedBy == uuid.Nil {
		return nil, fmt.Errorf("创建人不能为空")
	}
	if len(req.ResumeIDs) == 0 {
		return nil, fmt.Errorf("至少需要选择一份简历")
	}

	// 版本兼容性检查
	if req.AgentVersion != "" {
		currentVersion := u.matcher.Version()
		if req.AgentVersion != currentVersion {
			return nil, fmt.Errorf("智能匹配代理版本不兼容，请求版本: %s，当前系统版本: %s", req.AgentVersion, currentVersion)
		}
	}

	// 如果用户没有提供维度权重，使用默认权重
	dimensionWeights := req.DimensionWeights
	if len(dimensionWeights) == 0 {
		dimensionWeights = map[string]float64{
			"skill":          domain.DefaultDimensionWeights.Skill,
			"responsibility": domain.DefaultDimensionWeights.Responsibility,
			"experience":     domain.DefaultDimensionWeights.Experience,
			"education":      domain.DefaultDimensionWeights.Education,
			"industry":       domain.DefaultDimensionWeights.Industry,
			"basic":          domain.DefaultDimensionWeights.Basic,
		}
	}

	taskID := uuid.New()
	entity := &db.ScreeningTask{
		ID:               taskID,
		JobPositionID:    req.JobPositionID,
		CreatedBy:        req.CreatedBy,
		Status:           string(consts.ScreeningTaskStatusPending),
		Notes:            req.Notes,
		ResumeTotal:      len(req.ResumeIDs),
		ResumeProcessed:  0,
		ResumeSucceeded:  0,
		ResumeFailed:     0,
		DimensionWeights: convertDimensionWeightsToMap(dimensionWeights),
		LlmConfig:        req.LLMConfig,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if _, err := u.repo.CreateScreeningTask(ctx, entity); err != nil {
		return nil, fmt.Errorf("创建筛选任务失败: %w", err)
	}

	for _, resumeID := range req.ResumeIDs {
		taskResume := &db.ScreeningTaskResume{
			TaskID:   taskID,
			ResumeID: resumeID,
			Status:   string(consts.ScreeningTaskResumeStatusPending),
		}
		if _, err := u.repo.CreateScreeningTaskResume(ctx, taskResume); err != nil {
			u.logger.Error("创建任务简历关联失败", slog.Any("task_id", taskID), slog.Any("resume_id", resumeID), slog.Any("err", err))
			return nil, fmt.Errorf("创建任务简历关联失败: %w", err)
		}
	}

	return &domain.CreateScreeningTaskResp{TaskID: taskID}, nil
}

// StartScreeningTask 启动任务
func (u *ScreeningUsecase) StartScreeningTask(ctx context.Context, req *domain.StartScreeningTaskReq) (*domain.StartScreeningTaskResp, error) {
	if req == nil || req.TaskID == uuid.Nil {
		return nil, fmt.Errorf("任务ID不能为空")
	}

	task, err := u.repo.GetScreeningTask(ctx, req.TaskID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, errcode.ErrScreeningTaskNotFound
		}
		return nil, fmt.Errorf("获取任务信息失败: %w", err)
	}
	if task.Status == string(consts.ScreeningTaskStatusRunning) {
		return &domain.StartScreeningTaskResp{TaskID: req.TaskID}, nil
	}
	if task.Status == string(consts.ScreeningTaskStatusCompleted) {
		return &domain.StartScreeningTaskResp{TaskID: req.TaskID}, nil
	}

	now := time.Now()
	if errTask := u.repo.UpdateScreeningTask(ctx, task.ID, map[string]any{
		"status":     consts.ScreeningTaskStatusRunning,
		"started_at": now,
	}); errTask != nil {
		return nil, fmt.Errorf("更新任务状态失败: %w", err)
	}

	jobDetail, err := u.jobUsecase.GetByID(ctx, task.JobPositionID.String())
	if err != nil {
		if updateErr := u.repo.UpdateScreeningTask(ctx, task.ID, map[string]any{
			"status":      consts.ScreeningTaskStatusFailed,
			"finished_at": time.Now(),
		}); updateErr != nil {
			u.logger.Error("更新任务状态失败", slog.Any("task_id", task.ID), slog.Any("err", updateErr))
		}
		return nil, fmt.Errorf("获取岗位信息失败: %w", err)
	}

	resumeFilter := &domain.ScreeningTaskResumeFilter{
		TaskID: &task.ID,
	}
	taskResumes, _, err := u.repo.ListScreeningTaskResumes(ctx, resumeFilter)
	if err != nil {
		return nil, fmt.Errorf("获取任务简历列表失败: %w", err)
	}
	if len(taskResumes) == 0 {
		return nil, fmt.Errorf("任务下没有待处理的简历")
	}

	// 启动后台异步处理简历
	go u.processScreeningTaskAsync(task, taskResumes, jobDetail)

	return &domain.StartScreeningTaskResp{TaskID: task.ID}, nil
}

// CancelScreeningTask 取消任务
func (u *ScreeningUsecase) CancelScreeningTask(ctx context.Context, req *domain.CancelScreeningTaskReq) (*domain.CancelScreeningTaskResp, error) {
	if req == nil || req.TaskID == uuid.Nil {
		return nil, errcode.ErrInvalidParam.Wrap(fmt.Errorf("任务ID不能为空"))
	}

	// 获取任务信息
	task, err := u.repo.GetScreeningTask(ctx, req.TaskID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, errcode.ErrScreeningTaskNotFound
		}
		return nil, fmt.Errorf("获取任务信息失败: %w", err)
	}

	// 检查任务状态，只有运行中的任务可以取消
	if task.Status != string(consts.ScreeningTaskStatusRunning) {
		return nil, errcode.ErrInvalidParam.Wrap(fmt.Errorf("只有运行中的任务可以取消，当前状态: %s", task.Status))
	}

	// 调用取消函数停止正在运行的任务
	if cancelFunc, ok := u.taskContexts.LoadAndDelete(req.TaskID); ok {
		if cancel, ok := cancelFunc.(context.CancelFunc); ok {
			cancel() // 取消正在运行的上下文
			u.logger.Info("成功取消正在运行的任务上下文", slog.Any("task_id", req.TaskID))
		}
	}

	// 更新任务状态为已取消
	now := time.Now()
	if updateErr := u.repo.UpdateScreeningTask(ctx, req.TaskID, map[string]any{
		"status":      consts.ScreeningTaskStatusCancelled,
		"finished_at": now,
	}); updateErr != nil {
		u.logger.Error("更新任务取消状态失败", slog.Any("err", err), slog.Any("task_id", req.TaskID))
		return nil, fmt.Errorf("取消任务失败: %w", err)
	}

	// 更新所有待处理的简历状态为已取消
	resumeFilter := &domain.ScreeningTaskResumeFilter{
		TaskID: &req.TaskID,
	}
	taskResumes, _, err := u.repo.ListScreeningTaskResumes(ctx, resumeFilter)
	if err != nil {
		u.logger.Warn("获取任务简历列表失败", slog.Any("task_id", req.TaskID), slog.Any("err", err))
	} else {
		for _, taskResume := range taskResumes {
			// 只更新待处理和处理中的简历状态
			if taskResume.Status == string(consts.ScreeningTaskResumeStatusPending) ||
				taskResume.Status == string(consts.ScreeningTaskResumeStatusRunning) {
				if updateErr := u.repo.UpdateScreeningTaskResume(ctx, taskResume.TaskID, taskResume.ResumeID, map[string]any{
					"status":        consts.ScreeningTaskResumeStatusCancelled,
					"error_message": "任务已被取消",
					"processed_at":  now,
				}); updateErr != nil {
					u.logger.Warn("更新简历取消状态失败",
						slog.Any("task_id", taskResume.TaskID),
						slog.Any("resume_id", taskResume.ResumeID),
						slog.Any("err", updateErr))
				}
			}
		}
	}

	u.logger.Info("成功取消筛选任务", slog.Any("task_id", req.TaskID))

	return &domain.CancelScreeningTaskResp{
		TaskID:  req.TaskID,
		Message: "任务已成功取消",
	}, nil
}

// DeleteScreeningTask 删除任务
func (u *ScreeningUsecase) DeleteScreeningTask(ctx context.Context, req *domain.DeleteScreeningTaskReq) (*domain.DeleteScreeningTaskResp, error) {
	if req == nil || req.TaskID == uuid.Nil {
		return nil, errcode.ErrInvalidParam.Wrap(fmt.Errorf("任务ID不能为空"))
	}

	// 获取任务信息
	task, err := u.repo.GetScreeningTask(ctx, req.TaskID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, errcode.ErrScreeningTaskNotFound
		}
		return nil, fmt.Errorf("获取任务信息失败: %w", err)
	}

	// 检查任务状态，正在运行的任务不能删除
	if task.Status == string(consts.ScreeningTaskStatusRunning) {
		return nil, errcode.ErrScreeningTaskRunning
	}

	// 执行删除操作
	if err := u.repo.DeleteScreeningTask(ctx, req.TaskID); err != nil {
		u.logger.Error("删除筛选任务失败", slog.Any("err", err), slog.Any("task_id", req.TaskID))
		return nil, errcode.ErrScreeningTaskDeleteFailed.Wrap(err)
	}

	u.logger.Info("成功删除筛选任务", slog.Any("task_id", req.TaskID), slog.String("status", string(task.Status)))

	return &domain.DeleteScreeningTaskResp{
		TaskID:  req.TaskID,
		Message: "任务已成功删除",
	}, nil
}

// GetScreeningTask 获取任务详情
func (u *ScreeningUsecase) GetScreeningTask(ctx context.Context, req *domain.GetScreeningTaskReq) (*domain.GetScreeningTaskResp, error) {
	if req == nil || req.TaskID == uuid.Nil {
		return nil, fmt.Errorf("任务ID不能为空")
	}

	task, err := u.repo.GetScreeningTask(ctx, req.TaskID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, errcode.ErrScreeningTaskNotFound
		}
		return nil, fmt.Errorf("获取任务信息失败: %w", err)
	}

	resumes, _, err := u.repo.ListScreeningTaskResumes(ctx, &domain.ScreeningTaskResumeFilter{
		TaskID: &req.TaskID,
	})
	if err != nil {
		return nil, fmt.Errorf("获取任务简历失败: %w", err)
	}

	taskModel := toScreeningTask(task)
	resumeModels := make([]*domain.ScreeningTaskResume, 0, len(resumes))
	for _, entity := range resumes {
		resumeModels = append(resumeModels, toScreeningTaskResume(entity))
	}

	var metricModel *domain.ScreeningRunMetric
	metric, err := u.repo.GetScreeningRunMetric(ctx, req.TaskID)
	if err != nil && !db.IsNotFound(err) {
		u.logger.Warn("获取运行指标失败", slog.Any("task_id", req.TaskID), slog.Any("err", err))
	}
	if metric != nil {
		metricModel = toScreeningRunMetric(metric)
	}

	return &domain.GetScreeningTaskResp{
		Task:    taskModel,
		Resumes: resumeModels,
		Metrics: metricModel,
	}, nil
}

// ListScreeningTasks 列表任务
func (u *ScreeningUsecase) ListScreeningTasks(ctx context.Context, req *domain.ListScreeningTasksReq) (*domain.ListScreeningTasksResp, error) {
	if req == nil {
		req = &domain.ListScreeningTasksReq{}
		req.Page = 1
		req.Size = 20
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}

	filter := &domain.ScreeningTaskFilter{
		JobPositionID: req.JobPositionID,
		Status:        req.Status,
		CreatedBy:     req.CreatedBy,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		Page:          req.Page,
		PageSize:      req.Size,
	}

	tasks, pageInfo, err := u.repo.ListScreeningTasks(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("获取任务列表失败: %w", err)
	}

	items := make([]*domain.ScreeningTask, 0, len(tasks))
	for _, entity := range tasks {
		items = append(items, toScreeningTask(entity))
	}

	resp := &domain.ListScreeningTasksResp{
		Items:    items,
		PageInfo: pageInfo,
	}
	return resp, nil
}

// GetScreeningResult 获取单个结果
func (u *ScreeningUsecase) GetScreeningResult(ctx context.Context, req *domain.GetScreeningResultReq) (*domain.GetScreeningResultResp, error) {
	if req == nil || req.TaskID == uuid.Nil || req.ResumeID == uuid.Nil {
		return nil, fmt.Errorf("任务ID和简历ID不能为空")
	}

	entity, err := u.repo.GetScreeningResult(ctx, req.TaskID, req.ResumeID)
	if err != nil {
		return nil, err
	}

	result, err := toScreeningResult(entity)
	if err != nil {
		return nil, fmt.Errorf("解析筛选结果失败: %w", err)
	}

	return &domain.GetScreeningResultResp{Result: result}, nil
}

// ListScreeningResults 列表查询结果
func (u *ScreeningUsecase) ListScreeningResults(ctx context.Context, req *domain.ListScreeningResultsReq) (*domain.ListScreeningResultsResp, error) {
	if req == nil {
		req = &domain.ListScreeningResultsReq{}
		req.Page = 1
		req.Size = 20
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}

	filter := &domain.ScreeningResultFilter{
		TaskID:   req.TaskID,
		ResumeID: req.ResumeID,
		MinScore: req.MinScore,
		MaxScore: req.MaxScore,
		Page:     req.Page,
		PageSize: req.Size,
	}

	entities, pageInfo, err := u.repo.ListScreeningResults(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("获取筛选结果失败: %w", err)
	}

	results := make([]*domain.ScreeningResult, 0, len(entities))
	for _, entity := range entities {
		result, convErr := toScreeningResult(entity)
		if convErr != nil {
			u.logger.Warn("解析筛选结果失败", slog.Any("task_id", entity.TaskID), slog.Any("resume_id", entity.ResumeID), slog.Any("err", convErr))
			continue
		}
		results = append(results, result)
	}

	return &domain.ListScreeningResultsResp{
		Items:    results,
		PageInfo: pageInfo,
	}, nil
}

// GetScreeningMetrics 获取任务指标
func (u *ScreeningUsecase) GetScreeningMetrics(ctx context.Context, req *domain.GetScreeningMetricsReq) (*domain.GetScreeningMetricsResp, error) {
	if req == nil || req.TaskID == uuid.Nil {
		return nil, fmt.Errorf("任务ID不能为空")
	}

	metric, err := u.repo.GetScreeningRunMetric(ctx, req.TaskID)
	if err != nil {
		if db.IsNotFound(err) {
			return &domain.GetScreeningMetricsResp{Metrics: nil}, nil
		}
		return nil, fmt.Errorf("获取运行指标失败: %w", err)
	}

	return &domain.GetScreeningMetricsResp{
		Metrics: toScreeningRunMetric(metric),
	}, nil
}

// GetTaskProgress 获取任务进度
func (u *ScreeningUsecase) GetTaskProgress(ctx context.Context, req *domain.GetTaskProgressReq) (*domain.GetTaskProgressResp, error) {
	if req == nil || req.TaskID == uuid.Nil {
		return nil, fmt.Errorf("任务ID不能为空")
	}

	// 获取任务基本信息
	task, err := u.repo.GetScreeningTask(ctx, req.TaskID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, errcode.ErrScreeningTaskNotFound
		}
		return nil, fmt.Errorf("获取任务信息失败: %w", err)
	}

	// 计算进度百分比
	var progressPercent float64
	if task.ResumeTotal > 0 {
		progressPercent = float64(task.ResumeProcessed) / float64(task.ResumeTotal) * 100
	}

	// 估算完成时间（如果任务正在运行且有进度）
	var estimatedFinish *time.Time
	if task.Status == string(consts.ScreeningTaskStatusRunning) && !task.StartedAt.IsZero() && task.ResumeProcessed > 0 {
		elapsed := time.Since(task.StartedAt)
		avgTimePerResume := elapsed / time.Duration(task.ResumeProcessed)
		remainingResumes := task.ResumeTotal - task.ResumeProcessed
		estimatedDuration := avgTimePerResume * time.Duration(remainingResumes)
		estimated := time.Now().Add(estimatedDuration)
		estimatedFinish = &estimated
	}

	return &domain.GetTaskProgressResp{
		TaskID:          task.ID,
		Status:          consts.ScreeningTaskStatus(task.Status),
		ResumeTotal:     task.ResumeTotal,
		ResumeProcessed: task.ResumeProcessed,
		ResumeSucceeded: task.ResumeSucceeded,
		ResumeFailed:    task.ResumeFailed,
		ProgressPercent: progressPercent,
		StartedAt:       timeToPtr(task.StartedAt),
		EstimatedFinish: estimatedFinish,
	}, nil
}

// GetResumeProgress 获取单个简历匹配进度
func (u *ScreeningUsecase) GetResumeProgress(ctx context.Context, req *domain.GetResumeProgressReq) (*domain.GetResumeProgressResp, error) {
	if req == nil || req.TaskID == uuid.Nil || req.ResumeID == uuid.Nil {
		return nil, fmt.Errorf("任务ID和简历ID不能为空")
	}

	// 获取任务简历关联信息
	taskResume, err := u.repo.GetScreeningTaskResume(ctx, req.TaskID, req.ResumeID)
	if err != nil {
		return nil, fmt.Errorf("获取简历进度信息失败: %w", err)
	}

	return &domain.GetResumeProgressResp{
		TaskID:       taskResume.TaskID,
		ResumeID:     taskResume.ResumeID,
		Status:       consts.ScreeningTaskResumeStatus(taskResume.Status),
		Score:        float64ToPtr(taskResume.Score),
		ErrorMessage: taskResume.ErrorMessage,
		ProcessedAt:  timeToPtr(taskResume.ProcessedAt),
		CreatedAt:    taskResume.CreatedAt,
		UpdatedAt:    taskResume.UpdatedAt,
	}, nil
}

// GetNodeRuns 获取节点运行记录
func (u *ScreeningUsecase) GetNodeRuns(ctx context.Context, req *domain.GetNodeRunsReq) (*domain.GetNodeRunsResp, error) {
	// 验证任务和简历是否存在
	taskResume, err := u.repo.GetScreeningTaskResume(ctx, req.TaskID, req.ResumeID)
	if err != nil {
		u.logger.Error("failed to get task resume", "error", err, "task_id", req.TaskID, "resume_id", req.ResumeID)
		return nil, errcode.ErrScreeningTaskNotFound
	}

	// 构建查询请求，不使用分页，获取所有节点运行记录
	listReq := &domain.ListNodeRunsReq{
		TaskID:       &req.TaskID,
		TaskResumeID: &taskResume.ID,
	}

	// 查询节点运行记录
	nodeRuns, _, err := u.nodeRunRepo.List(ctx, &domain.ListNodeRunsRepoReq{ListNodeRunsReq: listReq})
	if err != nil {
		u.logger.Error("failed to list node runs", "error", err, "task_id", req.TaskID, "resume_id", req.ResumeID)
		return nil, errcode.ErrInvalidParam
	}

	// 转换为Key-Value格式的Agent状态映射
	agentStatus := make(map[string]*domain.ScreeningNodeRun)

	// 初始化所有Agent状态为nil
	agentStatus[domain.TaskMetaDataNode] = nil
	agentStatus[domain.DispatcherNode] = nil
	agentStatus[domain.BasicInfoAgent] = nil
	agentStatus[domain.SkillAgent] = nil
	agentStatus[domain.ResponsibilityAgent] = nil
	agentStatus[domain.ExperienceAgent] = nil
	agentStatus[domain.EducationAgent] = nil
	agentStatus[domain.IndustryAgent] = nil
	agentStatus[domain.AggregatorAgent] = nil

	// 填充实际的节点运行记录
	for _, nodeRun := range nodeRuns {
		domainNodeRun := new(domain.ScreeningNodeRun).From(nodeRun)
		agentStatus[nodeRun.NodeKey] = domainNodeRun
	}

	return &domain.GetNodeRunsResp{
		AgentStatus: agentStatus,
	}, nil
}

// =========================
// 转换方法
// =========================

func toScreeningTask(entity *db.ScreeningTask) *domain.ScreeningTask {
	if entity == nil {
		return nil
	}

	task := &domain.ScreeningTask{}
	task.From(entity)
	return task
}

func toScreeningTaskResume(entity *db.ScreeningTaskResume) *domain.ScreeningTaskResume {
	if entity == nil {
		return nil
	}

	taskResume := &domain.ScreeningTaskResume{}
	taskResume.From(entity)
	return taskResume
}

func toScreeningResult(entity *db.ScreeningResult) (*domain.ScreeningResult, error) {
	if entity == nil {
		return nil, nil
	}

	result := &domain.ScreeningResult{}
	result.From(entity)
	return result, nil
}

func toScreeningRunMetric(entity *db.ScreeningRunMetric) *domain.ScreeningRunMetric {
	if entity == nil {
		return nil
	}

	metric := &domain.ScreeningRunMetric{}
	metric.From(entity)
	return metric
}

// 辅助函数：转换 map[string]float64 到 map[string]interface{}
func convertDimensionWeightsToMap(weights map[string]float64) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range weights {
		result[k] = v
	}
	return result
}

// 辅助函数：转换 histogram map[string]float64 到 map[string]interface{}
func convertHistogramToMap(histogram map[string]float64) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range histogram {
		result[k] = v
	}
	return result
}

// 辅助函数：时间指针转换
func timeToPtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// 辅助函数：float64指针转换
func float64ToPtr(f float64) *float64 {
	if f == 0 {
		return nil
	}
	return &f
}

// updateResumeRankings 更新任务内简历的排名
func (u *ScreeningUsecase) updateResumeRankings(ctx context.Context, taskID uuid.UUID) error {
	// 获取任务内所有已完成的简历及其分数
	filter := &domain.ScreeningTaskResumeFilter{
		TaskID: &taskID,
		Status: (*consts.ScreeningTaskResumeStatus)(&[]consts.ScreeningTaskResumeStatus{consts.ScreeningTaskResumeStatusCompleted}[0]),
	}
	
	taskResumes, _, err := u.repo.ListScreeningTaskResumes(ctx, filter)
	if err != nil {
		return fmt.Errorf("获取任务简历列表失败: %w", err)
	}

	if len(taskResumes) == 0 {
		return nil // 没有已完成的简历，无需排名
	}

	// 按分数降序排序简历
	type resumeScore struct {
		resumeID uuid.UUID
		score    float64
	}

	var resumeScores []resumeScore
	for _, tr := range taskResumes {
		if tr.Score != 0 { // 只处理有分数的简历
			resumeScores = append(resumeScores, resumeScore{
				resumeID: tr.ResumeID,
				score:    tr.Score,
			})
		}
	}

	if len(resumeScores) == 0 {
		return nil // 没有有效分数的简历
	}

	// 按分数降序排序
	for i := 0; i < len(resumeScores)-1; i++ {
		for j := i + 1; j < len(resumeScores); j++ {
			if resumeScores[i].score < resumeScores[j].score {
				resumeScores[i], resumeScores[j] = resumeScores[j], resumeScores[i]
			}
		}
	}

	// 构建排名映射，处理相同分数的情况
	rankings := make(map[uuid.UUID]int)
	currentRank := 1
	
	for i, rs := range resumeScores {
		if i > 0 && resumeScores[i-1].score != rs.score {
			currentRank = i + 1 // 分数不同时，排名为当前位置+1
		}
		rankings[rs.resumeID] = currentRank
	}

	// 批量更新排名
	return u.repo.BatchUpdateScreeningTaskResumeRankings(ctx, taskID, rankings)
}
