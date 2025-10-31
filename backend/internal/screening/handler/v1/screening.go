package v1

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/errcode"
	"github.com/chaitin/WhaleHire/backend/internal/middleware"
	"github.com/chaitin/WhaleHire/backend/pkg/web"
)

// ScreeningHandler 智能筛选 HTTP 处理器
type ScreeningHandler struct {
	usecase domain.ScreeningUsecase
	logger  *slog.Logger
}

// NewScreeningHandler 注册路由
func NewScreeningHandler(
	w *web.Web,
	usecase domain.ScreeningUsecase,
	auth *middleware.AuthMiddleware,
	logger *slog.Logger,
) *ScreeningHandler {
	handler := &ScreeningHandler{
		usecase: usecase,
		logger:  logger.With("module", "screening_handler"),
	}

	group := w.Group("/api/v1/screening")
	group.Use(auth.UserAuth())
	group.POST("/tasks", web.BindHandler(handler.CreateTask))
	group.POST("/tasks/:id/start", web.BaseHandler(handler.StartTask))
	group.POST("/tasks/:id/cancel", web.BaseHandler(handler.CancelTask))
	group.DELETE("/tasks/:id", web.BaseHandler(handler.DeleteTask))
	group.GET("/tasks", web.BindHandler(handler.ListTasks, web.WithPage()))
	group.GET("/tasks/:id", web.BaseHandler(handler.GetTask))
	group.GET("/tasks/:id/progress", web.BaseHandler(handler.GetTaskProgress))
	group.GET("/tasks/:id/metrics", web.BaseHandler(handler.GetMetrics))
	group.GET("/tasks/:task_id/results/:resume_id", web.BaseHandler(handler.GetResult))
	group.GET("/tasks/:task_id/resumes/:resume_id/progress", web.BaseHandler(handler.GetResumeProgress))
	group.GET("/tasks/:task_id/resumes/:resume_id/node-runs", web.BaseHandler(handler.GetNodeRuns))
	group.GET("/results", web.BindHandler(handler.ListResults, web.WithPage()))
	group.POST("/weights/preview", web.BindHandler(handler.PreviewWeights))

	// Weight template routes
	group.POST("/weights/templates", web.BindHandler(handler.CreateWeightTemplate))
	group.GET("/weights/templates", web.BindHandler(handler.ListWeightTemplates, web.WithPage()))
	group.GET("/weights/templates/:id", web.BaseHandler(handler.GetWeightTemplate))
	group.PUT("/weights/templates/:id", web.BindHandler(handler.UpdateWeightTemplate))
	group.DELETE("/weights/templates/:id", web.BaseHandler(handler.DeleteWeightTemplate))

	return handler
}

// CreateTask 创建筛选任务
//
//	@Tags			Screening
//	@Summary		创建筛选任务
//	@Description	创建智能筛选任务，支持多份简历批量筛选
//	@ID				create-screening-task
//	@Accept			json
//	@Produce		json
//	@Param			param	body		domain.CreateScreeningTaskReq	true	"创建筛选任务参数"
//	@Success		200		{object}	web.Resp{data=domain.CreateScreeningTaskResp}
//	@Router			/api/v1/screening/tasks [post]
func (h *ScreeningHandler) CreateTask(c *web.Context, req domain.CreateScreeningTaskReq) error {
	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission
	}
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "当前用户ID格式不正确")
	}
	req.CreatedBy = userID

	if len(req.ResumeIDs) == 0 {
		return errcode.ErrInvalidParam.WithData("message", "请至少选择一份简历")
	}

	resp, err := h.usecase.CreateScreeningTask(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("创建筛选任务失败", slog.Any("err", err), slog.String("user_id", user.ID))
		return err
	}
	return c.Success(resp)
}

// StartTask 启动筛选任务
//
//	@Tags			Screening
//	@Summary		启动筛选任务
//	@Description	启动指定的筛选任务，开始执行智能筛选流程，返回任务ID、岗位ID和简历ID数组用于后续调试和请求
//	@ID				start-screening-task
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"任务ID"
//	@Success		200	{object}	web.Resp{data=domain.StartScreeningTaskResp}
//	@Router			/api/v1/screening/tasks/{id}/start [post]
func (h *ScreeningHandler) StartTask(c *web.Context) error {
	taskID, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "任务ID格式不正确")
	}

	resp, err := h.usecase.StartScreeningTask(c.Request().Context(), &domain.StartScreeningTaskReq{
		TaskID: taskID,
	})
	if err != nil {
		h.logger.Error("启动筛选任务失败", slog.Any("err", err), slog.Any("task_id", taskID))
		return err
	}
	return c.Success(resp)
}

