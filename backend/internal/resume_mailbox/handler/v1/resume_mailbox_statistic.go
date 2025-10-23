package v1

import (
	"log/slog"
	"time"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/errcode"
	"github.com/chaitin/WhaleHire/backend/internal/middleware"
	"github.com/chaitin/WhaleHire/backend/pkg/web"
	"github.com/google/uuid"
)

type ResumeMailboxStatisticHandler struct {
	usecase domain.ResumeMailboxStatisticUsecase
	logger  *slog.Logger
}

func NewResumeMailboxStatisticHandler(
	w *web.Web,
	usecase domain.ResumeMailboxStatisticUsecase,
	logger *slog.Logger,
	auth *middleware.AuthMiddleware,
) *ResumeMailboxStatisticHandler {
	h := &ResumeMailboxStatisticHandler{
		usecase: usecase,
		logger:  logger,
	}

	// 注册路由
	group := w.Group("/api/v1/resume-mailbox-statistics")
	group.Use(auth.UserAuth())

	// 统计数据管理
	group.GET("", web.BaseHandler(h.ListStatistics, web.WithPage()))
	group.GET("/:mailbox_id/:date", web.BaseHandler(h.GetStatistic))
	group.POST("", web.BindHandler(h.UpsertStatistic))
	group.DELETE("/:mailbox_id/:date", web.BaseHandler(h.DeleteStatistic))

	// 统计汇总
	group.GET("/summary", web.BaseHandler(h.GetSummary))

	return h
}

// ListStatistics 获取统计列表
//
//	@Tags			Resume Mailbox Statistics
//	@Summary		获取简历邮箱统计列表
//	@Description	分页获取简历邮箱同步统计数据，支持按邮箱ID、日期范围等条件筛选
//	@ID				list-resume-mailbox-statistics
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"页码，默认为1"
//	@Param			page_size	query		int		false	"每页数量，默认为20"
//	@Param			mailbox_id	query		string	false	"邮箱设置ID"
//	@Param			date_from	query		string	false	"开始日期，格式：2006-01-02"
//	@Param			date_to		query		string	false	"结束日期，格式：2006-01-02"
//	@Param			range		query		string	false	"预设时间范围：7d、30d、90d"
//	@Success		200			{object}	web.Resp{data=domain.ListResumeMailboxStatisticsResponse}
//	@Failure		400			{object}	web.Resp{}	"参数错误"
//	@Failure		401			{object}	web.Resp{}	"未授权"
//	@Failure		500			{object}	web.Resp{}	"服务器错误"
//	@Router			/api/v1/resume-mailbox-statistics [get]
func (h *ResumeMailboxStatisticHandler) ListStatistics(ctx *web.Context) error {
	// 获取当前用户
	user := middleware.GetUser(ctx)
	if user == nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to get user")
		return errcode.ErrPermission
	}

	req := &domain.ListResumeMailboxStatisticsRequest{
		Pagination: *ctx.Page(),
	}

	// 解析查询参数
	if mailboxIDStr := ctx.QueryParam("mailbox_id"); mailboxIDStr != "" {
		mailboxID, err := uuid.Parse(mailboxIDStr)
		if err != nil {
			h.logger.ErrorContext(ctx.Request().Context(), "Invalid mailbox_id format",
				slog.String("mailbox_id", mailboxIDStr),
			)
			return errcode.ErrInvalidParam.WithData("message", "Invalid mailbox_id format")
		}
		req.MailboxID = &mailboxID
	}

	if dateFromStr := ctx.QueryParam("date_from"); dateFromStr != "" {
		dateFrom, err := time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			h.logger.ErrorContext(ctx.Request().Context(), "Invalid date_from format",
				slog.String("date_from", dateFromStr),
			)
			return errcode.ErrInvalidParam.WithData("message", "Invalid date_from format, expected: 2006-01-02")
		}
		req.DateFrom = &dateFrom
	}

	if dateToStr := ctx.QueryParam("date_to"); dateToStr != "" {
		dateTo, err := time.Parse("2006-01-02", dateToStr)
		if err != nil {
			h.logger.ErrorContext(ctx.Request().Context(), "Invalid date_to format",
				slog.String("date_to", dateToStr),
			)
			return errcode.ErrInvalidParam.WithData("message", "Invalid date_to format, expected: 2006-01-02")
		}
		req.DateTo = &dateTo
	}

	if rangeStr := ctx.QueryParam("range"); rangeStr != "" {
		req.Range = &rangeStr
	}

	statistics, err := h.usecase.List(ctx.Request().Context(), req)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to list mailbox statistics",
			slog.String("error", err.Error()),
		)
		return err
	}

	return ctx.Success(statistics)
}

