package v1

import (
	"log/slog"

	"github.com/chaitin/WhaleHire/backend/pkg/web"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/errcode"
	"github.com/chaitin/WhaleHire/backend/internal/middleware"
)

type DepartmentHandler struct {
	usecase domain.DepartmentUsecase
	logger  *slog.Logger
}

func NewDepartmentHandler(
	w *web.Web,
	usecase domain.DepartmentUsecase,
	auth *middleware.AuthMiddleware,
	logger *slog.Logger,
) *DepartmentHandler {
	h := &DepartmentHandler{
		usecase: usecase,
		logger:  logger,
	}

	g := w.Group("/api/v1/departments")
	g.Use(auth.UserAuth())
	g.POST("", web.BindHandler(h.Create))
	g.GET("", web.BindHandler(h.List, web.WithPage()))
	g.GET("/:id", web.BaseHandler(h.Get))
	g.PUT("/:id", web.BindHandler(h.Update))
	g.DELETE("/:id", web.BaseHandler(h.Delete))

	return h
}

// Create 创建部门
//
//	@Summary		创建部门
//	@Description	创建一个新的部门
//	@Tags			部门管理
//	@Accept			json
//	@Produce		json
//	@Param			request	body		domain.CreateDepartmentReq	true	"部门信息"
//	@Success		200		{object}	web.Resp{data=domain.Department}
//	@Router			/api/v1/departments [post]
func (h *DepartmentHandler) Create(c *web.Context, req domain.CreateDepartmentReq) error {
	dept, err := h.usecase.Create(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("failed to create department", "error", err)
		return err
	}
	return c.Success(dept)
}

// Update 更新部门
//
//	@Summary		更新部门
//	@Description	根据 ID 更新部门信息
//	@Tags			部门管理
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"部门 ID"
//	@Param			request	body		domain.UpdateDepartmentReq	true	"部门信息"
//	@Success		200		{object}	web.Resp{data=domain.Department}
//	@Router			/api/v1/departments/{id} [put]
func (h *DepartmentHandler) Update(c *web.Context, req domain.UpdateDepartmentReq) error {
	req.ID = c.Param("id")
	if req.ID == "" {
		return errcode.ErrDepartmentRequired
	}

	dept, err := h.usecase.Update(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("failed to update department", "error", err, "department_id", req.ID)
		return err
	}

	return c.Success(dept)
}

// Get 获取部门详情
//
//	@Summary		获取部门详情
//	@Description	根据 ID 获取部门详细信息
//	@Tags			部门管理
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"部门 ID"
//	@Success		200	{object}	web.Resp{data=domain.Department}
//	@Router			/api/v1/departments/{id} [get]
func (h *DepartmentHandler) Get(c *web.Context) error {
	id := c.Param("id")
	if id == "" {
		return errcode.ErrDepartmentRequired
	}

	dept, err := h.usecase.GetByID(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("failed to get department", "error", err, "department_id", id)
		return err
	}

	return c.Success(dept)
}

// Delete 删除部门
//
//	@Summary		删除部门
//	@Description	根据 ID 删除部门
//	@Tags			部门管理
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"部门 ID"
//	@Success		200	{object}	web.Resp{}
//	@Router			/api/v1/departments/{id} [delete]
func (h *DepartmentHandler) Delete(c *web.Context) error {
	id := c.Param("id")
	if id == "" {
		return errcode.ErrDepartmentRequired
	}

	if err := h.usecase.Delete(c.Request().Context(), id); err != nil {
		h.logger.Error("failed to delete department", "error", err, "department_id", id)
		return err
	}

	return c.Success(nil)
}

// List 获取部门列表
//
//	@Summary		获取部门列表
//	@Description	分页获取部门列表，支持筛选条件
//	@Tags			部门管理
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"页码"
//	@Param			size		query		int		false	"每页数量"
//	@Param			keyword		query		string	false	"部门名称关键词"
//	@Param			parent_id	query		string	false	"上级部门 ID"
//	@Success		200			{object}	web.Resp{data=domain.ListDepartmentResp}
//	@Router			/api/v1/departments [get]
func (h *DepartmentHandler) List(c *web.Context, req domain.ListDepartmentReq) error {
	req.Page = c.Page().Page
	req.Size = c.Page().Size

	resp, err := h.usecase.List(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("failed to list departments", "error", err)
		return err
	}

	return c.Success(resp)
}
