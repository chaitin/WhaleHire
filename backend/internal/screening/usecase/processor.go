package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/screeningresult"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/internal/screening/service"
	"github.com/chaitin/WhaleHire/backend/internal/screening/util"
)

// ProcessResult 单个简历处理结果
type ProcessResult struct {
	ResumeID     uuid.UUID
	Success      bool
	Score        float64
	MatchLevel   consts.MatchLevel
	TokenInput   int64
	TokenOutput  int64
	ErrorMessage string
}

// ResultCollector 结果收集器，用于无锁并发处理
type ResultCollector struct {
	mu          sync.Mutex
	processed   int
	succeeded   int
	failed      int
	scoreSum    float64
	scoreCnt    float64
	histogram   map[string]float64
	tokenInput  int64
	tokenOutput int64
}

// NewResultCollector 创建新的结果收集器
func NewResultCollector() *ResultCollector {
	return &ResultCollector{
		histogram: make(map[string]float64),
	}
}

// CollectResult 收集单个处理结果
func (rc *ResultCollector) CollectResult(result ProcessResult) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if result.Success {
		rc.succeeded++
		rc.scoreSum += result.Score
		rc.scoreCnt++
		rc.histogram[string(result.MatchLevel)]++
	} else {
		rc.failed++
	}

	// processed 应该等于 succeeded + failed，避免重复计数
	rc.processed = rc.succeeded + rc.failed

	rc.tokenInput += result.TokenInput
	rc.tokenOutput += result.TokenOutput
}

// GetStats 获取当前统计信息
func (rc *ResultCollector) GetStats() (int, int, int, float64, float64, map[string]float64, int64, int64) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// 复制 histogram 避免并发访问问题
	histogramCopy := make(map[string]float64)
	for k, v := range rc.histogram {
		histogramCopy[k] = v
	}

	return rc.processed, rc.succeeded, rc.failed, rc.scoreSum, rc.scoreCnt, histogramCopy, rc.tokenInput, rc.tokenOutput
}

func (u *ScreeningUsecase) processConcurrentResumes(
	ctx context.Context,
	task *db.ScreeningTask,
	taskResumes []*db.ScreeningTaskResume,
	jobDetail *domain.JobProfileDetail,
	collector *ResultCollector,
) {
	executor := util.TaskExecutor[*db.ScreeningTaskResume, ProcessResult]{
		MaxWorkers: 10,
		TaskFunc: func(ctx context.Context, item *db.ScreeningTaskResume) (ProcessResult, error) {
			return u.processResumeItem(ctx, task, item, jobDetail), nil
		},
		OnResult: func(item *db.ScreeningTaskResume, res ProcessResult, err error) {
			collector.CollectResult(res)

			processed, succeeded, failed, _, _, _, _, _ := collector.GetStats()
			if dbErr := u.repo.UpdateScreeningTask(ctx, task.ID, map[string]any{
				"resume_processed": processed,
				"resume_succeeded": succeeded,
				"resume_failed":    failed,
			}); dbErr != nil {
				u.logger.Warn("更新任务进度失败",
					slog.Any("task_id", task.ID),
					slog.Any("resume_id", item.ResumeID),
					slog.Any("err", dbErr))
			}
		},
	}

	executor.Run(ctx, taskResumes)
}

