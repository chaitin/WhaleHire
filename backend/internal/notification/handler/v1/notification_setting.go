package v1

import (
	"log/slog"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/errcode"
	"github.com/chaitin/WhaleHire/backend/internal/middleware"
	"github.com/chaitin/WhaleHire/backend/pkg/web"
)

// NotificationSettingHandler 通知设置 HTTP 处理器
type NotificationSettingHandler struct {
	usecase domain.NotificationSettingUsecase
	logger  *slog.Logger
}

// NewNotificationSettingHandler 创建通知设置处理器
func NewNotificationSettingHandler(
	w *web.Web,
	usecase domain.NotificationSettingUsecase,
	logger *slog.Logger,
	auth *middleware.AuthMiddleware,
) *NotificationSettingHandler {
	h := &NotificationSettingHandler{
		usecase: usecase,
		logger:  logger,
	}

	// 注册路由
	g := w.Group("/api/v1/notification-settings")
	g.Use(auth.UserAuth())

	g.POST("", web.BindHandler(h.CreateSetting))
	g.GET("", web.BindHandler(h.ListSettings))
	g.GET("/enabled", web.BindHandler(h.GetEnabledSettings))
	g.GET("/:id", web.BindHandler(h.GetSetting))
	g.PUT("/:id", web.BindHandler(h.UpdateSetting))
	g.DELETE("/:id", web.BindHandler(h.DeleteSetting))

	return h
}

// CreateSetting 创建通知设置
//
//	@Tags			NotificationSetting
//	@Summary		创建通知设置
//	@Description	创建新的通知设置配置，支持钉钉等多种通知渠道
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			param	body		domain.CreateSettingRequest	true	"创建通知设置请求参数"
//	@Success		200		{object}	web.Resp{data=domain.NotificationSetting}
//	@Failure		400		{object}	web.Resp{}	"请求参数错误"
//	@Failure		401		{object}	web.Resp{}	"未授权"
//	@Failure		500		{object}	web.Resp{}	"创建通知设置失败"
//	@Router			/api/v1/notification-settings [post]
func (h *NotificationSettingHandler) CreateSetting(ctx *web.Context, req domain.CreateSettingRequest) error {
	setting := &domain.NotificationSetting{
		Channel:        req.Channel,
		Enabled:        req.Enabled,
		DingTalkConfig: &req.DingTalkConfig,
		MaxRetry:       req.MaxRetry,
		Timeout:        req.Timeout,
		Description:    req.Description,
	}

	if err := h.usecase.CreateSetting(ctx.Request().Context(), setting); err != nil {
		h.logger.Error("create notification setting failed", "error", err)
		return errcode.ErrNotificationSettingCreateFailed
	}

	return ctx.Success(setting)
}

// GetSetting 获取通知设置
//
//	@Tags			NotificationSetting
//	@Summary		获取通知设置详情
//	@Description	根据ID获取指定通知设置的详细信息
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"通知设置ID"
//	@Success		200	{object}	web.Resp{data=domain.NotificationSetting}
//	@Failure		400	{object}	web.Resp{}	"请求参数错误或ID格式无效"
//	@Failure		401	{object}	web.Resp{}	"未授权"
//	@Failure		404	{object}	web.Resp{}	"通知设置不存在"
//	@Failure		500	{object}	web.Resp{}	"获取通知设置失败"
//	@Router			/api/v1/notification-settings/{id} [get]
func (h *NotificationSettingHandler) GetSetting(ctx *web.Context, req domain.GetSettingRequest) error {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "无效的ID格式")
	}

	setting, err := h.usecase.GetSetting(ctx.Request().Context(), id)
	if err != nil {
		h.logger.Error("get notification setting failed", "id", req.ID, "error", err)
		return errcode.ErrNotificationSettingGetFailed
	}

	if setting == nil {
		return errcode.ErrNotificationSettingNotFound
	}

	return ctx.Success(setting)
}

