package scheduler

import (
	"context"
)

// Servicer 将调度器封装为 service.Servicer 接口
type Servicer struct {
	scheduler *Scheduler
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewServicer 创建调度器服务包装器
func NewServicer(s *Scheduler) *Servicer {
	return &Servicer{scheduler: s}
}

// Name 返回服务名称
func (s *Servicer) Name() string {
	return "Resume Mailbox Scheduler"
}

// Start 启动调度器并阻塞直到 Stop 被调用
func (s *Servicer) Start() error {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s.scheduler.Start(s.ctx)
}

// Stop 停止调度器
func (s *Servicer) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	return s.scheduler.Stop(context.Background())
}
