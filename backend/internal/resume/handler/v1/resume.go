package v1

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/chaitin/WhaleHire/backend/pkg/web"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/errcode"
	"github.com/chaitin/WhaleHire/backend/internal/middleware"
	"github.com/redis/go-redis/v9"
)

type ResumeHandler struct {
	usecase               domain.ResumeUsecase
	jobApplicationUsecase domain.JobApplicationUsecase
	redis                 *redis.Client
	logger                *slog.Logger
}

func NewResumeHandler(
	w *web.Web,
	usecase domain.ResumeUsecase,
	jobApplicationUsecase domain.JobApplicationUsecase,
	redis *redis.Client,
	auth *middleware.AuthMiddleware,
	logger *slog.Logger,
) *ResumeHandler {
	h := &ResumeHandler{
		usecase:               usecase,
		jobApplicationUsecase: jobApplicationUsecase,
		redis:                 redis,
		logger:                logger,
	}

	// 简历管理API路由
	g := w.Group("/api/v1/resume")

	// 需要用户认证的路由
	g.Use(auth.UserAuth())
	g.POST("/upload", web.BaseHandler(h.Upload))
	g.POST("/batch-upload", web.BaseHandler(h.BatchUpload))
	g.GET("/batch-upload/:task_id/status", web.BaseHandler(h.GetBatchUploadStatus))
	g.POST("/batch-upload/:task_id/cancel", web.BaseHandler(h.CancelBatchUpload))
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
//	@Description	上传简历文件并进行解析，支持选择多个岗位
//	@ID				upload-resume
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file				formData	file	true	"简历文件"
//	@Param			job_position_ids	formData	string	false	"岗位ID列表，多个ID用逗号分隔"
//	@Param			source				formData	string	true	"申请来源类型，必需参数，可选值：email（邮箱采集）、manual（手动上传）"
//	@Param			notes				formData	string	false	"备注信息"
//	@Success		200					{object}	web.Resp{data=domain.Resume}
//	@Router			/api/v1/resume/upload [post]
func (h *ResumeHandler) Upload(c *web.Context) error {
	// 获取当前用户
	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission.Wrap(fmt.Errorf("user not found"))
	}

	// 获取上传的文件
	file, header, err := c.Request().FormFile("file")
	if err != nil {
		h.logger.Error("failed to get form file", "error", err)
		return err
	}
	defer file.Close()

	// 获取岗位ID列表
	var jobPositionIDs []string
	if jobPositionIDsStr := c.Request().FormValue("job_position_ids"); jobPositionIDsStr != "" {
		jobPositionIDs = strings.Split(strings.TrimSpace(jobPositionIDsStr), ",")
		// 清理空字符串
		var cleanIDs []string
		for _, id := range jobPositionIDs {
			if trimmedID := strings.TrimSpace(id); trimmedID != "" {
				cleanIDs = append(cleanIDs, trimmedID)
			}
		}
		jobPositionIDs = cleanIDs
	}

	// 获取并校验必需的 source 参数
	sourceStr := c.Request().FormValue("source")
	if sourceStr == "" {
		return errcode.ErrInvalidParam.WithData("message", fmt.Errorf("source parameter is required"))
	}

	// 校验 source 参数的有效性
	sourceType := consts.ResumeSourceType(sourceStr)
	if !sourceType.IsValid() {
		return errcode.ErrInvalidParam.WithData("message", fmt.Errorf("invalid source type: %s, valid values are: %v", sourceStr, consts.ResumeSourceType("").Values()))
	}

	// 获取其他可选参数
	var notes *string
	if notesStr := c.Request().FormValue("notes"); notesStr != "" {
		notes = &notesStr
	}

	// 构建上传请求
	req := &domain.UploadResumeReq{
		UploaderID:     user.ID,
		File:           file,
		Filename:       header.Filename,
		JobPositionIDs: jobPositionIDs,
		Source:         &sourceStr,
		Notes:          notes,
	}

	// 调用业务逻辑
	resume, err := h.usecase.Upload(c.Request().Context(), req)
	if err != nil {
		h.logger.Error("failed to upload resume", "error", err, "user_id", user.ID)
		return err
	}

	// 如果指定了岗位，创建关联关系
	if len(jobPositionIDs) > 0 {
		jobAppReq := &domain.CreateJobApplicationsReq{
			ResumeID:       resume.ID,
			JobPositionIDs: jobPositionIDs,
			Source:         &sourceStr,
			Notes:          notes,
		}

		_, err := h.jobApplicationUsecase.CreateJobApplications(c.Request().Context(), jobAppReq)
		if err != nil {
			h.logger.Error("failed to create job applications", "error", err, "resume_id", resume.ID)
			// 不返回错误，因为简历已经上传成功，只是关联关系创建失败
			h.logger.Warn("resume uploaded successfully but job applications creation failed", "resume_id", resume.ID)
		} else {
			h.logger.Info("job applications created successfully", "resume_id", resume.ID, "job_position_count", len(jobPositionIDs))
		}
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
	// 设置分页参数
	req.Page = c.Page().Page
	req.Size = c.Page().Size

	resp, err := h.usecase.List(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("failed to list resumes", "error", err)
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
	// 设置分页参数
	req.Page = c.Page().Page
	req.Size = c.Page().Size

	resp, err := h.usecase.Search(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("failed to search resumes", "error", err, "keywords", req.Keywords)
		return err
	}

	return c.Success(resp)
}

// Update 更新简历
//
//	@Tags			Resume
//	@Summary		更新简历
//	@Description	更新简历信息，支持更新基本信息、教育经历、工作经历、技能和岗位关联
//	@ID				update-resume
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"简历ID"
//	@Param			param	body		domain.UpdateResumeReq	true	"更新参数，支持更新基本信息、子表数据（教育经历、工作经历、技能）和岗位关联信息"
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
		return errcode.ErrPermission.Wrap(fmt.Errorf("user not found"))
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
		return errcode.ErrPermission.Wrap(fmt.Errorf("user not found"))
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
		return errcode.ErrPermission.Wrap(fmt.Errorf("user not found"))
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
		return errcode.ErrPermission.Wrap(fmt.Errorf("user not found"))
	}

	progress, err := h.usecase.GetParseProgress(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("failed to get resume parse progress", "error", err, "resume_id", id, "user_id", user.ID)
		return err
	}

	h.logger.Info("resume parse progress retrieved", "resume_id", id, "user_id", user.ID, "status", progress.Status)
	return c.Success(progress)
}

