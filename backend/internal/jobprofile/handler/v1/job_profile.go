package v1

import (
	"log/slog"

	"github.com/chaitin/WhaleHire/backend/pkg/web"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/errcode"
	"github.com/chaitin/WhaleHire/backend/internal/middleware"
)

type JobProfileHandler struct {
	usecase domain.JobProfileUsecase
	logger  *slog.Logger
}

func NewJobProfileHandler(
	w *web.Web,
	usecase domain.JobProfileUsecase,
	auth *middleware.AuthMiddleware,
	logger *slog.Logger,
) *JobProfileHandler {
	h := &JobProfileHandler{
		usecase: usecase,
		logger:  logger,
	}

	g := w.Group("/api/v1/job-profiles")
	g.Use(auth.UserAuth())
	g.POST("", web.BindHandler(h.Create))
	g.GET("", web.BindHandler(h.List, web.WithPage()))
	g.GET("/search", web.BindHandler(h.Search, web.WithPage()))
	g.GET("/:id", web.BaseHandler(h.Get))
	g.PUT("/:id", web.BindHandler(h.Update))
	g.DELETE("/:id", web.BaseHandler(h.Delete))

	skillGroup := w.Group("/api/v1/job-skills/meta")
	skillGroup.Use(auth.UserAuth())
	skillGroup.POST("", web.BindHandler(h.CreateSkillMeta))
	skillGroup.GET("", web.BindHandler(h.ListSkillMeta, web.WithPage()))
	skillGroup.DELETE("/:id", web.BaseHandler(h.DeleteSkillMeta))

	return h
}

// Create 创建岗位画像
//
//	@Tags			JobProfile
//	@Summary		创建岗位画像
//	@Description	创建岗位画像并返回详情。工作性质可选值：full_time/part_time/internship/outsourcing；职位状态可选值：draft/published
//	@ID				create-job-profile
//	@Accept			json
//	@Produce		json
//	@Param			param	body		domain.CreateJobProfileReq	true	"岗位画像信息"
//	@Success		200		{object}	web.Resp{data=domain.JobProfileDetail}
//	@Router			/api/v1/job-profiles [post]
func (h *JobProfileHandler) Create(c *web.Context, req domain.CreateJobProfileReq) error {
	// 校验请求参数中的常量值
	if err := req.Validate(); err != nil {
		h.logger.Error("invalid request parameters", "error", err)
		return errcode.ErrInvalidParam.WithData("message", err.Error())
	}

	// 获取当前用户
	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission.WithData("message", "user not found")
	}

	// 设置创建者ID
	req.CreatedBy = &user.ID

	profile, err := h.usecase.Create(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("failed to create job profile", "error", err)
		return err
	}
	return c.Success(profile)
}

// Update 更新岗位画像
//
//	@Tags			JobProfile
//	@Summary		更新岗位画像
//	@Description	根据ID更新岗位画像及相关信息。工作性质可选值：full_time/part_time/internship/outsourcing；职位状态可选值：draft/published
//	@ID				update-job-profile
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"岗位画像ID"
//	@Param			param	body		domain.UpdateJobProfileReq	true	"岗位画像更新信息"
//	@Success		200		{object}	web.Resp{data=domain.JobProfileDetail}
//	@Router			/api/v1/job-profiles/{id} [put]
func (h *JobProfileHandler) Update(c *web.Context, req domain.UpdateJobProfileReq) error {
	req.ID = c.Param("id")
	if req.ID == "" {
		return errcode.ErrJobProfileRequired
	}

	// 校验请求参数中的常量值
	if err := req.Validate(); err != nil {
		h.logger.Error("invalid request parameters", "error", err, "job_id", req.ID)
		return errcode.ErrInvalidParam.WithData("message", err.Error())
	}

	profile, err := h.usecase.Update(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("failed to update job profile", "error", err, "job_id", req.ID)
		return err
	}

	return c.Success(profile)
}

// Delete 删除岗位画像
//
//	@Tags			JobProfile
//	@Summary		删除岗位画像
//	@Description	根据ID删除岗位画像
//	@ID				delete-job-profile
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"岗位画像ID"
//	@Success		200	{object}	web.Resp{}
//	@Router			/api/v1/job-profiles/{id} [delete]
func (h *JobProfileHandler) Delete(c *web.Context) error {
	id := c.Param("id")
	if id == "" {
		return errcode.ErrJobProfileRequired
	}

	if err := h.usecase.Delete(c.Request().Context(), id); err != nil {
		h.logger.Error("failed to delete job profile", "error", err, "job_id", id)
		return err
	}

	return c.Success(nil)
}

