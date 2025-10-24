package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/resumemailboxstatistic"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/google/uuid"
)

// ResumeMailboxStatisticRepo 邮箱统计仓储实现
type ResumeMailboxStatisticRepo struct {
	client *db.Client
}

// NewResumeMailboxStatisticRepo 创建邮箱统计仓储实例
func NewResumeMailboxStatisticRepo(client *db.Client) domain.ResumeMailboxStatisticRepo {
	return &ResumeMailboxStatisticRepo{
		client: client,
	}
}

// GetByMailboxIDAndDate 根据邮箱ID和日期获取统计
func (r *ResumeMailboxStatisticRepo) GetByMailboxIDAndDate(ctx context.Context, mailboxID uuid.UUID, date time.Time) (*db.ResumeMailboxStatistic, error) {
	// 将日期对齐到零点
	alignedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	entity, err := r.client.ResumeMailboxStatistic.Query().
		Where(
			resumemailboxstatistic.MailboxID(mailboxID),
			resumemailboxstatistic.Date(alignedDate),
		).
		WithMailbox().
		Only(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, domain.ErrResumeMailboxStatisticNotFound
		}
		return nil, fmt.Errorf("failed to get resume mailbox statistic: %w", err)
	}

	return entity, nil
}

// List 获取统计列表
func (r *ResumeMailboxStatisticRepo) List(ctx context.Context, req *domain.ListResumeMailboxStatisticsRequest) ([]*db.ResumeMailboxStatistic, *db.PageInfo, error) {
	query := r.client.ResumeMailboxStatistic.Query().
		WithMailbox()

	// 添加过滤条件
	if req.MailboxID != nil {
		query = query.Where(resumemailboxstatistic.MailboxID(*req.MailboxID))
	}

	// 处理日期范围筛选
	if req.DateFrom != nil {
		alignedDateFrom := time.Date(req.DateFrom.Year(), req.DateFrom.Month(), req.DateFrom.Day(), 0, 0, 0, 0, req.DateFrom.Location())
		query = query.Where(resumemailboxstatistic.DateGTE(alignedDateFrom))
	}

	if req.DateTo != nil {
		alignedDateTo := time.Date(req.DateTo.Year(), req.DateTo.Month(), req.DateTo.Day(), 23, 59, 59, 999999999, req.DateTo.Location())
		query = query.Where(resumemailboxstatistic.DateLTE(alignedDateTo))
	}

	// 处理预设时间范围
	if req.Range != nil {
		now := time.Now()
		var startDate time.Time
		switch *req.Range {
		case "7d":
			startDate = now.AddDate(0, 0, -7)
		case "30d":
			startDate = now.AddDate(0, 0, -30)
		case "90d":
			startDate = now.AddDate(0, 0, -90)
		}
		if !startDate.IsZero() {
			alignedStartDate := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
			query = query.Where(resumemailboxstatistic.DateGTE(alignedStartDate))
		}
	}

	// 分页查询
	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.Size
	if size <= 0 {
		size = 10
	}

	entities, pageInfo, err := query.
		Order(db.Desc(resumemailboxstatistic.FieldDate)).
		Page(ctx, page, size)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list resume mailbox statistics: %w", err)
	}

	return entities, pageInfo, nil
}

// Upsert 更新或创建统计记录
func (r *ResumeMailboxStatisticRepo) Upsert(ctx context.Context, req *domain.UpsertResumeMailboxStatisticRequest) (*db.ResumeMailboxStatistic, error) {
	// 将日期对齐到零点
	alignedDate := time.Date(req.Date.Year(), req.Date.Month(), req.Date.Day(), 0, 0, 0, 0, req.Date.Location())

	// 尝试查找现有记录
	existing, err := r.client.ResumeMailboxStatistic.Query().
		Where(
			resumemailboxstatistic.MailboxID(req.MailboxID),
			resumemailboxstatistic.Date(alignedDate),
		).
		Only(ctx)

	if err != nil && !db.IsNotFound(err) {
		return nil, fmt.Errorf("failed to check existing statistic: %w", err)
	}

	if existing != nil {
		// 更新现有记录 - 使用累加逻辑
		builder := r.client.ResumeMailboxStatistic.UpdateOneID(existing.ID)

		if req.SyncedEmails != nil {
			builder = builder.AddSyncedEmails(*req.SyncedEmails)
		}
		if req.ParsedResumes != nil {
			builder = builder.AddParsedResumes(*req.ParsedResumes)
		}
		if req.FailedResumes != nil {
			builder = builder.AddFailedResumes(*req.FailedResumes)
		}
		if req.SkippedAttachments != nil {
			builder = builder.AddSkippedAttachments(*req.SkippedAttachments)
		}
		// 最后同步耗时使用覆盖逻辑，因为这是最新的耗时记录
		if req.LastSyncDurationMs != nil {
			builder = builder.SetLastSyncDurationMs(*req.LastSyncDurationMs)
		}

		entity, err := builder.Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to update resume mailbox statistic: %w", err)
		}

		// 加载关联数据
		entity, err = r.client.ResumeMailboxStatistic.Query().
			Where(resumemailboxstatistic.ID(entity.ID)).
			WithMailbox().
			Only(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load updated resume mailbox statistic: %w", err)
		}

		return entity, nil
	} else {
		// 创建新记录
		builder := r.client.ResumeMailboxStatistic.Create().
			SetMailboxID(req.MailboxID).
			SetDate(alignedDate)

		if req.SyncedEmails != nil {
			builder = builder.SetSyncedEmails(*req.SyncedEmails)
		}
		if req.ParsedResumes != nil {
			builder = builder.SetParsedResumes(*req.ParsedResumes)
		}
		if req.FailedResumes != nil {
			builder = builder.SetFailedResumes(*req.FailedResumes)
		}
		if req.SkippedAttachments != nil {
			builder = builder.SetSkippedAttachments(*req.SkippedAttachments)
		}
		if req.LastSyncDurationMs != nil {
			builder = builder.SetLastSyncDurationMs(*req.LastSyncDurationMs)
		}

		entity, err := builder.Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create resume mailbox statistic: %w", err)
		}

		// 加载关联数据
		entity, err = r.client.ResumeMailboxStatistic.Query().
			Where(resumemailboxstatistic.ID(entity.ID)).
			WithMailbox().
			Only(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load created resume mailbox statistic: %w", err)
		}

		return entity, nil
	}
}