// CancelTask 取消筛选任务
//
//	@Tags			Screening
//	@Summary		取消筛选任务
//	@Description	取消正在进行中的筛选任务，停止剩余简历的匹配过程
//	@ID				cancel-screening-task
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"任务ID"
//	@Success		200	{object}	web.Resp{data=domain.CancelScreeningTaskResp}
//	@Router			/api/v1/screening/tasks/{id}/cancel [post]
func (h *ScreeningHandler) CancelTask(c *web.Context) error {
	taskID, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "任务ID格式不正确")
	}

	resp, err := h.usecase.CancelScreeningTask(c.Request().Context(), &domain.CancelScreeningTaskReq{
		TaskID: taskID,
	})
	if err != nil {
		h.logger.Error("取消筛选任务失败", slog.Any("err", err), slog.Any("task_id", taskID))
		return err
	}
	return c.Success(resp)
}

// DeleteTask 删除筛选任务
//
//	@Tags			Screening
//	@Summary		删除筛选任务
//	@Description	删除指定的筛选任务，正在进行中的任务不能删除
//	@ID				delete-screening-task
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"任务ID"
//	@Success		200	{object}	web.Resp{data=domain.DeleteScreeningTaskResp}
//	@Router			/api/v1/screening/tasks/{id} [delete]
func (h *ScreeningHandler) DeleteTask(c *web.Context) error {
	taskID, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "任务ID格式不正确")
	}

	resp, err := h.usecase.DeleteScreeningTask(c.Request().Context(), &domain.DeleteScreeningTaskReq{
		TaskID: taskID,
	})
	if err != nil {
		h.logger.Error("删除筛选任务失败", slog.Any("err", err), slog.Any("task_id", taskID))
		return err
	}

	return c.Success(resp)
}

// GetTask 查询任务详情
//
//	@Tags			Screening
//	@Summary		查询任务详情
//	@Description	根据任务ID查询筛选任务的详细信息
//	@ID				get-screening-task
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"任务ID"
//	@Success		200	{object}	web.Resp{data=domain.GetScreeningTaskResp}
//	@Router			/api/v1/screening/tasks/{id} [get]
func (h *ScreeningHandler) GetTask(c *web.Context) error {
	taskID, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "任务ID格式不正确")
	}

	resp, err := h.usecase.GetScreeningTask(c.Request().Context(), &domain.GetScreeningTaskReq{
		TaskID: taskID,
	})
	if err != nil {
		h.logger.Error("查询筛选任务失败", slog.Any("err", err), slog.Any("task_id", taskID))
		return err
	}
	return c.Success(resp)
}

// ListTasks 分页查询任务
//
//	@Tags			Screening
//	@Summary		分页查询任务
//	@Description	分页查询筛选任务列表，支持按状态筛选
//	@ID				list-screening-tasks
//	@Accept			json
//	@Produce		json
//	@Param			page			query		int							false	"页码，默认为1"
//	@Param			size			query		int							false	"每页大小，默认为10"
//	@Param			job_position_id	query		string						false	"职位ID过滤条件"
//	@Param			status			query		consts.ScreeningTaskStatus	false	"任务状态过滤条件"
//	@Param			created_by		query		string						false	"创建者用户ID过滤条件"
//	@Param			start_time		query		string						false	"任务创建时间范围起始时间，格式：2006-01-02T15:04:05Z07:00"
//	@Param			end_time		query		string						false	"任务创建时间范围结束时间，格式：2006-01-02T15:04:05Z07:00"
//	@Success		200				{object}	web.Resp{data=domain.ListScreeningTasksResp}
//	@Router			/api/v1/screening/tasks [get]
func (h *ScreeningHandler) ListTasks(c *web.Context, req domain.ListScreeningTasksReq) error {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}
	resp, err := h.usecase.ListScreeningTasks(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("获取筛选任务列表失败", slog.Any("err", err))
		return err
	}
	return c.Success(resp)
}

