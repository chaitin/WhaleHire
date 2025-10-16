package domain

import (
	"context"
	"time"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/pkg/web"
	"github.com/google/uuid"
)

// ScreeningUsecase 筛选业务用例接口
type ScreeningUsecase interface {
	CreateScreeningTask(ctx context.Context, req *CreateScreeningTaskReq) (*CreateScreeningTaskResp, error)
	StartScreeningTask(ctx context.Context, req *StartScreeningTaskReq) (*StartScreeningTaskResp, error)
	CancelScreeningTask(ctx context.Context, req *CancelScreeningTaskReq) (*CancelScreeningTaskResp, error)
	DeleteScreeningTask(ctx context.Context, req *DeleteScreeningTaskReq) (*DeleteScreeningTaskResp, error)
	GetScreeningTask(ctx context.Context, req *GetScreeningTaskReq) (*GetScreeningTaskResp, error)
	ListScreeningTasks(ctx context.Context, req *ListScreeningTasksReq) (*ListScreeningTasksResp, error)
	GetScreeningResult(ctx context.Context, req *GetScreeningResultReq) (*GetScreeningResultResp, error)
	ListScreeningResults(ctx context.Context, req *ListScreeningResultsReq) (*ListScreeningResultsResp, error)
	GetScreeningMetrics(ctx context.Context, req *GetScreeningMetricsReq) (*GetScreeningMetricsResp, error)
	GetTaskProgress(ctx context.Context, req *GetTaskProgressReq) (*GetTaskProgressResp, error)
	GetResumeProgress(ctx context.Context, req *GetResumeProgressReq) (*GetResumeProgressResp, error)
	GetNodeRuns(ctx context.Context, req *GetNodeRunsReq) (*GetNodeRunsResp, error)
}

// ScreeningRepo 筛选数据访问接口
type ScreeningRepo interface {
	// 基础CRUD
	CreateScreeningTask(ctx context.Context, task *db.ScreeningTask) (*db.ScreeningTask, error)
	GetScreeningTask(ctx context.Context, id uuid.UUID) (*db.ScreeningTask, error)
	ListScreeningTasks(ctx context.Context, filter *ScreeningTaskFilter) ([]*db.ScreeningTask, *db.PageInfo, error)
	UpdateScreeningTask(ctx context.Context, id uuid.UUID, updates map[string]any) error
	DeleteScreeningTask(ctx context.Context, id uuid.UUID) error

	CreateScreeningTaskResume(ctx context.Context, taskResume *db.ScreeningTaskResume) (*db.ScreeningTaskResume, error)
	GetScreeningTaskResume(ctx context.Context, taskID, resumeID uuid.UUID) (*db.ScreeningTaskResume, error)
	ListScreeningTaskResumes(ctx context.Context, filter *ScreeningTaskResumeFilter) ([]*db.ScreeningTaskResume, *db.PageInfo, error)
	UpdateScreeningTaskResume(ctx context.Context, taskID, resumeID uuid.UUID, updates map[string]any) error

	CreateScreeningResult(ctx context.Context, result *db.ScreeningResult) (*db.ScreeningResult, error)
	GetScreeningResult(ctx context.Context, taskID, resumeID uuid.UUID) (*db.ScreeningResult, error)
	ListScreeningResults(ctx context.Context, filter *ScreeningResultFilter) ([]*db.ScreeningResult, *db.PageInfo, error)

	CreateScreeningRunMetric(ctx context.Context, metric *db.ScreeningRunMetric) (*db.ScreeningRunMetric, error)
	GetScreeningRunMetric(ctx context.Context, taskID uuid.UUID) (*db.ScreeningRunMetric, error)
}

// =========================
// 请求与响应模型
// =========================

