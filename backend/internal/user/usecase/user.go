package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/GoYoko/web"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/ent/types"
	"github.com/chaitin/WhaleHire/backend/errcode"
	"github.com/chaitin/WhaleHire/backend/pkg/cvt"
	"github.com/chaitin/WhaleHire/backend/pkg/oauth"
	"github.com/chaitin/WhaleHire/backend/pkg/session"
)

type UserUsecase struct {
	cfg     *config.Config
	redis   *redis.Client
	repo    domain.UserRepo
	logger  *slog.Logger
	session *session.Session
}

func NewUserUsecase(
	cfg *config.Config,
	redis *redis.Client,
	repo domain.UserRepo,
	logger *slog.Logger,
	session *session.Session,
) domain.UserUsecase {
	u := &UserUsecase{
		cfg:     cfg,
		redis:   redis,
		repo:    repo,
		logger:  logger,
		session: session,
	}
	return u
}

func (u *UserUsecase) InitAdmin(ctx context.Context) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.cfg.Admin.Password), bcrypt.DefaultCost)
	if err != nil {
		u.logger.Error("generate admin password", "error", err)
		return err
	}
	return u.repo.InitAdmin(ctx, u.cfg.Admin.User, string(hash))
}

func (u *UserUsecase) List(ctx context.Context, req domain.ListReq) (*domain.ListUserResp, error) {
	users, p, err := u.repo.List(ctx, &req.Pagination)
	if err != nil {
		return nil, err
	}

	ids := cvt.Iter(users, func(_ int, u *db.User) string { return u.ID.String() })
	m, err := u.getUserActive(ctx, ids)
	if err != nil {
		return nil, err
	}

	return &domain.ListUserResp{
		PageInfo: p,
		Users: cvt.Iter(users, func(_ int, e *db.User) *domain.User {
			return cvt.From(e, &domain.User{
				LastActiveAt: m[e.ID.String()],
			})
		}),
	}, nil
}

func (u *UserUsecase) getUserActive(ctx context.Context, ids []string) (map[string]int64, error) {
	m := make(map[string]int64)
	for _, id := range ids {
		key := fmt.Sprintf(consts.UserActiveKeyFmt, id)
		if t, err := u.redis.Get(ctx, key).Int64(); err != nil && !errors.Is(err, redis.Nil) {
			u.logger.With("key", key).With("error", err).Warn("get user active time failed")
		} else {
			m[id] = t
		}
	}

	return m, nil
}

// AdminList implements domain.UserUsecase.
func (u *UserUsecase) AdminList(ctx context.Context, page *web.Pagination) (*domain.ListAdminUserResp, error) {
	admins, p, err := u.repo.AdminList(ctx, page)
	if err != nil {
		return nil, err
	}

	ids := cvt.Iter(admins, func(_ int, u *db.Admin) string { return u.ID.String() })
	m, err := u.getAdminActive(ctx, ids)
	if err != nil {
		return nil, err
	}

	return &domain.ListAdminUserResp{
		PageInfo: p,
		Users: cvt.Iter(admins, func(_ int, e *db.Admin) *domain.AdminUser {
			return cvt.From(e, &domain.AdminUser{
				LastActiveAt: m[e.ID.String()],
			})
		}),
	}, nil
}

func (u *UserUsecase) getAdminActive(ctx context.Context, ids []string) (map[string]int64, error) {
	m := make(map[string]int64)
	for _, id := range ids {
		key := fmt.Sprintf(consts.AdminActiveKeyFmt, id)
		if t, err := u.redis.Get(ctx, key).Int64(); err != nil && !errors.Is(err, redis.Nil) {
			u.logger.With("key", key).With("error", err).Warn("get admin active time failed")
		} else {
			m[id] = t
		}
	}

	return m, nil
}

// AdminLoginHistory implements domain.UserUsecase.
func (u *UserUsecase) AdminLoginHistory(ctx context.Context, page *web.Pagination) (*domain.ListAdminLoginHistoryResp, error) {
	histories, p, err := u.repo.AdminLoginHistory(ctx, page)
	if err != nil {
		return nil, err
	}

	return &domain.ListAdminLoginHistoryResp{
		PageInfo: p,
		LoginHistories: cvt.Iter(histories, func(_ int, e *db.AdminLoginHistory) *domain.AdminLoginHistory {
			return cvt.From(e, &domain.AdminLoginHistory{})
		}),
	}, nil
}

