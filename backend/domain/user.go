package domain

import (
	"context"

	"github.com/GoYoko/web"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/ent/types"
	"github.com/chaitin/WhaleHire/backend/pkg/cvt"
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
	GetSetting(ctx context.Context) (*Setting, error)
	UpdateSetting(ctx context.Context, req *UpdateSettingReq) (*Setting, error)
	GetPermissions(ctx context.Context, id uuid.UUID) (*Permissions, error)
	OAuthSignUpOrIn(ctx context.Context, req *OAuthSignUpOrInReq) (*OAuthURLResp, error)
	OAuthCallback(ctx *web.Context, req *OAuthCallbackReq) error
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
	GetSetting(ctx context.Context) (*db.Setting, error)
	UpdateSetting(ctx context.Context, fn func(*db.Setting, *db.SettingUpdateOne)) (*db.Setting, error)
	OAuthRegister(ctx context.Context, platform consts.UserPlatform, inviteCode string, req *OAuthUserInfo) (*db.User, error)
	OAuthLogin(ctx context.Context, platform consts.UserPlatform, req *OAuthUserInfo) (*db.User, error)
	SignUpOrIn(ctx context.Context, platform consts.UserPlatform, req *OAuthUserInfo) (*db.User, error)
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

type UpdateSettingReq struct {
	EnableSSO            *bool             `json:"enable_sso"`             // 是否开启SSO
	ForceTwoFactorAuth   *bool             `json:"force_two_factor_auth"`  // 是否强制两步验证
	DisablePasswordLogin *bool             `json:"disable_password_login"` // 是否禁用密码登录
	EnableAutoLogin      *bool             `json:"enable_auto_login"`      // 是否开启自动登录
	DingtalkOAuth        *DingtalkOAuthReq `json:"dingtalk_oauth"`         // 钉钉OAuth配置
	CustomOAuth          *CustomOAuthReq   `json:"custom_oauth"`           // 自定义OAuth配置
	BaseURL              *string           `json:"base_url"`               // base url 配置，为了支持前置代理
}

type DingtalkOAuthReq struct {
	Enable       *bool   `json:"enable"`        // 钉钉OAuth开关
	ClientID     *string `json:"client_id"`     // 钉钉客户端ID
	ClientSecret *string `json:"client_secret"` // 钉钉客户端密钥
}

type DingtalkOAuth struct {
	Enable       bool   `json:"enable"`        // 钉钉OAuth开关
	ClientID     string `json:"client_id"`     // 钉钉客户端ID
	ClientSecret string `json:"client_secret"` // 钉钉客户端密钥
}

func (d *DingtalkOAuth) From(e *types.DingtalkOAuth) *DingtalkOAuth {
	if e == nil {
		d.Enable = false
		return d
	}

	d.Enable = e.Enable
	d.ClientID = e.ClientID
	d.ClientSecret = e.ClientSecret
	return d
}

type CustomOAuthReq struct {
	Enable         *bool    `json:"enable"`           // 自定义OAuth开关
	ClientID       *string  `json:"client_id"`        // 自定义客户端ID
	ClientSecret   *string  `json:"client_secret"`    // 自定义客户端密钥
	AuthorizeURL   *string  `json:"authorize_url"`    // 自定义OAuth授权URL
	AccessTokenURL *string  `json:"access_token_url"` // 自定义OAuth访问令牌URL
	UserInfoURL    *string  `json:"userinfo_url"`     // 自定义OAuth用户信息URL
	Scopes         []string `json:"scopes"`           // 自定义OAuth Scope列表
	IDField        *string  `json:"id_field"`         // 用户信息回包中的ID字段名
	NameField      *string  `json:"name_field"`       // 用户信息回包中的用户名字段名
	AvatarField    *string  `json:"avatar_field"`     // 用户信息回包中的头像URL字段名
	EmailField     *string  `json:"email_field"`      // 用户信息回包中的邮箱字段名
}

type CustomOAuth struct {
	Enable         bool     `json:"enable"`           // 自定义OAuth开关
	ClientID       string   `json:"client_id"`        // 自定义客户端ID
	ClientSecret   string   `json:"client_secret"`    // 自定义客户端密钥
	AuthorizeURL   string   `json:"authorize_url"`    // 自定义OAuth授权URL
	AccessTokenURL string   `json:"access_token_url"` // 自定义OAuth访问令牌URL
	UserInfoURL    string   `json:"userinfo_url"`     // 自定义OAuth用户信息URL
	Scopes         []string `json:"scopes"`           // 自定义OAuth Scope列表
	IDField        string   `json:"id_field"`         // 用户信息回包中的ID字段名
	NameField      string   `json:"name_field"`       // 用户信息回包中的用户名字段名
	AvatarField    string   `json:"avatar_field"`     // 用户信息回包中的头像URL字段名
	EmailField     string   `json:"email_field"`      // 用户信息回包中的邮箱字段名
}

func (c *CustomOAuth) From(e *types.CustomOAuth) *CustomOAuth {
	if e == nil {
		c.Enable = false
		return c
	}

	c.Enable = e.Enable
	c.ClientID = e.ClientID
	c.AuthorizeURL = e.AuthorizeURL
	c.AccessTokenURL = e.AccessTokenURL
	c.UserInfoURL = e.UserInfoURL
	c.Scopes = e.Scopes
	c.IDField = e.IDField
	c.NameField = e.NameField
	c.AvatarField = e.AvatarField
	c.EmailField = e.EmailField
	return c
}

type Setting struct {
	EnableSSO            bool          `json:"enable_sso"`             // 是否开启SSO
	ForceTwoFactorAuth   bool          `json:"force_two_factor_auth"`  // 是否强制两步验证
	DisablePasswordLogin bool          `json:"disable_password_login"` // 是否禁用密码登录
	EnableAutoLogin      bool          `json:"enable_auto_login"`      // 是否开启自动登录
	DingtalkOAuth        DingtalkOAuth `json:"dingtalk_oauth"`         // 钉钉OAuth接入
	CustomOAuth          CustomOAuth   `json:"custom_oauth"`           // 自定义OAuth接入
	BaseURL              string        `json:"base_url,omitempty"`     // base url 配置，为了支持前置代理
	CreatedAt            int64         `json:"created_at"`             // 创建时间
	UpdatedAt            int64         `json:"updated_at"`             // 更新时间
}

func (s *Setting) From(e *db.Setting) *Setting {
	if e == nil {
		return s
	}

	s.EnableSSO = e.EnableSSO
	s.ForceTwoFactorAuth = e.ForceTwoFactorAuth
	s.DisablePasswordLogin = e.DisablePasswordLogin
	s.EnableAutoLogin = e.EnableAutoLogin
	s.DingtalkOAuth = *cvt.From(e.DingtalkOauth, &DingtalkOAuth{})
	s.CustomOAuth = *cvt.From(e.CustomOauth, &CustomOAuth{})
	s.BaseURL = e.BaseURL
	s.CreatedAt = e.CreatedAt.Unix()
	s.UpdatedAt = e.UpdatedAt.Unix()

	return s
}