// processResumeItem 处理单个简历项
func (u *ScreeningUsecase) processResumeItem(
	ctx context.Context,
	task *db.ScreeningTask,
	item *db.ScreeningTaskResume,
	jobDetail *domain.JobProfileDetail,
) ProcessResult {
	result := ProcessResult{
		ResumeID: item.ResumeID,
		Success:  false,
	}

	// 更新简历状态为运行中
	if err := u.repo.UpdateScreeningTaskResume(ctx, task.ID, item.ResumeID, map[string]any{
		"status": consts.ScreeningTaskResumeStatusRunning,
	}); err != nil {
		u.logger.Warn("更新简历状态失败", slog.Any("task_id", task.ID), slog.Any("resume_id", item.ResumeID), slog.Any("err", err))
	}

	resumeDetail, err := u.resumeUsecase.GetByID(ctx, item.ResumeID.String())
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("获取简历信息失败: %v", err)
		if updateErr := u.repo.UpdateScreeningTaskResume(ctx, task.ID, item.ResumeID, map[string]any{
			"status":        consts.ScreeningTaskResumeStatusFailed,
			"error_message": result.ErrorMessage,
			"processed_at":  time.Now(),
		}); updateErr != nil {
			u.logger.Error("更新简历状态失败", slog.Any("task_id", task.ID), slog.Any("resume_id", item.ResumeID), slog.Any("err", updateErr))
		}
		return result
	}

	// 转换 DimensionWeights 类型
	dimensionWeights := make(map[string]float64)
	if task.DimensionWeights != nil {
		for k, v := range task.DimensionWeights {
			if f, ok := v.(float64); ok {
				dimensionWeights[k] = f
			}
		}
	}

	matchReq := &service.MatchRequest{
		TaskID:           task.ID,
		ResumeID:         item.ResumeID,
		JobProfile:       jobDetail,
		Resume:           resumeDetail,
		DimensionWeights: dimensionWeights,
		LLMConfig:        task.LlmConfig,
	}

	matchResult, err := u.matcher.Match(ctx, matchReq)
	if err != nil {
		result.ErrorMessage = err.Error()
		u.logger.Error("执行智能匹配失败", slog.Any("task_id", task.ID), slog.Any("resume_id", item.ResumeID), slog.Any("err", err))
		if updateErr := u.repo.UpdateScreeningTaskResume(ctx, task.ID, item.ResumeID, map[string]any{
			"status":        consts.ScreeningTaskResumeStatusFailed,
			"error_message": result.ErrorMessage,
			"processed_at":  time.Now(),
		}); updateErr != nil {
			u.logger.Error("更新简历状态失败", slog.Any("task_id", task.ID), slog.Any("resume_id", item.ResumeID), slog.Any("err", updateErr))
		}
		return result
	}

	if matchResult.Match == nil {
		result.ErrorMessage = "智能匹配返回空结果"
		if updateErr := u.repo.UpdateScreeningTaskResume(ctx, task.ID, item.ResumeID, map[string]any{
			"status":        consts.ScreeningTaskResumeStatusFailed,
			"error_message": result.ErrorMessage,
			"processed_at":  time.Now(),
		}); updateErr != nil {
			u.logger.Error("更新简历状态失败", slog.Any("task_id", task.ID), slog.Any("resume_id", item.ResumeID), slog.Any("err", updateErr))
		}
		return result
	}

	// 汇总token
	for _, usage := range matchResult.TokenUsages {
		if usage == nil {
			continue
		}
		result.TokenInput += int64(usage.PromptTokens)
		result.TokenOutput += int64(usage.CompletionTokens)
	}

	result.Success = true
	result.Score = matchResult.Match.OverallScore
	result.MatchLevel = toMatchLevel(matchResult.Match.OverallScore)

	resultEntity, err := buildResultEntity(task, item.ResumeID, matchResult)
	if err != nil {
		u.logger.Error("构建筛选结果失败", slog.Any("task_id", task.ID), slog.Any("resume_id", item.ResumeID), slog.Any("err", err))
	} else {
		if _, err := u.repo.CreateScreeningResult(ctx, resultEntity); err != nil {
			u.logger.Error("保存筛选结果失败", slog.Any("task_id", task.ID), slog.Any("resume_id", item.ResumeID), slog.Any("err", err))
		}
	}

	if err := u.repo.UpdateScreeningTaskResume(ctx, task.ID, item.ResumeID, map[string]any{
		"status":       consts.ScreeningTaskResumeStatusCompleted,
		"score":        matchResult.Match.OverallScore,
		"processed_at": time.Now(),
	}); err != nil {
		u.logger.Warn("更新简历完成状态失败", slog.Any("task_id", task.ID), slog.Any("resume_id", item.ResumeID), slog.Any("err", err))
	}

	return result
}

