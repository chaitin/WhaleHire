package v1

import (
	"log/slog"

	"github.com/chaitin/WhaleHire/backend/pkg/web"

	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/errcode"
	"github.com/chaitin/WhaleHire/backend/internal/middleware"
)

type JobApplicationHandler struct {
	usecase       domain.JobApplicationUsecase
	resumeUsecase domain.ResumeUsecase
	logger        *slog.Logger
}

func NewJobApplicationHandler(
	w *web.Web,
	usecase domain.JobApplicationUsecase,
	resumeUsecase domain.ResumeUsecase,
	auth *middleware.AuthMiddleware,
	logger *slog.Logger,
) *JobApplicationHandler {
	h := &JobApplicationHandler{
		usecase:       usecase,
		resumeUsecase: resumeUsecase,
		logger:        logger,
	}

	g := w.Group("/api/v1/job-applications")
	g.Use(auth.UserAuth())
	g.POST("", web.BindHandler(h.Create))
	g.PUT("", web.BindHandler(h.Update))
	g.GET("/resume/:resume_id", web.BaseHandler(h.GetByResumeID))
	g.GET("/job-position/:job_position_id", web.BindHandler(h.GetByJobPositionID))

	return h
}

// Create 创建简历与岗位的关联关系
//
//	@Tags			JobApplication
//	@Summary		创建简历与岗位的关联关系
//	@Description	为指定简历创建与多个岗位的关联关系
//	@ID				create-job-applications
//	@Accept			json
//	@Produce		json
//	@Param			param	body		domain.CreateJobApplicationsReq	true	"创建关联关系请求"
//	@Success		200		{object}	web.Resp{data=[]domain.JobApplication}
//	@Router			/api/v1/job-applications [post]
func (h *JobApplicationHandler) Create(c *web.Context, req domain.CreateJobApplicationsReq) error {
	if req.ResumeID == "" {
		return errcode.ErrInvalidParam
	}

	if len(req.JobPositionIDs) == 0 {
		return errcode.ErrInvalidParam
	}

	applications, err := h.usecase.CreateJobApplications(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("failed to create job applications", "error", err, "resume_id", req.ResumeID)
		return err
	}

	return c.Success(applications)
}

// Update 更新简历与岗位的关联关系
//
//	@Tags			JobApplication
//	@Summary		更新简历与岗位的关联关系
//	@Description	更新指定简历与多个岗位的关联关系
//	@ID				update-job-applications
//	@Accept			json
//	@Produce		json
//	@Param			param	body		domain.UpdateJobApplicationsReq	true	"更新关联关系请求"
//	@Success		200		{object}	web.Resp{data=[]domain.JobApplication}
//	@Router			/api/v1/job-applications [put]
func (h *JobApplicationHandler) Update(c *web.Context, req domain.UpdateJobApplicationsReq) error {
	if req.ResumeID == "" {
		return errcode.ErrInvalidParam
	}

	if len(req.JobPositionIDs) == 0 {
		return errcode.ErrInvalidParam
	}

	applications, err := h.usecase.UpdateJobApplications(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("failed to update job applications", "error", err, "resume_id", req.ResumeID)
		return err
	}

	return c.Success(applications)
}

// GetByResumeID 根据简历ID获取关联的岗位信息
//
//	@Tags			JobApplication
//	@Summary		根据简历ID获取关联的岗位信息
//	@Description	获取指定简历关联的所有岗位申请记录
//	@ID				get-job-applications-by-resume
//	@Accept			json
//	@Produce		json
//	@Param			resume_id	path		string	true	"简历ID"
//	@Success		200			{object}	web.Resp{data=[]domain.JobApplication}
//	@Router			/api/v1/job-applications/resume/{resume_id} [get]
func (h *JobApplicationHandler) GetByResumeID(c *web.Context) error {
	resumeID := c.Param("resume_id")
	if resumeID == "" {
		return errcode.ErrInvalidParam
	}

	applications, err := h.usecase.GetJobPositionsByResumeID(c.Request().Context(), resumeID)
	if err != nil {
		h.logger.Error("failed to get job applications by resume ID", "error", err, "resume_id", resumeID)
		return err
	}

	return c.Success(applications)
}

// GetByJobPositionID 根据岗位ID获取关联的简历信息
//
//	@Tags			JobApplication
//	@Summary		根据岗位ID获取关联的简历信息
//	@Description	获取指定岗位关联的所有简历申请记录，支持分页和筛选
//	@ID				get-job-applications-by-job-position
//	@Accept			json
//	@Produce		json
//	@Param			job_position_id	path		string	true	"岗位ID"
//	@Param			page			query		int		false	"页码，默认为1"
//	@Param			size			query		int		false	"每页数量，默认为10"
//	@Param			status			query		string	false	"申请状态筛选，可选值：applied, reviewing, interviewed, rejected, accepted"
//	@Param			source			query		string	false	"申请来源筛选，可选值：direct, referral, headhunter, email"
//	@Success		200				{object}	web.Resp{data=domain.GetResumesByJobPositionIDResp}
//	@Router			/api/v1/job-applications/job-position/{job_position_id} [get]
func (h *JobApplicationHandler) GetByJobPositionID(c *web.Context, req domain.GetResumesByJobPositionIDReq) error {
	jobPositionID := c.Param("job_position_id")
	if jobPositionID == "" {
		return errcode.ErrInvalidParam
	}

	// 设置岗位ID
	req.JobPositionID = jobPositionID

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}

	// 获取岗位关联的简历申请记录
	h.logger.Info("starting to get resumes by job position ID",
		"job_position_id", jobPositionID,
		"page", req.Page,
		"size", req.Size,
		"status", req.Status,
		"source", req.Source)

	applications, total, err := h.usecase.GetResumesByJobPositionID(c.Request().Context(), jobPositionID, &req)
	if err != nil {
		h.logger.Error("failed to get resumes by job position ID", "error", err, "job_position_id", jobPositionID)
		return err
	}

	// 提取简历ID列表并获取简历详细信息
	resumes := make([]*domain.Resume, 0, len(applications))
	for _, app := range applications {
		if app.ResumeID != "" {
			resumeDetail, err := h.resumeUsecase.GetByID(c.Request().Context(), app.ResumeID)
			if err != nil {
				h.logger.Error("failed to get resume detail", "error", err, "resume_id", app.ResumeID)
				continue // 跳过获取失败的简历
			}
			resumes = append(resumes, resumeDetail.Resume)
		}
	}

	// 构建响应
	resp := &domain.GetResumesByJobPositionIDResp{
		PageInfo: &db.PageInfo{
			HasNextPage: int64((req.Page-1)*req.Size)+int64(len(resumes)) < total,
			TotalCount:  total,
		},
		Resumes: resumes,
	}

	return c.Success(resp)
}
