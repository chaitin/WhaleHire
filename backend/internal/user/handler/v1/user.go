package v1

import (
	"context"
	"log/slog"
	"time"

	"github.com/GoYoko/web"
	"golang.org/x/time/rate"

	"github.com/ptonlix/whalehire/backend/config"
	"github.com/ptonlix/whalehire/backend/consts"
	"github.com/ptonlix/whalehire/backend/domain"
	"github.com/ptonlix/whalehire/backend/errcode"
	"github.com/ptonlix/whalehire/backend/internal/middleware"
	"github.com/ptonlix/whalehire/backend/pkg/session"
)

type UserHandler struct {
	usecase domain.UserUsecase
	session *session.Session
	logger  *slog.Logger
	cfg     *config.Config
	limiter *rate.Limiter
}

func NewUserHandler(
	w *web.Web,
	usecase domain.UserUsecase,
	auth *middleware.AuthMiddleware,
	active *middleware.ActiveMiddleware,
	readonly *middleware.ReadOnlyMiddleware,
	session *session.Session,
	logger *slog.Logger,
	cfg *config.Config,
) *UserHandler {
	u := &UserHandler{
		usecase: usecase,
		session: session,
		logger:  logger,
		cfg:     cfg,
		limiter: rate.NewLimiter(rate.Every(10*time.Second), 1),
	}

	// admin
	admin := w.Group("/api/v1/admin")
	admin.POST("/login", web.BindHandler(u.AdminLogin))
	admin.GET("/role", web.BaseHandler(u.ListRole))

	admin.Use(auth.Auth(), active.Active("admin"), readonly.Guard())
	admin.GET("/profile", web.BaseHandler(u.AdminProfile))
	admin.GET("/list", web.BaseHandler(u.AdminList, web.WithPage()))
	admin.GET("/login-history", web.BaseHandler(u.AdminLoginHistory, web.WithPage()))
	admin.POST("/create", web.BindHandler(u.CreateAdmin))
	admin.POST("/logout", web.BaseHandler(u.AdminLogout))
	admin.DELETE("/delete", web.BaseHandler(u.DeleteAdmin))
	admin.POST("/role", web.BindHandler(u.GrantRole))

	// user
	g := w.Group("/api/v1/user")
	g.POST("/register", web.BindHandler(u.Register))
	g.POST("/login", web.BindHandler(u.Login))

	g.Use(readonly.Guard())
	g.GET("/profile", web.BaseHandler(u.Profile), auth.UserAuth())
	g.PUT("/profile", web.BindHandler(u.UpdateProfile), auth.UserAuth())
	g.POST("/logout", web.BaseHandler(u.Logout), auth.UserAuth())

	g.Use(auth.Auth(), active.Active("admin"))

	g.PUT("/update", web.BindHandler(u.Update))
	g.DELETE("/delete", web.BaseHandler(u.Delete))
	g.GET("/list", web.BindHandler(u.List, web.WithPage()))
	g.GET("/login-history", web.BaseHandler(u.LoginHistory, web.WithPage()))

	return u
}

// Login 用户登录
//
//	@Tags			User
//	@Summary		用户登录
//	@Description	用户登录
//	@ID				login
//	@Accept			json
//	@Produce		json
//	@Param			param	body		domain.LoginReq	true	"登录参数"
//	@Success		200		{object}	web.Resp{data=domain.LoginResp}
//	@Router			/api/v1/user/login [post]
func (h *UserHandler) Login(c *web.Context, req domain.LoginReq) error {
	req.IP = c.RealIP()
	resp, err := h.usecase.Login(c.Request().Context(), &req)
	if err != nil {
		return err
	}
	if req.Source == consts.LoginSourceBrowser {
		if _, err := h.session.Save(c, consts.UserSessionName, c.Request().Host, resp.User); err != nil {
			return err
		}
	}
	return c.Success(resp)
}

// Logout 用户登出
//
//	@Tags			User
//	@Summary		用户登出
//	@Description	用户登出
//	@ID				logout
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	web.Resp{}
//	@Router			/api/v1/user/logout [post]
func (h *UserHandler) Logout(c *web.Context) error {
	if err := h.session.Del(c, consts.UserSessionName); err != nil {
		return err
	}
	return c.Success(nil)
}

// Update 更新用户
//
//	@Tags			User
//	@Summary		更新用户
//	@Description	更新用户
//	@ID				update-user
//	@Accept			json
//	@Produce		json
//	@Param			param	body		domain.UpdateUserReq	true	"更新用户参数"
//	@Success		200		{object}	web.Resp{data=domain.User}
//	@Router			/api/v1/user/update [put]
func (h *UserHandler) Update(c *web.Context, req domain.UpdateUserReq) error {
	resp, err := h.usecase.Update(c.Request().Context(), &req)
	if err != nil {
		return err
	}
	return c.Success(resp)
}

