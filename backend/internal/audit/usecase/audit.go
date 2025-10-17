package usecase

import (
	"context"
	"log/slog"

	"github.com/chaitin/WhaleHire/backend/domain"
)

// AuditUsecase 审计日志用例
type AuditUsecase struct {
	repo   domain.AuditRepo
	logger *slog.Logger
}

// NewAuditUsecase 创建审计日志用例
func NewAuditUsecase(repo domain.AuditRepo, logger *slog.Logger) *AuditUsecase {
	return &AuditUsecase{
		repo:   repo,
		logger: logger,
	}
}

// List 获取审计日志列表
func (u *AuditUsecase) List(ctx context.Context, req *domain.ListAuditLogReq) (*domain.ListAuditLogResp, error) {
	logs, pageInfo, err := u.repo.List(ctx, req)
	if err != nil {
		u.logger.With("error", err).Error("failed to list audit logs")
		return nil, err
	}

	return &domain.ListAuditLogResp{
		PageInfo:  pageInfo,
		AuditLogs: logs,
	}, nil
}

// GetByID 根据ID获取审计日志详情
func (u *AuditUsecase) GetByID(ctx context.Context, id string) (*domain.AuditLog, error) {
	log, err := u.repo.GetByID(ctx, id)
	if err != nil {
		u.logger.With("error", err).With("id", id).Error("failed to get audit log by id")
		return nil, err
	}

	return log, nil
}

// GetStats 获取审计日志统计信息
func (u *AuditUsecase) GetStats(ctx context.Context, req *domain.AuditStatsReq) (*domain.AuditStatsResp, error) {
	stats, err := u.repo.GetStats(ctx, req)
	if err != nil {
		u.logger.With("error", err).Error("failed to get audit stats")
		return nil, err
	}

	return stats, nil
}

// Delete 删除审计日志（软删除）
func (u *AuditUsecase) Delete(ctx context.Context, id string) error {
	err := u.repo.Delete(ctx, id)
	if err != nil {
		u.logger.With("error", err).With("id", id).Error("failed to delete audit log")
		return err
	}

	return nil
}

// BatchDelete 批量删除审计日志
func (u *AuditUsecase) BatchDelete(ctx context.Context, ids []string) error {
	err := u.repo.BatchDelete(ctx, ids)
	if err != nil {
		u.logger.With("error", err).With("ids", ids).Error("failed to batch delete audit logs")
		return err
	}

	return nil
}

// CleanupOldLogs 清理过期的审计日志
func (u *AuditUsecase) CleanupOldLogs(ctx context.Context, days int) error {
	err := u.repo.CleanupOldLogs(ctx, days)
	if err != nil {
		u.logger.With("error", err).With("days", days).Error("failed to cleanup old audit logs")
		return err
	}

	return nil
}
