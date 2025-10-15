package repo

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/screeningresult"
	"github.com/chaitin/WhaleHire/backend/db/screeningrunmetric"
	"github.com/chaitin/WhaleHire/backend/db/screeningtask"
	"github.com/chaitin/WhaleHire/backend/db/screeningtaskresume"
	"github.com/chaitin/WhaleHire/backend/domain"
)

// ScreeningRepo 智能筛选仓储实现
type ScreeningRepo struct {
	db *db.Client
}

// NewScreeningRepo 构建仓储
func NewScreeningRepo(client *db.Client) domain.ScreeningRepo {
	return &ScreeningRepo{db: client}
}

// CreateScreeningTask 新建任务
func (r *ScreeningRepo) CreateScreeningTask(ctx context.Context, task *db.ScreeningTask) (*db.ScreeningTask, error) {
	if task == nil {
		return nil, fmt.Errorf("task is nil")
	}

	builder := r.db.ScreeningTask.Create().
		SetID(task.ID).
		SetJobPositionID(task.JobPositionID).
		SetCreatedBy(task.CreatedBy).
		SetStatus(task.Status).
		SetResumeTotal(task.ResumeTotal).
		SetResumeProcessed(task.ResumeProcessed).
		SetResumeSucceeded(task.ResumeSucceeded).
		SetResumeFailed(task.ResumeFailed)

	if len(task.DimensionWeights) > 0 {
		builder = builder.SetDimensionWeights(task.DimensionWeights)
	}
	if len(task.LlmConfig) > 0 {
		builder = builder.SetLlmConfig(task.LlmConfig)
	}
	if task.Notes != "" {
		builder = builder.SetNotes(task.Notes)
	}
	if task.AgentVersion != "" {
		builder = builder.SetAgentVersion(task.AgentVersion)
	}
	if !task.StartedAt.IsZero() {
		builder = builder.SetStartedAt(task.StartedAt)
	}
	if !task.FinishedAt.IsZero() {
		builder = builder.SetFinishedAt(task.FinishedAt)
	}

	entity, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create screening task failed: %w", err)
	}
	return entity, nil
}

// GetScreeningTask 查询任务
func (r *ScreeningRepo) GetScreeningTask(ctx context.Context, id uuid.UUID) (*db.ScreeningTask, error) {
	task, err := r.db.ScreeningTask.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// ListScreeningTasks 查询任务列表
func (r *ScreeningRepo) ListScreeningTasks(ctx context.Context, filter *domain.ScreeningTaskFilter) ([]*db.ScreeningTask, *db.PageInfo, error) {
	query := r.db.ScreeningTask.Query()

	if filter != nil {
		if filter.JobPositionID != nil {
			query = query.Where(screeningtask.JobPositionID(*filter.JobPositionID))
		}
		if filter.Status != nil {
			query = query.Where(screeningtask.StatusEQ(string(*filter.Status)))
		}
		if filter.CreatedBy != nil {
			query = query.Where(screeningtask.CreatedBy(*filter.CreatedBy))
		}
		if filter.StartTime != nil {
			query = query.Where(screeningtask.CreatedAtGTE(*filter.StartTime))
		}
		if filter.EndTime != nil {
			query = query.Where(screeningtask.CreatedAtLTE(*filter.EndTime))
		}
	}

	// 排序
	query = query.Order(screeningtask.ByCreatedAt(sql.OrderDesc()))

	// 使用 db 层的分页方法
	page := 1
	pageSize := 20
	if filter != nil {
		if filter.Page > 0 {
			page = filter.Page
		}
		if filter.PageSize > 0 {
			pageSize = filter.PageSize
		}
	}

	items, pageInfo, err := query.Page(ctx, page, pageSize)
	if err != nil {
		return nil, nil, fmt.Errorf("list screening tasks failed: %w", err)
	}

	return items, pageInfo, nil
}

// UpdateScreeningTask 更新任务
func (r *ScreeningRepo) UpdateScreeningTask(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	builder := r.db.ScreeningTask.UpdateOneID(id)
	if err := applyTaskUpdates(builder, updates); err != nil {
		return err
	}
	if err := builder.Exec(ctx); err != nil {
		return fmt.Errorf("update screening task failed: %w", err)
	}
	return nil
}

// DeleteScreeningTask 删除任务
func (r *ScreeningRepo) DeleteScreeningTask(ctx context.Context, id uuid.UUID) error {
	// 使用事务确保数据一致性
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = fmt.Errorf("rollback failed: %w (original error: %v)", rollbackErr, err)
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("commit failed: %w", commitErr)
			}
		}
	}()

	// 删除相关的筛选结果
	if _, err = tx.ScreeningResult.Delete().Where(screeningresult.TaskID(id)).Exec(ctx); err != nil {
		return fmt.Errorf("delete screening results failed: %w", err)
	}

	// 删除相关的任务简历关联
	if _, err = tx.ScreeningTaskResume.Delete().Where(screeningtaskresume.TaskID(id)).Exec(ctx); err != nil {
		return fmt.Errorf("delete screening task resumes failed: %w", err)
	}

	// 删除相关的运行指标
	if _, err = tx.ScreeningRunMetric.Delete().Where(screeningrunmetric.TaskID(id)).Exec(ctx); err != nil {
		return fmt.Errorf("delete screening run metrics failed: %w", err)
	}

	// 删除任务本身
	if err = tx.ScreeningTask.DeleteOneID(id).Exec(ctx); err != nil {
		return fmt.Errorf("delete screening task failed: %w", err)
	}

	return nil
}

