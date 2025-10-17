package repo

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/auditlog"
	"github.com/chaitin/WhaleHire/backend/domain"
)

// AuditRepo 审计日志仓储实现
type AuditRepo struct {
	db     *db.Client
	logger *slog.Logger
}

// NewAuditRepo 创建审计日志仓储
func NewAuditRepo(db *db.Client, logger *slog.Logger) domain.AuditRepo {
	return &AuditRepo{
		db:     db,
		logger: logger,
	}
}

// Create 创建审计日志
func (r *AuditRepo) Create(ctx context.Context, log *domain.AuditLog) error {
	builder := r.db.AuditLog.Create().
		SetOperatorType(log.OperatorType).
		SetOperationType(log.OperationType).
		SetResourceType(log.ResourceType).
		SetMethod(log.RequestMethod).
		SetPath(log.RequestPath).
		SetStatusCode(log.ResponseStatus).
		SetStatus(log.Status).
		SetIP(log.IP).
		SetDurationMs(log.Duration)

	// 设置可选字段
	if log.OperatorID != nil {
		operatorID, err := uuid.Parse(*log.OperatorID)
		if err == nil {
			builder = builder.SetOperatorID(operatorID)
		}
	}

	if log.OperatorName != nil {
		builder = builder.SetOperatorName(*log.OperatorName)
	}

	if log.ResourceID != nil {
		builder = builder.SetResourceID(*log.ResourceID)
	}

	if log.ResourceName != nil {
		builder = builder.SetResourceName(*log.ResourceName)
	}

	if log.RequestQuery != nil {
		builder = builder.SetQueryParams(*log.RequestQuery)
	}

	if log.RequestBody != nil {
		builder = builder.SetRequestBody(*log.RequestBody)
	}

	if log.ResponseBody != nil {
		builder = builder.SetResponseBody(*log.ResponseBody)
	}

	if log.UserAgent != nil {
		builder = builder.SetUserAgent(*log.UserAgent)
	}

	if log.Country != nil {
		builder = builder.SetCountry(*log.Country)
	}

	if log.Province != nil {
		builder = builder.SetProvince(*log.Province)
	}

	if log.City != nil {
		builder = builder.SetCity(*log.City)
	}

	if log.ISP != nil {
		builder = builder.SetIsp(*log.ISP)
	}

	if log.SessionID != nil {
		builder = builder.SetSessionID(*log.SessionID)
	}

	if log.RequestID != nil {
		builder = builder.SetTraceID(*log.RequestID)
	}

	if log.ErrorMessage != nil {
		builder = builder.SetErrorMessage(*log.ErrorMessage)
	}

	_, err := builder.Save(ctx)
	if err != nil {
		r.logger.With("error", err).Error("failed to create audit log")
		return err
	}

	return nil
}

// List 获取审计日志列表
func (r *AuditRepo) List(ctx context.Context, req *domain.ListAuditLogReq) ([]*domain.AuditLog, *db.PageInfo, error) {
	query := r.db.AuditLog.Query().Where(auditlog.DeletedAtIsNil())

	// 应用过滤条件
	if req.OperatorType != nil {
		query = query.Where(auditlog.OperatorTypeEQ(*req.OperatorType))
	}

	if req.OperationType != nil {
		query = query.Where(auditlog.OperationTypeEQ(*req.OperationType))
	}

	if req.ResourceType != nil {
		query = query.Where(auditlog.ResourceTypeEQ(*req.ResourceType))
	}

	if req.Status != nil {
		query = query.Where(auditlog.StatusEQ(*req.Status))
	}

	if req.IP != nil {
		query = query.Where(auditlog.IPEQ(*req.IP))
	}

	if req.OperatorID != nil {
		operatorID, err := uuid.Parse(*req.OperatorID)
		if err == nil {
			query = query.Where(auditlog.OperatorIDEQ(operatorID))
		}
	}

	if req.StartTime != nil && req.StartTime.Time != nil {
		query = query.Where(auditlog.CreatedAtGTE(*req.StartTime.Time))
	}

	if req.EndTime != nil && req.EndTime.Time != nil {
		query = query.Where(auditlog.CreatedAtLTE(*req.EndTime.Time))
	}

	if req.Search != nil && *req.Search != "" {
		searchTerm := *req.Search
		query = query.Where(
			auditlog.Or(
				auditlog.OperatorNameContains(searchTerm),
				auditlog.ResourceNameContains(searchTerm),
				auditlog.PathContains(searchTerm),
			),
		)
	}

	// 计算总数
	total, err := query.Count(ctx)
	if err != nil {
		r.logger.With("error", err).Error("failed to count audit logs")
		return nil, nil, err
	}

	// 应用分页和排序
	offset := (req.Page - 1) * req.Size
	logs, err := query.
		Order(auditlog.ByCreatedAt(sql.OrderDesc())).
		Offset(offset).
		Limit(req.Size).
		All(ctx)
	if err != nil {
		r.logger.With("error", err).Error("failed to list audit logs")
		return nil, nil, err
	}

	// 转换为领域对象
	result := make([]*domain.AuditLog, len(logs))
	for i, log := range logs {
		result[i] = new(domain.AuditLog).From(log)
	}

	pageInfo := &db.PageInfo{
		TotalCount:  int64(total),
		HasNextPage: offset+len(logs) < total,
	}

	return result, pageInfo, nil
}

