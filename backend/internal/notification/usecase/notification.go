package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/internal/queue"
)

// notificationUsecase 通知用例实现
type notificationUsecase struct {
	repo        domain.NotificationEventRepo
	settingRepo domain.NotificationSettingUsecase

	producer queue.Producer
	logger   *slog.Logger
}

// NewNotificationUsecase 创建通知用例实例
func NewNotificationUsecase(
	repo domain.NotificationEventRepo,
	settingRepo domain.NotificationSettingUsecase,
	producer queue.Producer,
	logger *slog.Logger,
) domain.NotificationUsecase {
	return &notificationUsecase{
		repo:        repo,
		settingRepo: settingRepo,
		producer:    producer,
		logger:      logger,
	}
}

// PublishEvent 发布通知事件
func (u *notificationUsecase) PublishEvent(ctx context.Context, payload domain.NotificationPayload) error {
	return u.PublishEventWithDelay(ctx, payload, 0)
}

// PublishEventWithDelay 延迟发布通知事件
func (u *notificationUsecase) PublishEventWithDelay(ctx context.Context, payload domain.NotificationPayload, delay time.Duration) error {
	// 生成幂等键
	idempotencyKey := u.generateIdempotencyKey(payload)

	// 检查是否已存在相同的事件
	if existingEvent, err := u.repo.GetByTraceID(ctx, idempotencyKey); err == nil && existingEvent != nil {
		u.logger.InfoContext(ctx, "Event already exists, skipping",
			slog.String("trace_id", idempotencyKey),
			slog.String("event_id", existingEvent.ID.String()),
		)
		return nil
	}

	// 获取启用的通知设置
	enabledSettings, err := u.settingRepo.GetEnabledSettings(ctx)
	if err != nil {
		u.logger.ErrorContext(ctx, "Failed to get enabled notification settings",
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to get enabled notification settings: %w", err)
	}

	// 如果没有启用的通知设置，记录日志并返回
	if len(enabledSettings) == 0 {
		u.logger.InfoContext(ctx, "No enabled notification settings found, skipping event",
			slog.String("trace_id", idempotencyKey),
		)
		return nil
	}

	// 使用第一个启用的通知设置（后续可以根据业务需求选择特定的通道）
	setting := enabledSettings[0]

	// 根据不同类型的 Channel 确定 Target
	var target string
	switch setting.Channel {
	case consts.NotificationChannelDingTalk:
		target = setting.DingTalkConfig.WebhookURL
	default:
	}

	// 创建通知事件
	event := &domain.NotificationEvent{
		ID:         uuid.New(),
		EventType:  payload.GetEventType(),
		Channel:    setting.Channel,
		Status:     consts.NotificationStatusPending,
		Payload:    payload.GetPayload(),
		TemplateID: payload.GetTemplateID(),
		Target:     target,
		RetryCount: 0,
		MaxRetry:   setting.MaxRetry,
		Timeout:    setting.Timeout,
		TraceID:    idempotencyKey,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 如果有延迟，设置计划投递时间
	if delay > 0 {
		scheduledAt := time.Now().Add(delay)
		event.ScheduledAt = &scheduledAt
	}

	// 保存到数据库
	if err := u.repo.Create(ctx, event); err != nil {
		u.logger.ErrorContext(ctx, "Failed to create notification event",
			slog.String("error", err.Error()),
			slog.String("event_type", string(event.EventType)),
			slog.String("trace_id", event.TraceID),
		)
		return fmt.Errorf("failed to create notification event: %w", err)
	}

	// 发送到消息队列
	queueData := map[string]interface{}{
		"event_id":    event.ID.String(),
		"event_type":  string(event.EventType),
		"channel":     string(event.Channel),
		"payload":     event.Payload,
		"template_id": event.TemplateID,
		"target":      event.Target,
		"trace_id":    event.TraceID,
		"created_at":  event.CreatedAt.Unix(),
	}

	// 如果有延迟，添加延迟信息
	if event.ScheduledAt != nil {
		queueData["scheduled_at"] = event.ScheduledAt.Unix()
	}

	streamName := "notification:events"
	if err := u.producer.Publish(ctx, streamName, queueData); err != nil {
		u.logger.ErrorContext(ctx, "Failed to publish event to queue",
			slog.String("error", err.Error()),
			slog.String("event_id", event.ID.String()),
			slog.String("stream", streamName),
		)
		return fmt.Errorf("failed to publish event to queue: %w", err)
	}

	u.logger.InfoContext(ctx, "Notification event published successfully",
		slog.String("event_id", event.ID.String()),
		slog.String("event_type", string(event.EventType)),
		slog.String("channel", string(event.Channel)),
		slog.String("trace_id", event.TraceID),
	)

	return nil
}

// MarkAsDelivered 标记事件为已投递
func (u *notificationUsecase) MarkAsDelivered(ctx context.Context, eventID uuid.UUID) error {
	if err := u.repo.MarkAsDelivered(ctx, eventID); err != nil {
		u.logger.ErrorContext(ctx, "Failed to mark event as delivered",
			slog.String("error", err.Error()),
			slog.String("event_id", eventID.String()),
		)
		return fmt.Errorf("failed to mark event as delivered: %w", err)
	}

	u.logger.InfoContext(ctx, "Event marked as delivered",
		slog.String("event_id", eventID.String()),
	)

	return nil
}

// MarkAsFailed 标记事件为失败
func (u *notificationUsecase) MarkAsFailed(ctx context.Context, eventID uuid.UUID, errorMsg string) error {
	if err := u.repo.UpdateStatus(ctx, eventID, consts.NotificationStatusFailed, errorMsg); err != nil {
		u.logger.ErrorContext(ctx, "Failed to mark event as failed",
			slog.String("error", err.Error()),
			slog.String("event_id", eventID.String()),
			slog.String("error_msg", errorMsg),
		)
		return fmt.Errorf("failed to mark event as failed: %w", err)
	}

	u.logger.InfoContext(ctx, "Event marked as failed",
		slog.String("event_id", eventID.String()),
		slog.String("error_msg", errorMsg),
	)

	return nil
}

// RetryEvent 重试事件
func (u *notificationUsecase) RetryEvent(ctx context.Context, eventID uuid.UUID) error {
	// 获取事件
	dbEvent, err := u.repo.GetByID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get event: %w", err)
	}

	// 检查是否可以重试
	if dbEvent.RetryCount >= dbEvent.MaxRetry {
		return fmt.Errorf("event has exceeded max retry count")
	}

	// 重新发布到队列
	queueData := map[string]interface{}{
		"event_id":    dbEvent.ID.String(),
		"event_type":  string(dbEvent.EventType),
		"channel":     string(dbEvent.Channel),
		"payload":     dbEvent.Payload,
		"template_id": dbEvent.TemplateID,
		"target":      dbEvent.Target,
		"trace_id":    dbEvent.TraceID,
		"retry_count": dbEvent.RetryCount + 1,
	}

	streamName := "notification:events"
	if err := u.producer.Publish(ctx, streamName, queueData); err != nil {
		return fmt.Errorf("failed to republish event to queue: %w", err)
	}

	u.logger.InfoContext(ctx, "Event republished for retry",
		slog.String("event_id", dbEvent.ID.String()),
		slog.Int("retry_count", dbEvent.RetryCount+1),
	)

	return nil
}

// GetEventsByStatus 根据状态获取事件
func (u *notificationUsecase) GetEventsByStatus(ctx context.Context, status consts.NotificationStatus, limit int) ([]*domain.NotificationEvent, error) {
	dbEvents, err := u.repo.GetEventsByStatus(ctx, status, limit)
	if err != nil {
		return nil, err
	}

	// 转换为 domain 类型
	events := make([]*domain.NotificationEvent, 0, len(dbEvents))
	for _, dbEvent := range dbEvents {
		event := &domain.NotificationEvent{}
		events = append(events, event.From(dbEvent))
	}

	return events, nil
}

// GetPendingEvents 获取待处理事件
func (u *notificationUsecase) GetPendingEvents(ctx context.Context, limit int) ([]*domain.NotificationEvent, error) {
	dbEvents, err := u.repo.GetPendingEvents(ctx, limit)
	if err != nil {
		return nil, err
	}

	// 转换为 domain 类型
	events := make([]*domain.NotificationEvent, 0, len(dbEvents))
	for _, dbEvent := range dbEvents {
		event := &domain.NotificationEvent{}
		events = append(events, event.From(dbEvent))
	}

	return events, nil
}

// generateIdempotencyKey 生成幂等键
func (u *notificationUsecase) generateIdempotencyKey(payload domain.NotificationPayload) string {
	// 根据事件类型和负载生成唯一键
	payloadData := payload.GetPayload()

	switch payload.GetEventType() {
	case consts.NotificationEventTypeResumeParseCompleted:
		if resumeID, ok := payloadData["resume_id"]; ok {
			return fmt.Sprintf("resume_parse_%v", resumeID)
		}
	case consts.NotificationEventTypeJobMatchingCompleted:
		if resumeID, ok := payloadData["resume_id"]; ok {
			if jobID, ok := payloadData["job_id"]; ok {
				return fmt.Sprintf("job_matching_%v_%v", resumeID, jobID)
			}
		}
	case consts.NotificationEventTypeScreeningTaskCompleted:
		if taskID, ok := payloadData["task_id"]; ok {
			return fmt.Sprintf("screening_task_%v", taskID)
		}
	}

	// 默认使用时间戳和随机数
	return fmt.Sprintf("%s_%d_%s", payload.GetEventType(), time.Now().UnixNano(), uuid.New().String()[:8])
}