// CreateScreeningTaskResume 新建任务简历关联
func (r *ScreeningRepo) CreateScreeningTaskResume(ctx context.Context, resume *db.ScreeningTaskResume) (*db.ScreeningTaskResume, error) {
	if resume == nil {
		return nil, fmt.Errorf("resume is nil")
	}

	builder := r.db.ScreeningTaskResume.Create().
		SetTaskID(resume.TaskID).
		SetResumeID(resume.ResumeID).
		SetStatus(resume.Status)

	if resume.Ranking != 0 {
		builder = builder.SetRanking(resume.Ranking)
	}
	if resume.Score != 0 {
		builder = builder.SetScore(resume.Score)
	}
	if resume.ErrorMessage != "" {
		builder = builder.SetErrorMessage(resume.ErrorMessage)
	}
	if !resume.ProcessedAt.IsZero() {
		builder = builder.SetProcessedAt(resume.ProcessedAt)
	}

	entity, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create screening task resume failed: %w", err)
	}
	return entity, nil
}

// GetScreeningTaskResume 查询任务简历关联
func (r *ScreeningRepo) GetScreeningTaskResume(ctx context.Context, taskID, resumeID uuid.UUID) (*db.ScreeningTaskResume, error) {
	entity, err := r.db.ScreeningTaskResume.Query().
		Where(screeningtaskresume.TaskID(taskID)).
		Where(screeningtaskresume.ResumeID(resumeID)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// ListScreeningTaskResumes 查询任务简历关联列表
func (r *ScreeningRepo) ListScreeningTaskResumes(ctx context.Context, filter *domain.ScreeningTaskResumeFilter) ([]*db.ScreeningTaskResume, *db.PageInfo, error) {
	query := r.db.ScreeningTaskResume.Query()

	if filter != nil {
		if filter.TaskID != nil {
			query = query.Where(screeningtaskresume.TaskID(*filter.TaskID))
		}
		if filter.ResumeID != nil {
			query = query.Where(screeningtaskresume.ResumeID(*filter.ResumeID))
		}
		if filter.Status != nil {
			query = query.Where(screeningtaskresume.StatusEQ(string(*filter.Status)))
		}
	}

	// 排序
	query = query.Order(screeningtaskresume.ByCreatedAt(sql.OrderDesc()))

	// 使用 db 层的分页方法
	page := 1
	pageSize := 20
	if filter != nil {
		if filter.Page > 0 {
			page = filter.Page
		}
		if filter.PageSize > 0 {
			pageSize = filter.PageSize
		}
	}

	items, pageInfo, err := query.Page(ctx, page, pageSize)
	if err != nil {
		return nil, nil, fmt.Errorf("list screening task resumes failed: %w", err)
	}

	return items, pageInfo, nil
}

// UpdateScreeningTaskResume 更新任务简历
func (r *ScreeningRepo) UpdateScreeningTaskResume(ctx context.Context, taskID, resumeID uuid.UUID, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	query := r.db.ScreeningTaskResume.Update().
		Where(
			screeningtaskresume.TaskID(taskID),
			screeningtaskresume.ResumeID(resumeID),
		)

	if err := applyTaskResumeUpdates(query, updates); err != nil {
		return err
	}

	if _, err := query.Save(ctx); err != nil {
		return fmt.Errorf("update task resume failed: %w", err)
	}
	return nil
}

// CreateScreeningResult 新建筛选结果
func (r *ScreeningRepo) CreateScreeningResult(ctx context.Context, result *db.ScreeningResult) (*db.ScreeningResult, error) {
	if result == nil {
		return nil, fmt.Errorf("result is nil")
	}

	builder := r.db.ScreeningResult.Create().
		SetTaskID(result.TaskID).
		SetJobPositionID(result.JobPositionID).
		SetResumeID(result.ResumeID).
		SetOverallScore(result.OverallScore).
		SetMatchLevel(result.MatchLevel)

	if len(result.DimensionScores) > 0 {
		builder = builder.SetDimensionScores(result.DimensionScores)
	}
	if len(result.BasicDetail) > 0 {
		builder = builder.SetBasicDetail(result.BasicDetail)
	}
	if len(result.EducationDetail) > 0 {
		builder = builder.SetEducationDetail(result.EducationDetail)
	}
	if len(result.ExperienceDetail) > 0 {
		builder = builder.SetExperienceDetail(result.ExperienceDetail)
	}
	if len(result.IndustryDetail) > 0 {
		builder = builder.SetIndustryDetail(result.IndustryDetail)
	}
	if len(result.ResponsibilityDetail) > 0 {
		builder = builder.SetResponsibilityDetail(result.ResponsibilityDetail)
	}
	if len(result.SkillDetail) > 0 {
		builder = builder.SetSkillDetail(result.SkillDetail)
	}
	if len(result.Recommendations) > 0 {
		builder = builder.SetRecommendations(result.Recommendations)
	}
	if result.TraceID != "" {
		builder = builder.SetTraceID(result.TraceID)
	}
	if len(result.RuntimeMetadata) > 0 {
		builder = builder.SetRuntimeMetadata(result.RuntimeMetadata)
	}
	if len(result.SubAgentVersions) > 0 {
		builder = builder.SetSubAgentVersions(result.SubAgentVersions)
	}
	if !result.MatchedAt.IsZero() {
		builder = builder.SetMatchedAt(result.MatchedAt)
	}

	entity, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create screening result failed: %w", err)
	}
	return entity, nil
}

// GetScreeningResult 查询筛选结果
func (r *ScreeningRepo) GetScreeningResult(ctx context.Context, taskID, resumeID uuid.UUID) (*db.ScreeningResult, error) {
	entity, err := r.db.ScreeningResult.Query().
		Where(screeningresult.TaskID(taskID)).
		Where(screeningresult.ResumeID(resumeID)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// ListScreeningResults 查询筛选结果列表
func (r *ScreeningRepo) ListScreeningResults(ctx context.Context, filter *domain.ScreeningResultFilter) ([]*db.ScreeningResult, *db.PageInfo, error) {
	query := r.db.ScreeningResult.Query()

	if filter != nil {
		if filter.TaskID != nil {
			query = query.Where(screeningresult.TaskID(*filter.TaskID))
		}
		if filter.ResumeID != nil {
			query = query.Where(screeningresult.ResumeID(*filter.ResumeID))
		}
		if filter.MinScore != nil {
			query = query.Where(screeningresult.OverallScoreGTE(*filter.MinScore))
		}
		if filter.MaxScore != nil {
			query = query.Where(screeningresult.OverallScoreLTE(*filter.MaxScore))
		}
	}

	// 排序
	query = query.Order(screeningresult.ByOverallScore(sql.OrderDesc()))

	// 使用 db 层的分页方法
	page := 1
	pageSize := 20
	if filter != nil {
		if filter.Page > 0 {
			page = filter.Page
		}
		if filter.PageSize > 0 {
			pageSize = filter.PageSize
		}
	}

	items, pageInfo, err := query.Page(ctx, page, pageSize)
	if err != nil {
		return nil, nil, fmt.Errorf("list screening results failed: %w", err)
	}

	return items, pageInfo, nil
}

// CreateScreeningRunMetric 新建运行指标
func (r *ScreeningRepo) CreateScreeningRunMetric(ctx context.Context, metric *db.ScreeningRunMetric) (*db.ScreeningRunMetric, error) {
	if metric == nil {
		return nil, fmt.Errorf("metric is nil")
	}

	builder := r.db.ScreeningRunMetric.Create().
		SetTaskID(metric.TaskID).
		SetAvgScore(metric.AvgScore).
		SetTokensInput(metric.TokensInput).
		SetTokensOutput(metric.TokensOutput).
		SetTotalCost(metric.TotalCost)

	if len(metric.Histogram) > 0 {
		builder = builder.SetHistogram(metric.Histogram)
	}

	entity, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create screening run metric failed: %w", err)
	}
	return entity, nil
}

// GetScreeningRunMetric 查询运行指标
func (r *ScreeningRepo) GetScreeningRunMetric(ctx context.Context, taskID uuid.UUID) (*db.ScreeningRunMetric, error) {
	entity, err := r.db.ScreeningRunMetric.Query().
		Where(screeningrunmetric.TaskID(taskID)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// =========================
// 辅助函数
// =========================

func floatMapToAny(input map[string]float64) map[string]any {
	if len(input) == 0 {
		return nil
	}
	result := make(map[string]any, len(input))
	for k, v := range input {
		result[k] = v
	}
	return result
}

func applyTaskUpdates(builder *db.ScreeningTaskUpdateOne, updates map[string]any) error {
	for key, value := range updates {
		switch key {
		case "status":
			if status, ok := value.(consts.ScreeningTaskStatus); ok {
				builder.SetStatus(string(status))
			} else if str, ok := value.(string); ok && str != "" {
				builder.SetStatus(str)
			}
		case "dimension_weights":
			switch v := value.(type) {
			case map[string]float64:
				builder.SetDimensionWeights(floatMapToAny(v))
			case map[string]any:
				builder.SetDimensionWeights(v)
			}
		case "llm_config":
			if cfg, ok := value.(map[string]any); ok {
				builder.SetLlmConfig(cfg)
			}
		case "notes":
			if str, ok := value.(string); ok {
				builder.SetNotes(str)
			}
		case "resume_total":
			builder.SetResumeTotal(asInt(value))
		case "resume_processed":
			builder.SetResumeProcessed(asInt(value))
		case "resume_succeeded":
			builder.SetResumeSucceeded(asInt(value))
		case "resume_failed":
			builder.SetResumeFailed(asInt(value))
		case "agent_version":
			if str, ok := value.(string); ok {
				builder.SetAgentVersion(str)
			}
		case "started_at":
			switch v := value.(type) {
			case time.Time:
				builder.SetStartedAt(v)
			case *time.Time:
				if v != nil {
					builder.SetStartedAt(*v)
				}
			}
		case "finished_at":
			switch v := value.(type) {
			case time.Time:
				builder.SetFinishedAt(v)
			case *time.Time:
				if v != nil {
					builder.SetFinishedAt(*v)
				}
			}
		}
	}
	return nil
}

func applyTaskResumeUpdates(builder *db.ScreeningTaskResumeUpdate, updates map[string]any) error {
	for key, value := range updates {
		switch key {
		case "status":
			if status, ok := value.(consts.ScreeningTaskResumeStatus); ok {
				builder.SetStatus(string(status))
			} else if str, ok := value.(string); ok && str != "" {
				builder.SetStatus(str)
			}
		case "ranking":
			if value == nil {
				builder.ClearRanking()
				continue
			}
			builder.SetRanking(asInt(value))
		case "score":
			if value == nil {
				builder.ClearScore()
				continue
			}
			switch v := value.(type) {
			case float64:
				builder.SetScore(v)
			case float32:
				builder.SetScore(float64(v))
			case int:
				builder.SetScore(float64(v))
			}
		case "error_message":
			if str, ok := value.(string); ok {
				builder.SetErrorMessage(str)
			}
		case "processed_at":
			switch v := value.(type) {
			case time.Time:
				builder.SetProcessedAt(v)
			case *time.Time:
				if v != nil {
					builder.SetProcessedAt(*v)
				}
			}
		}
	}
	return nil
}

func asInt(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}
