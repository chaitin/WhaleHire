package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/pkg/web"
)

// ScreeningNodeRunUsecase 筛选节点运行业务用例接口
type ScreeningNodeRunUsecase interface {
	CreateNodeRun(ctx context.Context, req *CreateNodeRunReq) (*ScreeningNodeRun, error)
	UpdateNodeRun(ctx context.Context, req *UpdateNodeRunReq) (*ScreeningNodeRun, error)
	GetNodeRun(ctx context.Context, id uuid.UUID) (*ScreeningNodeRun, error)
	ListNodeRuns(ctx context.Context, req *ListNodeRunsReq) (*ListNodeRunsResp, error)
	GetNodeRunsByTaskResume(ctx context.Context, taskResumeID uuid.UUID) ([]*ScreeningNodeRun, error)
}

// ScreeningNodeRunRepo 筛选节点运行仓储接口
type ScreeningNodeRunRepo interface {
	Create(ctx context.Context, req *CreateNodeRunRepoReq) (*db.ScreeningNodeRun, error)
	Update(ctx context.Context, id uuid.UUID, fn func(tx *db.Tx, current *db.ScreeningNodeRun, updater *db.ScreeningNodeRunUpdateOne) error) (*db.ScreeningNodeRun, error)
	GetByID(ctx context.Context, id uuid.UUID) (*db.ScreeningNodeRun, error)
	List(ctx context.Context, req *ListNodeRunsRepoReq) ([]*db.ScreeningNodeRun, *db.PageInfo, error)
	GetByTaskResumeID(ctx context.Context, taskResumeID uuid.UUID) ([]*db.ScreeningNodeRun, error)
	GetByTaskResumeAndNode(ctx context.Context, taskResumeID uuid.UUID, nodeKey string, attemptNo int) (*db.ScreeningNodeRun, error)
}