// ListResults 分页查询筛选结果
//
//	@Tags			Screening
//	@Summary		分页查询筛选结果
//	@Description	分页查询筛选结果列表，支持按任务ID、简历ID等条件筛选
//	@ID				list-screening-results
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"页码，默认为1"
//	@Param			size		query		int		false	"每页大小，默认为10"
//	@Param			task_id		query		string	false	"筛选任务ID过滤条件"
//	@Param			resume_id	query		string	false	"简历ID过滤条件"
//	@Param			min_score	query		number	false	"最小匹配分数过滤条件"
//	@Param			max_score	query		number	false	"最大匹配分数过滤条件"
//	@Success		200			{object}	web.Resp{data=domain.ListScreeningResultsResp}
//	@Router			/api/v1/screening/results [get]
func (h *ScreeningHandler) ListResults(c *web.Context, req domain.ListScreeningResultsReq) error {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}

	resp, err := h.usecase.ListScreeningResults(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("获取筛选结果列表失败", slog.Any("err", err))
		return err
	}
	return c.Success(resp)
}

// GetResult 查询单条筛选结果
//
//	@Tags			Screening
//	@Summary		查询单条筛选结果
//	@Description	根据任务ID和简历ID查询具体的筛选结果详情
//	@ID				get-screening-result
//	@Accept			json
//	@Produce		json
//	@Param			task_id		path		string	true	"任务ID"
//	@Param			resume_id	path		string	true	"简历ID"
//	@Success		200			{object}	web.Resp{data=domain.GetScreeningResultResp}
//	@Router			/api/v1/screening/tasks/{task_id}/results/{resume_id} [get]
func (h *ScreeningHandler) GetResult(c *web.Context) error {
	taskID, err := parseUUIDParam(c.Param("task_id"))
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "任务ID格式不正确")
	}
	resumeID, err := parseUUIDParam(c.Param("resume_id"))
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "简历ID格式不正确")
	}

	resp, err := h.usecase.GetScreeningResult(c.Request().Context(), &domain.GetScreeningResultReq{
		TaskID:   taskID,
		ResumeID: resumeID,
	})
	if err != nil {
		h.logger.Error("获取筛选结果失败", slog.Any("err", err), slog.Any("task_id", taskID), slog.Any("resume_id", resumeID))
		return err
	}
	return c.Success(resp)
}

// GetTaskProgress 查询任务进度
//
//	@Tags			Screening
//	@Summary		查询任务进度
//	@Description	根据任务ID查询筛选任务的执行进度信息
//	@ID				get-task-progress
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"任务ID"
//	@Success		200	{object}	web.Resp{data=domain.GetTaskProgressResp}
//	@Router			/api/v1/screening/tasks/{id}/progress [get]
func (h *ScreeningHandler) GetTaskProgress(c *web.Context) error {
	taskID, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "任务ID格式不正确")
	}

	resp, err := h.usecase.GetTaskProgress(c.Request().Context(), &domain.GetTaskProgressReq{
		TaskID: taskID,
	})
	if err != nil {
		h.logger.Error("获取任务进度失败", slog.Any("err", err), slog.Any("task_id", taskID))
		return err
	}
	return c.Success(resp)
}

// GetResumeProgress 查询单个简历匹配进度
//
//	@Tags			Screening
//	@Summary		查询单个简历匹配进度
//	@Description	根据任务ID和简历ID查询单个简历的匹配进度信息
//	@ID				get-screening-resume-progress
//	@Accept			json
//	@Produce		json
//	@Param			task_id		path		string	true	"任务ID"
//	@Param			resume_id	path		string	true	"简历ID"
//	@Success		200			{object}	web.Resp{data=domain.GetResumeProgressResp}
//	@Router			/api/v1/screening/tasks/{task_id}/resumes/{resume_id}/progress [get]
func (h *ScreeningHandler) GetResumeProgress(c *web.Context) error {
	taskID, err := parseUUIDParam(c.Param("task_id"))
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "任务ID格式不正确")
	}
	resumeID, err := parseUUIDParam(c.Param("resume_id"))
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "简历ID格式不正确")
	}

	resp, err := h.usecase.GetResumeProgress(c.Request().Context(), &domain.GetResumeProgressReq{
		TaskID:   taskID,
		ResumeID: resumeID,
	})
	if err != nil {
		h.logger.Error("获取简历匹配进度失败", slog.Any("err", err), slog.Any("task_id", taskID), slog.Any("resume_id", resumeID))
		return err
	}
	return c.Success(resp)
}

