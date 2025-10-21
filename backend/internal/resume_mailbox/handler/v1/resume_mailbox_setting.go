package v1

import (
	"log/slog"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/errcode"
	"github.com/chaitin/WhaleHire/backend/internal/middleware"
	"github.com/chaitin/WhaleHire/backend/pkg/web"
	"github.com/google/uuid"
)

type ResumeMailboxSettingHandler struct {
	usecase domain.ResumeMailboxSettingUsecase
	logger  *slog.Logger
}

func NewResumeMailboxSettingHandler(
	w *web.Web,
	usecase domain.ResumeMailboxSettingUsecase,
	logger *slog.Logger,
	auth *middleware.AuthMiddleware,
) *ResumeMailboxSettingHandler {
	h := &ResumeMailboxSettingHandler{
		usecase: usecase,
		logger:  logger,
	}

	// 注册路由
	group := w.Group("/api/v1/resume-mailbox-settings")
	group.Use(auth.UserAuth())

	// 邮箱设置管理
	group.POST("", web.BindHandler(h.CreateSetting))
	group.GET("", web.BaseHandler(h.ListSettings, web.WithPage()))
	group.GET("/:id", web.BaseHandler(h.GetSetting))
	group.PUT("/:id", web.BindHandler(h.UpdateSetting))
	group.DELETE("/:id", web.BaseHandler(h.DeleteSetting))

	// 邮箱连接测试和状态管理
	group.POST("/:id/test", web.BaseHandler(h.TestConnection))
	group.PUT("/:id/status", web.BindHandler(h.UpdateStatus))

	return h
}

// CreateSetting 创建邮箱设置
//
//	@Tags			Resume Mailbox
//	@Summary		创建简历邮箱设置
//	@Description	创建新的简历邮箱设置，用于配置邮箱连接信息和同步参数
//	@ID				create-resume-mailbox-setting
//	@Accept			json
//	@Produce		json
//	@Param			setting	body		domain.CreateResumeMailboxSettingRequest	true	"邮箱设置参数"
//	@Success		200		{object}	web.Resp{data=domain.ResumeMailboxSetting}
//	@Failure		400		{object}	web.Resp{}	"参数错误"
//	@Failure		401		{object}	web.Resp{}	"未授权"
//	@Failure		500		{object}	web.Resp{}	"服务器错误"
//	@Router			/api/v1/resume-mailbox-settings [post]
func (h *ResumeMailboxSettingHandler) CreateSetting(ctx *web.Context, req domain.CreateResumeMailboxSettingRequest) error {
	// 获取当前用户
	user := middleware.GetUser(ctx)
	if user == nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to get user")
		return errcode.ErrPermission
	}

	userID, err := uuid.Parse(user.ID)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Invalid user ID format",
			slog.String("user_id", user.ID),
		)
		return errcode.ErrInvalidParam.WithData("message", "Invalid user ID format")
	}

	req.UploaderID = userID

	setting, err := h.usecase.Create(ctx.Request().Context(), &req)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to create mailbox setting",
			slog.String("error", err.Error()),
		)
		return err
	}

	return ctx.Success(setting)
}

// GetSetting 获取邮箱设置详情
//
//	@Tags			Resume Mailbox
//	@Summary		获取简历邮箱设置详情
//	@Description	根据设置ID获取指定的简历邮箱设置详细信息
//	@ID				get-resume-mailbox-setting
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"邮箱设置ID"
//	@Success		200	{object}	web.Resp{data=domain.ResumeMailboxSetting}
//	@Failure		400	{object}	web.Resp{}	"参数错误"
//	@Failure		401	{object}	web.Resp{}	"未授权"
//	@Failure		404	{object}	web.Resp{}	"设置不存在"
//	@Failure		500	{object}	web.Resp{}	"服务器错误"
//	@Router			/api/v1/resume-mailbox-settings/{id} [get]
func (h *ResumeMailboxSettingHandler) GetSetting(ctx *web.Context) error {
	// 获取当前用户
	user := middleware.GetUser(ctx)
	if user == nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to get user")
		return errcode.ErrPermission
	}

	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Invalid ID format",
			slog.String("id", idStr),
		)
		return errcode.ErrInvalidParam.WithData("message", "Invalid ID format")
	}

	setting, err := h.usecase.GetByID(ctx.Request().Context(), id)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to get mailbox setting",
			slog.String("error", err.Error()),
		)
		return err
	}

	return ctx.Success(setting)
}