// CreateScreeningTaskReq 创建筛选任务请求
type CreateScreeningTaskReq struct {
	// JobPositionID 职位ID，指定要筛选的目标职位
	JobPositionID uuid.UUID `json:"job_position_id" validate:"required"`
	// ResumeIDs 简历ID列表，指定要筛选的简历，至少需要一个简历
	ResumeIDs []uuid.UUID `json:"resume_ids" validate:"required,min=1"`
	// CreatedBy 创建者用户ID，标识任务创建人
	CreatedBy uuid.UUID `json:"created_by"`
	// Notes 任务备注信息，可选的任务说明或描述
	Notes string `json:"notes,omitempty"`
	// DimensionWeights 维度权重配置，用于控制各个匹配维度的重要性
	// 支持的维度包括: skill(技能,默认0.35), responsibility(职责,默认0.20), experience(经验,默认0.20),
	// education(教育,默认0.15), industry(行业,默认0.07), basic(基本信息,默认0.03)
	// 如果不提供，系统将使用默认权重配置
	// 示例: {"skill": 0.4, "responsibility": 0.25, "experience": 0.2, "education": 0.1, "industry": 0.05, "basic": 0.0}
	DimensionWeights map[string]float64 `json:"dimension_weights,omitempty"`
	// LLMConfig 用户自定义LLM配置，支持OpenAI等模型配置
	// 示例: {"model_type": "openai", "model_name": "gpt-4", "api_key": "sk-xxx", "base_url": "https://api.openai.com/v1"}
	// 如果不提供，系统将使用默认配置
	LLMConfig map[string]any `json:"llm_config,omitempty"`
	// AgentVersion 智能匹配代理版本号，用于版本兼容性检查
	// 必须与当前系统的匹配服务版本一致才能执行任务
	// 示例: "1.0.0"
	AgentVersion string `json:"agent_version,omitempty"`
}

// CreateScreeningTaskResp 创建筛选任务响应
type CreateScreeningTaskResp struct {
	// TaskID 创建成功的筛选任务ID
	TaskID uuid.UUID `json:"task_id"`
}

// StartScreeningTaskReq 启动筛选任务请求
type StartScreeningTaskReq struct {
	// TaskID 要启动的筛选任务ID
	TaskID uuid.UUID `json:"task_id" validate:"required"`
}

// StartScreeningTaskResp 启动筛选任务响应
type StartScreeningTaskResp struct {
	// TaskID 已启动的筛选任务ID
	TaskID uuid.UUID `json:"task_id"`
}

// CancelScreeningTaskReq 取消筛选任务请求
type CancelScreeningTaskReq struct {
	// TaskID 要取消的筛选任务ID
	TaskID uuid.UUID `json:"task_id" validate:"required"`
}

// CancelScreeningTaskResp 取消筛选任务响应
type CancelScreeningTaskResp struct {
	// TaskID 已取消的筛选任务ID
	TaskID uuid.UUID `json:"task_id"`
	// Message 取消操作的结果消息
	Message string `json:"message"`
}

// DeleteScreeningTaskReq 删除筛选任务请求
type DeleteScreeningTaskReq struct {
	// TaskID 要删除的筛选任务ID
	TaskID uuid.UUID `json:"task_id" validate:"required"`
}

// DeleteScreeningTaskResp 删除筛选任务响应
type DeleteScreeningTaskResp struct {
	// TaskID 已删除的筛选任务ID
	TaskID uuid.UUID `json:"task_id"`
	// Message 删除操作的结果消息
	Message string `json:"message"`
}

// GetScreeningTaskReq 获取筛选任务详情请求
type GetScreeningTaskReq struct {
	// TaskID 要查询的筛选任务ID
	TaskID uuid.UUID `json:"task_id" validate:"required"`
}

// GetScreeningTaskResp 获取筛选任务详情响应
type GetScreeningTaskResp struct {
	// Task 筛选任务基本信息
	Task *ScreeningTask `json:"task"`
	// Resumes 任务关联的简历列表及其处理状态
	Resumes []*ScreeningTaskResume `json:"resumes"`
	// Metrics 任务运行指标数据，可选
	Metrics *ScreeningRunMetric `json:"metrics,omitempty"`
}

// ListScreeningTasksReq 获取筛选任务列表请求
type ListScreeningTasksReq struct {
	// JobPositionID 职位ID过滤条件，可选
	JobPositionID *uuid.UUID `json:"job_position_id,omitempty"`
	// Status 任务状态过滤条件，可选
	Status *consts.ScreeningTaskStatus `json:"status,omitempty"`
	// CreatedBy 创建者用户ID过滤条件，可选
	CreatedBy *uuid.UUID `json:"created_by,omitempty"`
	// StartTime 任务创建时间范围起始时间，可选
	StartTime *time.Time `json:"start_time,omitempty"`
	// EndTime 任务创建时间范围结束时间，可选
	EndTime *time.Time `json:"end_time,omitempty"`
	web.Pagination
}

