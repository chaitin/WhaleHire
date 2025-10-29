package v1

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/errcode"
	"github.com/chaitin/WhaleHire/backend/internal/middleware"
	"github.com/chaitin/WhaleHire/backend/pkg/web"
)

type UniversityHandler struct {
	usecase domain.UniversityUsecase
	logger  *slog.Logger
}

func NewUniversityHandler(
	w *web.Web,
	usecase domain.UniversityUsecase,
	auth *middleware.AuthMiddleware,
	logger *slog.Logger,
) *UniversityHandler {
	h := &UniversityHandler{
		usecase: usecase,
		logger:  logger,
	}

	g := w.Group("/api/v1/universities")
	g.Use(auth.UserAuth())
	g.POST("", web.BindHandler(h.Create))
	g.GET("", web.BindHandler(h.List, web.WithPage()))
	g.GET("/:id", web.BaseHandler(h.Get))
	g.PUT("/:id", web.BindHandler(h.Update))
	g.DELETE("/:id", web.BaseHandler(h.Delete))
	g.GET("/search", web.BaseHandler(h.Search))
	g.POST("/batch-match", web.BindHandler(h.BatchMatch))
	g.POST("/import-csv", web.BaseHandler(h.ImportFromCSV))

	return h
}

// Create 创建高校
//
//	@Summary		创建高校
//	@Description	创建一个新的高校
//	@Tags			University
//	@Accept			json
//	@Produce		json
//	@Param			request	body		domain.CreateUniversityReq	true	"高校信息"
//	@Success		200		{object}	web.Resp{data=domain.University}
//	@Router			/api/v1/universities [post]
func (h *UniversityHandler) Create(c *web.Context, req domain.CreateUniversityReq) error {
	university, err := h.usecase.Create(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("failed to create university", "error", err)
		return err
	}

	return c.Success(university)
}

// List 获取高校列表
//
//	@Summary		获取高校列表
//	@Description	分页获取高校列表，支持筛选条件
//	@Tags			University
//	@Accept			json
//	@Produce		json
//	@Param			page					query		int		false	"页码"
//	@Param			size					query		int		false	"每页数量"
//	@Param			keyword					query		string	false	"高校名称关键词"
//	@Param			country					query		string	false	"国家"
//	@Param			is_double_first_class	query		bool	false	"是否双一流高校"
//	@Param			is_project_985			query		bool	false	"是否985高校"
//	@Param			is_project_211			query		bool	false	"是否211高校"
//	@Param			is_qs_top100			query		bool	false	"是否QS Top 100高校"
//	@Success		200						{object}	web.Resp{data=domain.ListUniversityResp}
//	@Router			/api/v1/universities [get]
func (h *UniversityHandler) List(c *web.Context, req domain.ListUniversityReq) error {
	req.Page = c.Page().Page
	req.Size = c.Page().Size

	resp, err := h.usecase.List(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("failed to list universities", "error", err)
		return err
	}

	return c.Success(resp)
}

// Get 获取高校详情
//
//	@Summary		获取高校详情
//	@Description	根据 ID 获取高校详细信息
//	@Tags			University
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"高校 ID"
//	@Success		200	{object}	web.Resp{data=domain.University}
//	@Router			/api/v1/universities/{id} [get]
func (h *UniversityHandler) Get(c *web.Context) error {
	id := c.Param("id")
	if id == "" {
		return errcode.ErrInvalidParam
	}

	university, err := h.usecase.GetByID(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("failed to get university", "error", err, "university_id", id)
		return err
	}

	return c.Success(university)
}

// Update 更新高校
//
//	@Summary		更新高校
//	@Description	根据 ID 更新高校信息
//	@Tags			University
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"高校 ID"
//	@Param			request	body		domain.UpdateUniversityReq	true	"高校信息"
//	@Success		200		{object}	web.Resp{data=domain.University}
//	@Router			/api/v1/universities/{id} [put]
func (h *UniversityHandler) Update(c *web.Context, req domain.UpdateUniversityReq) error {
	req.ID = c.Param("id")
	if req.ID == "" {
		return errcode.ErrInvalidParam
	}

	university, err := h.usecase.Update(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("failed to update university", "error", err, "university_id", req.ID)
		return err
	}

	return c.Success(university)
}