// BatchUpload 批量上传简历
//
//	@Tags			Resume
//	@Summary		批量上传简历
//	@Description	批量上传简历文件并进行解析，支持选择多个岗位
//	@ID				batch-upload-resume
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			files				formData	file	true	"简历文件列表"
//	@Param			job_position_ids	formData	string	false	"岗位ID列表，多个ID用逗号分隔"
//	@Param			source				formData	string	true	"申请来源类型，必需参数，可选值：email（邮箱采集）、manual（手动上传）"
//	@Param			notes				formData	string	false	"备注信息"
//	@Success		200					{object}	web.Resp{data=domain.BatchUploadTask}
//	@Router			/api/v1/resume/batch-upload [post]
func (h *ResumeHandler) BatchUpload(c *web.Context) error {
	// 获取当前用户
	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission.Wrap(fmt.Errorf("user not found"))
	}

	// 解析multipart form
	err := c.Request().ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		h.logger.Error("failed to parse multipart form", "error", err)
		return errcode.ErrInvalidParam.Wrap(err)
	}

	// 获取上传的文件列表
	files := c.Request().MultipartForm.File["files"]
	if len(files) == 0 {
		return errcode.ErrInvalidParam.Wrap(fmt.Errorf("no files provided"))
	}

	// 获取岗位ID列表
	var jobPositionIDs []string
	if jobPositionIDsStr := c.Request().FormValue("job_position_ids"); jobPositionIDsStr != "" {
		jobPositionIDs = strings.Split(strings.TrimSpace(jobPositionIDsStr), ",")
		// 清理空字符串
		var cleanIDs []string
		for _, id := range jobPositionIDs {
			if trimmedID := strings.TrimSpace(id); trimmedID != "" {
				cleanIDs = append(cleanIDs, trimmedID)
			}
		}
		jobPositionIDs = cleanIDs
	}

	// 获取并校验必需的 source 参数
	sourceStr := c.Request().FormValue("source")
	if sourceStr == "" {
		return errcode.ErrInvalidParam.WithData("message", fmt.Errorf("source parameter is required"))
	}

	// 校验 source 参数的有效性
	sourceType := consts.ResumeSourceType(sourceStr)
	if !sourceType.IsValid() {
		return errcode.ErrInvalidParam.WithData("message", fmt.Errorf("invalid source type: %s, valid values are: %v", sourceStr, consts.ResumeSourceType("").Values()))
	}

	// 获取其他可选参数
	var notes *string
	if notesStr := c.Request().FormValue("notes"); notesStr != "" {
		notes = &notesStr
	}

	// 构建批量上传请求
	var fileInfos []*domain.BatchUploadFileInfo
	for _, fileHeader := range files {
		file, openErr := fileHeader.Open()
		if openErr != nil {
			h.logger.Error("failed to open file", "filename", fileHeader.Filename, "error", openErr)
			continue
		}
		defer file.Close()
		fileInfos = append(fileInfos, &domain.BatchUploadFileInfo{
			File:     file,
			Filename: fileHeader.Filename,
		})
	}

	req := &domain.BatchUploadResumeReq{
		UploaderID:     user.ID,
		Files:          fileInfos,
		JobPositionIDs: jobPositionIDs,
		Source:         &sourceStr,
		Notes:          notes,
	}

	// 调用业务逻辑
	task, err := h.usecase.BatchUpload(c.Request().Context(), req)
	if err != nil {
		h.logger.Error("failed to batch upload resumes", "error", err)
		return err
	}

	return c.Success(task)
}