// GetMetrics 查询任务运行指标
//
//	@Tags			Screening
//	@Summary		查询任务运行指标
//	@Description	根据任务ID查询筛选任务的运行指标和统计信息
//	@ID				get-screening-metrics
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"任务ID"
//	@Success		200	{object}	web.Resp{data=domain.GetScreeningMetricsResp}
//	@Router			/api/v1/screening/tasks/{id}/metrics [get]
func (h *ScreeningHandler) GetMetrics(c *web.Context) error {
	taskID, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "任务ID格式不正确")
	}

	resp, err := h.usecase.GetScreeningMetrics(c.Request().Context(), &domain.GetScreeningMetricsReq{
		TaskID: taskID,
	})
	if err != nil {
		h.logger.Error("获取筛选任务指标失败", slog.Any("err", err), slog.Any("task_id", taskID))
		return err
	}
	return c.Success(resp)
}

// GetNodeRuns 查询节点运行记录
//
//	@Tags			Screening
//	@Summary		查询节点运行记录
//	@Description	根据任务ID和简历ID查询匹配过程中的节点运行记录
//	@ID				get-screening-node-runs
//	@Accept			json
//	@Produce		json
//	@Param			task_id		path		string	true	"任务ID"
//	@Param			resume_id	path		string	true	"简历ID"
//	@Success		200			{object}	web.Resp{data=domain.GetNodeRunsResp}
//	@Router			/api/v1/screening/tasks/{task_id}/resumes/{resume_id}/node-runs [get]
func (h *ScreeningHandler) GetNodeRuns(c *web.Context) error {
	taskID, err := parseUUIDParam(c.Param("task_id"))
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "任务ID格式不正确")
	}
	resumeID, err := parseUUIDParam(c.Param("resume_id"))
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "简历ID格式不正确")
	}

	resp, err := h.usecase.GetNodeRuns(c.Request().Context(), &domain.GetNodeRunsReq{
		TaskID:   taskID,
		ResumeID: resumeID,
	})
	if err != nil {
		h.logger.Error("获取节点运行记录失败", slog.Any("err", err), slog.Any("task_id", taskID), slog.Any("resume_id", resumeID))
		return err
	}
	return c.Success(resp)
}

// PreviewWeights 预览权重
//
//	@Tags			Screening
//	@Summary		预览权重
//	@Description	根据岗位信息自动推理维度权重，支持用户预览和确认
//	@ID				preview-weights
//	@Accept			json
//	@Produce		json
//	@Param			param	body		domain.PreviewWeightsReq	true	"预览权重参数"
//	@Success		200		{object}	web.Resp{data=domain.PreviewWeightsResp}
//	@Router			/api/v1/screening/weights/preview [post]
func (h *ScreeningHandler) PreviewWeights(c *web.Context, req domain.PreviewWeightsReq) error {
	if req.JobPositionID == uuid.Nil {
		return errcode.ErrInvalidParam.WithData("message", "岗位ID不能为空")
	}

	resp, err := h.usecase.PreviewWeights(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("预览权重失败", slog.Any("err", err), slog.String("job_position_id", req.JobPositionID.String()))
		return err
	}
	return c.Success(resp)
}

func parseUUIDParam(value string) (uuid.UUID, error) {
	if value == "" {
		return uuid.Nil, fmt.Errorf("参数不能为空")
	}
	return uuid.Parse(value)
}

// CreateWeightTemplate 创建权重模板
//
//	@Tags			Screening
//	@Summary		创建权重模板
//	@Description	创建权重配置模板，支持保存常用权重配置以便复用
//	@ID				create-weight-template
//	@Accept			json
//	@Produce		json
//	@Param			param	body		domain.CreateWeightTemplateReq	true	"创建权重模板参数"
//	@Success		200		{object}	web.Resp{data=domain.WeightTemplateResp}
//	@Router			/api/v1/screening/weights/templates [post]
func (h *ScreeningHandler) CreateWeightTemplate(c *web.Context, req domain.CreateWeightTemplateReq) error {
	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission
	}
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "当前用户ID格式不正确")
	}

	resp, err := h.usecase.CreateWeightTemplate(c.Request().Context(), &req, userID)
	if err != nil {
		h.logger.Error("创建权重模板失败", slog.Any("err", err), slog.String("user_id", user.ID))
		return errcode.ErrWeightTemplateCreateFailed.WithData("message", err.Error())
	}
	return c.Success(resp)
}