// ListScreeningTasksResp 获取筛选任务列表响应
type ListScreeningTasksResp struct {
	// Items 筛选任务列表
	Items []*ScreeningTask `json:"items"`
	*db.PageInfo
}

// GetScreeningResultReq 获取筛选结果请求
type GetScreeningResultReq struct {
	// TaskID 筛选任务ID
	TaskID uuid.UUID `json:"task_id" validate:"required"`
	// ResumeID 简历ID
	ResumeID uuid.UUID `json:"resume_id" validate:"required"`
}

// GetScreeningResultResp 获取筛选结果响应
type GetScreeningResultResp struct {
	// Result 筛选结果详情
	Result *ScreeningResult `json:"result"`
}

// ListScreeningResultsReq 获取筛选结果列表请求
type ListScreeningResultsReq struct {
	// TaskID 筛选任务ID过滤条件，可选
	TaskID *uuid.UUID `json:"task_id,omitempty"`
	// ResumeID 简历ID过滤条件，可选
	ResumeID *uuid.UUID `json:"resume_id,omitempty"`
	// MinScore 最小匹配分数过滤条件，可选
	MinScore *float64 `json:"min_score,omitempty"`
	// MaxScore 最大匹配分数过滤条件，可选
	MaxScore *float64 `json:"max_score,omitempty"`
	web.Pagination
}

// ListScreeningResultsResp 获取筛选结果列表响应
type ListScreeningResultsResp struct {
	// Items 筛选结果列表
	Items []*ScreeningResult `json:"results"`
	*db.PageInfo
}

// GetScreeningMetricsReq 获取筛选指标请求
type GetScreeningMetricsReq struct {
	// TaskID 筛选任务ID
	TaskID uuid.UUID `json:"task_id" validate:"required"`
}

// GetScreeningMetricsResp 获取筛选指标响应
type GetScreeningMetricsResp struct {
	// Metrics 筛选任务运行指标数据
	Metrics *ScreeningRunMetric `json:"metrics"`
}

// GetTaskProgressReq 获取任务进度请求
type GetTaskProgressReq struct {
	// TaskID 筛选任务ID
	TaskID uuid.UUID `json:"task_id" validate:"required"`
}

// GetTaskProgressResp 获取任务进度响应
type GetTaskProgressResp struct {
	// TaskID 筛选任务ID
	TaskID uuid.UUID `json:"task_id"`
	// Status 任务当前状态
	Status consts.ScreeningTaskStatus `json:"status"`
	// ResumeTotal 简历总数
	ResumeTotal int `json:"resume_total"`
	// ResumeProcessed 已处理简历数量
	ResumeProcessed int `json:"resume_processed"`
	// ResumeSucceeded 处理成功简历数量
	ResumeSucceeded int `json:"resume_succeeded"`
	// ResumeFailed 处理失败简历数量
	ResumeFailed int `json:"resume_failed"`
	// ProgressPercent 任务完成百分比 (0-100)
	ProgressPercent float64 `json:"progress_percent"`
	// StartedAt 任务开始时间，可选
	StartedAt *time.Time `json:"started_at,omitempty"`
	// EstimatedFinish 预计完成时间，可选
	EstimatedFinish *time.Time `json:"estimated_finish,omitempty"`
}

// GetResumeProgressReq 获取简历处理进度请求
type GetResumeProgressReq struct {
	// TaskID 筛选任务ID
	TaskID uuid.UUID `json:"task_id" validate:"required"`
	// ResumeID 简历ID
	ResumeID uuid.UUID `json:"resume_id" validate:"required"`
}

