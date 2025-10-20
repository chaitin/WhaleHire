package repo

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/notificationevent"
	"github.com/chaitin/WhaleHire/backend/domain"
)

// notificationEventRepo 通知事件仓储实现
type notificationEventRepo struct {
	client *db.Client
}

// NewNotificationEventRepo 创建通知事件仓储实例
func NewNotificationEventRepo(client *db.Client) domain.NotificationEventRepo {
	return &notificationEventRepo{
		client: client,
	}
}

// Create 创建通知事件
func (r *notificationEventRepo) Create(ctx context.Context, event *domain.NotificationEvent) error {
	_, err := r.client.NotificationEvent.Create().
		SetID(event.ID).
		SetEventType(event.EventType).
		SetChannel(event.Channel).
		SetStatus(event.Status).
		SetPayload(event.Payload).
		SetTemplateID(event.TemplateID).
		SetTarget(event.Target).
		SetRetryCount(event.RetryCount).
		SetMaxRetry(event.MaxRetry).
		SetNillableLastError(&event.LastError).
		SetNillableTraceID(&event.TraceID).
		SetNillableScheduledAt(event.ScheduledAt).
		SetCreatedAt(event.CreatedAt).
		SetUpdatedAt(event.UpdatedAt).
		Save(ctx)

	return err
}

// GetByID 根据ID获取通知事件
func (r *notificationEventRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.NotificationEvent, error) {
	event, err := r.client.NotificationEvent.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.toDomainEvent(event), nil
}

// GetByTraceID 根据TraceID获取通知事件
func (r *notificationEventRepo) GetByTraceID(ctx context.Context, traceID string) (*domain.NotificationEvent, error) {
	event, err := r.client.NotificationEvent.Query().
		Where(notificationevent.TraceID(traceID)).
		First(ctx)
	if err != nil {
		return nil, err
	}

	return r.toDomainEvent(event), nil
}

// GetPendingEvents 获取待处理的通知事件
func (r *notificationEventRepo) GetPendingEvents(ctx context.Context, limit int) ([]*domain.NotificationEvent, error) {
	events, err := r.client.NotificationEvent.Query().
		Where(
			notificationevent.Status(consts.NotificationStatusPending),
			notificationevent.Or(
				notificationevent.ScheduledAtIsNil(),
				notificationevent.ScheduledAtLTE(time.Now()),
			),
		).
		Order(notificationevent.ByCreatedAt()).
		Limit(limit).
		All(ctx)

	if err != nil {
		return nil, err
	}

	return r.toDomainEvents(events), nil
}

// GetRetryableEvents 获取可重试的通知事件
func (r *notificationEventRepo) GetRetryableEvents(ctx context.Context, limit int) ([]*domain.NotificationEvent, error) {
	events, err := r.client.NotificationEvent.Query().
		Where(
			notificationevent.Status(consts.NotificationStatusFailed),
			notificationevent.And(
				notificationevent.RetryCountLT(3), // 使用默认最大重试次数
			),
		).
		Order(notificationevent.ByUpdatedAt()).
		Limit(limit).
		All(ctx)

	if err != nil {
		return nil, err
	}

	return r.toDomainEvents(events), nil
}

// Update 更新通知事件
func (r *notificationEventRepo) Update(ctx context.Context, event *domain.NotificationEvent) error {
	_, err := r.client.NotificationEvent.UpdateOneID(event.ID).
		SetEventType(event.EventType).
		SetChannel(event.Channel).
		SetStatus(event.Status).
		SetPayload(event.Payload).
		SetTemplateID(event.TemplateID).
		SetTarget(event.Target).
		SetRetryCount(event.RetryCount).
		SetMaxRetry(event.MaxRetry).
		SetNillableLastError(&event.LastError).
		SetNillableTraceID(&event.TraceID).
		SetNillableScheduledAt(event.ScheduledAt).
		SetNillableDeliveredAt(event.DeliveredAt).
		SetUpdatedAt(event.UpdatedAt).
		Save(ctx)

	return err
}