// Delete 删除用户
//
//	@Tags			User
//	@Summary		删除用户
//	@Description	删除用户
//	@ID				delete-user
//	@Accept			json
//	@Produce		json
//	@Param			id	query		string	true	"用户ID"
//	@Success		200	{object}	web.Resp{data=nil}
//	@Router			/api/v1/user/delete [delete]
func (h *UserHandler) Delete(c *web.Context) error {
	err := h.usecase.Delete(c.Request().Context(), c.QueryParam("id"))
	if err != nil {
		return err
	}
	return c.Success(nil)
}

// DeleteAdmin 删除管理员
//
//	@Tags			Admin
//	@Summary		删除管理员
//	@Description	删除管理员
//	@ID				delete-admin
//	@Accept			json
//	@Produce		json
//	@Param			id	query		string	true	"管理员ID"
//	@Success		200	{object}	web.Resp{data=nil}
//	@Router			/api/v1/admin/delete [delete]
func (h *UserHandler) DeleteAdmin(c *web.Context) error {
	err := h.usecase.DeleteAdmin(c.Request().Context(), c.QueryParam("id"))
	if err != nil {
		return err
	}
	return c.Success(nil)
}

// AdminLogin 管理员登录
//
//	@Tags			Admin
//	@Summary		管理员登录
//	@Description	管理员登录
//	@ID				admin-login
//	@Accept			json
//	@Produce		json
//	@Param			param	body		domain.LoginReq	true	"登录参数"
//	@Success		200		{object}	web.Resp{data=domain.AdminUser}
//	@Router			/api/v1/admin/login [post]
func (h *UserHandler) AdminLogin(c *web.Context, req domain.LoginReq) error {
	req.IP = c.RealIP()
	resp, err := h.usecase.AdminLogin(c.Request().Context(), &req)
	if err != nil {
		return err
	}

	h.logger.With("header", c.Request().Header).With("host", c.Request().Host).Info("admin login", "username", resp.Username)
	if _, err := h.session.Save(c, consts.SessionName, c.Request().Host, resp); err != nil {
		return err
	}
	return c.Success(resp)
}

// AdminLogout 管理员登出
//
//	@Tags			Admin
//	@Summary		管理员登出
//	@Description	管理员登出
//	@ID				admin-logout
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	web.Resp{}
//	@Router			/api/v1/admin/logout [post]
func (h *UserHandler) AdminLogout(c *web.Context) error {
	if err := h.session.Del(c, consts.SessionName); err != nil {
		return err
	}
	return c.Success(nil)
}

// AdminProfile 管理员信息
//
//	@Tags			Admin
//	@Summary		管理员信息
//	@Description	管理员信息
//	@ID				admin-profile
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	web.Resp{data=domain.AdminUser}
//	@Router			/api/v1/admin/profile [get]
func (h *UserHandler) AdminProfile(c *web.Context) error {
	user := middleware.GetAdmin(c)
	return c.Success(user)
}

