package domain

import (
	"context"

	"github.com/GoYoko/web"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
)

type UserUsecase interface {
	Login(ctx context.Context, req *LoginReq) (*LoginResp, error)
	Update(ctx context.Context, req *UpdateUserReq) (*User, error)
	ProfileUpdate(ctx context.Context, req *ProfileUpdateReq) (*User, error)
	Delete(ctx context.Context, id string) error
	InitAdmin(ctx context.Context) error
	AdminLogin(ctx context.Context, req *LoginReq) (*AdminUser, error)
	DeleteAdmin(ctx context.Context, id string) error
	CreateAdmin(ctx context.Context, req *CreateAdminReq) (*AdminUser, error)
	List(ctx context.Context, req ListReq) (*ListUserResp, error)
	AdminList(ctx context.Context, page *web.Pagination) (*ListAdminUserResp, error)
	LoginHistory(ctx context.Context, page *web.Pagination) (*ListLoginHistoryResp, error)
	AdminLoginHistory(ctx context.Context, page *web.Pagination) (*ListAdminLoginHistoryResp, error)
	Register(ctx context.Context, req *RegisterReq) (*User, error)
	GetUserCount(ctx context.Context) (int64, error)
	ListRole(ctx context.Context) ([]*Role, error)
	GrantRole(ctx context.Context, req *GrantRoleReq) error
	GetPermissions(ctx context.Context, id uuid.UUID) (*Permissions, error)
}

type UserRepo interface {
	List(ctx context.Context, page *web.Pagination) ([]*db.User, *db.PageInfo, error)
	Update(ctx context.Context, id string, fn func(*db.Tx, *db.User, *db.UserUpdateOne) error) (*db.User, error)
	Delete(ctx context.Context, id string) error
	InitAdmin(ctx context.Context, username, password string) error
	CreateUser(ctx context.Context, user *db.User) (*db.User, error)
	CreateAdmin(ctx context.Context, admin *db.Admin, roleID int64) (*db.Admin, error)
	DeleteAdmin(ctx context.Context, id string) error
	AdminByName(ctx context.Context, username string) (*db.Admin, error)
	GetByName(ctx context.Context, username string) (*db.User, error)
	AdminList(ctx context.Context, page *web.Pagination) ([]*db.Admin, *db.PageInfo, error)
	UserLoginHistory(ctx context.Context, page *web.Pagination) ([]*db.UserLoginHistory, *db.PageInfo, error)
	AdminLoginHistory(ctx context.Context, page *web.Pagination) ([]*db.AdminLoginHistory, *db.PageInfo, error)
	SaveUserLoginHistory(ctx context.Context, userID, ip string) error
	SaveAdminLoginHistory(ctx context.Context, adminID, ip string) error
	GetUserCount(ctx context.Context) (int64, error)
	ListRole(ctx context.Context) ([]*db.Role, error)
	GrantRole(ctx context.Context, req *GrantRoleReq) error
	GetPermissions(ctx context.Context, id uuid.UUID) (*Permissions, error)
}

type ProfileUpdateReq struct {
	UID         string  `json:"-"`
	Username    *string `json:"username"`     // 用户名
	Password    *string `json:"password"`     // 密码
	OldPassword *string `json:"old_password"` // 旧密码
	Avatar      *string `json:"avatar"`       // 头像
}

type UpdateUserReq struct {
	ID       string             `json:"id" validate:"required"` // 用户ID
	Status   *consts.UserStatus `json:"status"`                 // 用户状态 active: 正常 locked: 锁定 inactive: 禁用
	Password *string            `json:"password"`               // 重置密码
}

type CreateAdminReq struct {
	Username string `json:"username" validate:"required"` // 用户名
	Password string `json:"password" validate:"required"` // 密码
	RoleID   int64  `json:"role_id" validate:"required"`  // 角色ID
}

type LoginReq struct {
	Source    consts.LoginSource `json:"source" validate:"required" default:"browser"` // 登录来源 browser: 浏览器;
	SessionID string             `json:"session_id"`                                   // 会话Id插件登录时必填
	Username  string             `json:"username"`                                     // 用户名
	Password  string             `json:"password"`                                     // 密码
	IP        string             `json:"-"`                                            // IP地址
}

type AdminLoginReq struct {
	Account  string `json:"account"`  // 用户名
	Password string `json:"password"` // 密码
}

