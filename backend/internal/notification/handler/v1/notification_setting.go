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
	
	// 新增路由：按名称和通道类型查询
	g.GET("/by-name/:name/channel/:channel", web.BindHandler(h.GetSettingByNameAndChannel))
	g.GET("/by-channel/:channel", web.BindHandler(h.GetSettingsByChannel))

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
		Name:           req.Name,
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
		Name:           req.Name,
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
//	@Description	分页获取通知设置列表，支持按渠道、启用状态和关键词过滤
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int							false	"页码，默认为1"
//	@Param			size	query		int							false	"每页大小，默认为20"
//	@Param			channel	query		consts.NotificationChannel	false	"通知渠道过滤"
//	@Param			enabled	query		bool						false	"启用状态过滤"
//	@Param			keyword	query		string						false	"关键词搜索（名称或描述）"
//	@Success		200		{object}	web.Resp{data=domain.ListSettingsResponse}
//	@Failure		401		{object}	web.Resp{}	"未授权"
//	@Failure		500		{object}	web.Resp{}	"获取通知设置列表失败"
//	@Router			/api/v1/notification-settings [get]
func (h *NotificationSettingHandler) ListSettings(ctx *web.Context, req domain.ListSettingsRequest) error {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}

	resp, err := h.usecase.List(ctx.Request().Context(), &req)
	if err != nil {
		h.logger.Error("list notification settings failed", "error", err)
		return errcode.ErrNotificationSettingGetFailed
	}

	return ctx.Success(resp)
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
	})
}

// GetSettingByNameAndChannel 根据名称和通道类型获取通知设置
//
//	@Tags			NotificationSetting
//	@Summary		根据名称和通道类型获取通知设置
//	@Description	根据配置名称和通道类型获取指定通知设置的详细信息
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			name	path		string	true	"配置名称"
//	@Param			channel	path		string	true	"通道类型"
//	@Success		200		{object}	web.Resp{data=domain.NotificationSetting}
//	@Failure		400		{object}	web.Resp{}	"请求参数错误"
//	@Failure		401		{object}	web.Resp{}	"未授权"
//	@Failure		404		{object}	web.Resp{}	"通知设置不存在"
//	@Failure		500		{object}	web.Resp{}	"获取通知设置失败"
//	@Router			/api/v1/notification-settings/by-name/{name}/channel/{channel} [get]
func (h *NotificationSettingHandler) GetSettingByNameAndChannel(ctx *web.Context, req domain.GetSettingByNameAndChannelRequest) error {
	setting, err := h.usecase.GetSettingByNameAndChannel(ctx.Request().Context(), req.Name, req.Channel)
	if err != nil {
		h.logger.Error("get notification setting by name and channel failed", "name", req.Name, "channel", string(req.Channel), "error", err)
		return errcode.ErrNotificationSettingGetFailed
	}

	if setting == nil {
		return errcode.ErrNotificationSettingNotFound
	}

	return ctx.Success(setting)
}

// GetSettingsByChannel 根据通道类型获取所有通知设置
//
//	@Tags			NotificationSetting
//	@Summary		根据通道类型获取所有通知设置
//	@Description	根据通道类型获取该类型下的所有通知设置列表
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			channel	path		string	true	"通道类型"
//	@Success		200		{object}	web.Resp{data=domain.GetSettingsByChannelResponse}
//	@Failure		400		{object}	web.Resp{}	"请求参数错误"
//	@Failure		401		{object}	web.Resp{}	"未授权"
//	@Failure		500		{object}	web.Resp{}	"获取通知设置失败"
//	@Router			/api/v1/notification-settings/by-channel/{channel} [get]
func (h *NotificationSettingHandler) GetSettingsByChannel(ctx *web.Context, req domain.GetSettingsByChannelRequest) error {
	settings, err := h.usecase.GetSettingsByChannel(ctx.Request().Context(), req.Channel)
	if err != nil {
		h.logger.Error("get notification settings by channel failed", "channel", string(req.Channel), "error", err)
		return errcode.ErrNotificationSettingGetFailed
	}

	return ctx.Success(domain.GetSettingsByChannelResponse{
		Settings: settings,
	})
}