// UpdateStatus 更新通知事件状态
func (r *notificationEventRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status consts.NotificationStatus, errorMsg string) error {
	update := r.client.NotificationEvent.UpdateOneID(id).
		SetStatus(status).
		SetUpdatedAt(time.Now())

	if errorMsg != "" {
		update = update.SetLastError(errorMsg)
	}

	_, err := update.Save(ctx)
	return err
}

// MarkAsDelivered 标记为已投递
func (r *notificationEventRepo) MarkAsDelivered(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := r.client.NotificationEvent.UpdateOneID(id).
		SetStatus(consts.NotificationStatusDelivered).
		SetDeliveredAt(now).
		SetUpdatedAt(now).
		ClearLastError().
		Save(ctx)

	return err
}

// IncrementRetryCount 增加重试次数
func (r *notificationEventRepo) IncrementRetryCount(ctx context.Context, id uuid.UUID, errorMsg string) error {
	event, err := r.client.NotificationEvent.Get(ctx, id)
	if err != nil {
		return err
	}

	_, err = r.client.NotificationEvent.UpdateOneID(id).
		SetRetryCount(event.RetryCount + 1).
		SetLastError(errorMsg).
		SetStatus(consts.NotificationStatusFailed).
		SetUpdatedAt(time.Now()).
		Save(ctx)

	return err
}

// GetEventsByType 根据事件类型获取事件列表
func (r *notificationEventRepo) GetEventsByType(ctx context.Context, eventType consts.NotificationEventType, limit int) ([]*domain.NotificationEvent, error) {
	events, err := r.client.NotificationEvent.Query().
		Where(notificationevent.EventType(eventType)).
		Order(notificationevent.ByCreatedAt(sql.OrderDesc())).
		Limit(limit).
		All(ctx)

	if err != nil {
		return nil, err
	}

	return r.toDomainEvents(events), nil
}

// GetEventsByStatus 根据状态获取事件列表
func (r *notificationEventRepo) GetEventsByStatus(ctx context.Context, status consts.NotificationStatus, limit int) ([]*domain.NotificationEvent, error) {
	events, err := r.client.NotificationEvent.Query().
		Where(notificationevent.Status(status)).
		Order(notificationevent.ByCreatedAt(sql.OrderDesc())).
		Limit(limit).
		All(ctx)

	if err != nil {
		return nil, err
	}

	return r.toDomainEvents(events), nil
}

// DeleteOldEvents 删除旧事件
func (r *notificationEventRepo) DeleteOldEvents(ctx context.Context, before time.Time) (int, error) {
	return r.client.NotificationEvent.Delete().
		Where(notificationevent.CreatedAtLT(before)).
		Exec(ctx)
}

// toDomainEvent 转换为领域事件
func (r *notificationEventRepo) toDomainEvent(event *db.NotificationEvent) *domain.NotificationEvent {
	var scheduledAt *time.Time
	var deliveredAt *time.Time

	if !event.ScheduledAt.IsZero() {
		scheduledAt = &event.ScheduledAt
	}

	if !event.DeliveredAt.IsZero() {
		deliveredAt = &event.DeliveredAt
	}

	return &domain.NotificationEvent{
		ID:          event.ID,
		EventType:   event.EventType,
		Channel:     event.Channel,
		Status:      event.Status,
		Payload:     event.Payload,
		TemplateID:  event.TemplateID,
		Target:      event.Target,
		RetryCount:  event.RetryCount,
		MaxRetry:    event.MaxRetry,
		LastError:   event.LastError,
		TraceID:     event.TraceID,
		CreatedAt:   event.CreatedAt,
		ScheduledAt: scheduledAt,
		DeliveredAt: deliveredAt,
		UpdatedAt:   event.UpdatedAt,
	}
}

// toDomainEvents 转换为领域事件列表
func (r *notificationEventRepo) toDomainEvents(events []*db.NotificationEvent) []*domain.NotificationEvent {
	result := make([]*domain.NotificationEvent, len(events))
	for i, event := range events {
		result[i] = r.toDomainEvent(event)
	}
	return result
}
