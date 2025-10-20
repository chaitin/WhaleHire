package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
)

// NotificationUsecase 通知用例接口
type NotificationUsecase interface {
	// PublishEvent 发布通知事件
	PublishEvent(ctx context.Context, payload NotificationPayload) error
	// PublishEventWithDelay 延迟发布通知事件
	PublishEventWithDelay(ctx context.Context, payload NotificationPayload, delay time.Duration) error
	// MarkAsDelivered 标记事件为已投递
	MarkAsDelivered(ctx context.Context, eventID uuid.UUID) error
	// MarkAsFailed 标记事件为失败
	MarkAsFailed(ctx context.Context, eventID uuid.UUID, errorMsg string) error
	// RetryEvent 重试事件
	RetryEvent(ctx context.Context, eventID uuid.UUID) error
	// GetEventsByStatus 根据状态获取事件
	GetEventsByStatus(ctx context.Context, status consts.NotificationStatus, limit int) ([]*NotificationEvent, error)
	// GetPendingEvents 获取待处理事件
	GetPendingEvents(ctx context.Context, limit int) ([]*NotificationEvent, error)
}

// NotificationEventRepo 通知事件仓储接口
type NotificationEventRepo interface {
	// Create 创建通知事件
	Create(ctx context.Context, event *NotificationEvent) error
	// GetByID 根据ID获取通知事件
	GetByID(ctx context.Context, id uuid.UUID) (*NotificationEvent, error)
	// GetByTraceID 根据TraceID获取通知事件
	GetByTraceID(ctx context.Context, traceID string) (*NotificationEvent, error)
	// GetPendingEvents 获取待处理的通知事件
	GetPendingEvents(ctx context.Context, limit int) ([]*NotificationEvent, error)
	// GetRetryableEvents 获取可重试的通知事件
	GetRetryableEvents(ctx context.Context, limit int) ([]*NotificationEvent, error)
	// Update 更新通知事件
	Update(ctx context.Context, event *NotificationEvent) error
	// UpdateStatus 更新通知事件状态
	UpdateStatus(ctx context.Context, id uuid.UUID, status consts.NotificationStatus, errorMsg string) error
	// MarkAsDelivered 标记为已投递
	MarkAsDelivered(ctx context.Context, id uuid.UUID) error
	// IncrementRetryCount 增加重试次数
	IncrementRetryCount(ctx context.Context, id uuid.UUID, errorMsg string) error
	// GetEventsByType 根据事件类型获取事件列表
	GetEventsByType(ctx context.Context, eventType consts.NotificationEventType, limit int) ([]*NotificationEvent, error)
	// GetEventsByStatus 根据状态获取事件列表
	GetEventsByStatus(ctx context.Context, status consts.NotificationStatus, limit int) ([]*NotificationEvent, error)
	// DeleteOldEvents 删除旧事件
	DeleteOldEvents(ctx context.Context, before time.Time) (int, error)
}

// NotificationEvent 通知事件领域实体
type NotificationEvent struct {
	ID          uuid.UUID                    `json:"id"`
	EventType   consts.NotificationEventType `json:"event_type"`
	Channel     consts.NotificationChannel   `json:"channel"`
	Status      consts.NotificationStatus    `json:"status"`
	Payload     map[string]interface{}       `json:"payload"`
	TemplateID  string                       `json:"template_id"`
	Target      string                       `json:"target"`
	RetryCount  int                          `json:"retry_count"`
	MaxRetry    int                          `json:"max_retry"`
	Timeout     int                          `json:"timeout"`
	LastError   string                       `json:"last_error"`
	TraceID     string                       `json:"trace_id"`
	CreatedAt   time.Time                    `json:"created_at"`
	ScheduledAt *time.Time                   `json:"scheduled_at"`
	DeliveredAt *time.Time                   `json:"delivered_at"`
	UpdatedAt   time.Time                    `json:"updated_at"`
}

// NotificationPayload 通知负载接口
type NotificationPayload interface {
	GetEventType() consts.NotificationEventType
	GetTemplateID() string
	GetPayload() map[string]interface{}
}

// ResumeParseCompletedPayload 简历解析完成事件载荷
type ResumeParseCompletedPayload struct {
	ResumeID   uuid.UUID `json:"resume_id"`
	UserID     uuid.UUID `json:"user_id"`
	FileName   string    `json:"file_name"`
	ParsedAt   time.Time `json:"parsed_at"`
	Success    bool      `json:"success"`
	ErrorMsg   string    `json:"error_msg,omitempty"`
	WebhookURL string    `json:"webhook_url"`
}

func (p ResumeParseCompletedPayload) GetEventType() consts.NotificationEventType {
	return consts.NotificationEventTypeResumeParseCompleted
}

func (p ResumeParseCompletedPayload) GetTemplateID() string {
	return "resume_parse_completed"
}

func (p ResumeParseCompletedPayload) GetTarget() string {
	return p.WebhookURL
}

func (p ResumeParseCompletedPayload) GetPayload() map[string]interface{} {
	return map[string]interface{}{
		"resume_id": p.ResumeID,
		"user_id":   p.UserID,
		"file_name": p.FileName,
		"parsed_at": p.ParsedAt,
		"success":   p.Success,
		"error_msg": p.ErrorMsg,
	}
}