// LoginHistory implements domain.UserUsecase.
func (u *UserUsecase) LoginHistory(ctx context.Context, page *web.Pagination) (*domain.ListLoginHistoryResp, error) {
	histories, p, err := u.repo.UserLoginHistory(ctx, page)
	if err != nil {
		return nil, err
	}

	return &domain.ListLoginHistoryResp{
		PageInfo: p,
		LoginHistories: cvt.Iter(histories, func(_ int, e *db.UserLoginHistory) *domain.UserLoginHistory {
			return cvt.From(e, &domain.UserLoginHistory{}).From(e)
		}),
	}, nil
}

// Register implements domain.UserUsecase.
func (u *UserUsecase) Register(ctx context.Context, req *domain.RegisterReq) (*domain.User, error) {

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &db.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hash),
		Status:   consts.UserStatusActive,
		Platform: consts.UserPlatformEmail,
	}
	n, err := u.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return cvt.From(n, &domain.User{}), nil
}

func (u *UserUsecase) Login(ctx context.Context, req *domain.LoginReq) (*domain.LoginResp, error) {
	user, err := u.repo.GetByName(ctx, req.Username)
	if err != nil {
		return nil, errcode.ErrUserNotFound.Wrap(err)
	}
	if user.Status != consts.UserStatusActive {
		return nil, errcode.ErrUserLock
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errcode.ErrPassword.Wrap(err)
	}

	switch req.Source {
	case consts.LoginSourceBrowser:
		if err := u.repo.SaveUserLoginHistory(ctx, user.ID.String(), req.IP); err != nil {
			u.logger.With("error", err).Error("save user login history")
		}
		return &domain.LoginResp{
			RedirectURL: "",
			User:        cvt.From(user, &domain.User{}),
		}, nil
	}

	return nil, fmt.Errorf("invalid login source")
}

