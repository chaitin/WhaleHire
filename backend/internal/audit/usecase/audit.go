package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/errcode"
)

// AuditUsecase 审计日志用例
type AuditUsecase struct {
	repo   domain.AuditRepo
	logger *slog.Logger
}

// NewAuditUsecase 创建审计日志用例
func NewAuditUsecase(repo domain.AuditRepo, logger *slog.Logger) domain.AuditUsecase {
	return &AuditUsecase{
		repo:   repo,
		logger: logger,
	}
}

// List 获取审计日志列表
func (u *AuditUsecase) List(ctx context.Context, req *domain.ListAuditLogReq) (*domain.ListAuditLogResp, error) {

	// 时间范围验证
	if req.StartTime != nil && req.EndTime != nil &&
		req.StartTime.After(*req.EndTime) {
		return nil, errcode.ErrInvalidParam.WithData("message", "开始时间不能晚于结束时间")
	}

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
	if id == "" {
		return nil, errcode.ErrInvalidParam.WithData("message", "审计日志ID不能为空")
	}

	log, err := u.repo.GetByID(ctx, id)
	if err != nil {
		u.logger.With("error", err).With("id", id).Error("failed to get audit log by id")
		return nil, err
	}

	return log, nil
}

// GetStats 获取审计日志统计信息
func (u *AuditUsecase) GetStats(ctx context.Context, req *domain.AuditStatsReq) (*domain.AuditStatsResp, error) {
	// 参数验证
	if req.StartTime != nil && req.EndTime != nil && req.StartTime.After(*req.EndTime) {
		return nil, errcode.ErrInvalidParam.WithData("message", "开始时间不能晚于结束时间")
	}

	// 如果没有指定时间范围，默认查询最近7天
	if req.StartTime == nil && req.EndTime == nil {
		now := time.Now()
		endTime := now
		startTime := now.AddDate(0, 0, -7)
		req.StartTime = &startTime
		req.EndTime = &endTime
	}

	stats, err := u.repo.GetStats(ctx, req)
	if err != nil {
		u.logger.With("error", err).Error("failed to get audit stats")
		return nil, err
	}

	return stats, nil
}

// Delete 删除审计日志（软删除）
func (u *AuditUsecase) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errcode.ErrInvalidParam.WithData("message", "审计日志ID不能为空")
	}

	// 检查日志是否存在
	_, err := u.repo.GetByID(ctx, id)
	if err != nil {
		u.logger.With("error", err).With("id", id).Error("audit log not found for deletion")
		return errcode.ErrInvalidParam.WithData("message", "审计日志不存在")
	}

	err = u.repo.Delete(ctx, id)
	if err != nil {
		u.logger.With("error", err).With("id", id).Error("failed to delete audit log")
		return err
	}

	u.logger.With("id", id).Info("audit log deleted successfully")
	return nil
}

// BatchDelete 批量删除审计日志
func (u *AuditUsecase) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return errcode.ErrInvalidParam.WithData("message", "删除ID列表不能为空")
	}

	if len(ids) > 100 {
		return errcode.ErrInvalidParam.WithData("message", "批量删除数量不能超过100条")
	}

	// 验证所有ID格式
	for _, id := range ids {
		if id == "" {
			return errcode.ErrInvalidParam.WithData("message", "审计日志ID不能为空")
		}
	}

	err := u.repo.BatchDelete(ctx, ids)
	if err != nil {
		u.logger.With("error", err).With("ids", ids).Error("failed to batch delete audit logs")
		return err
	}

	u.logger.With("count", len(ids)).Info("audit logs batch deleted successfully")
	return nil
}

// CleanupOldLogs 清理过期的审计日志
func (u *AuditUsecase) CleanupOldLogs(ctx context.Context, days int) error {
	if days <= 0 {
		return errcode.ErrInvalidParam.WithData("message", "清理天数必须大于0")
	}

	if days < 30 {
		return errcode.ErrInvalidParam.WithData("message", "为了数据安全，清理天数不能少于30天")
	}

	err := u.repo.CleanupOldLogs(ctx, days)
	if err != nil {
		u.logger.With("error", err).With("days", days).Error("failed to cleanup old audit logs")
		return err
	}

	u.logger.With("days", days).Info("old audit logs cleaned up successfully")
	return nil
}