// GetByID 根据ID获取审计日志
func (r *AuditRepo) GetByID(ctx context.Context, id string) (*domain.AuditLog, error) {
	logID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid audit log id: %w", err)
	}

	log, err := r.db.AuditLog.Query().
		Where(auditlog.IDEQ(logID)).
		Where(auditlog.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		r.logger.With("error", err).With("id", id).Error("failed to get audit log by id")
		return nil, err
	}

	return new(domain.AuditLog).From(log), nil
}

// GetStats 获取审计日志统计信息
func (r *AuditRepo) GetStats(ctx context.Context, req *domain.AuditStatsReq) (*domain.AuditStatsResp, error) {
	query := r.db.AuditLog.Query().Where(auditlog.DeletedAtIsNil())

	// 应用时间范围过滤
	if req.StartTime != nil {
		query = query.Where(auditlog.CreatedAtGTE(*req.StartTime))
	}

	if req.EndTime != nil {
		query = query.Where(auditlog.CreatedAtLTE(*req.EndTime))
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		r.logger.With("error", err).Error("failed to count audit logs for stats")
		return nil, err
	}

	// 获取成功数量
	successCount, err := query.Where(auditlog.StatusEQ(consts.AuditLogStatusSuccess)).Count(ctx)
	if err != nil {
		r.logger.With("error", err).Error("failed to count successful audit logs")
		return nil, err
	}

	// 获取失败数量
	failedCount, err := query.Where(auditlog.StatusEQ(consts.AuditLogStatusFailed)).Count(ctx)
	if err != nil {
		r.logger.With("error", err).Error("failed to count failed audit logs")
		return nil, err
	}

	// 按操作者类型统计
	operatorTypeStats := make(map[consts.OperatorType]int64)
	for _, operatorType := range []consts.OperatorType{consts.OperatorTypeUser, consts.OperatorTypeAdmin} {
		count, err := query.Where(auditlog.OperatorTypeEQ(operatorType)).Count(ctx)
		if err != nil {
			r.logger.With("error", err).With("operator_type", operatorType).Error("failed to count by operator type")
			return nil, err
		}
		operatorTypeStats[operatorType] = int64(count)
	}

	// 按操作类型统计
	operationStats := make(map[consts.OperationType]int64)
	for _, operationType := range []consts.OperationType{
		consts.OperationTypeCreate, consts.OperationTypeUpdate, consts.OperationTypeDelete,
		consts.OperationTypeView, consts.OperationTypeLogin, consts.OperationTypeLogout,
	} {
		count, err := query.Where(auditlog.OperationTypeEQ(operationType)).Count(ctx)
		if err != nil {
			r.logger.With("error", err).With("operation_type", operationType).Error("failed to count by operation type")
			return nil, err
		}
		operationStats[operationType] = int64(count)
	}

	// 按资源类型统计
	resourceStats := make(map[consts.ResourceType]int64)
	for _, resourceType := range []consts.ResourceType{
		consts.ResourceTypeUser, consts.ResourceTypeAdmin, consts.ResourceTypeRole,
		consts.ResourceTypeDepartment, consts.ResourceTypeJobPosition, consts.ResourceTypeResume,
		consts.ResourceTypeScreening, consts.ResourceTypeSetting, consts.ResourceTypeAttachment,
		consts.ResourceTypeConversation, consts.ResourceTypeMessage,
	} {
		count, err := query.Where(auditlog.ResourceTypeEQ(resourceType)).Count(ctx)
		if err != nil {
			r.logger.With("error", err).With("resource_type", resourceType).Error("failed to count by resource type")
			return nil, err
		}
		resourceStats[resourceType] = int64(count)
	}

	// 按小时统计（过去24小时）
	hourlyStats := make([]domain.HourlyAuditStat, 0, 24)
	now := time.Now()
	for i := 23; i >= 0; i-- {
		hourStart := now.Add(-time.Duration(i) * time.Hour).Truncate(time.Hour)
		hourEnd := hourStart.Add(time.Hour)

		hourQuery := query.Where(
			auditlog.CreatedAtGTE(hourStart),
			auditlog.CreatedAtLT(hourEnd),
		)

		totalCount, err := hourQuery.Count(ctx)
		if err != nil {
			r.logger.With("error", err).With("hour", hourStart.Hour()).Error("failed to count hourly stats")
			return nil, err
		}

		successCount, err := hourQuery.Where(auditlog.StatusEQ(consts.AuditLogStatusSuccess)).Count(ctx)
		if err != nil {
			r.logger.With("error", err).With("hour", hourStart.Hour()).Error("failed to count hourly success stats")
			return nil, err
		}

		failedCount, err := hourQuery.Where(auditlog.StatusEQ(consts.AuditLogStatusFailed)).Count(ctx)
		if err != nil {
			r.logger.With("error", err).With("hour", hourStart.Hour()).Error("failed to count hourly failed stats")
			return nil, err
		}

		hourlyStats = append(hourlyStats, domain.HourlyAuditStat{
			Hour:         hourStart.Hour(),
			Count:        int64(totalCount),
			SuccessCount: int64(successCount),
			FailedCount:  int64(failedCount),
		})
	}

	return &domain.AuditStatsResp{
		TotalCount:        int64(total),
		SuccessCount:      int64(successCount),
		FailedCount:       int64(failedCount),
		OperatorTypeStats: operatorTypeStats,
		OperationStats:    operationStats,
		ResourceStats:     resourceStats,
		HourlyStats:       hourlyStats,
	}, nil
}