// ScreeningNodeRun 筛选节点运行信息
// @Description 筛选节点运行记录，包含节点执行的详细信息和状态
type ScreeningNodeRun struct {
	// ID 节点运行记录唯一标识符
	ID uuid.UUID `json:"id"`
	// TaskID 关联的筛选任务ID
	TaskID uuid.UUID `json:"task_id"`
	// TaskResumeID 关联的任务简历ID
	TaskResumeID uuid.UUID `json:"task_resume_id"`
	// NodeKey 节点类型标识，支持的类型包括：TaskMetaDataNode、DispatcherNode、BasicInfoAgent、SkillAgent、ResponsibilityAgent、ExperienceAgent、EducationAgent、IndustryAgent、AggregatorAgent
	NodeKey string `json:"node_key"`
	// Status 节点运行状态：pending(待处理)、running(运行中)、completed(已完成)、failed(失败)
	Status consts.ScreeningNodeRunStatus `json:"status"`
	// AttemptNo 尝试次数，从1开始
	AttemptNo int `json:"attempt_no"`
	// TraceID 链路追踪ID，用于调试和监控
	TraceID *string `json:"trace_id,omitempty"`
	// AgentVersion Agent版本号
	AgentVersion *string `json:"agent_version,omitempty"`
	// ModelName 使用的模型名称
	ModelName *string `json:"model_name,omitempty"`
	// ModelProvider 模型提供商
	ModelProvider *string `json:"model_provider,omitempty"`
	// LLMParams LLM调用参数快照
	LLMParams map[string]interface{} `json:"llm_params,omitempty" swaggertype:"object"`
	// InputPayload 节点输入数据快照
	InputPayload map[string]interface{} `json:"input_payload,omitempty" swaggertype:"object"`
	// OutputPayload 节点运行输出数据，根据不同节点类型包含不同的数据结构
	// 各节点输出数据结构映射关系：
	// - TaskMetaDataNode: TaskMetaData 任务元数据，包含任务ID、简历ID、匹配任务ID和维度权重
	// - DispatcherNode: map[string]any 分发数据，包含各Agent的输入数据分发结果
	// - BasicInfoAgent: BasicMatchDetail 基本信息匹配详情，包含地点、薪资、部门等基础信息匹配结果
	// - EducationAgent: EducationMatchDetail 教育背景匹配详情，包含学历、专业、学校匹配分析
	// - ExperienceAgent: ExperienceMatchDetail 工作经验匹配详情，包含年限、职位、行业经验匹配
	// - IndustryAgent: IndustryMatchDetail 行业背景匹配详情，包含行业和公司匹配分析
	// - ResponsibilityAgent: ResponsibilityMatchDetail 职责匹配详情，包含岗位职责与工作经历的匹配分析
	// - SkillAgent: SkillMatchDetail 技能匹配详情，包含技能匹配、缺失技能和LLM分析结果
	// - AggregatorAgent: JobResumeMatch 综合匹配结果，包含各维度匹配详情和最终综合评分
	OutputPayload map[string]interface{} `json:"output_payload,omitempty" swaggertype:"object"`
	// ErrorMessage 错误信息，当Status为failed时包含具体错误描述
	ErrorMessage *string `json:"error_message,omitempty"`
	// TokensInput 输入token数量
	TokensInput *int64 `json:"tokens_input,omitempty"`
	// TokensOutput 输出token数量
	TokensOutput *int64 `json:"tokens_output,omitempty"`
	// TotalCost 调用总成本
	TotalCost *float64 `json:"total_cost,omitempty"`
	// StartedAt 开始执行时间
	StartedAt *time.Time `json:"started_at,omitempty"`
	// FinishedAt 完成执行时间
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	// DurationMs 执行耗时(毫秒)
	DurationMs *int `json:"duration_ms,omitempty"`
	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateNodeRunReq 创建节点运行请求
type CreateNodeRunReq struct {
	TaskID        uuid.UUID              `json:"task_id" validate:"required"`
	TaskResumeID  uuid.UUID              `json:"task_resume_id" validate:"required"`
	NodeKey       string                 `json:"node_key" validate:"required"`
	AttemptNo     int                    `json:"attempt_no"`
	TraceID       *string                `json:"trace_id,omitempty"`
	AgentVersion  *string                `json:"agent_version,omitempty"`
	ModelName     *string                `json:"model_name,omitempty"`
	ModelProvider *string                `json:"model_provider,omitempty"`
	LLMParams     map[string]interface{} `json:"llm_params,omitempty"`
	InputPayload  map[string]interface{} `json:"input_payload,omitempty"`
}

// CreateNodeRunRepoReq 创建节点运行仓储请求
type CreateNodeRunRepoReq struct {
	TaskID        uuid.UUID
	TaskResumeID  uuid.UUID
	NodeKey       string
	Status        consts.ScreeningNodeRunStatus
	AttemptNo     int
	TraceID       *string
	AgentVersion  *string
	ModelName     *string
	ModelProvider *string
	LLMParams     map[string]interface{}
	InputPayload  map[string]interface{}
	OutputPayload map[string]interface{}
}

// UpdateNodeRunReq 更新节点运行请求
type UpdateNodeRunReq struct {
	ID            uuid.UUID                      `json:"id" validate:"required"`
	Status        *consts.ScreeningNodeRunStatus `json:"status,omitempty"`
	OutputPayload map[string]interface{}         `json:"output_payload,omitempty"`
	ErrorMessage  *string                        `json:"error_message,omitempty"`
	TokensInput   *int64                         `json:"tokens_input,omitempty"`
	TokensOutput  *int64                         `json:"tokens_output,omitempty"`
	TotalCost     *float64                       `json:"total_cost,omitempty"`
	StartedAt     *time.Time                     `json:"started_at,omitempty"`
	FinishedAt    *time.Time                     `json:"finished_at,omitempty"`
	DurationMs    *int                           `json:"duration_ms,omitempty"`
}

// ListNodeRunsReq 查询节点运行列表请求
type ListNodeRunsReq struct {
	web.Pagination
	TaskID       *uuid.UUID                     `json:"task_id,omitempty" query:"task_id"`
	TaskResumeID *uuid.UUID                     `json:"task_resume_id,omitempty" query:"task_resume_id"`
	NodeKey      *string                        `json:"node_key,omitempty" query:"node_key"`
	Status       *consts.ScreeningNodeRunStatus `json:"status,omitempty" query:"status"`
}

// ListNodeRunsRepoReq 查询节点运行列表仓储请求
type ListNodeRunsRepoReq struct {
	*ListNodeRunsReq
}

// ListNodeRunsResp 查询节点运行列表响应
type ListNodeRunsResp struct {
	Items []*ScreeningNodeRun `json:"items"`
	*db.PageInfo
}

// CreateScreeningNodeRunReq 创建节点运行记录请求
type CreateScreeningNodeRunReq struct {
	TaskID        uuid.UUID  `json:"task_id" validate:"required"`
	TaskResumeID  uuid.UUID  `json:"task_resume_id" validate:"required"`
	NodeKey       string     `json:"node_key" validate:"required"`
	Status        string     `json:"status" validate:"required"`
	AttemptNo     int        `json:"attempt_no" validate:"required"`
	TraceID       string     `json:"trace_id" validate:"required"`
	AgentVersion  *string    `json:"agent_version,omitempty"`
	ModelName     *string    `json:"model_name,omitempty"`
	ModelProvider *string    `json:"model_provider,omitempty"`
	LlmParams     *string    `json:"llm_params,omitempty"`
	InputPayload  *string    `json:"input_payload,omitempty"`
	OutputPayload *string    `json:"output_payload,omitempty"`
	ErrorMessage  *string    `json:"error_message,omitempty"`
	TokensInput   *int       `json:"tokens_input,omitempty"`
	TokensOutput  *int       `json:"tokens_output,omitempty"`
	TotalCost     *float64   `json:"total_cost,omitempty"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	DurationMs    *int64     `json:"duration_ms,omitempty"`
}

// UpdateScreeningNodeRunReq 更新节点运行记录请求
type UpdateScreeningNodeRunReq struct {
	Status        *string    `json:"status,omitempty"`
	OutputPayload *string    `json:"output_payload,omitempty"`
	ErrorMessage  *string    `json:"error_message,omitempty"`
	TokensInput   *int       `json:"tokens_input,omitempty"`
	TokensOutput  *int       `json:"tokens_output,omitempty"`
	TotalCost     *float64   `json:"total_cost,omitempty"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	DurationMs    *int64     `json:"duration_ms,omitempty"`
}

// ListScreeningNodeRunReq 查询节点运行记录列表请求
type ListScreeningNodeRunReq struct {
	TaskID       *uuid.UUID      `json:"task_id,omitempty"`
	TaskResumeID *uuid.UUID      `json:"task_resume_id,omitempty"`
	NodeKey      *string         `json:"node_key,omitempty"`
	Status       *string         `json:"status,omitempty"`
	TraceID      *string         `json:"trace_id,omitempty"`
	StartTime    *time.Time      `json:"start_time,omitempty"`
	EndTime      *time.Time      `json:"end_time,omitempty"`
	Pagination   *web.Pagination `json:"pagination,omitempty"`
}

// ListScreeningNodeRunResp 查询节点运行记录列表响应
type ListScreeningNodeRunResp struct {
	Items      []*ScreeningNodeRun `json:"items"`
	Pagination *db.PageInfo        `json:"pagination"`
}

// From 从数据库实体转换
func (s *ScreeningNodeRun) From(e *db.ScreeningNodeRun) *ScreeningNodeRun {
	if e == nil {
		return s
	}
	s.ID = e.ID
	s.TaskID = e.TaskID
	s.TaskResumeID = e.TaskResumeID
	s.NodeKey = e.NodeKey
	s.Status = consts.ScreeningNodeRunStatus(e.Status)
	s.AttemptNo = e.AttemptNo

	// 处理可选字段的指针转换
	if e.TraceID != "" {
		s.TraceID = &e.TraceID
	}
	if e.AgentVersion != "" {
		s.AgentVersion = &e.AgentVersion
	}
	if e.ModelName != "" {
		s.ModelName = &e.ModelName
	}
	if e.ModelProvider != "" {
		s.ModelProvider = &e.ModelProvider
	}
	if e.ErrorMessage != "" {
		s.ErrorMessage = &e.ErrorMessage
	}

	s.LLMParams = e.LlmParams
	s.InputPayload = e.InputPayload
	s.OutputPayload = e.OutputPayload

	// 处理数值字段的指针转换
	if e.TokensInput != 0 {
		s.TokensInput = &e.TokensInput
	}
	if e.TokensOutput != 0 {
		s.TokensOutput = &e.TokensOutput
	}
	if e.TotalCost != 0 {
		s.TotalCost = &e.TotalCost
	}
	if e.DurationMs != 0 {
		s.DurationMs = &e.DurationMs
	}

	// 处理时间字段的指针转换
	if !e.StartedAt.IsZero() {
		s.StartedAt = &e.StartedAt
	}
	if !e.FinishedAt.IsZero() {
		s.FinishedAt = &e.FinishedAt
	}

	s.CreatedAt = e.CreatedAt
	s.UpdatedAt = e.UpdatedAt
	return s
}
