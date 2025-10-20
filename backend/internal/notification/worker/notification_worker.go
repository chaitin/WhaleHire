package worker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/internal/notification/adapter"
	"github.com/chaitin/WhaleHire/backend/internal/queue"
)

// NotificationWorker 通知工作器
type NotificationWorker struct {
	consumer        queue.Consumer
	repo            domain.NotificationEventRepo
	dingTalkAdapter *adapter.DingTalkAdapter
	logger          *slog.Logger
}

// NewNotificationWorker 创建通知工作器
func NewNotificationWorker(
	consumer queue.Consumer,
	repo domain.NotificationEventRepo,
	dingTalkAdapter *adapter.DingTalkAdapter,
	logger *slog.Logger,
) *NotificationWorker {
	return &NotificationWorker{
		consumer:        consumer,
		repo:            repo,
		dingTalkAdapter: dingTalkAdapter,
		logger:          logger,
	}
}

// Start 启动工作器
func (w *NotificationWorker) Start(ctx context.Context) error {
	w.logger.InfoContext(ctx, "Starting notification worker")

	// 订阅通知事件流
	streamName := "notification:events"
	consumerGroup := "notification-worker"
	consumerName := "worker-1"

	msgCh, err := w.consumer.Subscribe(ctx, streamName, consumerGroup, consumerName)
	if err != nil {
		return fmt.Errorf("failed to subscribe to notification stream: %w", err)
	}

	// 启动消息处理协程
	go w.processMessages(ctx, msgCh)

	w.logger.InfoContext(ctx, "Notification worker started successfully")
	return nil
}

// Stop 停止工作器
func (w *NotificationWorker) Stop(ctx context.Context) error {
	w.logger.InfoContext(ctx, "Stopping notification worker")
	// 关闭消费者连接，释放资源
	if err := w.consumer.Close(); err != nil {
		w.logger.ErrorContext(ctx, "Failed to close queue consumer", slog.String("error", err.Error()))
		return err
	}
	return nil
}

// processMessages 处理消息
func (w *NotificationWorker) processMessages(ctx context.Context, msgCh <-chan queue.Message) {
	for {
		select {
		case <-ctx.Done():
			w.logger.InfoContext(ctx, "Notification worker stopped")
			return
		case msg := <-msgCh:
			if err := w.handleMessage(ctx, msg); err != nil {
				w.logger.ErrorContext(ctx, "Failed to handle message",
					slog.String("message_id", msg.ID),
					slog.String("error", err.Error()))
			}
		}
	}
}

// handleMessage 处理单个消息
func (w *NotificationWorker) handleMessage(ctx context.Context, msg queue.Message) error {
	w.logger.InfoContext(ctx, "Processing notification message",
		slog.String("message_id", msg.ID),
		slog.String("stream", msg.Stream),
	)

	// 解析消息数据
	eventID, ok := msg.Data["event_id"].(string)
	if !ok {
		w.logger.ErrorContext(ctx, "Invalid event_id in message",
			slog.String("message_id", msg.ID),
		)
		w.ackMessage(ctx, msg)
		return nil
	}

	// 解析UUID
	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		w.logger.ErrorContext(ctx, "Invalid event_id format",
			slog.String("message_id", msg.ID),
			slog.String("event_id", eventID),
			slog.String("error", err.Error()),
		)
		w.ackMessage(ctx, msg)
		return nil
	}

	// 获取通知事件
	event, err := w.repo.GetByID(ctx, eventUUID)
	if err != nil {
		w.logger.ErrorContext(ctx, "Failed to get notification event",
			slog.String("message_id", msg.ID),
			slog.String("event_id", eventID),
			slog.String("error", err.Error()),
		)
		w.ackMessage(ctx, msg)
		return err
	}

	if event == nil {
		w.logger.ErrorContext(ctx, "Notification event not found",
			slog.String("message_id", msg.ID),
			slog.String("event_id", eventID),
		)
		w.ackMessage(ctx, msg)
		return fmt.Errorf("notification event not found: %s", eventID)
	}

	// 检查事件状态
	if event.Status == consts.NotificationStatusDelivered {
		w.logger.InfoContext(ctx, "Event already delivered, skipping",
			slog.String("message_id", msg.ID),
			slog.String("event_id", eventID),
		)
		w.ackMessage(ctx, msg)
		return nil
	}

	// 发送通知
	if err := w.sendNotification(ctx, event); err != nil {
		w.handleSendError(ctx, msg, event, err)
		return err
	}

	// 标记为已投递
	if err := w.repo.MarkAsDelivered(ctx, event.ID); err != nil {
		w.logger.ErrorContext(ctx, "Failed to mark event as delivered",
			slog.String("message_id", msg.ID),
			slog.String("event_id", eventID),
			slog.String("error", err.Error()),
		)
		return err
	}

	w.logger.InfoContext(ctx, "Notification sent successfully",
		slog.String("message_id", msg.ID),
		slog.String("event_id", eventID),
		slog.String("event_type", string(event.EventType)),
		slog.String("channel", string(event.Channel)),
	)

	w.ackMessage(ctx, msg)
	return nil
}

// sendNotification 发送通知
func (w *NotificationWorker) sendNotification(ctx context.Context, event *domain.NotificationEvent) error {
	switch event.Channel {
	case consts.NotificationChannelDingTalk:
		return w.dingTalkAdapter.SendNotification(ctx, event)
	default:
		return fmt.Errorf("unsupported notification channel: %s", event.Channel)
	}
}

// handleSendError 处理发送错误
func (w *NotificationWorker) handleSendError(ctx context.Context, msg queue.Message, event *domain.NotificationEvent, sendErr error) {
	w.logger.ErrorContext(ctx, "Failed to send notification",
		slog.String("message_id", msg.ID),
		slog.String("event_id", event.ID.String()),
		slog.String("event_type", string(event.EventType)),
		slog.String("channel", string(event.Channel)),
		slog.String("error", sendErr.Error()),
	)

	// 增加重试次数
	if err := w.repo.IncrementRetryCount(ctx, event.ID, sendErr.Error()); err != nil {
		w.logger.ErrorContext(ctx, "Failed to increment retry count",
			slog.String("message_id", msg.ID),
			slog.String("event_id", event.ID.String()),
			slog.String("error", err.Error()),
		)
	}

	// 确认消息，避免无限重试
	w.ackMessage(ctx, msg)
}

// ackMessage 确认消息
func (w *NotificationWorker) ackMessage(ctx context.Context, msg queue.Message) {
	streamName := "notification:events"
	consumerGroup := "notification-worker"

	if err := w.consumer.Ack(ctx, streamName, consumerGroup, msg.ID); err != nil {
		w.logger.ErrorContext(ctx, "Failed to ack message",
			slog.String("message_id", msg.ID),
			slog.String("stream", msg.Stream),
			slog.String("error", err.Error()),
		)
	}
}