// GetStatistic 获取指定邮箱和日期的统计
//
//	@Tags			Resume Mailbox Statistics
//	@Summary		获取指定邮箱和日期的统计数据
//	@Description	根据邮箱ID和日期获取指定的简历邮箱同步统计数据
//	@ID				get-resume-mailbox-statistic
//	@Accept			json
//	@Produce		json
//	@Param			mailbox_id	path		string	true	"邮箱设置ID"
//	@Param			date		path		string	true	"日期，格式：2006-01-02"
//	@Success		200			{object}	web.Resp{data=domain.ResumeMailboxStatistic}
//	@Failure		400			{object}	web.Resp{}	"参数错误"
//	@Failure		401			{object}	web.Resp{}	"未授权"
//	@Failure		404			{object}	web.Resp{}	"统计数据不存在"
//	@Failure		500			{object}	web.Resp{}	"服务器错误"
//	@Router			/api/v1/resume-mailbox-statistics/{mailbox_id}/{date} [get]
func (h *ResumeMailboxStatisticHandler) GetStatistic(ctx *web.Context) error {
	// 获取当前用户
	user := middleware.GetUser(ctx)
	if user == nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to get user")
		return errcode.ErrPermission
	}

	mailboxIDStr := ctx.Param("mailbox_id")
	mailboxID, err := uuid.Parse(mailboxIDStr)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Invalid mailbox_id format",
			slog.String("mailbox_id", mailboxIDStr),
		)
		return errcode.ErrInvalidParam.WithData("message", "Invalid mailbox_id format")
	}

	dateStr := ctx.Param("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Invalid date format",
			slog.String("date", dateStr),
		)
		return errcode.ErrInvalidParam.WithData("message", "Invalid date format, expected: 2006-01-02")
	}

	statistic, err := h.usecase.GetByMailboxIDAndDate(ctx.Request().Context(), mailboxID, date)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to get mailbox statistic",
			slog.String("error", err.Error()),
		)
		return err
	}

	return ctx.Success(statistic)
}

// UpsertStatistic 更新或创建统计记录
//
//	@Tags			Resume Mailbox Statistics
//	@Summary		更新或创建统计记录
//	@Description	更新或创建简历邮箱同步统计数据，如果指定日期的记录存在则更新，否则创建新记录
//	@ID				upsert-resume-mailbox-statistic
//	@Accept			json
//	@Produce		json
//	@Param			statistic	body		domain.UpsertResumeMailboxStatisticRequest	true	"统计数据参数"
//	@Success		200			{object}	web.Resp{data=domain.ResumeMailboxStatistic}
//	@Failure		400			{object}	web.Resp{}	"参数错误"
//	@Failure		401			{object}	web.Resp{}	"未授权"
//	@Failure		500			{object}	web.Resp{}	"服务器错误"
//	@Router			/api/v1/resume-mailbox-statistics [post]
func (h *ResumeMailboxStatisticHandler) UpsertStatistic(ctx *web.Context, req domain.UpsertResumeMailboxStatisticRequest) error {
	// 获取当前用户
	user := middleware.GetUser(ctx)
	if user == nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to get user")
		return errcode.ErrPermission
	}

	statistic, err := h.usecase.Upsert(ctx.Request().Context(), &req)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to upsert mailbox statistic",
			slog.String("error", err.Error()),
		)
		return err
	}

	return ctx.Success(statistic)
}