func (u *UserUsecase) AdminLogin(ctx context.Context, req *domain.LoginReq) (*domain.AdminUser, error) {
	admin, err := u.repo.AdminByName(ctx, req.Username)
	if err != nil {
		return nil, errcode.ErrUserNotFound.Wrap(err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		return nil, errcode.ErrPassword.Wrap(err)
	}

	if err := u.repo.SaveAdminLoginHistory(ctx, admin.ID.String(), req.IP); err != nil {
		return nil, err
	}
	return cvt.From(admin, &domain.AdminUser{}), nil
}

func (u *UserUsecase) CreateAdmin(ctx context.Context, req *domain.CreateAdminReq) (*domain.AdminUser, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	admin := &db.Admin{
		Username: req.Username,
		Password: string(hash),
	}
	n, err := u.repo.CreateAdmin(ctx, admin, req.RoleID)
	if err != nil {
		return nil, err
	}
	return &domain.AdminUser{
		ID:        n.ID,
		Username:  n.Username,
		CreatedAt: n.CreatedAt.Unix(),
	}, nil
}

func (u *UserUsecase) Update(ctx context.Context, req *domain.UpdateUserReq) (*domain.User, error) {
	user, err := u.repo.Update(ctx, req.ID, func(tx *db.Tx, old *db.User, up *db.UserUpdateOne) error {
		if req.Status != nil {
			up.SetStatus(*req.Status)
		}
		if req.Password != nil {
			hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			up.SetPassword(string(hash))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return cvt.From(user, &domain.User{}), nil
}

func (u *UserUsecase) Delete(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}

func (u *UserUsecase) DeleteAdmin(ctx context.Context, id string) error {
	return u.repo.DeleteAdmin(ctx, id)
}

func (u *UserUsecase) ProfileUpdate(ctx context.Context, req *domain.ProfileUpdateReq) (*domain.User, error) {
	user, err := u.repo.Update(ctx, req.UID, func(_ *db.Tx, old *db.User, uuo *db.UserUpdateOne) error {
		if req.Avatar != nil {
			uuo.SetAvatarURL(*req.Avatar)
		}

		if req.Username != nil {
			uuo.SetUsername(*req.Username)
		}

		if req.Password != nil && req.OldPassword != nil {
			if err := bcrypt.CompareHashAndPassword([]byte(old.Password), []byte(*req.OldPassword)); err != nil {
				return errcode.ErrPassword.Wrap(err)
			}

			hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
			if err != nil {
				return errcode.ErrPassword.Wrap(err)
			}
			uuo.SetPassword(string(hash))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return cvt.From(user, &domain.User{}), nil
}

func (u *UserUsecase) GetUserCount(ctx context.Context) (int64, error) {
	return u.repo.GetUserCount(ctx)
}

func (u *UserUsecase) ListRole(ctx context.Context) ([]*domain.Role, error) {
	roles, err := u.repo.ListRole(ctx)
	if err != nil {
		return nil, err
	}
	return cvt.Iter(roles, func(i int, role *db.Role) *domain.Role {
		return cvt.From(role, &domain.Role{})
	}), nil
}

func (u *UserUsecase) GrantRole(ctx context.Context, req *domain.GrantRoleReq) error {
	return u.repo.GrantRole(ctx, req)
}

func (u *UserUsecase) GetPermissions(ctx context.Context, id uuid.UUID) (*domain.Permissions, error) {
	return u.repo.GetPermissions(ctx, id)
}

func (u *UserUsecase) getOAuthConfig(baseURL string, setting *db.Setting, platform consts.UserPlatform) (*domain.OAuthConfig, error) {
	cfg := domain.OAuthConfig{
		Debug:       u.cfg.Debug,
		Platform:    platform,
		RedirectURI: fmt.Sprintf("%s/oauth/callback", baseURL),
	}

	switch platform {
	case consts.UserPlatformDingTalk:
		if setting.DingtalkOauth == nil || !setting.DingtalkOauth.Enable {
			return nil, errcode.ErrDingtalkNotEnabled
		}
		cfg.ClientID = setting.DingtalkOauth.ClientID
		cfg.ClientSecret = setting.DingtalkOauth.ClientSecret
	case consts.UserPlatformCustom:
		if setting.CustomOauth == nil || !setting.CustomOauth.Enable {
			return nil, errcode.ErrCustomNotEnabled
		}
		cfg.ClientID = setting.CustomOauth.ClientID
		cfg.ClientSecret = setting.CustomOauth.ClientSecret
		cfg.AuthorizeURL = setting.CustomOauth.AuthorizeURL
		cfg.Scopes = setting.CustomOauth.Scopes
		cfg.TokenURL = setting.CustomOauth.AccessTokenURL
		cfg.UserInfoURL = setting.CustomOauth.UserInfoURL
		cfg.IDField = setting.CustomOauth.IDField
		cfg.NameField = setting.CustomOauth.NameField
		cfg.AvatarField = setting.CustomOauth.AvatarField
		cfg.EmailField = setting.CustomOauth.EmailField
	default:
		return nil, errcode.ErrUnsupportedPlatform
	}

	return &cfg, nil
}

func (u *UserUsecase) GetSetting(ctx context.Context) (*domain.Setting, error) {
	s, err := u.repo.GetSetting(ctx)
	if err != nil {
		return nil, err
	}
	return cvt.From(s, &domain.Setting{}), nil
}

func (u *UserUsecase) OAuthSignUpOrIn(ctx context.Context, req *domain.OAuthSignUpOrInReq) (*domain.OAuthURLResp, error) {
	u.logger.With("req", req).Debug("OAuthSignUpOrIn request")
	setting, err := u.repo.GetSetting(ctx)
	if err != nil {
		return nil, err
	}
	cfg, err := u.getOAuthConfig(req.BaseURL, setting, req.Platform)
	if err != nil {
		return nil, err
	}

	u.logger.With("cfg", cfg).Debug("OAuth config created")

	oauth, err := oauth.NewOAuther(*cfg)
	if err != nil {
		return nil, err
	}
	state, url := oauth.GetAuthorizeURL()

	session := &domain.OAuthState{
		Source:      req.Source,
		SessionID:   req.SessionID,
		Kind:        req.OAuthKind(),
		Platform:    req.Platform,
		RedirectURL: req.RedirectURL,
		InviteCode:  req.InviteCode,
	}
	b, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}
	if err := u.redis.Set(ctx, fmt.Sprintf("oauth:state:%s", state), b, 15*time.Minute).Err(); err != nil {
		return nil, err
	}

	return &domain.OAuthURLResp{
		URL: url,
	}, nil
}

func (u *UserUsecase) FetchUserInfo(ctx context.Context, req *domain.OAuthCallbackReq, session *domain.OAuthState) (*domain.OAuthUserInfo, error) {
	setting, err := u.repo.GetSetting(ctx)
	if err != nil {
		u.logger.With("error", err).Warn("failed to get setting in FetchUserInfo")
		return nil, err
	}

	cfg, err := u.getOAuthConfig(req.BaseURL, setting, session.Platform)
	if err != nil {
		u.logger.With("error", err).With("platform", session.Platform).Warn("failed to get OAuth config")
		return nil, err
	}

	oauth, err := oauth.NewOAuther(*cfg)
	if err != nil {
		u.logger.With("error", err).With("config", cfg).Warn("failed to create OAuth client")
		return nil, err
	}
	userInfo, err := oauth.GetUserInfo(req.Code)
	if err != nil {
		u.logger.With("error", err).With("code", req.Code).With("platform", session.Platform).Warn("failed to get user info from OAuth provider")
		return nil, err
	}
	return userInfo, nil
}

func (u *UserUsecase) OAuthCallback(c *web.Context, req *domain.OAuthCallbackReq) (*domain.OAuthCallbackResp, error) {
	ctx := c.Request().Context()
	req.IP = c.RealIP()
	s, err := u.GetSetting(ctx)
	if err != nil {
		return nil, err
	}
	req.BaseURL = u.cfg.GetBaseURL(c.Request(), s)
	b, err := u.redis.Get(ctx, fmt.Sprintf("oauth:state:%s", req.State)).Result()
	if err != nil {
		return nil, err
	}
	var session domain.OAuthState
	if err := json.Unmarshal([]byte(b), &session); err != nil {
		return nil, err
	}

	switch session.Kind {
	case consts.OAuthKindInvite:
		setting, err := u.repo.GetSetting(ctx)
		if err != nil {
			return nil, err
		}
		_, redirect, err := u.WithOAuthCallback(ctx, req, &session, func(ctx context.Context, s *domain.OAuthState, oui *domain.OAuthUserInfo) (*db.User, error) {
			if setting.EnableAutoLogin {
				return u.repo.SignUpOrIn(ctx, s.Platform, oui)
			}
			return u.repo.OAuthRegister(ctx, s.Platform, s.InviteCode, oui)
		})
		if err != nil {
			return nil, err
		}
		return &domain.OAuthCallbackResp{RedirectURL: redirect}, nil

	case consts.OAuthKindLogin:
		setting, err := u.repo.GetSetting(ctx)
		if err != nil {
			return nil, err
		}
		user, redirect, err := u.WithOAuthCallback(ctx, req, &session, func(ctx context.Context, s *domain.OAuthState, oui *domain.OAuthUserInfo) (*db.User, error) {
			if setting.EnableAutoLogin {
				return u.repo.SignUpOrIn(ctx, s.Platform, oui)
			}
			return u.repo.OAuthLogin(ctx, s.Platform, oui)
		})
		if err != nil {
			return nil, err
		}
		u.logger.With("session", session).With("platform", session.Source).Debug("OAuthCallback login session")
		if session.Source == consts.LoginSourceBrowser {
			resUser := cvt.From(user, &domain.User{})
			u.logger.With("user", resUser).With("host", c.Request().Host).DebugContext(ctx, "save user session")
			if _, err := u.session.Save(c, consts.UserSessionName, c.Request().Host, resUser); err != nil {
				return nil, err
			}
		}

		return &domain.OAuthCallbackResp{RedirectURL: redirect}, nil

	default:
		return nil, errcode.ErrOAuthStateInvalid
	}
}

type OAuthUserRepoHandle func(context.Context, *domain.OAuthState, *domain.OAuthUserInfo) (*db.User, error)

func (u *UserUsecase) WithOAuthCallback(ctx context.Context, req *domain.OAuthCallbackReq, session *domain.OAuthState, handle OAuthUserRepoHandle) (*db.User, string, error) {
	info, err := u.FetchUserInfo(ctx, req, session)
	if err != nil {
		return nil, "", err
	}

	user, err := handle(ctx, session, info)
	if err != nil {
		return nil, "", err
	}

	redirect := session.RedirectURL

	// 添加详细的调试日志
	u.logger.With("session", session).With("redirect", redirect).With("redirect_empty", redirect == "").With("redirect_length", len(redirect)).Debug("oauth callback redirect analysis")

	// 如果redirect为空，记录警告
	if redirect == "" {
		u.logger.Warn("redirect URL is empty in OAuth callback", "session", session)
	}

	return user, redirect, nil
}

func (u *UserUsecase) UpdateSetting(ctx context.Context, req *domain.UpdateSettingReq) (*domain.Setting, error) {
	s, err := u.repo.UpdateSetting(ctx, func(old *db.Setting, up *db.SettingUpdateOne) {
		if req.EnableSSO != nil {
			up.SetEnableSSO(*req.EnableSSO)
		}
		if req.ForceTwoFactorAuth != nil {
			up.SetForceTwoFactorAuth(*req.ForceTwoFactorAuth)
		}
		if req.DisablePasswordLogin != nil {
			up.SetDisablePasswordLogin(*req.DisablePasswordLogin)
		}
		if req.EnableAutoLogin != nil {
			up.SetEnableAutoLogin(*req.EnableAutoLogin)
		}
		if req.DingtalkOAuth != nil {
			dingtalk := cvt.NilWithDefault(old.DingtalkOauth, &types.DingtalkOAuth{})
			if req.DingtalkOAuth.Enable != nil {
				dingtalk.Enable = *req.DingtalkOAuth.Enable
			}
			if req.DingtalkOAuth.ClientID != nil {
				dingtalk.ClientID = *req.DingtalkOAuth.ClientID
			}
			if req.DingtalkOAuth.ClientSecret != nil {
				dingtalk.ClientSecret = *req.DingtalkOAuth.ClientSecret
			}
			up.SetDingtalkOauth(dingtalk)
		}
		if req.CustomOAuth != nil {
			custom := cvt.NilWithDefault(old.CustomOauth, &types.CustomOAuth{})
			if req.CustomOAuth.Enable != nil {
				custom.Enable = *req.CustomOAuth.Enable
			}
			if req.CustomOAuth.ClientID != nil {
				custom.ClientID = *req.CustomOAuth.ClientID
			}
			if req.CustomOAuth.ClientSecret != nil {
				custom.ClientSecret = *req.CustomOAuth.ClientSecret
			}
			if req.CustomOAuth.AuthorizeURL != nil {
				custom.AuthorizeURL = *req.CustomOAuth.AuthorizeURL
			}
			if req.CustomOAuth.AccessTokenURL != nil {
				custom.AccessTokenURL = *req.CustomOAuth.AccessTokenURL
			}
			if req.CustomOAuth.UserInfoURL != nil {
				custom.UserInfoURL = *req.CustomOAuth.UserInfoURL
			}
			if req.CustomOAuth.Scopes != nil {
				custom.Scopes = req.CustomOAuth.Scopes
			}
			if req.CustomOAuth.IDField != nil {
				custom.IDField = *req.CustomOAuth.IDField
			}
			if req.CustomOAuth.NameField != nil {
				custom.NameField = *req.CustomOAuth.NameField
			}
			if req.CustomOAuth.AvatarField != nil {
				custom.AvatarField = *req.CustomOAuth.AvatarField
			}
			if req.CustomOAuth.EmailField != nil {
				custom.EmailField = *req.CustomOAuth.EmailField
			}
			up.SetCustomOauth(custom)
		}
		if req.BaseURL != nil {
			up.SetBaseURL(*req.BaseURL)
		}
	})
	if err != nil {
		return nil, err
	}
	return cvt.From(s, &domain.Setting{}), nil
}