// UpdateSetting 更新通知设置
//
//	@Tags			NotificationSetting
//	@Summary		更新通知设置
//	@Description	根据ID更新指定通知设置的配置信息
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"通知设置ID"
//	@Param			param	body		domain.UpdateSettingRequest	true	"更新通知设置请求参数"
//	@Success		200		{object}	web.Resp{data=domain.NotificationSetting}
//	@Failure		400		{object}	web.Resp{}	"请求参数错误或ID格式无效"
//	@Failure		401		{object}	web.Resp{}	"未授权"
//	@Failure		404		{object}	web.Resp{}	"通知设置不存在"
//	@Failure		500		{object}	web.Resp{}	"更新通知设置失败"
//	@Router			/api/v1/notification-settings/{id} [put]
func (h *NotificationSettingHandler) UpdateSetting(ctx *web.Context, req domain.UpdateSettingRequest) error {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "无效的ID格式")
	}

	setting := &domain.NotificationSetting{
		ID:             id,
		Channel:        req.Channel,
		Enabled:        req.Enabled,
		DingTalkConfig: &req.DingTalkConfig,
		MaxRetry:       req.MaxRetry,
		Timeout:        req.Timeout,
		Description:    req.Description,
	}

	if err := h.usecase.UpdateSetting(ctx.Request().Context(), setting); err != nil {
		h.logger.Error("update notification setting failed", "id", idStr, "error", err)
		return errcode.ErrNotificationSettingUpdateFailed
	}

	return ctx.Success(setting)
}

// DeleteSetting 删除通知设置
//
//	@Tags			NotificationSetting
//	@Summary		删除通知设置
//	@Description	根据ID删除指定的通知设置配置
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"通知设置ID"
//	@Success		200	{object}	web.Resp{}
//	@Failure		400	{object}	web.Resp{}	"请求参数错误或ID格式无效"
//	@Failure		401	{object}	web.Resp{}	"未授权"
//	@Failure		404	{object}	web.Resp{}	"通知设置不存在"
//	@Failure		500	{object}	web.Resp{}	"删除通知设置失败"
//	@Router			/api/v1/notification-settings/{id} [delete]
func (h *NotificationSettingHandler) DeleteSetting(ctx *web.Context, req domain.DeleteSettingRequest) error {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "无效的ID格式")
	}

	if err := h.usecase.DeleteSetting(ctx.Request().Context(), id); err != nil {
		h.logger.Error("delete notification setting failed", "id", req.ID, "error", err)
		return errcode.ErrNotificationSettingDeleteFailed
	}

	return ctx.Success(nil)
}

// ListSettings 获取通知设置列表
//
//	@Tags			NotificationSetting
//	@Summary		获取通知设置列表
//	@Description	获取所有通知设置的列表，包含每个设置的基本信息和配置状态
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	web.Resp{data=domain.ListSettingsResponse}
//	@Failure		401	{object}	web.Resp{}	"未授权"
//	@Failure		500	{object}	web.Resp{}	"获取通知设置列表失败"
//	@Router			/api/v1/notification-settings [get]
func (h *NotificationSettingHandler) ListSettings(ctx *web.Context, req domain.ListSettingsRequest) error {
	settings, err := h.usecase.List(ctx.Request().Context())
	if err != nil {
		h.logger.Error("list notification settings failed", "error", err)
		return errcode.ErrNotificationSettingGetFailed
	}

	return ctx.Success(domain.ListSettingsResponse{
		Settings: settings,
		Total:    len(settings),
	})
}

// GetEnabledSettings 获取启用的通知设置列表
//
//	@Tags			NotificationSetting
//	@Summary		获取启用的通知设置列表
//	@Description	获取所有已启用状态的通知设置列表，用于系统通知功能
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	web.Resp{data=domain.GetEnabledSettingsResponse}
//	@Failure		401	{object}	web.Resp{}	"未授权"
//	@Failure		500	{object}	web.Resp{}	"获取启用通知设置失败"
//	@Router			/api/v1/notification-settings/enabled [get]
func (h *NotificationSettingHandler) GetEnabledSettings(ctx *web.Context, req domain.GetEnabledSettingsRequest) error {
	settings, err := h.usecase.GetEnabledSettings(ctx.Request().Context())
	if err != nil {
		h.logger.Error("get enabled notification settings failed", "error", err)
		return errcode.ErrNotificationSettingGetFailed
	}

	return ctx.Success(domain.GetEnabledSettingsResponse{
		Settings: settings,
		Total:    len(settings),
	})
}
