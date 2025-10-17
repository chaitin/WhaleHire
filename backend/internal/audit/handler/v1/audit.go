package v1

import (
	"log/slog"
	"strconv"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/errcode"
	"github.com/chaitin/WhaleHire/backend/internal/middleware"
	"github.com/chaitin/WhaleHire/backend/pkg/web"
)

// AuditHandler 审计日志HTTP处理器
type AuditHandler struct {
	usecase domain.AuditUsecase
	logger  *slog.Logger
}

// NewAuditHandler 创建审计日志处理器并注册路由
func NewAuditHandler(
	w *web.Web,
	usecase domain.AuditUsecase,
	auth *middleware.AuthMiddleware,
	logger *slog.Logger,
) *AuditHandler {
	handler := &AuditHandler{
		usecase: usecase,
		logger:  logger.With("module", "audit_handler"),
	}

	// 注册路由
	group := w.Group("/api/v1/audit")
	group.Use(auth.UserAuth()) // 需要用户认证

	// 审计日志查询接口
	group.GET("/logs", web.BindHandler(handler.ListLogs, web.WithPage()))
	group.GET("/logs/:id", web.BaseHandler(handler.GetLog))
	group.GET("/stats", web.BindHandler(handler.GetStats))
	group.GET("/operation-summary", web.BaseHandler(handler.GetOperationSummary))
	group.GET("/security-alerts", web.BaseHandler(handler.GetSecurityAlerts))

	// 审计日志管理接口（需要管理员权限）
	group.Use(auth.Auth())
	group.DELETE("/logs/:id", web.BaseHandler(handler.DeleteLog))
	group.DELETE("/logs/batch", web.BindHandler(handler.BatchDeleteLogs))
	group.DELETE("/logs/cleanup", web.BindHandler(handler.CleanupOldLogs))

	return handler
}

// ListLogs 获取审计日志列表
//
//	@Tags			Audit
//	@Summary		获取审计日志列表
//	@Description	分页获取审计日志列表，支持多种过滤条件
//	@ID				list-audit-logs
//	@Accept			json
//	@Produce		json
//	@Param			page			query		int		false	"页码"
//	@Param			size			query		int		false	"每页数量"
//	@Param			operator_type	query		string	false	"操作者类型"
//	@Param			operator_id		query		string	false	"操作者ID"
//	@Param			operation_type	query		string	false	"操作类型"
//	@Param			resource_type	query		string	false	"资源类型"
//	@Param			resource_id		query		string	false	"资源ID"
//	@Param			status			query		string	false	"状态"
//	@Param			ip				query		string	false	"IP地址"
//	@Param			start_time		query		string	false	"开始时间"
//	@Param			end_time		query		string	false	"结束时间"
//	@Param			search			query		string	false	"搜索关键词"
//	@Success		200				{object}	web.Resp{data=domain.ListAuditLogResp}
//	@Router			/api/v1/audit/logs [get]
func (h *AuditHandler) ListLogs(c *web.Context, req domain.ListAuditLogReq) error {
	// 设置分页参数
	req.Page = c.Page().Page
	req.Size = c.Page().Size

	// 解析时间参数
	resp, err := h.usecase.List(c.Request().Context(), &req)
	if err != nil {
		h.logger.With("error", err).Error("failed to list audit logs")
		return err
	}

	return c.Success(resp)
}

// GetLog 获取审计日志详情
//
//	@Tags			Audit
//	@Summary		获取审计日志详情
//	@Description	根据ID获取单个审计日志的详细信息
//	@ID				get-audit-log
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"审计日志ID"
//	@Success		200	{object}	web.Resp{data=domain.AuditLog}
//	@Router			/api/v1/audit/logs/{id} [get]
func (h *AuditHandler) GetLog(c *web.Context) error {
	id := c.Param("id")
	if id == "" {
		return errcode.ErrInvalidParam.WithData("message", "审计日志ID不能为空")
	}

	log, err := h.usecase.GetByID(c.Request().Context(), id)
	if err != nil {
		h.logger.With("error", err).With("id", id).Error("failed to get audit log")
		return err
	}

	return c.Success(log)
}

// GetStats 获取审计统计信息
//
//	@Tags			Audit
//	@Summary		获取审计统计信息
//	@Description	获取审计日志的统计信息，包括总数、成功失败数、各类型统计等
//	@ID				get-audit-stats
//	@Accept			json
//	@Produce		json
//	@Param			start_time	query		string	false	"开始时间"
//	@Param			end_time	query		string	false	"结束时间"
//	@Success		200			{object}	web.Resp{data=domain.AuditStatsResp}
//	@Router			/api/v1/audit/stats [get]
func (h *AuditHandler) GetStats(c *web.Context, req domain.AuditStatsReq) error {
	stats, err := h.usecase.GetStats(c.Request().Context(), &req)
	if err != nil {
		h.logger.With("error", err).Error("failed to get audit stats")
		return err
	}

	return c.Success(stats)
}

