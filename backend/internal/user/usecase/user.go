package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/GoYoko/web"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/errcode"
	"github.com/chaitin/WhaleHire/backend/pkg/cvt"
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
