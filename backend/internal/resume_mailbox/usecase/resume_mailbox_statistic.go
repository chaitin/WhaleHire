package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/google/uuid"
)

// ResumeMailboxStatisticUsecase 邮箱统计用例实现
type ResumeMailboxStatisticUsecase struct {
	repo domain.ResumeMailboxStatisticRepo
}

// NewResumeMailboxStatisticUsecase 创建邮箱统计用例实例
func NewResumeMailboxStatisticUsecase(
	repo domain.ResumeMailboxStatisticRepo,
) domain.ResumeMailboxStatisticUsecase {
	return &ResumeMailboxStatisticUsecase{
		repo: repo,
	}
}

// GetByMailboxIDAndDate 根据邮箱ID和日期获取统计
func (u *ResumeMailboxStatisticUsecase) GetByMailboxIDAndDate(ctx context.Context, mailboxID uuid.UUID, date time.Time) (*domain.ResumeMailboxStatistic, error) {
	dbStatistic, err := u.repo.GetByMailboxIDAndDate(ctx, mailboxID, date)
	if err != nil {
		return nil, err
	}

	// 转换为 domain 类型
	statistic := &domain.ResumeMailboxStatistic{}
	statistic.From(dbStatistic)

	return statistic, nil
}

// List 获取统计列表
func (u *ResumeMailboxStatisticUsecase) List(ctx context.Context, req *domain.ListResumeMailboxStatisticsRequest) (*domain.ListResumeMailboxStatisticsResponse, error) {
	// 设置默认分页参数
	if req.Size <= 0 {
		req.Size = 20
	}
	if req.Size > 100 {
		req.Size = 100
	}

	// 验证日期范围
	if req.DateFrom != nil && req.DateTo != nil && req.DateFrom.After(*req.DateTo) {
		return nil, fmt.Errorf("date_from cannot be after date_to")
	}

	// 验证预设时间范围
	if req.Range != nil {
		validRanges := map[string]bool{
			"7d":  true,
			"30d": true,
			"90d": true,
		}
		if !validRanges[*req.Range] {
			return nil, fmt.Errorf("invalid range: %s, must be one of: 7d, 30d, 90d", *req.Range)
		}
	}

	dbStatistics, pageInfo, err := u.repo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list resume mailbox statistics: %w", err)
	}

	// 转换为 domain 类型
	items := make([]*domain.ResumeMailboxStatistic, len(dbStatistics))
	for i, dbStatistic := range dbStatistics {
		statistic := &domain.ResumeMailboxStatistic{}
		statistic.From(dbStatistic)
		items[i] = statistic
	}

	return &domain.ListResumeMailboxStatisticsResponse{
		Items:    items,
		PageInfo: pageInfo,
	}, nil
}

// Upsert 更新或创建统计记录
func (u *ResumeMailboxStatisticUsecase) Upsert(ctx context.Context, req *domain.UpsertResumeMailboxStatisticRequest) (*domain.ResumeMailboxStatistic, error) {
	// 验证必要字段
	if req.MailboxID == uuid.Nil {
		return nil, fmt.Errorf("mailbox_id is required")
	}
	if req.Date.IsZero() {
		return nil, fmt.Errorf("date is required")
	}

	// 验证数值字段不能为负数
	if req.SyncedEmails != nil && *req.SyncedEmails < 0 {
		return nil, fmt.Errorf("synced_emails cannot be negative")
	}
	if req.ParsedResumes != nil && *req.ParsedResumes < 0 {
		return nil, fmt.Errorf("parsed_resumes cannot be negative")
	}
	if req.FailedResumes != nil && *req.FailedResumes < 0 {
		return nil, fmt.Errorf("failed_resumes cannot be negative")
	}
	if req.SkippedAttachments != nil && *req.SkippedAttachments < 0 {
		return nil, fmt.Errorf("skipped_attachments cannot be negative")
	}
	if req.LastSyncDurationMs != nil && *req.LastSyncDurationMs < 0 {
		return nil, fmt.Errorf("last_sync_duration_ms cannot be negative")
	}

	dbStatistic, err := u.repo.Upsert(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert resume mailbox statistic: %w", err)
	}

	// 转换为 domain 类型
	statistic := &domain.ResumeMailboxStatistic{}
	statistic.From(dbStatistic)

	return statistic, nil
}

// Delete 删除统计记录
func (u *ResumeMailboxStatisticUsecase) Delete(ctx context.Context, mailboxID uuid.UUID, date time.Time) error {
	if mailboxID == uuid.Nil {
		return fmt.Errorf("mailbox_id is required")
	}
	if date.IsZero() {
		return fmt.Errorf("date is required")
	}

	err := u.repo.Delete(ctx, mailboxID, date)
	if err != nil {
		return fmt.Errorf("failed to delete resume mailbox statistic: %w", err)
	}

	return nil
}

// GetSummary 获取统计汇总数据
func (u *ResumeMailboxStatisticUsecase) GetSummary(ctx context.Context, req *domain.GetMailboxStatisticsSummaryRequest) (*domain.MailboxStatisticsSummary, error) {
	// 验证日期范围
	if req.DateFrom != nil && req.DateTo != nil && req.DateFrom.After(*req.DateTo) {
		return nil, fmt.Errorf("date_from cannot be after date_to")
	}

	// 验证预设时间范围
	if req.Range != nil {
		validRanges := map[string]bool{
			"7d":  true,
			"30d": true,
			"90d": true,
		}
		if !validRanges[*req.Range] {
			return nil, fmt.Errorf("invalid range: %s, must be one of: 7d, 30d, 90d", *req.Range)
		}
	}

	summary, err := u.repo.GetSummary(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get mailbox statistics summary: %w", err)
	}

	return summary, nil
}
