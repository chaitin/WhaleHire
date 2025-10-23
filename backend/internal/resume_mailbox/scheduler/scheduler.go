package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/domain"
)

const (
	defaultSyncIntervalMinutes = 15
	minSyncIntervalMinutes     = 5
	defaultSyncTimeout         = 2 * time.Minute
)

// Scheduler 负责按配置定期触发邮箱同步任务
type Scheduler struct {
	settingRepo domain.ResumeMailboxSettingRepo
	syncUsecase domain.ResumeMailboxSyncUsecase
	logger      *slog.Logger

	mu        sync.Mutex
	jobs      map[uuid.UUID]*mailboxJob
	pending   map[uuid.UUID]*domain.ResumeMailboxSetting
	started   bool
	baseCtx   context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	interval  time.Duration
	syncLimit time.Duration
}

type mailboxJob struct {
	mailboxID uuid.UUID
	interval  time.Duration
	cancel    context.CancelFunc
}

var _ domain.ResumeMailboxScheduler = (*Scheduler)(nil)

// NewScheduler 创建调度器实例
func NewScheduler(
	settingRepo domain.ResumeMailboxSettingRepo,
	syncUsecase domain.ResumeMailboxSyncUsecase,
	logger *slog.Logger,
) *Scheduler {
	if logger == nil {
		logger = slog.Default()
	}

	return &Scheduler{
		settingRepo: settingRepo,
		syncUsecase: syncUsecase,
		logger:      logger,
		jobs:        make(map[uuid.UUID]*mailboxJob),
		pending:     make(map[uuid.UUID]*domain.ResumeMailboxSetting),
		interval:    time.Duration(defaultSyncIntervalMinutes) * time.Minute,
		syncLimit:   defaultSyncTimeout,
	}
}

// Start 启动调度器主循环，阻塞直到 Stop 被调用
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		<-ctx.Done()
		return nil
	}
	baseCtx, cancel := context.WithCancel(ctx)
	s.baseCtx = baseCtx
	s.cancel = cancel
	s.started = true
	pending := s.collectPendingLocked()
	s.mu.Unlock()

	s.logger.InfoContext(baseCtx, "简历邮箱调度器启动")

	if err := s.bootstrap(baseCtx); err != nil {
		s.logger.ErrorContext(baseCtx, "加载邮箱调度配置失败", slog.String("error", err.Error()))
	}

	for _, setting := range pending {
		if err := s.Upsert(baseCtx, setting); err != nil {
			s.logger.ErrorContext(baseCtx, "恢复待调度邮箱失败",
				slog.String("mailbox_id", setting.ID.String()),
				slog.String("error", err.Error()),
			)
		}
	}

	<-baseCtx.Done()

	s.logger.Info("简历邮箱调度器停止中")
	s.stopAllJobs()
	s.wg.Wait()
	s.logger.Info("简历邮箱调度器已停止")

	return nil
}

// Stop 请求调度器停止，等待 Start 返回
func (s *Scheduler) Stop(context.Context) error {
	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
	}
	s.mu.Unlock()
	return nil
}

// Upsert 注册或更新邮箱调度任务
func (s *Scheduler) Upsert(ctx context.Context, setting *domain.ResumeMailboxSetting) error {
	if setting == nil {
		return errors.New("缺少邮箱配置")
	}

	mailboxID := setting.ID
	status := setting.Status

	s.mu.Lock()
	if !s.started {
		s.pending[mailboxID] = cloneSetting(setting)
		s.mu.Unlock()
		return nil
	}

	baseCtx := s.baseCtx
	if baseCtx == nil {
		baseCtx = context.Background()
	}

	if status != string(consts.MailboxStatusEnabled) {
		if job, ok := s.jobs[mailboxID]; ok {
			job.cancel()
			s.logger.InfoContext(ctx, "已停用邮箱同步任务",
				slog.String("mailbox_id", mailboxID.String()),
				slog.String("email", setting.EmailAddress),
			)
			delete(s.jobs, mailboxID)
		}
		s.mu.Unlock()
		return nil
	}

	interval := s.resolveInterval(setting)
	if job, ok := s.jobs[mailboxID]; ok {
		if job.interval == interval {
			s.mu.Unlock()
			return nil
		}
		job.cancel()
		s.logger.InfoContext(ctx, "邮箱同步任务重新调度",
			slog.String("mailbox_id", mailboxID.String()),
			slog.String("email", setting.EmailAddress),
			slog.Duration("interval", interval),
			slog.String("reason", "同步间隔更新"),
		)
		delete(s.jobs, mailboxID)
	}

	jobCtx, cancel := context.WithCancel(baseCtx)
	job := &mailboxJob{
		mailboxID: mailboxID,
		interval:  interval,
		cancel:    cancel,
	}
	s.jobs[mailboxID] = job
	s.wg.Add(1)
	go s.runJob(jobCtx, job, setting.EmailAddress)

	s.logger.InfoContext(ctx, "邮箱同步任务已启动",
		slog.String("mailbox_id", mailboxID.String()),
		slog.String("email", setting.EmailAddress),
		slog.Duration("interval", interval),
	)

	s.mu.Unlock()
	return nil
}