// UpdateSetting 更新邮箱设置
//
//	@Tags			Resume Mailbox
//	@Summary		更新简历邮箱设置
//	@Description	更新指定ID的简历邮箱设置信息，包括邮箱连接参数和同步配置
//	@ID				update-resume-mailbox-setting
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string										true	"邮箱设置ID"
//	@Param			setting	body		domain.UpdateResumeMailboxSettingRequest	true	"更新的邮箱设置参数"
//	@Success		200		{object}	web.Resp{data=domain.ResumeMailboxSetting}
//	@Failure		400		{object}	web.Resp{}	"参数错误"
//	@Failure		401		{object}	web.Resp{}	"未授权"
//	@Failure		404		{object}	web.Resp{}	"设置不存在"
//	@Failure		500		{object}	web.Resp{}	"服务器错误"
//	@Router			/api/v1/resume-mailbox-settings/{id} [put]
func (h *ResumeMailboxSettingHandler) UpdateSetting(ctx *web.Context, req domain.UpdateResumeMailboxSettingRequest) error {
	// 获取当前用户
	user := middleware.GetUser(ctx)
	if user == nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to get user")
		return errcode.ErrPermission
	}

	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Invalid ID format",
			slog.String("id", idStr),
		)
		return errcode.ErrInvalidParam.WithData("message", "Invalid ID format")
	}

	setting, err := h.usecase.Update(ctx.Request().Context(), id, &req)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to update mailbox setting",
			slog.String("error", err.Error()),
		)
		return err
	}

	return ctx.Success(setting)
}

// DeleteSetting 删除邮箱设置
//
//	@Tags			Resume Mailbox
//	@Summary		删除简历邮箱设置
//	@Description	根据设置ID删除指定的简历邮箱设置，删除后将停止该邮箱的简历同步
//	@ID				delete-resume-mailbox-setting
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string		true	"邮箱设置ID"
//	@Success		200	{object}	web.Resp{}	"删除成功"
//	@Failure		400	{object}	web.Resp{}	"参数错误"
//	@Failure		401	{object}	web.Resp{}	"未授权"
//	@Failure		404	{object}	web.Resp{}	"设置不存在"
//	@Failure		500	{object}	web.Resp{}	"服务器错误"
//	@Router			/api/v1/resume-mailbox-settings/{id} [delete]
func (h *ResumeMailboxSettingHandler) DeleteSetting(ctx *web.Context) error {
	// 获取当前用户
	user := middleware.GetUser(ctx)
	if user == nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to get user")
		return errcode.ErrPermission
	}

	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Invalid ID format",
			slog.String("id", idStr),
		)
		return errcode.ErrInvalidParam.WithData("message", "Invalid ID format")
	}

	err = h.usecase.Delete(ctx.Request().Context(), id)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to delete mailbox setting",
			slog.String("error", err.Error()),
		)
		return err
	}

	return ctx.Success(nil)
}