// GetBatchUploadStatus 获取批量上传任务状态
//
//	@Tags			Resume
//	@Summary		获取批量上传任务状态
//	@Description	获取批量上传任务的进度和状态信息
//	@ID				get-batch-upload-status
//	@Produce		json
//	@Param			task_id	path		string	true	"任务ID"
//	@Success		200		{object}	web.Resp{data=domain.BatchUploadTask}
//	@Router			/api/v1/resume/batch-upload/{task_id}/status [get]
func (h *ResumeHandler) GetBatchUploadStatus(c *web.Context) error {
	taskID := c.Param("task_id")
	if taskID == "" {
		return errcode.ErrInvalidParam.Wrap(fmt.Errorf("task_id is required"))
	}

	// 调用业务逻辑
	task, err := h.usecase.GetBatchUploadStatus(c.Request().Context(), taskID)
	if err != nil {
		h.logger.Error("failed to get batch upload status", "task_id", taskID, "error", err)
		return err
	}

	return c.Success(task)
}

// CancelBatchUpload 取消批量上传任务
//
//	@Tags			Resume
//	@Summary		取消批量上传任务
//	@Description	取消正在进行的批量上传任务
//	@ID				cancel-batch-upload
//	@Produce		json
//	@Param			task_id	path		string	true	"任务ID"
//	@Success		200		{object}	web.Resp{data=string}
//	@Router			/api/v1/resume/batch-upload/{task_id}/cancel [post]
func (h *ResumeHandler) CancelBatchUpload(c *web.Context) error {
	taskID := c.Param("task_id")
	if taskID == "" {
		return errcode.ErrInvalidParam.Wrap(fmt.Errorf("task_id is required"))
	}

	// 调用业务逻辑
	err := h.usecase.CancelBatchUpload(c.Request().Context(), taskID)
	if err != nil {
		h.logger.Error("failed to cancel batch upload", "task_id", taskID, "error", err)
		return err
	}

	return c.Success("Task cancelled successfully")
}
