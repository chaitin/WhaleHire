package worker

import (
	"context"
	"log/slog"
)

// WorkerServicer 将 NotificationWorker 封装为符合 service.Servicer 的后台服务，
// 以便统一由 Service 生命周期管理，支持优雅启动与关闭。
type WorkerServicer struct {
	w      *NotificationWorker
	ctx    context.Context
	cancel context.CancelFunc
}

// NewServicer 创建一个基于 NotificationWorker 的后台服务封装。
func NewServicer(w *NotificationWorker) *WorkerServicer {
	return &WorkerServicer{w: w}
}

func (ws *WorkerServicer) Name() string { return "Notification Worker" }

// Start 启动通知 Worker，并阻塞直到 Stop 被调用以符合 service.Servicer 的约定（Start 不应主动返回）。
func (ws *WorkerServicer) Start() error {
	ws.ctx, ws.cancel = context.WithCancel(context.Background())
	if err := ws.w.Start(ws.ctx); err != nil {
		return err
	}
	// 阻塞等待停止信号，防止提前返回导致 Service 认为该服务已退出
	<-ws.ctx.Done()
	return nil
}

// Stop 触发优雅关闭：取消上下文 + 调用 Worker.Stop 做附加清理。
func (ws *WorkerServicer) Stop() error {
	if ws.cancel != nil {
		ws.cancel()
	}
	// 传入已取消的上下文用于日志和资源清理
	if err := ws.w.Stop(ws.ctx); err != nil {
		// 使用 Worker 本身的 logger 记录错误（若存在）
		if ws.w != nil && ws.w.logger != nil {
			ws.w.logger.Error("通知 Worker 停止失败", slog.String("error", err.Error()))
		}
		return err
	}
	return nil
}