// GetResumeProgressResp 获取简历处理进度响应
type GetResumeProgressResp struct {
	// TaskID 筛选任务ID
	TaskID uuid.UUID `json:"task_id"`
	// ResumeID 简历ID
	ResumeID uuid.UUID `json:"resume_id"`
	// Status 简历处理状态
	Status consts.ScreeningTaskResumeStatus `json:"status"`
	// Score 匹配分数，处理成功时有值
	Score *float64 `json:"score,omitempty"`
	// ErrorMessage 错误信息，处理失败时有值
	ErrorMessage string `json:"error_message,omitempty"`
	// ProcessedAt 处理完成时间，可选
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	// CreatedAt 记录创建时间
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 记录更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// GetNodeRunsReq 获取节点运行记录请求
type GetNodeRunsReq struct {
	// TaskID 筛选任务ID
	TaskID uuid.UUID `json:"task_id" validate:"required"`
	// ResumeID 简历ID
	ResumeID uuid.UUID `json:"resume_id" validate:"required"`
}

// GetNodeRunsResp 获取节点运行记录响应
type GetNodeRunsResp struct {
	// AgentStatus Agent状态映射，Key为Agent名称，Value为对应的节点运行记录
	AgentStatus map[string]*ScreeningNodeRun `json:"agent_status"`
}

// =========================
// API 展示模型
// =========================

// ScreeningTask 筛选任务展示模型
type ScreeningTask struct {
	// ID 筛选任务唯一标识符
	ID uuid.UUID `json:"id"`
	// JobPositionID 关联的职位ID
	JobPositionID uuid.UUID `json:"job_position_id"`
	// CreatedBy 任务创建者用户ID
	CreatedBy uuid.UUID `json:"created_by"`
	// CreatorName 任务创建者用户名，可选
	CreatorName *string `json:"creator_name,omitempty"`
	// Status 任务当前状态
	Status consts.ScreeningTaskStatus `json:"status"`
	// Notes 任务备注信息，可选
	Notes string `json:"notes,omitempty"`
	// DimensionWeights 维度权重配置，可选
	DimensionWeights map[string]float64 `json:"dimension_weights,omitempty"`
	// LLMConfig 用户自定义LLM配置，可选
	LLMConfig map[string]any `json:"llm_config,omitempty"`
	// ResumeTotal 简历总数
	ResumeTotal int `json:"resume_total"`
	// ResumeProcessed 已处理简历数量
	ResumeProcessed int `json:"resume_processed"`
	// ResumeSucceeded 处理成功简历数量
	ResumeSucceeded int `json:"resume_succeeded"`
	// ResumeFailed 处理失败简历数量
	ResumeFailed int `json:"resume_failed"`
	// AgentVersion 智能匹配代理版本号，可选
	AgentVersion string `json:"agent_version,omitempty"`
	// StartedAt 任务开始时间，可选
	StartedAt *time.Time `json:"started_at,omitempty"`
	// FinishedAt 任务完成时间，可选
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	// CreatedAt 记录创建时间
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 记录更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// From 方法用于从 db 实体转换
func (st *ScreeningTask) From(dbTask *db.ScreeningTask) *ScreeningTask {
	if dbTask == nil {
		return nil
	}

	st.ID = dbTask.ID
	st.JobPositionID = dbTask.JobPositionID
	st.CreatedBy = dbTask.CreatedBy

	// 从 Creator 边获取用户名
	if dbTask.Edges.Creator != nil {
		st.CreatorName = &dbTask.Edges.Creator.Username
	}

	st.Status = consts.ScreeningTaskStatus(dbTask.Status)
	st.Notes = dbTask.Notes

	// 转换 DimensionWeights
	if dbTask.DimensionWeights != nil {
		st.DimensionWeights = make(map[string]float64)
		for k, v := range dbTask.DimensionWeights {
			if f, ok := v.(float64); ok {
				st.DimensionWeights[k] = f
			}
		}
	}

	st.LLMConfig = dbTask.LlmConfig
	st.ResumeTotal = dbTask.ResumeTotal
	st.ResumeProcessed = dbTask.ResumeProcessed
	st.ResumeSucceeded = dbTask.ResumeSucceeded
	st.ResumeFailed = dbTask.ResumeFailed
	st.AgentVersion = dbTask.AgentVersion

	// 处理时间指针类型
	if !dbTask.StartedAt.IsZero() {
		st.StartedAt = &dbTask.StartedAt
	}
	if !dbTask.FinishedAt.IsZero() {
		st.FinishedAt = &dbTask.FinishedAt
	}

	st.CreatedAt = dbTask.CreatedAt
	st.UpdatedAt = dbTask.UpdatedAt

	return st
}

// ScreeningTaskResume 筛选任务简历关联展示模型
type ScreeningTaskResume struct {
	// TaskID 筛选任务ID
	TaskID uuid.UUID `json:"task_id"`
	// ResumeID 简历ID
	ResumeID uuid.UUID `json:"resume_id"`
	// Status 简历处理状态
	Status consts.ScreeningTaskResumeStatus `json:"status"`
	// Ranking 简历在任务中的排名，可选
	Ranking *int `json:"ranking,omitempty"`
	// Score 匹配分数，处理成功时有值
	Score *float64 `json:"score,omitempty"`
	// ErrorMessage 错误信息，处理失败时有值
	ErrorMessage string `json:"error_message,omitempty"`
	// ProcessedAt 处理完成时间，可选
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	// CreatedAt 记录创建时间
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 记录更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// From 方法用于从 db 实体转换
func (str *ScreeningTaskResume) From(dbTaskResume *db.ScreeningTaskResume) *ScreeningTaskResume {
	if dbTaskResume == nil {
		return nil
	}

	str.TaskID = dbTaskResume.TaskID
	str.ResumeID = dbTaskResume.ResumeID
	str.Status = consts.ScreeningTaskResumeStatus(dbTaskResume.Status)
	str.ErrorMessage = dbTaskResume.ErrorMessage

	// 处理指针类型
	if dbTaskResume.Ranking != 0 {
		str.Ranking = &dbTaskResume.Ranking
	}
	if dbTaskResume.Score != 0 {
		str.Score = &dbTaskResume.Score
	}
	if !dbTaskResume.ProcessedAt.IsZero() {
		str.ProcessedAt = &dbTaskResume.ProcessedAt
	}

	str.CreatedAt = dbTaskResume.CreatedAt
	str.UpdatedAt = dbTaskResume.UpdatedAt

	return str
}

// ScreeningResult 筛选结果展示模型
type ScreeningResult struct {
	// TaskID 筛选任务ID
	TaskID uuid.UUID `json:"task_id"`
	// JobPositionID 职位ID
	JobPositionID uuid.UUID `json:"job_position_id"`
	// ResumeID 简历ID
	ResumeID uuid.UUID `json:"resume_id"`
	// OverallScore 总体匹配分数 (0-100)
	OverallScore float64 `json:"overall_score"`
	// MatchLevel 匹配等级 (非常匹配/高匹配/一般匹配/低匹配/不匹配)
	MatchLevel consts.MatchLevel `json:"match_level"`
	// DimensionScores 各维度匹配分数，可选
	DimensionScores map[string]float64 `json:"dimension_scores,omitempty"`
	// BasicDetail 基本信息匹配详情，包含地点、薪资等基础信息的匹配分析，可选
	BasicDetail *BasicMatchDetail `json:"basic_detail,omitempty" swaggertype:"object"`
	// EducationDetail 教育背景匹配详情，包含学历、专业、院校等教育信息的匹配分析，可选
	EducationDetail *EducationMatchDetail `json:"education_detail,omitempty" swaggertype:"object"`
	// ExperienceDetail 工作经验匹配详情，包含工作年限、职位、行业等经验信息的匹配分析，可选
	ExperienceDetail *ExperienceMatchDetail `json:"experience_detail,omitempty" swaggertype:"object"`
	// IndustryDetail 行业背景匹配详情，包含行业经验、公司背景等行业信息的匹配分析，可选
	IndustryDetail *IndustryMatchDetail `json:"industry_detail,omitempty" swaggertype:"object"`
	// Responsibility 职责匹配详情，包含工作职责、项目经验等职责信息的匹配分析，可选
	Responsibility *ResponsibilityMatchDetail `json:"responsibility_detail,omitempty" swaggertype:"object"`
	// SkillDetail 技能匹配详情，包含技术技能、工具使用等技能信息的匹配分析，可选
	SkillDetail *SkillMatchDetail `json:"skill_detail,omitempty" swaggertype:"object"`
	// Recommendations 推荐建议列表，可选
	Recommendations []string `json:"recommendations,omitempty"`
	// TraceID 追踪ID，用于调试和日志关联，可选
	TraceID string `json:"trace_id,omitempty"`
	// RuntimeMetadata 运行时元数据，可选
	RuntimeMetadata map[string]any `json:"runtime_metadata,omitempty"`
	// SubAgentVersions 子代理版本信息，可选
	SubAgentVersions map[string]any `json:"sub_agent_versions,omitempty"`
	// MatchedAt 匹配完成时间
	MatchedAt time.Time `json:"matched_at"`
	// CreatedAt 记录创建时间
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 记录更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// From 方法用于从 db 实体转换
func (sr *ScreeningResult) From(dbResult *db.ScreeningResult) *ScreeningResult {
	if dbResult == nil {
		return nil
	}

	sr.TaskID = dbResult.TaskID
	sr.JobPositionID = dbResult.JobPositionID
	sr.ResumeID = dbResult.ResumeID
	sr.OverallScore = dbResult.OverallScore
	sr.MatchLevel = consts.MatchLevel(dbResult.MatchLevel)

	// 转换 DimensionScores
	if dbResult.DimensionScores != nil {
		sr.DimensionScores = make(map[string]float64)
		for k, v := range dbResult.DimensionScores {
			if f, ok := v.(float64); ok {
				sr.DimensionScores[k] = f
			}
		}
	}

	// 转换各个详细匹配信息
	sr.BasicDetail = convertToBasicMatchDetail(dbResult.BasicDetail)
	sr.SkillDetail = convertToSkillMatchDetail(dbResult.SkillDetail)
	sr.Responsibility = convertToResponsibilityMatchDetail(dbResult.ResponsibilityDetail)
	sr.ExperienceDetail = convertToExperienceMatchDetail(dbResult.ExperienceDetail)
	sr.EducationDetail = convertToEducationMatchDetail(dbResult.EducationDetail)
	sr.IndustryDetail = convertToIndustryMatchDetail(dbResult.IndustryDetail)

	sr.Recommendations = dbResult.Recommendations
	sr.TraceID = dbResult.TraceID
	sr.RuntimeMetadata = dbResult.RuntimeMetadata
	sr.SubAgentVersions = dbResult.SubAgentVersions
	sr.MatchedAt = dbResult.MatchedAt
	sr.CreatedAt = dbResult.CreatedAt
	sr.UpdatedAt = dbResult.UpdatedAt

	return sr
}

// ScreeningRunMetric 筛选任务运行指标展示模型
type ScreeningRunMetric struct {
	// TaskID 筛选任务ID
	TaskID uuid.UUID `json:"task_id"`
	// AvgScore 平均匹配分数
	AvgScore float64 `json:"avg_score"`
	// Histogram 分数分布直方图，可选
	Histogram map[string]float64 `json:"histogram,omitempty"`
	// TokensInput 输入Token数量
	TokensInput int64 `json:"tokens_input"`
	// TokensOutput 输出Token数量
	TokensOutput int64 `json:"tokens_output"`
	// TotalCost 总成本费用
	TotalCost float64 `json:"total_cost"`
	// CreatedAt 记录创建时间
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 记录更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// From 方法用于从 db 实体转换
func (srm *ScreeningRunMetric) From(dbMetric *db.ScreeningRunMetric) *ScreeningRunMetric {
	if dbMetric == nil {
		return nil
	}

	srm.TaskID = dbMetric.TaskID
	srm.AvgScore = dbMetric.AvgScore

	// 转换 Histogram
	if dbMetric.Histogram != nil {
		srm.Histogram = make(map[string]float64)
		for k, v := range dbMetric.Histogram {
			if f, ok := v.(float64); ok {
				srm.Histogram[k] = f
			}
		}
	}

	srm.TokensInput = dbMetric.TokensInput
	srm.TokensOutput = dbMetric.TokensOutput
	srm.TotalCost = dbMetric.TotalCost
	srm.CreatedAt = dbMetric.CreatedAt
	srm.UpdatedAt = dbMetric.UpdatedAt

	return srm
}

// =========================
// 数据过滤模型
// =========================

type ScreeningTaskFilter struct {
	JobPositionID *uuid.UUID
	Status        *consts.ScreeningTaskStatus
	CreatedBy     *uuid.UUID
	StartTime     *time.Time
	EndTime       *time.Time
	Page          int
	PageSize      int
}

type ScreeningTaskResumeFilter struct {
	TaskID   *uuid.UUID
	ResumeID *uuid.UUID
	Status   *consts.ScreeningTaskResumeStatus
	Page     int
	PageSize int
}

type ScreeningResultFilter struct {
	TaskID   *uuid.UUID
	ResumeID *uuid.UUID
	MinScore *float64
	MaxScore *float64
	Page     int
	PageSize int
}