// GetOperationSummary 获取操作摘要统计
func (u *AuditUsecase) GetOperationSummary(ctx context.Context, operatorID string, days int) (*domain.OperationSummary, error) {
	if operatorID == "" {
		return nil, errcode.ErrInvalidParam.WithData("message", "操作者ID不能为空")
	}

	if days <= 0 || days > 365 {
		days = 30 // 默认30天
	}

	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)

	// 构建查询请求
	req := &domain.ListAuditLogReq{
		OperatorID: operatorID,
		StartTime:  &startTime,
		EndTime:    &endTime,
	}
	req.Page = 1
	req.Size = 1000 // 获取足够多的数据用于统计

	logs, _, err := u.repo.List(ctx, req)
	if err != nil {
		u.logger.With("error", err).With("operator_id", operatorID).Error("failed to get operation summary")
		return nil, err
	}

	// 统计操作摘要
	summary := &domain.OperationSummary{
		OperatorID:    operatorID,
		Days:          days,
		TotalCount:    int64(len(logs)),
		SuccessCount:  0,
		FailedCount:   0,
		OperationMap:  make(map[consts.OperationType]int64),
		ResourceMap:   make(map[consts.ResourceType]int64),
		DailyActivity: make([]domain.DailyActivity, 0),
	}

	// 按天统计活动
	dailyMap := make(map[string]*domain.DailyActivity)

	for _, log := range logs {
		// 统计成功失败
		if log.Status == consts.AuditLogStatusSuccess {
			summary.SuccessCount++
		} else {
			summary.FailedCount++
		}

		// 统计操作类型
		summary.OperationMap[log.OperationType]++

		// 统计资源类型
		summary.ResourceMap[log.ResourceType]++

		// 按天统计
		dayKey := time.Unix(log.CreatedAt, 0).Format("2006-01-02")
		if daily, exists := dailyMap[dayKey]; exists {
			daily.Count++
			if log.Status == consts.AuditLogStatusSuccess {
				daily.SuccessCount++
			} else {
				daily.FailedCount++
			}
		} else {
			daily := &domain.DailyActivity{
				Date:         dayKey,
				Count:        1,
				SuccessCount: 0,
				FailedCount:  0,
			}
			if log.Status == consts.AuditLogStatusSuccess {
				daily.SuccessCount = 1
			} else {
				daily.FailedCount = 1
			}
			dailyMap[dayKey] = daily
		}
	}

	// 转换为切片
	for _, daily := range dailyMap {
		summary.DailyActivity = append(summary.DailyActivity, *daily)
	}

	return summary, nil
}

// GetSecurityAlerts 获取安全告警
func (u *AuditUsecase) GetSecurityAlerts(ctx context.Context, hours int) ([]*domain.SecurityAlert, error) {
	if hours <= 0 || hours > 168 { // 最多7天
		hours = 24 // 默认24小时
	}

	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	// 获取最近的审计日志
	req := &domain.ListAuditLogReq{
		StartTime: &startTime,
		EndTime:   &endTime,
	}
	req.Page = 1
	req.Size = 1000

	logs, _, err := u.repo.List(ctx, req)
	if err != nil {
		u.logger.With("error", err).Error("failed to get logs for security alerts")
		return nil, err
	}

	alerts := make([]*domain.SecurityAlert, 0)

	// 统计每个IP的失败次数
	ipFailureMap := make(map[string]int)
	ipOperatorMap := make(map[string]string)

	for _, log := range logs {
		if log.Status == consts.AuditLogStatusFailed {
			ipFailureMap[log.IP]++
			if log.OperatorID != nil {
				ipOperatorMap[log.IP] = *log.OperatorID
			}
		}
	}

	// 检查异常IP（失败次数超过阈值）
	for ip, failureCount := range ipFailureMap {
		if failureCount >= 10 { // 失败次数阈值
			alert := &domain.SecurityAlert{
				Type:        "suspicious_ip",
				Level:       "high",
				Title:       "可疑IP地址",
				Description: fmt.Sprintf("IP地址 %s 在过去 %d 小时内失败操作 %d 次", ip, hours, failureCount),
				IP:          ip,
				Count:       int64(failureCount),
				CreatedAt:   time.Now().Unix(),
			}
			if operatorID, exists := ipOperatorMap[ip]; exists {
				alert.OperatorID = &operatorID
			}
			alerts = append(alerts, alert)
		}
	}

	// 统计每个用户的操作频率
	operatorCountMap := make(map[string]int)
	for _, log := range logs {
		if log.OperatorID != nil {
			operatorCountMap[*log.OperatorID]++
		}
	}

	// 检查异常用户（操作频率过高）
	for operatorID, count := range operatorCountMap {
		if count >= 100 { // 操作次数阈值
			alert := &domain.SecurityAlert{
				Type:        "high_frequency_user",
				Level:       "medium",
				Title:       "高频操作用户",
				Description: fmt.Sprintf("用户 %s 在过去 %d 小时内操作 %d 次", operatorID, hours, count),
				OperatorID:  &operatorID,
				Count:       int64(count),
				CreatedAt:   time.Now().Unix(),
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}