// Remove 移除邮箱调度任务
func (s *Scheduler) Remove(ctx context.Context, mailboxID uuid.UUID) error {
	s.mu.Lock()
	if !s.started {
		delete(s.pending, mailboxID)
		s.mu.Unlock()
		return nil
	}

	if job, ok := s.jobs[mailboxID]; ok {
		job.cancel()
		s.logger.InfoContext(ctx, "邮箱同步任务已移除",
			slog.String("mailbox_id", mailboxID.String()),
		)
		delete(s.jobs, mailboxID)
	}

	s.mu.Unlock()
	return nil
}

func (s *Scheduler) runJob(ctx context.Context, job *mailboxJob, email string) {
	defer s.wg.Done()

	// 首次立即执行一次同步
	s.executeSync(ctx, job.mailboxID, email)

	ticker := time.NewTicker(job.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.InfoContext(ctx, "邮箱同步任务已停止",
				slog.String("mailbox_id", job.mailboxID.String()),
				slog.String("email", email),
			)
			return
		case <-ticker.C:
			s.executeSync(ctx, job.mailboxID, email)
		}
	}
}

func (s *Scheduler) executeSync(ctx context.Context, mailboxID uuid.UUID, email string) {
	runCtx, cancel := context.WithTimeout(ctx, s.syncLimit)
	defer cancel()

	start := time.Now()
	result, err := s.syncUsecase.SyncNow(runCtx, mailboxID)
	if err != nil {
		s.logger.ErrorContext(ctx, "邮箱自动同步失败",
			slog.String("mailbox_id", mailboxID.String()),
			slog.String("email", email),
			slog.String("error", err.Error()),
		)
		return
	}

	duration := time.Since(start)
	processed := 0
	success := 0
	failure := 0
	skipped := 0
	lastMessageID := ""

	if result != nil {
		processed = result.ProcessedEmails
		success = result.SuccessAttachments
		failure = result.FailedAttachments
		skipped = result.SkippedAttachments
		lastMessageID = result.LastMessageID
	}

	s.logger.InfoContext(ctx, "邮箱自动同步完成",
		slog.String("mailbox_id", mailboxID.String()),
		slog.String("email", email),
		slog.Int("processed_emails", processed),
		slog.Int("success_attachments", success),
		slog.Int("failed_attachments", failure),
		slog.Int("skipped_attachments", skipped),
		slog.Duration("duration", duration),
		slog.String("last_message_id", lastMessageID),
	)
}

func (s *Scheduler) resolveInterval(setting *domain.ResumeMailboxSetting) time.Duration {
	if setting != nil && setting.SyncIntervalMinutes != nil {
		minutes := *setting.SyncIntervalMinutes
		if minutes < minSyncIntervalMinutes {
			minutes = minSyncIntervalMinutes
		}
		return time.Duration(minutes) * time.Minute
	}
	return s.interval
}

func (s *Scheduler) bootstrap(ctx context.Context) error {
	settings, err := s.settingRepo.GetActiveSettings(ctx)
	if err != nil {
		return err
	}

	for _, entity := range settings {
		setting := (&domain.ResumeMailboxSetting{}).From(entity)
		if err := s.Upsert(ctx, setting); err != nil {
			s.logger.ErrorContext(ctx, "初始化邮箱调度失败",
				slog.String("mailbox_id", setting.ID.String()),
				slog.String("error", err.Error()),
			)
		}
	}

	return nil
}

// stopAllJobs 停止所有正在运行的邮箱同步任务
// 该方法会遍历所有活跃的任务，调用其取消函数并从任务映射中删除
// 注意：调用此方法前需要确保已获取互斥锁
func (s *Scheduler) stopAllJobs() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, job := range s.jobs {
		job.cancel()
		delete(s.jobs, id)
	}
}

// collectPendingLocked 收集并清空所有待处理的邮箱设置
// 返回待处理设置的深拷贝列表，同时清空内部的pending映射
// 如果没有待处理的设置，返回nil
// 注意：此方法假设调用者已经持有锁
func (s *Scheduler) collectPendingLocked() []*domain.ResumeMailboxSetting {
	if len(s.pending) == 0 {
		return nil
	}
	list := make([]*domain.ResumeMailboxSetting, 0, len(s.pending))
	for _, setting := range s.pending {
		list = append(list, cloneSetting(setting))
	}
	s.pending = make(map[uuid.UUID]*domain.ResumeMailboxSetting)
	return list
}

// cloneSetting 创建邮箱设置对象的深拷贝
// 对于指针字段（如SyncIntervalMinutes），会创建新的指针指向拷贝的值
// 如果输入为nil，返回nil
// 这样可以避免并发访问时的数据竞争问题
func cloneSetting(setting *domain.ResumeMailboxSetting) *domain.ResumeMailboxSetting {
	if setting == nil {
		return nil
	}
	copy := *setting
	if setting.SyncIntervalMinutes != nil {
		value := *setting.SyncIntervalMinutes
		copy.SyncIntervalMinutes = &value
	}
	return &copy
}