type LoginResp struct {
	RedirectURL string `json:"redirect_url"`   // 重定向URL
	User        *User  `json:"user,omitempty"` // 用户信息
}

type ListReq struct {
	web.Pagination

	Search string `json:"search" query:"search"` // 搜索
}

type RegisterReq struct {
	Username string `json:"username" validate:"required"` // 用户名
	Email    string `json:"email" validate:"required"`    // 邮箱
	Password string `json:"password" validate:"required"` // 密码
}

type ListLoginHistoryResp struct {
	*db.PageInfo

	LoginHistories []*UserLoginHistory `json:"login_histories"`
}

type ListAdminLoginHistoryResp struct {
	*db.PageInfo

	LoginHistories []*AdminLoginHistory `json:"login_histories"`
}

type UserLoginHistory struct {
	User          *User  `json:"user"`           // 用户信息
	ClientVersion string `json:"client_version"` // 客户端版本
	Device        string `json:"device"`         // 设备信息
	Hostname      string `json:"hostname"`       // 主机名
	ClientID      string `json:"client_id"`      // 插件ID vscode
	CreatedAt     int64  `json:"created_at"`     // 登录时间
}

func (l *UserLoginHistory) From(e *db.UserLoginHistory) *UserLoginHistory {
	if e == nil {
		return l
	}

	l.ClientVersion = e.ClientVersion
	l.Device = e.OsType.Name()
	l.Hostname = e.Hostname
	l.ClientID = e.ClientID
	l.CreatedAt = e.CreatedAt.Unix()

	return l
}

type AdminLoginHistory struct {
	User          *AdminUser `json:"user"`           // 用户信息
	ClientVersion string     `json:"client_version"` // 客户端版本
	Device        string     `json:"device"`         // 设备信息
	CreatedAt     int64      `json:"created_at"`     // 登录时间
}

func (l *AdminLoginHistory) From(e *db.AdminLoginHistory) *AdminLoginHistory {
	if e == nil {
		return l
	}

	l.ClientVersion = e.ClientVersion
	l.Device = e.Device
	l.CreatedAt = e.CreatedAt.Unix()

	return l
}

type ListUserResp struct {
	*db.PageInfo

	Users []*User `json:"users"`
}

type ListAdminUserResp struct {
	*db.PageInfo

	Users []*AdminUser `json:"users"`
}

type User struct {
	ID           string            `json:"id"`             // 用户ID
	Username     string            `json:"username"`       // 用户名
	Email        string            `json:"email"`          // 邮箱
	Status       consts.UserStatus `json:"status"`         // 用户状态 active: 正常 locked: 锁定 inactive: 禁用
	AvatarURL    string            `json:"avatar_url"`     // 头像URL
	CreatedAt    int64             `json:"created_at"`     // 创建时间
	IsDeleted    bool              `json:"is_deleted"`     // 是否删除
	LastActiveAt int64             `json:"last_active_at"` // 最后活跃时间
}

func (u *User) From(e *db.User) *User {
	if e == nil {
		return u
	}

	u.ID = e.ID.String()
	u.Username = e.Username
	u.Email = e.Email
	u.Status = e.Status
	u.AvatarURL = e.AvatarURL
	u.IsDeleted = !e.DeletedAt.IsZero()
	u.CreatedAt = e.CreatedAt.Unix()

	return u
}

type AdminUser struct {
	ID           uuid.UUID          `json:"id"`             // 用户ID
	Username     string             `json:"username"`       // 用户名
	LastActiveAt int64              `json:"last_active_at"` // 最后活跃时间
	Status       consts.AdminStatus `json:"status"`         // 用户状态 active: 正常 inactive: 禁用
	Role         *Role              `json:"role"`           // 角色
	CreatedAt    int64              `json:"created_at"`     // 创建时间
}

func (a *AdminUser) IsAdmin() bool {
	return a.Username == "admin"
}

func (a *AdminUser) From(e *db.Admin) *AdminUser {
	if e == nil {
		return a
	}

	a.ID = e.ID
	a.Username = e.Username
	a.Status = e.Status
	if e.Username == "admin" {
		a.Role = &Role{
			ID:   1,
			Name: "超级管理员",
		}
	}
	a.CreatedAt = e.CreatedAt.Unix()

	return a
}

type Permissions struct {
	AdminID uuid.UUID
	IsAdmin bool
	Roles   []*Role
	UserIDs []uuid.UUID
}