// Delete 删除审计日志（软删除）
func (r *AuditRepo) Delete(ctx context.Context, id string) error {
	logID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid audit log id: %w", err)
	}

	_, err = r.db.AuditLog.UpdateOneID(logID).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		r.logger.With("error", err).With("id", id).Error("failed to delete audit log")
		return err
	}

	return nil
}

// BatchDelete 批量删除审计日志
func (r *AuditRepo) BatchDelete(ctx context.Context, ids []string) error {
	logIDs := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		logID, err := uuid.Parse(id)
		if err != nil {
			return fmt.Errorf("invalid audit log id %s: %w", id, err)
		}
		logIDs[i] = logID
	}

	_, err := r.db.AuditLog.Update().
		Where(auditlog.IDIn(logIDs...)).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		r.logger.With("error", err).With("ids", ids).Error("failed to batch delete audit logs")
		return err
	}

	return nil
}

// CleanupOldLogs 清理过期的审计日志
func (r *AuditRepo) CleanupOldLogs(ctx context.Context, days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days)

	_, err := r.db.AuditLog.Delete().
		Where(auditlog.CreatedAtLT(cutoffTime)).
		Exec(ctx)
	if err != nil {
		r.logger.With("error", err).With("days", days).Error("failed to cleanup old audit logs")
		return err
	}

	return nil
}