// List 获取用户列表
//
//	@Tags			User
//	@Summary		获取用户列表
//	@Description	获取用户列表
//	@ID				list-user
//	@Accept			json
//	@Produce		json
//	@Param			page	query		web.Pagination	true	"分页"
//	@Success		200		{object}	web.Resp{data=domain.ListUserResp}
//	@Router			/api/v1/user/list [get]
func (h *UserHandler) List(c *web.Context, req domain.ListReq) error {
	resp, err := h.usecase.List(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.Success(resp)
}

// LoginHistory 获取用户登录历史
//
//	@Tags			User
//	@Summary		获取用户登录历史
//	@Description	获取用户登录历史
//	@ID				login-history
//	@Accept			json
//	@Produce		json
//	@Param			page	query		web.Pagination	true	"分页"
//	@Success		200		{object}	web.Resp{data=domain.ListLoginHistoryResp}
//	@Router			/api/v1/user/login-history [get]
func (h *UserHandler) LoginHistory(c *web.Context) error {
	resp, err := h.usecase.LoginHistory(c.Request().Context(), c.Page())
	if err != nil {
		return err
	}
	return c.Success(resp)
}

// Register 注册用户
//
//	@Tags			User
//	@Summary		注册用户
//	@Description	注册用户
//	@ID				register
//	@Accept			json
//	@Produce		json
//	@Param			param	body		domain.RegisterReq	true	"注册参数"
//	@Success		200		{object}	web.Resp{data=domain.User}
//	@Router			/api/v1/user/register [post]
func (h *UserHandler) Register(c *web.Context, req domain.RegisterReq) error {
	resp, err := h.usecase.Register(c.Request().Context(), &req)
	if err != nil {
		return err
	}

	return c.Success(resp)
}

// CreateAdmin 创建管理员
//
//	@Tags			Admin
//	@Summary		创建管理员
//	@Description	创建管理员
//	@ID				create-admin
//	@Accept			json
//	@Produce		json
//	@Param			param	body		domain.CreateAdminReq	true	"创建管理员参数"
//	@Success		200		{object}	web.Resp{data=domain.AdminUser}
//	@Router			/api/v1/admin/create [post]
func (h *UserHandler) CreateAdmin(c *web.Context, req domain.CreateAdminReq) error {
	user := middleware.GetAdmin(c)
	if user.Username != "admin" {
		return errcode.ErrPermission
	}
	resp, err := h.usecase.CreateAdmin(c.Request().Context(), &req)
	if err != nil {
		return err
	}
	return c.Success(resp)
}

// AdminList 获取管理员用户列表
//
//	@Tags			Admin
//	@Summary		获取管理员用户列表
//	@Description	获取管理员用户列表
//	@ID				list-admin-user
//	@Accept			json
//	@Produce		json
//	@Param			page	query		web.Pagination	true	"分页"
//	@Success		200		{object}	web.Resp{data=domain.ListAdminUserResp}
//	@Router			/api/v1/admin/list [get]
func (h *UserHandler) AdminList(c *web.Context) error {
	resp, err := h.usecase.AdminList(c.Request().Context(), c.Page())
	if err != nil {
		return err
	}
	return c.Success(resp)
}

// AdminLoginHistory 获取管理员登录历史
//
//	@Tags			Admin
//	@Summary		获取管理员登录历史
//	@Description	获取管理员登录历史
//	@ID				admin-login-history
//	@Accept			json
//	@Produce		json
//	@Param			page	query		web.Pagination	true	"分页"
//	@Success		200		{object}	web.Resp{data=domain.ListAdminLoginHistoryResp}
//	@Router			/api/v1/admin/login-history [get]
func (h *UserHandler) AdminLoginHistory(c *web.Context) error {
	resp, err := h.usecase.AdminLoginHistory(c.Request().Context(), c.Page())
	if err != nil {
		return err
	}
	return c.Success(resp)
}

// ListRole 获取系统角色列表
//
//	@Tags			Admin
//	@Summary		获取角色列表
//	@Description	获取角色列表
//	@ID				list-role
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	web.Resp{data=[]domain.Role}
//	@Router			/api/v1/admin/role [get]
func (h *UserHandler) ListRole(c *web.Context) error {
	roles, err := h.usecase.ListRole(c.Request().Context())
	if err != nil {
		return err
	}
	return c.Success(roles)
}

// GrantRole 授权角色
//
//	@Tags			Admin
//	@Summary		授权角色
//	@Description	授权角色
//	@ID				grant-role
//	@Accept			json
//	@Produce		json
//	@Param			param	body		domain.GrantRoleReq	true	"授权角色参数"
//	@Success		200		{object}	web.Resp
//	@Router			/api/v1/admin/role [post]
func (h *UserHandler) GrantRole(c *web.Context, req domain.GrantRoleReq) error {
	if err := h.usecase.GrantRole(c.Request().Context(), &req); err != nil {
		return err
	}
	return c.Success(nil)
}

// Profile 获取用户信息
//
//	@Tags			User Manage
//	@Summary		获取用户信息
//	@Description	获取用户信息
//	@ID				user-profile
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	web.Resp{data=domain.User}
//	@Failure		401	{object}	web.Resp{}
//	@Router			/api/v1/user/profile [get]
func (h *UserHandler) Profile(ctx *web.Context) error {
	return ctx.Success(middleware.GetUser(ctx))
}

// UpdateProfile 更新用户信息
//
//	@Tags			User Manage
//	@Summary		更新用户信息
//	@Description	更新用户信息
//	@ID				user-update-profile
//	@Accept			json
//	@Produce		json
//	@Param			req	body		domain.ProfileUpdateReq	true	"param"
//	@Success		200	{object}	web.Resp{data=domain.User}
//	@Failure		401	{object}	web.Resp{}
//	@Router			/api/v1/user/profile [put]
func (h *UserHandler) UpdateProfile(ctx *web.Context, req domain.ProfileUpdateReq) error {
	req.UID = middleware.GetUser(ctx).ID
	user, err := h.usecase.ProfileUpdate(ctx.Request().Context(), &req)
	if err != nil {
		return err
	}
	return ctx.Success(user)
}

func (h *UserHandler) InitAdmin() error {
	return h.usecase.InitAdmin(context.Background())
}