// Delete 删除高校
//
//	@Summary		删除高校
//	@Description	根据 ID 删除高校
//	@Tags			University
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"高校 ID"
//	@Success		200	{object}	web.Resp{}
//	@Router			/api/v1/universities/{id} [delete]
func (h *UniversityHandler) Delete(c *web.Context) error {
	id := c.Param("id")
	if id == "" {
		return errcode.ErrInvalidParam
	}

	if err := h.usecase.Delete(c.Request().Context(), id); err != nil {
		h.logger.Error("failed to delete university", "error", err, "university_id", id)
		return err
	}

	return c.Success(nil)
}

// Search 搜索高校
//
//	@Summary		搜索高校
//	@Description	根据名称搜索高校
//	@Tags			University
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	true	"搜索关键词"
//	@Param			limit	query		int		false	"返回数量限制"
//	@Success		200		{object}	web.Resp{data=[]domain.University}
//	@Router			/api/v1/universities/search [get]
func (h *UniversityHandler) Search(c *web.Context) error {
	name := c.QueryParam("name")
	if name == "" {
		return errcode.ErrInvalidParam.WithData("name", "不能为空")
	}

	limitStr := c.QueryParam("limit")
	limit := 10 // 默认限制
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if limit > 100 {
		limit = 100 // 最大限制
	}

	universities, err := h.usecase.SearchByName(c.Request().Context(), name, limit)
	if err != nil {
		h.logger.Error("failed to search universities", "error", err, "name", name)
		return err
	}

	return c.Success(universities)
}

// BatchMatchRequest 批量匹配请求
type BatchMatchRequest struct {
	Names []string `json:"names" validate:"required,min=1"`
}

// BatchMatch 批量匹配高校
//
//	@Summary		批量匹配高校
//	@Description	批量匹配高校名称
//	@Tags			高校管理
//	@Accept			json
//	@Produce		json
//	@Param			request	body		BatchMatchRequest	true	"匹配请求"
//	@Success		200		{object}	web.Resp{data=[]domain.UniversityMatch}
//	@Router			/api/v1/universities/batch-match [post]
func (h *UniversityHandler) BatchMatch(c *web.Context, req BatchMatchRequest) error {
	if len(req.Names) == 0 {
		return errcode.ErrInvalidParam
	}

	if len(req.Names) > 100 {
		return errcode.ErrInvalidParam
	}

	matches, err := h.usecase.BatchMatch(c.Request().Context(), req.Names)
	if err != nil {
		h.logger.Error("failed to batch match universities", "error", err)
		return err
	}

	return c.Success(matches)
}

// ImportFromCSV 从CSV文件导入高校数据
//
//	@Summary		从CSV文件导入高校数据
//	@Description	从上传的CSV文件导入高校数据
//	@Tags			University
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"CSV文件"
//	@Success		200		{object}	web.Resp{data=domain.ImportResult}
//	@Router			/api/v1/universities/import-csv [post]
func (h *UniversityHandler) ImportFromCSV(c *web.Context) error {
	// 获取上传的文件
	file, header, err := c.Request().FormFile("file")
	if err != nil {
		h.logger.Error("failed to get form file", "error", err)
		return errcode.ErrInvalidParam.WithData("file", "无法获取上传文件")
	}
	defer file.Close()

	// 验证文件类型
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		return errcode.ErrInvalidParam.WithData("file", "文件必须是CSV格式")
	}

	result, err := h.usecase.ImportFromCSV(c.Request().Context(), file)
	if err != nil {
		h.logger.Error("failed to import from CSV", "error", err, "filename", header.Filename)
		return errcode.ErrCSVParseError
	}

	h.logger.Info("CSV import completed",
		"filename", header.Filename,
		"total", result.Total,
		"success", result.Success,
		"failed", result.Failed)

	return c.Success(result)
}