// =========================
// 工具方法
// =========================

func buildResultEntity(task *db.ScreeningTask, resumeID uuid.UUID, result *service.MatchResult) (*db.ScreeningResult, error) {
	if result == nil || result.Match == nil {
		return nil, fmt.Errorf("匹配结果为空")
	}

	match := result.Match

	// 序列化详细匹配信息
	basicDetail, err := structToMap(match.BasicMatch)
	if err != nil {
		return nil, fmt.Errorf("序列化基本信息失败: %w", err)
	}

	educationDetail, err := structToMap(match.EducationMatch)
	if err != nil {
		return nil, fmt.Errorf("序列化教育背景失败: %w", err)
	}

	experienceDetail, err := structToMap(match.ExperienceMatch)
	if err != nil {
		return nil, fmt.Errorf("序列化工作经验失败: %w", err)
	}

	industryDetail, err := structToMap(match.IndustryMatch)
	if err != nil {
		return nil, fmt.Errorf("序列化行业匹配失败: %w", err)
	}

	responsibilityDetail, err := structToMap(match.ResponsibilityMatch)
	if err != nil {
		return nil, fmt.Errorf("序列化职责匹配失败: %w", err)
	}

	skillDetail, err := structToMap(match.SkillMatch)
	if err != nil {
		return nil, fmt.Errorf("序列化技能匹配失败: %w", err)
	}

	// 提取各维度分数
	dimensionScores := make(map[string]interface{})
	if result.DimensionMap != nil {
		for dimension, score := range result.DimensionMap {
			dimensionScores[dimension] = score
		}
	}

	// 提取 TraceID
	traceID := ""
	if match.TaskMetaData != nil && match.TaskMetaData.MatchTaskID != "" {
		traceID = match.TaskMetaData.MatchTaskID
	}

	// 序列化运行时元数据
	runtimeMetadata := result.Collector.SerializeTokenUsages()

	// 构建 SubAgentVersions 信息
	subAgentVersions := make(map[string]interface{})
	if result.SubAgentVersion != nil {
		for agent, version := range result.SubAgentVersion {
			subAgentVersions[agent] = version
		}
	}

	return &db.ScreeningResult{
		TaskID:               task.ID,
		JobPositionID:        task.JobPositionID,
		ResumeID:             resumeID,
		OverallScore:         match.OverallScore,
		MatchLevel:           screeningresult.MatchLevel(toMatchLevel(match.OverallScore)),
		DimensionScores:      dimensionScores,
		BasicDetail:          basicDetail,
		EducationDetail:      educationDetail,
		ExperienceDetail:     experienceDetail,
		IndustryDetail:       industryDetail,
		ResponsibilityDetail: responsibilityDetail,
		SkillDetail:          skillDetail,
		Recommendations:      match.Recommendations,
		TraceID:              traceID,
		RuntimeMetadata:      runtimeMetadata,
		SubAgentVersions:     subAgentVersions,
		MatchedAt:            match.MatchedAt,
	}, nil
}

func structToMap(obj any) (map[string]any, error) {
	if obj == nil {
		return nil, nil
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func safeAvg(sum, count float64) float64 {
	if count == 0 {
		return 0
	}
	return sum / count
}

// toMatchLevel 根据分数转换为匹配等级
// 分数范围：
// - 85-100：优秀匹配 (Excellent)
// - 70-84：良好匹配 (Good)
// - 55-69：一般匹配 (Fair)
// - 40-54：较差匹配 (Poor)
// - 0-39：不匹配 (No Match)
func toMatchLevel(score float64) consts.MatchLevel {
	switch {
	case score >= 85:
		return consts.MatchLevelExcellent
	case score >= 70:
		return consts.MatchLevelGood
	case score >= 55:
		return consts.MatchLevelFair
	case score >= 40:
		return consts.MatchLevelPoor
	default:
		return consts.MatchLevelNoMatch
	}
}