// JobMatchingCompletedPayload 岗位匹配完成通知负载
type JobMatchingCompletedPayload struct {
	ResumeID   uuid.UUID `json:"resume_id"`
	JobID      uuid.UUID `json:"job_id"`
	UserID     uuid.UUID `json:"user_id"`
	MatchScore float64   `json:"match_score"`
	MatchedAt  time.Time `json:"matched_at"`
	WebhookURL string    `json:"webhook_url"`
}

func (p JobMatchingCompletedPayload) GetEventType() consts.NotificationEventType {
	return consts.NotificationEventTypeJobMatchingCompleted
}

func (p JobMatchingCompletedPayload) GetTemplateID() string {
	return "job_matching_completed"
}

func (p JobMatchingCompletedPayload) GetTarget() string {
	return p.WebhookURL
}

func (p JobMatchingCompletedPayload) GetPayload() map[string]interface{} {
	return map[string]interface{}{
		"resume_id":   p.ResumeID,
		"job_id":      p.JobID,
		"user_id":     p.UserID,
		"match_score": p.MatchScore,
		"matched_at":  p.MatchedAt,
	}
}

// ScreeningTaskCompletedPayload 筛选任务完成通知负载
type ScreeningTaskCompletedPayload struct {
	TaskID      uuid.UUID `json:"task_id"`
	UserID      uuid.UUID `json:"user_id"`
	UserName    string    `json:"user_name"`
	JobID       uuid.UUID `json:"job_id"`
	JobName     string    `json:"job_name"`
	TotalCount  int       `json:"total_count"`
	PassedCount int       `json:"passed_count"`
	AvgScore    float64   `json:"avg_score"`
	CompletedAt time.Time `json:"completed_at"`
	WebhookURL  string    `json:"webhook_url"`
}

func (p ScreeningTaskCompletedPayload) GetEventType() consts.NotificationEventType {
	return consts.NotificationEventTypeScreeningTaskCompleted
}

func (p ScreeningTaskCompletedPayload) GetTemplateID() string {
	return "screening_task_completed"
}

func (p ScreeningTaskCompletedPayload) GetTarget() string {
	return p.WebhookURL
}

func (p ScreeningTaskCompletedPayload) GetPayload() map[string]interface{} {
	return map[string]interface{}{
		"task_id":      p.TaskID,
		"user_id":      p.UserID,
		"user_name":    p.UserName,
		"job_id":       p.JobID,
		"job_name":     p.JobName,
		"total_count":  p.TotalCount,
		"passed_count": p.PassedCount,
		"avg_score":    p.AvgScore,
		"completed_at": p.CompletedAt,
	}
}

// BatchResumeParseCompletedPayload 批量简历解析完成通知负载
type BatchResumeParseCompletedPayload struct {
	TaskID         string    `json:"task_id"`
	UploaderID     uuid.UUID `json:"uploader_id"`
	UploaderName   string    `json:"uploader_name"`
	TotalCount     int       `json:"total_count"`
	SuccessCount   int       `json:"success_count"`
	FailedCount    int       `json:"failed_count"`
	JobPositionIDs []string  `json:"job_position_ids,omitempty"`
	Source         string    `json:"source,omitempty"`
	CompletedAt    time.Time `json:"completed_at"`
	WebhookURL     string    `json:"webhook_url"`
}

func (p BatchResumeParseCompletedPayload) GetEventType() consts.NotificationEventType {
	return consts.NotificationEventTypeBatchResumeParseCompleted
}

func (p BatchResumeParseCompletedPayload) GetTemplateID() string {
	return "batch_resume_parse_completed"
}

func (p BatchResumeParseCompletedPayload) GetTarget() string {
	return p.WebhookURL
}

func (p BatchResumeParseCompletedPayload) GetPayload() map[string]interface{} {
	return map[string]interface{}{
		"task_id":          p.TaskID,
		"uploader_id":      p.UploaderID,
		"uploader_name":    p.UploaderName,
		"total_count":      p.TotalCount,
		"success_count":    p.SuccessCount,
		"failed_count":     p.FailedCount,
		"job_position_ids": p.JobPositionIDs,
		"source":           p.Source,
		"completed_at":     p.CompletedAt,
	}
}

// IsRetryable 判断事件是否可重试
func (e *NotificationEvent) IsRetryable() bool {
	return e.Status == consts.NotificationStatusFailed && e.RetryCount < e.MaxRetry
}

// CanDeliver 判断事件是否可以投递
func (e *NotificationEvent) CanDeliver() bool {
	return e.Status == consts.NotificationStatusPending || e.IsRetryable()
}

// MarkAsDelivering 标记为投递中
func (e *NotificationEvent) MarkAsDelivering() {
	e.Status = consts.NotificationStatusDelivering
	e.UpdatedAt = time.Now()
}

// MarkAsDelivered 标记为已投递
func (e *NotificationEvent) MarkAsDelivered() {
	now := time.Now()
	e.Status = consts.NotificationStatusDelivered
	e.DeliveredAt = &now
	e.UpdatedAt = now
	e.LastError = ""
}

// MarkAsFailed 标记为失败
func (e *NotificationEvent) MarkAsFailed(errorMsg string) {
	e.Status = consts.NotificationStatusFailed
	e.LastError = errorMsg
	e.RetryCount++
	e.UpdatedAt = time.Now()
}

// GetIdempotencyKey 获取幂等性键
func (e *NotificationEvent) GetIdempotencyKey() string {
	return e.ID.String()
}