// Search 搜索岗位画像
//
//	@Tags			JobProfile
//	@Summary		搜索岗位画像
//	@Description	根据关键词和条件搜索岗位画像，支持模糊搜索
//	@ID				search-job-profiles
//	@Accept			json
//	@Produce		json
//	@Param			page			query		web.Pagination	true	"分页参数"
//	@Param			keyword			query		string			false	"关键词"
//	@Param			department_id	query		string			false	"部门ID"
//	@Param			skill_ids		query		[]string		false	"技能ID列表"
//	@Param			locations		query		[]string		false	"工作地点列表"
//	@Param			salary_min		query		number			false	"最低薪资"
//	@Param			salary_max		query		number			false	"最高薪资"
//	@Success		200				{object}	web.Resp{data=domain.SearchJobProfileResp}
//	@Router			/api/v1/job-profiles/search [get]
func (h *JobProfileHandler) Search(c *web.Context, req domain.SearchJobProfileReq) error {
	req.Page = c.Page().Page
	req.Size = c.Page().Size

	profiles, err := h.usecase.Search(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("failed to search job profiles", "error", err)
		return err
	}

	return c.Success(profiles)
}

// Get 获取岗位画像详情
//
//	@Tags			JobProfile
//	@Summary		获取岗位画像详情
//	@Description	根据ID获取岗位画像详情
//	@ID				get-job-profile
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"岗位画像ID"
//	@Success		200	{object}	web.Resp{data=domain.JobProfileDetail}
//	@Router			/api/v1/job-profiles/{id} [get]
func (h *JobProfileHandler) Get(c *web.Context) error {
	id := c.Param("id")
	if id == "" {
		return errcode.ErrJobProfileRequired
	}

	profile, err := h.usecase.GetByID(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("failed to get job profile", "error", err, "job_id", id)
		return err
	}

	return c.Success(profile)
}

// List 获取岗位画像列表
//
//	@Tags			JobProfile
//	@Summary		获取岗位画像列表
//	@Description	获取岗位画像列表，支持分页和多条件筛选
//	@ID				list-job-profiles
//	@Accept			json
//	@Produce		json
//	@Param			page			query		web.Pagination	true	"分页参数"
//	@Param			keyword			query		string			false	"关键词"
//	@Param			department_id	query		string			false	"部门ID"
//	@Param			skill_ids		query		[]string		false	"技能ID列表"
//	@Param			locations		query		[]string		false	"工作地点列表"
//	@Success		200				{object}	web.Resp{data=domain.ListJobProfileResp}
//	@Router			/api/v1/job-profiles [get]
func (h *JobProfileHandler) List(c *web.Context, req domain.ListJobProfileReq) error {
	req.Page = c.Page().Page
	req.Size = c.Page().Size

	profiles, err := h.usecase.List(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("failed to list job profiles", "error", err)
		return err
	}

	return c.Success(profiles)
}

// CreateSkillMeta 创建岗位技能字典
//
//	@Tags			JobSkill
//	@Summary		创建岗位技能字典
//	@Description	创建岗位技能元信息
//	@ID				create-job-skill-meta
//	@Accept			json
//	@Produce		json
//	@Param			param	body		domain.CreateSkillMetaReq	true	"技能元信息"
//	@Success		200		{object}	web.Resp{data=domain.JobSkillMeta}
//	@Router			/api/v1/job-skills/meta [post]
func (h *JobProfileHandler) CreateSkillMeta(c *web.Context, req domain.CreateSkillMetaReq) error {
	skill, err := h.usecase.CreateSkillMeta(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("failed to create skill meta", "error", err)
		return err
	}

	return c.Success(skill)
}

// ListSkillMeta 获取岗位技能字典列表
//
//	@Tags			JobSkill
//	@Summary		获取岗位技能字典列表
//	@Description	获取岗位技能元信息列表，支持分页
//	@ID				list-job-skill-meta
//	@Accept			json
//	@Produce		json
//	@Param			page	query		web.Pagination	true	"分页参数"
//	@Param			keyword	query		string			false	"关键词"
//	@Success		200		{object}	web.Resp{data=domain.ListSkillMetaResp}
//	@Router			/api/v1/job-skills/meta [get]
func (h *JobProfileHandler) ListSkillMeta(c *web.Context, req domain.ListSkillMetaReq) error {
	req.Page = c.Page().Page
	req.Size = c.Page().Size

	skills, err := h.usecase.ListSkillMeta(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("failed to list skill meta", "error", err)
		return err
	}

	return c.Success(skills)
}

// DeleteSkillMeta 删除岗位技能字典
//
//	@Tags			JobSkill
//	@Summary		删除岗位技能字典
//	@Description	软删除岗位技能元信息
//	@ID				delete-job-skill-meta
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"技能元信息ID"
//	@Success		200	{object}	web.Resp
//	@Router			/api/v1/job-skills/meta/{id} [delete]
func (h *JobProfileHandler) DeleteSkillMeta(c *web.Context) error {
	id := c.Param("id")
	if id == "" {
		return errcode.ErrJobSkillMetaRequired
	}

	err := h.usecase.DeleteSkillMeta(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("failed to delete skill meta", "error", err, "id", id)
		return err
	}

	return c.Success(nil)
}