// DeleteStatistic 删除统计记录
//
//	@Tags			Resume Mailbox Statistics
//	@Summary		删除统计记录
//	@Description	根据邮箱ID和日期删除指定的简历邮箱同步统计数据
//	@ID				delete-resume-mailbox-statistic
//	@Accept			json
//	@Produce		json
//	@Param			mailbox_id	path		string		true	"邮箱设置ID"
//	@Param			date		path		string		true	"日期，格式：2006-01-02"
//	@Success		200			{object}	web.Resp{}	"删除成功"
//	@Failure		400			{object}	web.Resp{}	"参数错误"
//	@Failure		401			{object}	web.Resp{}	"未授权"
//	@Failure		404			{object}	web.Resp{}	"统计数据不存在"
//	@Failure		500			{object}	web.Resp{}	"服务器错误"
//	@Router			/api/v1/resume-mailbox-statistics/{mailbox_id}/{date} [delete]
func (h *ResumeMailboxStatisticHandler) DeleteStatistic(ctx *web.Context) error {
	// 获取当前用户
	user := middleware.GetUser(ctx)
	if user == nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to get user")
		return errcode.ErrPermission
	}

	mailboxIDStr := ctx.Param("mailbox_id")
	mailboxID, err := uuid.Parse(mailboxIDStr)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Invalid mailbox_id format",
			slog.String("mailbox_id", mailboxIDStr),
		)
		return errcode.ErrInvalidParam.WithData("message", "Invalid mailbox_id format")
	}

	dateStr := ctx.Param("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Invalid date format",
			slog.String("date", dateStr),
		)
		return errcode.ErrInvalidParam.WithData("message", "Invalid date format, expected: 2006-01-02")
	}

	err = h.usecase.Delete(ctx.Request().Context(), mailboxID, date)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to delete mailbox statistic",
			slog.String("error", err.Error()),
		)
		return err
	}

	return ctx.Success(nil)
}

// GetSummary 获取统计汇总数据
//
//	@Tags			Resume Mailbox Statistics
//	@Summary		获取统计汇总数据
//	@Description	获取简历邮箱同步统计的汇总数据，包括总计数据和成功率等指标
//	@ID				get-resume-mailbox-statistics-summary
//	@Accept			json
//	@Produce		json
//	@Param			mailbox_id	query		string	false	"邮箱设置ID"
//	@Param			date_from	query		string	false	"开始日期，格式：2006-01-02"
//	@Param			date_to		query		string	false	"结束日期，格式：2006-01-02"
//	@Param			range		query		string	false	"预设时间范围：7d、30d、90d"
//	@Success		200			{object}	web.Resp{data=domain.MailboxStatisticsSummary}
//	@Failure		400			{object}	web.Resp{}	"参数错误"
//	@Failure		401			{object}	web.Resp{}	"未授权"
//	@Failure		500			{object}	web.Resp{}	"服务器错误"
//	@Router			/api/v1/resume-mailbox-statistics/summary [get]
func (h *ResumeMailboxStatisticHandler) GetSummary(ctx *web.Context) error {

	req := &domain.GetMailboxStatisticsSummaryRequest{}

	// 解析查询参数
	if mailboxIDStr := ctx.QueryParam("mailbox_id"); mailboxIDStr != "" {
		mailboxID, err := uuid.Parse(mailboxIDStr)
		if err != nil {
			h.logger.ErrorContext(ctx.Request().Context(), "Invalid mailbox_id format",
				slog.String("mailbox_id", mailboxIDStr),
			)
			return errcode.ErrInvalidParam.WithData("message", "Invalid mailbox_id format")
		}
		req.MailboxID = &mailboxID
	}

	if dateFromStr := ctx.QueryParam("date_from"); dateFromStr != "" {
		dateFrom, err := time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			h.logger.ErrorContext(ctx.Request().Context(), "Invalid date_from format",
				slog.String("date_from", dateFromStr),
			)
			return errcode.ErrInvalidParam.WithData("message", "Invalid date_from format, expected: 2006-01-02")
		}
		req.DateFrom = &dateFrom
	}

	if dateToStr := ctx.QueryParam("date_to"); dateToStr != "" {
		dateTo, err := time.Parse("2006-01-02", dateToStr)
		if err != nil {
			h.logger.ErrorContext(ctx.Request().Context(), "Invalid date_to format",
				slog.String("date_to", dateToStr),
			)
			return errcode.ErrInvalidParam.WithData("message", "Invalid date_to format, expected: 2006-01-02")
		}
		req.DateTo = &dateTo
	}

	if rangeStr := ctx.QueryParam("range"); rangeStr != "" {
		req.Range = &rangeStr
	}

	summary, err := h.usecase.GetSummary(ctx.Request().Context(), req)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to get mailbox statistics summary",
			slog.String("error", err.Error()),
		)
		return err
	}

	return ctx.Success(summary)
}