// GetWeightTemplate 获取权重模板详情
//
//	@Tags			Screening
//	@Summary		获取权重模板详情
//	@Description	根据模板ID获取权重模板详情信息
//	@ID				get-weight-template
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"模板ID"
//	@Success		200	{object}	web.Resp{data=domain.WeightTemplateResp}
//	@Router			/api/v1/screening/weights/templates/{id} [get]
func (h *ScreeningHandler) GetWeightTemplate(c *web.Context) error {
	templateID, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "模板ID格式不正确")
	}

	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission
	}
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "当前用户ID格式不正确")
	}

	resp, err := h.usecase.GetWeightTemplate(c.Request().Context(), templateID, userID)
	if err != nil {
		h.logger.Error("获取权重模板详情失败", slog.Any("err", err), slog.Any("template_id", templateID))
		return err
	}
	return c.Success(resp)
}

// ListWeightTemplates 查询权重模板列表
//
//	@Tags			Screening
//	@Summary		查询权重模板列表
//	@Description	分页查询权重模板列表，支持按名称模糊搜索
//	@ID				list-weight-templates
//	@Accept			json
//	@Produce		json
//	@Param			page		query		web.Pagination	true	"分页参数"
//	@Param			name		query		string			false	"模板名称（模糊匹配）"
//	@Success		200			{object}	web.Resp{data=domain.ListWeightTemplatesResp}
//	@Router			/api/v1/screening/weights/templates [get]
func (h *ScreeningHandler) ListWeightTemplates(c *web.Context, req domain.ListWeightTemplatesReq) error {
	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission
	}
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "当前用户ID格式不正确")
	}

	// 设置分页参数
	req.Page = c.Page().Page
	req.Size = c.Page().Size

	resp, err := h.usecase.ListWeightTemplates(c.Request().Context(), &req, userID)
	if err != nil {
		h.logger.Error("查询权重模板列表失败", slog.Any("err", err), slog.String("user_id", user.ID))
		return err
	}
	return c.Success(resp)
}

// UpdateWeightTemplate 更新权重模板
//
//	@Tags			Screening
//	@Summary		更新权重模板
//	@Description	更新权重模板信息，仅创建者可以修改
//	@ID				update-weight-template
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"模板ID"
//	@Param			param	body		domain.UpdateWeightTemplateReq	true	"更新权重模板参数"
//	@Success		200		{object}	web.Resp{data=domain.WeightTemplateResp}
//	@Router			/api/v1/screening/weights/templates/{id} [put]
func (h *ScreeningHandler) UpdateWeightTemplate(c *web.Context, req domain.UpdateWeightTemplateReq) error {
	templateID, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "模板ID格式不正确")
	}

	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission
	}
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "当前用户ID格式不正确")
	}

	resp, err := h.usecase.UpdateWeightTemplate(c.Request().Context(), templateID, &req, userID)
	if err != nil {
		h.logger.Error("更新权重模板失败", slog.Any("err", err), slog.Any("template_id", templateID), slog.String("user_id", user.ID))
		return err
	}
	return c.Success(resp)
}

// DeleteWeightTemplate 删除权重模板
//
//	@Tags			Screening
//	@Summary		删除权重模板
//	@Description	删除权重模板（软删除），仅创建者可以删除
//	@ID				delete-weight-template
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"模板ID"
//	@Success		200	{object}	web.Resp
//	@Router			/api/v1/screening/weights/templates/{id} [delete]
func (h *ScreeningHandler) DeleteWeightTemplate(c *web.Context) error {
	templateID, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "模板ID格式不正确")
	}

	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission
	}
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return errcode.ErrInvalidParam.WithData("message", "当前用户ID格式不正确")
	}

	err = h.usecase.DeleteWeightTemplate(c.Request().Context(), templateID, userID)
	if err != nil {
		h.logger.Error("删除权重模板失败", slog.Any("err", err), slog.Any("template_id", templateID), slog.String("user_id", user.ID))
		return err
	}
	return c.Success(nil)
}
