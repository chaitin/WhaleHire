package v1

import (
	"log/slog"

	"github.com/GoYoko/web"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/errcode"
	"github.com/chaitin/WhaleHire/backend/internal/middleware"
)

type ResumeHandler struct {
	usecase domain.ResumeUsecase
	logger  *slog.Logger
}

func NewResumeHandler(
	w *web.Web,
	usecase domain.ResumeUsecase,
	auth *middleware.AuthMiddleware,
	logger *slog.Logger,
) *ResumeHandler {
	h := &ResumeHandler{
		usecase: usecase,
		logger:  logger,
	}

	// 简历管理API路由
	g := w.Group("/api/v1/resume")

	// 需要用户认证的路由
	g.Use(auth.UserAuth())
	g.POST("/upload", web.BaseHandler(h.Upload))
	g.GET("/:id", web.BaseHandler(h.GetByID))
	g.GET("/list", web.BindHandler(h.List, web.WithPage()))
	g.GET("/search", web.BindHandler(h.Search, web.WithPage()))
	g.PUT("/:id", web.BindHandler(h.Update))
	g.DELETE("/:id", web.BaseHandler(h.Delete))
	g.POST("/:id/reparse", web.BaseHandler(h.Reparse))
	g.GET("/:id/progress", web.BaseHandler(h.GetParseProgress))

	return h
}

// Upload 上传简历
//
//	@Tags			Resume
//	@Summary		上传简历
//	@Description	上传简历文件并进行解析
//	@ID				upload-resume
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"简历文件"
//	@Success		200		{object}	web.Resp{data=domain.Resume}
//	@Router			/api/v1/resume/upload [post]
func (h *ResumeHandler) Upload(c *web.Context) error {
	// 获取当前用户
	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission
	}

	// 获取上传的文件
	file, header, err := c.Request().FormFile("file")
	if err != nil {
		h.logger.Error("failed to get form file", "error", err)
		return err
	}
	defer file.Close()

	// 构建上传请求
	req := &domain.UploadResumeReq{
		UserID:   user.ID,
		File:     file,
		Filename: header.Filename,
	}

	// 调用业务逻辑
	resume, err := h.usecase.Upload(c.Request().Context(), req)
	if err != nil {
		h.logger.Error("failed to upload resume", "error", err, "user_id", user.ID)
		return err
	}

	h.logger.Info("resume uploaded successfully", "resume_id", resume.ID, "user_id", user.ID)
	return c.Success(resume)
}

// GetByID 获取简历详情
//
//	@Tags			Resume
//	@Summary		获取简历详情
//	@Description	根据ID获取简历详情
//	@ID				get-resume
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"简历ID"
//	@Success		200	{object}	web.Resp{data=domain.ResumeDetail}
//	@Router			/api/v1/resume/{id} [get]
func (h *ResumeHandler) GetByID(c *web.Context) error {
	id := c.Param("id")
	if id == "" {
		return web.NewBadRequestErr("简历ID不能为空")
	}

	resume, err := h.usecase.GetByID(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("failed to get resume", "error", err, "resume_id", id)
		return err
	}

	return c.Success(resume)
}

// List 获取简历列表
//
//	@Tags			Resume
//	@Summary		获取简历列表
//	@Description	获取简历列表，支持分页
//	@ID				list-resume
//	@Accept			json
//	@Produce		json
//	@Param			page	query		web.Pagination	true	"分页参数"
//	@Success		200		{object}	web.Resp{data=domain.ListResumeResp}
//	@Router			/api/v1/resume/list [get]
func (h *ResumeHandler) List(c *web.Context, req domain.ListResumeReq) error {
	// 获取当前用户
	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission
	}

	// 设置用户ID
	req.UserID = &user.ID
	req.Page = c.Page().Page
	req.Size = c.Page().Size

	resp, err := h.usecase.List(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("failed to list resumes", "error", err, "user_id", user.ID)
		return err
	}

	return c.Success(resp)
}

