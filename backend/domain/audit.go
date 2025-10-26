package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/pkg/web"
)

// AuditUsecase 审计日志用例接口
type AuditUsecase interface {
	List(ctx context.Context, req *ListAuditLogReq) (*ListAuditLogResp, error)
	GetByID(ctx context.Context, id string) (*AuditLog, error)
	GetStats(ctx context.Context, req *AuditStatsReq) (*AuditStatsResp, error)
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	CleanupOldLogs(ctx context.Context, days int) error
	GetOperationSummary(ctx context.Context, operatorID string, days int) (*OperationSummary, error)
	GetSecurityAlerts(ctx context.Context, hours int) ([]*SecurityAlert, error)
}

// AuditRepo 审计日志仓储接口
type AuditRepo interface {
	Create(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, req *ListAuditLogReq) ([]*AuditLog, *db.PageInfo, error)
	GetByID(ctx context.Context, id string) (*AuditLog, error)
	GetStats(ctx context.Context, req *AuditStatsReq) (*AuditStatsResp, error)
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	CleanupOldLogs(ctx context.Context, days int) error
}

// ListAuditLogReq 审计日志列表请求
type ListAuditLogReq struct {
	web.Pagination

	// 筛选条件
	OperatorType  consts.OperatorType   `json:"operator_type" query:"operator_type"`     // 操作者类型
	OperatorID    string                `json:"operator_id" query:"operator_id"`         // 操作者ID
	OperationType consts.OperationType  `json:"operation_type" query:"operation_type"`   // 操作类型
	ResourceType  consts.ResourceType   `json:"resource_type" query:"resource_type"`     // 资源类型
	ResourceID    string                `json:"resource_id" query:"resource_id"`         // 资源ID
	Status        consts.AuditLogStatus `json:"status" query:"status"`                   // 状态
	IP            string                `json:"ip" query:"ip"`                           // IP地址
	StartTime     *time.Time            `json:"start_time,omitempty" query:"start_time"` // 任务创建时间范围起始时间，可选
	EndTime       *time.Time            `json:"end_time,omitempty" query:"end_time"`     // 任务创建时间范围结束时间，可选
	Search        string                `json:"search" query:"search"`                   // 搜索关键词
}

// NormalizeEmptyPointers 规范化空指针字段，将指向空字符串的指针设置为nil
func (req *ListAuditLogReq) NormalizeEmptyPointers() {
	// 字符串字段已改为非指针类型，不需要处理
}

// ListAuditLogResp 审计日志列表响应
type ListAuditLogResp struct {
	*db.PageInfo
	AuditLogs []*AuditLog `json:"audit_logs"`
}

// AuditStatsReq 审计日志统计请求
type AuditStatsReq struct {
	StartTime *time.Time `json:"start_time" query:"start_time, omitempty"` // 开始时间
	EndTime   *time.Time `json:"end_time" query:"end_time, omitempty"`     // 结束时间
}

// AuditStatsResp 审计日志统计响应
type AuditStatsResp struct {
	TotalCount        int64                          `json:"total_count"`         // 总数
	SuccessCount      int64                          `json:"success_count"`       // 成功数
	FailedCount       int64                          `json:"failed_count"`        // 失败数
	OperatorTypeStats map[consts.OperatorType]int64  `json:"operator_type_stats"` // 操作者类型统计
	OperationStats    map[consts.OperationType]int64 `json:"operation_stats"`     // 操作类型统计
	ResourceStats     map[consts.ResourceType]int64  `json:"resource_stats"`      // 资源类型统计
	HourlyStats       []HourlyAuditStat              `json:"hourly_stats"`        // 按小时统计
}

// HourlyAuditStat 按小时统计
type HourlyAuditStat struct {
	Hour         int   `json:"hour"`          // 小时 (0-23)
	Count        int64 `json:"count"`         // 数量
	SuccessCount int64 `json:"success_count"` // 成功数量
	FailedCount  int64 `json:"failed_count"`  // 失败数量
}

// AuditLog 审计日志
type AuditLog struct {
	ID            string               `json:"id"`                      // 审计日志ID
	RequestID     *string              `json:"request_id,omitempty"`    // 请求ID
	SessionID     *string              `json:"session_id,omitempty"`    // 会话ID
	OperatorType  consts.OperatorType  `json:"operator_type"`           // 操作者类型
	OperatorID    *string              `json:"operator_id,omitempty"`   // 操作者ID
	OperatorName  *string              `json:"operator_name,omitempty"` // 操作者名称
	OperationType consts.OperationType `json:"operation_type"`          // 操作类型: create创建、update更新、delete删除、view查看（敏感数据）、login登录、logout登出
	// 资源类型: user用户、admin管理员、role角色、department部门、job_position职位、resume简历、
	// screening筛选任务、setting系统设置、attachment附件、conversation对话、message消息、
	// notification_setting通知设置、resume_mailbox_setting简历邮箱设置、resume_mailbox_statistic简历邮箱统计
	ResourceType   consts.ResourceType    `json:"resource_type"`
	ResourceID     *string                `json:"resource_id,omitempty"`   // 资源ID
	ResourceName   *string                `json:"resource_name,omitempty"` // 资源名称
	RequestMethod  string                 `json:"request_method"`          // 请求方法
	RequestPath    string                 `json:"request_path"`            // 请求路径
	RequestQuery   *string                `json:"request_query,omitempty"` // 请求查询参数
	RequestBody    *string                `json:"request_body,omitempty"`  // 请求体
	ResponseStatus int                    `json:"response_status"`         // 响应状态码
	ResponseBody   *string                `json:"response_body,omitempty"` // 响应体
	IP             string                 `json:"ip"`                      // IP地址
	UserAgent      *string                `json:"user_agent,omitempty"`    // 用户代理
	Country        *string                `json:"country,omitempty"`       // 国家
	Province       *string                `json:"province,omitempty"`      // 省份
	City           *string                `json:"city,omitempty"`          // 城市
	ISP            *string                `json:"isp,omitempty"`           // ISP
	Duration       int64                  `json:"duration"`                // 请求耗时(毫秒)
	Status         consts.AuditLogStatus  `json:"status"`                  // 状态
	ErrorMessage   *string                `json:"error_message,omitempty"` // 错误信息
	BusinessData   map[string]interface{} `json:"business_data,omitempty"` // 业务数据
	CreatedAt      int64                  `json:"created_at"`              // 创建时间
	UpdatedAt      int64                  `json:"updated_at"`              // 更新时间
}