// ListSettings 获取邮箱设置列表
//
//	@Tags			Resume Mailbox
//	@Summary		获取简历邮箱设置列表
//	@Description	分页获取当前用户的简历邮箱设置列表，支持按状态筛选
//	@ID				list-resume-mailbox-settings
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"页码，默认为1"
//	@Param			page_size	query		int		false	"每页数量，默认为10"
//	@Param			status		query		string	false	"状态筛选：active(激活)、inactive(未激活)"
//	@Success		200			{object}	web.Resp{data=domain.ListResumeMailboxSettingsResponse}
//	@Failure		400			{object}	web.Resp{}	"参数错误"
//	@Failure		401			{object}	web.Resp{}	"未授权"
//	@Failure		500			{object}	web.Resp{}	"服务器错误"
//	@Router			/api/v1/resume-mailbox-settings [get]
func (h *ResumeMailboxSettingHandler) ListSettings(ctx *web.Context) error {
	// 获取当前用户
	user := middleware.GetUser(ctx)
	if user == nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to get user")
		return errcode.ErrPermission
	}

	userID, err := uuid.Parse(user.ID)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Invalid user ID format",
			slog.String("user_id", user.ID),
		)
		return errcode.ErrInvalidParam.WithData("message", "Invalid user ID format")
	}

	req := &domain.ListResumeMailboxSettingsRequest{
		Pagination: *ctx.Page(),
		UploaderID: &userID,
	}

	settings, err := h.usecase.List(ctx.Request().Context(), req)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to list mailbox settings",
			slog.String("error", err.Error()),
		)
		return err
	}

	return ctx.Success(settings)
}

// TestConnection 测试邮箱连接
//
//	@Tags			Resume Mailbox
//	@Summary		测试简历邮箱连接
//	@Description	测试指定邮箱设置的连接状态，验证邮箱服务器配置是否正确
//	@ID				test-resume-mailbox-connection
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"邮箱设置ID"
//	@Success		200	{object}	web.Resp{data=domain.TestConnectionResponse}
//	@Failure		400	{object}	web.Resp{}	"参数错误"
//	@Failure		401	{object}	web.Resp{}	"未授权"
//	@Failure		404	{object}	web.Resp{}	"设置不存在"
//	@Failure		500	{object}	web.Resp{}	"连接测试失败"
//	@Router			/api/v1/resume-mailbox-settings/{id}/test-connection [post]
func (h *ResumeMailboxSettingHandler) TestConnection(ctx *web.Context) error {
	// 获取当前用户
	user := middleware.GetUser(ctx)
	if user == nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to get user")
		return errcode.ErrPermission
	}

	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Invalid ID format",
			slog.String("id", idStr),
		)
		return errcode.ErrInvalidParam.WithData("message", "Invalid ID format")
	}

	// 获取邮箱设置
	setting, err := h.usecase.GetByID(ctx.Request().Context(), id)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to get mailbox setting",
			slog.String("error", err.Error()),
		)
		return err
	}

	// 构建测试连接请求
	testReq := &domain.TestConnectionRequest{
		EmailAddress:        setting.EmailAddress,
		Protocol:            setting.Protocol,
		Host:                setting.Host,
		Port:                setting.Port,
		UseSsl:              setting.UseSsl,
		Folder:              &setting.Folder,
		AuthType:            setting.AuthType,
		EncryptedCredential: setting.EncryptedCredential,
	}

	err = h.usecase.TestConnection(ctx.Request().Context(), testReq)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to test connection",
			slog.String("error", err.Error()),
		)
		return err
	}

	return ctx.Success(map[string]string{"status": "success", "message": "Connection test successful"})
}

// UpdateStatus 更新邮箱设置状态
func (h *ResumeMailboxSettingHandler) UpdateStatus(ctx *web.Context, req domain.UpdateResumeMailboxSettingStatusRequest) error {
	// 获取当前用户
	user := middleware.GetUser(ctx)
	if user == nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to get user")
		return errcode.ErrPermission
	}

	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Invalid ID format",
			slog.String("id", idStr),
		)
		return errcode.ErrInvalidParam.WithData("message", "Invalid ID format")
	}

	err = h.usecase.UpdateStatus(ctx.Request().Context(), id, req.Status)
	if err != nil {
		h.logger.ErrorContext(ctx.Request().Context(), "Failed to update status",
			slog.String("error", err.Error()),
		)
		return err
	}

	return ctx.Success(map[string]string{"status": "success", "message": "Status updated successfully"})
}