// Delete 删除统计记录
func (r *ResumeMailboxStatisticRepo) Delete(ctx context.Context, mailboxID uuid.UUID, date time.Time) error {
	// 将日期对齐到零点
	alignedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	_, err := r.client.ResumeMailboxStatistic.Delete().
		Where(
			resumemailboxstatistic.MailboxID(mailboxID),
			resumemailboxstatistic.Date(alignedDate),
		).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete resume mailbox statistic: %w", err)
	}

	return nil
}

// GetSummary 获取统计汇总数据
func (r *ResumeMailboxStatisticRepo) GetSummary(ctx context.Context, req *domain.GetMailboxStatisticsSummaryRequest) (*domain.MailboxStatisticsSummary, error) {
	query := r.client.ResumeMailboxStatistic.Query()

	// 添加过滤条件
	if req.MailboxID != nil {
		query = query.Where(resumemailboxstatistic.MailboxID(*req.MailboxID))
	}

	// 处理日期范围筛选
	if req.DateFrom != nil {
		alignedDateFrom := time.Date(req.DateFrom.Year(), req.DateFrom.Month(), req.DateFrom.Day(), 0, 0, 0, 0, req.DateFrom.Location())
		query = query.Where(resumemailboxstatistic.DateGTE(alignedDateFrom))
	}

	if req.DateTo != nil {
		alignedDateTo := time.Date(req.DateTo.Year(), req.DateTo.Month(), req.DateTo.Day(), 23, 59, 59, 999999999, req.DateTo.Location())
		query = query.Where(resumemailboxstatistic.DateLTE(alignedDateTo))
	}

	// 处理预设时间范围
	if req.Range != nil {
		now := time.Now()
		var startDate time.Time
		switch *req.Range {
		case "7d":
			startDate = now.AddDate(0, 0, -7)
		case "30d":
			startDate = now.AddDate(0, 0, -30)
		case "90d":
			startDate = now.AddDate(0, 0, -90)
		}
		if !startDate.IsZero() {
			alignedStartDate := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
			query = query.Where(resumemailboxstatistic.DateGTE(alignedStartDate))
		}
	}

	// 获取所有符合条件的统计记录
	statistics, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics for summary: %w", err)
	}

	// 计算汇总数据
	summary := &domain.MailboxStatisticsSummary{}
	totalDuration := 0
	validDurationCount := 0

	for _, stat := range statistics {
		summary.TotalSyncedEmails += stat.SyncedEmails
		summary.TotalParsedResumes += stat.ParsedResumes
		summary.TotalFailedResumes += stat.FailedResumes
		summary.TotalSkippedAttachments += stat.SkippedAttachments

		if stat.LastSyncDurationMs > 0 {
			totalDuration += stat.LastSyncDurationMs
			validDurationCount++
		}
	}

	// 计算平均同步耗时
	if validDurationCount > 0 {
		summary.AvgSyncDurationMs = float64(totalDuration) / float64(validDurationCount)
	}

	// 计算成功率
	totalProcessed := summary.TotalParsedResumes + summary.TotalFailedResumes
	if totalProcessed > 0 {
		summary.SuccessRate = float64(summary.TotalParsedResumes) / float64(totalProcessed) * 100
	}

	// 计算所有解析的简历总数（成功+失败）
	summary.TotalResumes = summary.TotalParsedResumes + summary.TotalFailedResumes

	return summary, nil
}