// From 从数据库实体转换
func (a *AuditLog) From(e *db.AuditLog) *AuditLog {
	if e == nil {
		return a
	}

	a.ID = e.ID.String()
	if e.TraceID != "" {
		a.RequestID = &e.TraceID
	}
	if e.SessionID != "" {
		a.SessionID = &e.SessionID
	}
	a.OperatorType = e.OperatorType
	if e.OperatorID != uuid.Nil {
		operatorID := e.OperatorID.String()
		a.OperatorID = &operatorID
	}
	if e.OperatorName != "" {
		a.OperatorName = &e.OperatorName
	}
	a.OperationType = e.OperationType
	a.ResourceType = e.ResourceType
	if e.ResourceID != "" {
		a.ResourceID = &e.ResourceID
	}
	if e.ResourceName != "" {
		a.ResourceName = &e.ResourceName
	}
	a.RequestMethod = e.Method
	a.RequestPath = e.Path
	if e.QueryParams != "" {
		a.RequestQuery = &e.QueryParams
	}
	if e.RequestBody != "" {
		a.RequestBody = &e.RequestBody
	}
	a.ResponseStatus = e.StatusCode
	if e.ResponseBody != "" {
		a.ResponseBody = &e.ResponseBody
	}
	a.IP = e.IP
	if e.UserAgent != "" {
		a.UserAgent = &e.UserAgent
	}
	if e.Country != "" {
		a.Country = &e.Country
	}
	if e.Province != "" {
		a.Province = &e.Province
	}
	if e.City != "" {
		a.City = &e.City
	}
	if e.Isp != "" {
		a.ISP = &e.Isp
	}
	a.Duration = e.DurationMs
	a.Status = e.Status
	if e.ErrorMessage != "" {
		a.ErrorMessage = &e.ErrorMessage
	}
	// 解析业务数据JSON
	if e.BusinessData != "" {
		// 这里可以添加JSON解析逻辑
		a.BusinessData = map[string]interface{}{"raw": e.BusinessData}
	}
	a.CreatedAt = e.CreatedAt.Unix()
	a.UpdatedAt = e.UpdatedAt.Unix()

	return a
}

// OperationSummary 操作摘要统计
type OperationSummary struct {
	OperatorID    string                         `json:"operator_id"`    // 操作者ID
	Days          int                            `json:"days"`           // 统计天数
	TotalCount    int64                          `json:"total_count"`    // 总操作数
	SuccessCount  int64                          `json:"success_count"`  // 成功数
	FailedCount   int64                          `json:"failed_count"`   // 失败数
	OperationMap  map[consts.OperationType]int64 `json:"operation_map"`  // 操作类型统计
	ResourceMap   map[consts.ResourceType]int64  `json:"resource_map"`   // 资源类型统计
	DailyActivity []DailyActivity                `json:"daily_activity"` // 每日活动统计
}

// DailyActivity 每日活动统计
type DailyActivity struct {
	Date         string `json:"date"`          // 日期 (YYYY-MM-DD)
	Count        int64  `json:"count"`         // 操作数量
	SuccessCount int64  `json:"success_count"` // 成功数量
	FailedCount  int64  `json:"failed_count"`  // 失败数量
}

// SecurityAlert 安全告警
type SecurityAlert struct {
	Type        string  `json:"type"`                  // 告警类型
	Level       string  `json:"level"`                 // 告警级别 (high, medium, low)
	Title       string  `json:"title"`                 // 告警标题
	Description string  `json:"description"`           // 告警描述
	IP          string  `json:"ip,omitempty"`          // 相关IP地址
	OperatorID  *string `json:"operator_id,omitempty"` // 相关操作者ID
	Count       int64   `json:"count"`                 // 相关计数
	CreatedAt   int64   `json:"created_at"`            // 创建时间
}

// BatchDeleteAuditLogReq 批量删除审计日志请求
type BatchDeleteAuditLogReq struct {
	IDs []string `json:"ids" validate:"required,min=1,max=100"` // 要删除的日志ID列表
}

// CleanupOldLogsReq 清理旧日志请求
type CleanupOldLogsReq struct {
	Days int `json:"days" validate:"required,min=1,max=365"` // 保留天数
}