// GetOperationSummary 获取操作摘要
//
//	@Tags			Audit
//	@Summary		获取用户操作摘要
//	@Description	获取指定用户在指定天数内的操作摘要统计
//	@ID				get-operation-summary
//	@Accept			json
//	@Produce		json
//	@Param			operator_id	query		string	true	"操作者ID"
//	@Param			days		query		int		false	"统计天数，默认30天"
//	@Success		200			{object}	web.Resp{data=domain.OperationSummary}
//	@Router			/api/v1/audit/operation-summary [get]
func (h *AuditHandler) GetOperationSummary(c *web.Context) error {
	operatorID := c.QueryParam("operator_id")
	if operatorID == "" {
		return errcode.ErrInvalidParam.WithData("message", "操作者ID不能为空")
	}

	daysStr := c.QueryParam("days")
	days := 30 // 默认30天
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	summary, err := h.usecase.GetOperationSummary(c.Request().Context(), operatorID, days)
	if err != nil {
		h.logger.With("error", err).With("operator_id", operatorID).Error("failed to get operation summary")
		return err
	}

	return c.Success(summary)
}

// GetSecurityAlerts 获取安全告警
//
//	@Tags			Audit
//	@Summary		获取安全告警
//	@Description	获取最近时间内的安全告警信息
//	@ID				get-security-alerts
//	@Accept			json
//	@Produce		json
//	@Param			hours	query		int	false	"统计小时数，默认24小时"
//	@Success		200		{object}	web.Resp{data=[]domain.SecurityAlert}
//	@Router			/api/v1/audit/security-alerts [get]
func (h *AuditHandler) GetSecurityAlerts(c *web.Context) error {
	hoursStr := c.QueryParam("hours")
	hours := 24 // 默认24小时
	if hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 {
			hours = h
		}
	}

	alerts, err := h.usecase.GetSecurityAlerts(c.Request().Context(), hours)
	if err != nil {
		h.logger.With("error", err).Error("failed to get security alerts")
		return err
	}

	return c.Success(alerts)
}

// DeleteLog 删除审计日志
//
//	@Tags			Audit
//	@Summary		删除审计日志
//	@Description	软删除指定的审计日志
//	@ID				delete-audit-log
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"审计日志ID"
//	@Success		200	{object}	web.Resp
//	@Router			/api/v1/audit/logs/{id} [delete]
func (h *AuditHandler) DeleteLog(c *web.Context) error {
	id := c.Param("id")
	if id == "" {
		return errcode.ErrInvalidParam.WithData("message", "审计日志ID不能为空")
	}

	err := h.usecase.Delete(c.Request().Context(), id)
	if err != nil {
		h.logger.With("error", err).With("id", id).Error("failed to delete audit log")
		return err
	}

	return c.Success(nil)
}

// BatchDeleteLogs 批量删除审计日志
//
//	@Tags			Audit
//	@Summary		批量删除审计日志
//	@Description	根据ID列表批量删除审计日志
//	@ID				batch-delete-audit-logs
//	@Accept			json
//	@Produce		json
//	@Param			request	body		domain.BatchDeleteAuditLogReq	true	"批量删除请求"
//	@Success		200		{object}	web.Resp
//	@Router			/api/v1/audit/logs/batch [delete]
func (h *AuditHandler) BatchDeleteLogs(c *web.Context, req domain.BatchDeleteAuditLogReq) error {
	err := h.usecase.BatchDelete(c.Request().Context(), req.IDs)
	if err != nil {
		h.logger.With("error", err).Error("failed to batch delete audit logs")
		return err
	}

	return c.Success(nil)
}

// CleanupOldLogs 清理旧审计日志
//
//	@Tags			Audit
//	@Summary		清理旧审计日志
//	@Description	清理指定天数之前的审计日志
//	@ID				cleanup-old-audit-logs
//	@Accept			json
//	@Produce		json
//	@Param			request	body		domain.CleanupOldLogsReq	true	"清理请求"
//	@Success		200		{object}	web.Resp
//	@Router			/api/v1/audit/logs/cleanup [delete]
func (h *AuditHandler) CleanupOldLogs(c *web.Context, req domain.CleanupOldLogsReq) error {
	err := h.usecase.CleanupOldLogs(c.Request().Context(), req.Days)
	if err != nil {
		h.logger.With("error", err).Error("failed to cleanup old audit logs")
		return err
	}

	return c.Success(nil)
}