// Search 搜索简历
//
//	@Tags			Resume
//	@Summary		搜索简历
//	@Description	根据关键词搜索简历
//	@ID				search-resume
//	@Accept			json
//	@Produce		json
//	@Param			keywords	query		string			true	"搜索关键词"
//	@Param			page		query		web.Pagination	true	"分页参数"
//	@Success		200			{object}	web.Resp{data=domain.SearchResumeResp}
//	@Router			/api/v1/resume/search [get]
func (h *ResumeHandler) Search(c *web.Context, req domain.SearchResumeReq) error {
	// 获取当前用户
	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission
	}

	// 设置用户ID过滤
	req.UserID = &user.ID
	// 设置分页参数
	req.Page = c.Page().Page
	req.Size = c.Page().Size

	resp, err := h.usecase.Search(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("failed to search resumes", "error", err, "user_id", user.ID, "keywords", req.Keywords)
		return err
	}

	return c.Success(resp)
}

// Update 更新简历
//
//	@Tags			Resume
//	@Summary		更新简历
//	@Description	更新简历信息，支持更新基本信息、教育经历、工作经历和技能
//	@ID				update-resume
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"简历ID"
//	@Param			param	body		domain.UpdateResumeReq	true	"更新参数，支持更新基本信息和子表数据（教育经历、工作经历、技能）"
//	@Success		200		{object}	web.Resp{data=domain.Resume}
//	@Router			/api/v1/resume/{id} [put]
func (h *ResumeHandler) Update(c *web.Context, req domain.UpdateResumeReq) error {
	id := c.Param("id")
	if id == "" {
		return web.NewBadRequestErr("简历ID不能为空")
	}

	// 获取当前用户
	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission
	}

	req.ID = id

	resume, err := h.usecase.Update(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("failed to update resume", "error", err, "resume_id", id, "user_id", user.ID)
		return err
	}

	h.logger.Info("resume updated successfully", "resume_id", id, "user_id", user.ID)
	return c.Success(resume)
}

// Delete 删除简历
//
//	@Tags			Resume
//	@Summary		删除简历
//	@Description	删除简历
//	@ID				delete-resume
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"简历ID"
//	@Success		200	{object}	web.Resp{}
//	@Router			/api/v1/resume/{id} [delete]
func (h *ResumeHandler) Delete(c *web.Context) error {
	id := c.Param("id")
	if id == "" {
		return web.NewBadRequestErr("简历ID不能为空")
	}

	// 获取当前用户
	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission
	}

	err := h.usecase.Delete(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("failed to delete resume", "error", err, "resume_id", id, "user_id", user.ID)
		return err
	}

	h.logger.Info("resume deleted successfully", "resume_id", id, "user_id", user.ID)
	return c.Success(nil)
}

// Reparse 重新解析简历
//
//	@Tags			Resume
//	@Summary		重新解析简历
//	@Description	重新解析简历内容
//	@ID				reparse-resume
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"简历ID"
//	@Success		200	{object}	web.Resp{}
//	@Router			/api/v1/resume/{id}/reparse [post]
func (h *ResumeHandler) Reparse(c *web.Context) error {
	id := c.Param("id")
	if id == "" {
		return web.NewBadRequestErr("简历ID不能为空")
	}

	// 获取当前用户
	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission
	}

	err := h.usecase.Reparse(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("failed to reparse resume", "error", err, "resume_id", id, "user_id", user.ID)
		return err
	}

	h.logger.Info("resume reparse started", "resume_id", id, "user_id", user.ID)
	return c.Success(nil)
}

// GetParseProgress 获取简历解析进度
//
//	@Tags			Resume
//	@Summary		获取简历解析进度
//	@Description	获取简历解析进度信息
//	@ID				get-resume-progress
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"简历ID"
//	@Success		200	{object}	web.Resp{data=domain.ResumeParseProgress}
//	@Router			/api/v1/resume/{id}/progress [get]
func (h *ResumeHandler) GetParseProgress(c *web.Context) error {
	id := c.Param("id")
	if id == "" {
		return web.NewBadRequestErr("简历ID不能为空")
	}

	// 获取当前用户
	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission
	}

	progress, err := h.usecase.GetParseProgress(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("failed to get resume parse progress", "error", err, "resume_id", id, "user_id", user.ID)
		return err
	}

	h.logger.Info("resume parse progress retrieved", "resume_id", id, "user_id", user.ID, "status", progress.Status)
	return c.Success(progress)
}
