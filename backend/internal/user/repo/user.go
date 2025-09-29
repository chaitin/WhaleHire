package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/GoYoko/web"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/admin"
	"github.com/chaitin/WhaleHire/backend/db/adminloginhistory"
	"github.com/chaitin/WhaleHire/backend/db/adminrole"
	"github.com/chaitin/WhaleHire/backend/db/user"
	"github.com/chaitin/WhaleHire/backend/db/useridentity"
	"github.com/chaitin/WhaleHire/backend/db/userloginhistory"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/errcode"
	"github.com/chaitin/WhaleHire/backend/pkg/cvt"
	"github.com/chaitin/WhaleHire/backend/pkg/entx"
	"github.com/chaitin/WhaleHire/backend/pkg/ipdb"
)

type UserRepo struct {
	db    *db.Client
	ipdb  *ipdb.IPDB
	redis *redis.Client
	cfg   *config.Config
	cache *cache.Cache
}

func NewUserRepo(
	db *db.Client,
	ipdb *ipdb.IPDB,
	redis *redis.Client,
	cfg *config.Config,
) domain.UserRepo {
	cache := cache.New(5*time.Minute, 10*time.Minute)
	return &UserRepo{db: db, ipdb: ipdb, redis: redis, cfg: cfg, cache: cache}
}

func (r *UserRepo) InitAdmin(ctx context.Context, username, password string) error {
	_, err := r.AdminByName(ctx, username)
	if db.IsNotFound(err) {
		_, err = r.CreateAdmin(ctx, &db.Admin{
			Username: username,
			Password: password,
			Status:   consts.AdminStatusActive,
		}, 1)
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepo) CreateAdmin(ctx context.Context, admin *db.Admin, roleID int64) (*db.Admin, error) {
	var a *db.Admin
	err := entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		new, err := tx.Admin.Create().
			SetUsername(admin.Username).
			SetPassword(admin.Password).
			SetStatus(admin.Status).
			Save(ctx)
		if err != nil {
			return err
		}
		a = new
		return tx.AdminRole.Create().SetAdminID(new.ID).SetRoleID(roleID).Exec(ctx)
	})
	return a, err
}

func (r *UserRepo) AdminByName(ctx context.Context, username string) (*db.Admin, error) {
	admin, err := r.db.Admin.Query().
		WithRoles().
		Where(admin.Username(username)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return admin, nil

}

func (r *UserRepo) GetByName(ctx context.Context, username string) (*db.User, error) {
	return r.db.User.Query().
		Where(
			user.Or(
				user.Username(username),
				user.Email(username),
			),
		).
		Only(ctx)
}

func (r *UserRepo) CreateUser(ctx context.Context, user *db.User) (*db.User, error) {
	var res *db.User
	err := entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		if err := r.checkLimit(ctx, tx); err != nil {
			return err
		}
		u, err := tx.User.Create().
			SetUsername(user.Username).
			SetEmail(user.Email).
			SetPassword(user.Password).
			SetStatus(user.Status).
			SetPlatform(user.Platform).
			Save(ctx)
		if err != nil {
			return err
		}
		res = u
		return nil
	})
	return res, err
}

func (r *UserRepo) UserLoginHistory(ctx context.Context, page *web.Pagination) ([]*db.UserLoginHistory, *db.PageInfo, error) {
	ctx = entx.SkipSoftDelete(ctx)
	q := r.db.UserLoginHistory.Query().WithOwner().Order(userloginhistory.ByCreatedAt(sql.OrderDesc()))
	return q.Page(ctx, page.Page, page.Size)
}

func (r *UserRepo) AdminLoginHistory(ctx context.Context, page *web.Pagination) ([]*db.AdminLoginHistory, *db.PageInfo, error) {
	q := r.db.AdminLoginHistory.Query().WithOwner().Order(adminloginhistory.ByCreatedAt(sql.OrderDesc()))
	return q.Page(ctx, page.Page, page.Size)
}

func (r *UserRepo) AdminList(ctx context.Context, page *web.Pagination) ([]*db.Admin, *db.PageInfo, error) {
	q := r.db.Admin.Query().WithRoles().Order(admin.ByCreatedAt(sql.OrderDesc()))
	return q.Page(ctx, page.Page, page.Size)
}

func (r *UserRepo) List(ctx context.Context, page *web.Pagination) ([]*db.User, *db.PageInfo, error) {
	q := r.db.User.Query().Order(user.ByCreatedAt(sql.OrderDesc()))
	return q.Page(ctx, page.Page, page.Size)
}

func (r *UserRepo) Update(ctx context.Context, id string, fn func(*db.Tx, *db.User, *db.UserUpdateOne) error) (*db.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	var u *db.User
	err = entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		u, err = tx.User.Query().Where(user.ID(uid)).Only(ctx)
		if err != nil {
			return err
		}
		up := tx.User.UpdateOneID(u.ID)
		if err = fn(tx, u, up); err != nil {
			return err
		}
		return up.Exec(ctx)
	})
	return u, err
}

func (r *UserRepo) UpdateAdmin(ctx context.Context, id string, fn func(*db.Tx, *db.Admin, *db.AdminUpdateOne) error) (*db.Admin, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	var a *db.Admin
	err = entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		a, err = tx.Admin.Query().Where(admin.ID(uid)).Only(ctx)
		if err != nil {
			return err
		}
		up := tx.Admin.UpdateOneID(a.ID)
		if err = fn(tx, a, up); err != nil {
			return err
		}
		a, err = up.Save(ctx)
		return err
	})
	return a, err
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		user, err := tx.User.Query().
			WithIdentities().
			Where(user.ID(uid)).
			Only(ctx)
		if err != nil {
			return err
		}

		for _, v := range user.Edges.Identities {
			if _, err := tx.UserIdentity.Delete().Where(useridentity.ID(v.ID)).Exec(ctx); err != nil {
				return err
			}
		}
		return tx.User.DeleteOneID(uid).Exec(ctx)
	})
}

func (r *UserRepo) DeleteAdmin(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	admin, err := r.db.Admin.Get(ctx, uid)
	if err != nil {
		return err
	}
	if admin.Username == "admin" {
		return errors.New("admin cannot be deleted")
	}
	return r.db.Admin.DeleteOne(admin).Exec(ctx)
}

func (r *UserRepo) checkLimit(ctx context.Context, tx *db.Tx) error {
	count, err := tx.User.Query().Count(ctx)
	if err != nil {
		return err
	}
	if count >= r.cfg.Admin.Limit {
		return errcode.ErrUserLimit.Wrap(err)
	}
	return nil
}

func (r *UserRepo) SaveAdminLoginHistory(ctx context.Context, adminID string, ip string) error {
	uid, err := uuid.Parse(adminID)
	if err != nil {
		return err
	}
	addr, err := r.ipdb.Lookup(ip)
	if err != nil {
		return err
	}
	// 检查 addr 是否为 nil，防止 nil pointer dereference
	if addr == nil {
		addr = &domain.IPAddress{
			IP:       ip,
			Country:  "未知",
			Province: "未知",
			City:     "未知",
		}
	}
	_, err = r.db.AdminLoginHistory.Create().
		SetAdminID(uid).
		SetIP(ip).
		SetCity(addr.City).
		SetCountry(addr.Country).
		SetProvince(addr.Province).
		Save(ctx)
	return err
}

func (r *UserRepo) SaveUserLoginHistory(ctx context.Context, userID string, ip string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	addr, err := r.ipdb.Lookup(ip)
	if err != nil {
		return err
	}
	// 检查 addr 是否为 nil，防止 nil pointer dereference
	if addr == nil {
		addr = &domain.IPAddress{
			IP:       ip,
			Country:  "未知",
			Province: "未知",
			City:     "未知",
		}
	}
	c := r.db.UserLoginHistory.Create().
		SetUserID(uid).
		SetIP(ip).
		SetCity(addr.City).
		SetCountry(addr.Country).
		SetProvince(addr.Province)

	return c.Exec(ctx)
}

func (r *UserRepo) GetUserCount(ctx context.Context) (int64, error) {
	count, err := r.db.User.Query().Count(ctx)
	if err != nil {
		return 0, err
	}
	return int64(count), nil
}
func (r *UserRepo) ListRole(ctx context.Context) ([]*db.Role, error) {
	return r.db.Role.Query().All(ctx)
}

func (r *UserRepo) GrantRole(ctx context.Context, req *domain.GrantRoleReq) error {
	return entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		if _, err := tx.AdminRole.Delete().
			Where(adminrole.AdminID(req.AdminID)).
			Exec(ctx); err != nil {
			return err
		}
		for _, rid := range req.RoleIDs {
			if err := tx.AdminRole.Create().
				SetAdminID(req.AdminID).
				SetRoleID(rid).
				Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *UserRepo) GetPermissions(ctx context.Context, id uuid.UUID) (*domain.Permissions, error) {
	key := fmt.Sprintf("user_permissions:%s", id.String())
	if cached, found := r.cache.Get(key); found {
		return cached.(*domain.Permissions), nil
	}

	admin, err := r.db.Admin.Query().
		WithRoles().
		Where(admin.ID(id)).
		First(ctx)
	if err != nil {
		return nil, err
	}

	res := &domain.Permissions{
		AdminID: admin.ID,
		IsAdmin: admin.Username == "admin",
	}

	for _, r := range admin.Edges.Roles {
		if r.ID == 1 {
			res.IsAdmin = true
		}
		res.Roles = append(res.Roles, cvt.From(r, &domain.Role{}))
	}

	r.cache.Set(key, res, 5*time.Minute)
	return res, nil
}

func (r *UserRepo) GetSetting(ctx context.Context) (*db.Setting, error) {
	if b, err := r.redis.Get(ctx, "setting").Result(); err == nil {
		s := &db.Setting{}
		if err := json.Unmarshal([]byte(b), s); err == nil {
			return s, nil
		}
	}
	s, err := r.db.Setting.Query().First(ctx)
	if db.IsNotFound(err) {
		s, err = r.db.Setting.Create().
			SetEnableSSO(false).
			SetForceTwoFactorAuth(false).
			SetDisablePasswordLogin(false).
			Save(ctx)
	}
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	if err := r.redis.Set(ctx, "setting", b, time.Hour*24).Err(); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *UserRepo) UpdateSetting(ctx context.Context, fn func(*db.Setting, *db.SettingUpdateOne)) (*db.Setting, error) {
	var res *db.Setting
	err := entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		s, err := tx.Setting.Query().First(ctx)
		if err != nil {
			return err
		}
		up := tx.Setting.UpdateOneID(s.ID)
		fn(s, up)
		s, err = up.Save(ctx)
		if err != nil {
			return err
		}
		res = s
		return r.redis.Del(ctx, "setting").Err()
	})
	return res, err
}

func (r *UserRepo) updateUsername(ctx context.Context, tx *db.Tx, ui *db.UserIdentity, name string) error {
	if err := tx.UserIdentity.UpdateOneID(ui.ID).SetNickname(name).Exec(ctx); err != nil {
		return err
	}
	return tx.User.UpdateOneID(ui.UserID).SetUsername(name).Exec(ctx)
}

func (r *UserRepo) updateAvatar(ctx context.Context, tx *db.Tx, ui *db.UserIdentity, avatar string) error {
	if err := tx.UserIdentity.UpdateOneID(ui.ID).SetAvatarURL(avatar).Exec(ctx); err != nil {
		return err
	}
	return tx.User.UpdateOneID(ui.UserID).SetAvatarURL(avatar).Exec(ctx)
}

func (r *UserRepo) SignUpOrIn(ctx context.Context, platform consts.UserPlatform, req *domain.OAuthUserInfo) (*db.User, error) {
	var u *db.User
	err := entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		ui, err := tx.UserIdentity.Query().
			WithUser().
			Where(useridentity.Platform(platform), useridentity.IdentityID(req.ID)).
			First(ctx)
		if err == nil {
			u = ui.Edges.User
			if u.Status != consts.UserStatusActive {
				return errcode.ErrUserLock.Wrap(fmt.Errorf("user is locked"))
			}
			if ui.Nickname != req.Name {
				if err = r.updateUsername(ctx, tx, ui, req.Name); err != nil {
					return err
				}
			}
			if ui.AvatarURL != req.AvatarURL {
				if err = r.updateAvatar(ctx, tx, ui, req.AvatarURL); err != nil {
					return err
				}
			}
			return nil
		}
		if !db.IsNotFound(err) {
			return err
		}
		user, err := tx.User.Create().
			SetUsername(req.Name).
			SetEmail(req.Email).
			SetAvatarURL(req.AvatarURL).
			SetPlatform(platform).
			SetStatus(consts.UserStatusActive).
			Save(ctx)
		if err != nil {
			return err
		}
		_, err = tx.UserIdentity.Create().
			SetUserID(user.ID).
			SetPlatform(platform).
			SetIdentityID(req.ID).
			SetUnionID(req.UnionID).
			SetNickname(req.Name).
			SetAvatarURL(req.AvatarURL).
			SetEmail(req.Email).
			Save(ctx)
		if err != nil {
			return err
		}
		u = user
		return nil
	})
	return u, err
}

func (r *UserRepo) OAuthLogin(ctx context.Context, platform consts.UserPlatform, req *domain.OAuthUserInfo) (*db.User, error) {
	ui, err := r.db.UserIdentity.Query().
		WithUser().
		Where(useridentity.Platform(platform), useridentity.IdentityID(req.ID)).
		Where(useridentity.HasUser()).
		Only(ctx)
	if err != nil {
		return nil, errcode.ErrNotInvited.Wrap(err)
	}
	if ui.Edges.User.Status != consts.UserStatusActive {
		return nil, errcode.ErrUserLock.Wrap(fmt.Errorf("user is locked"))
	}
	if ui.Nickname != req.Name {
		if err = entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
			return r.updateUsername(ctx, tx, ui, req.Name)
		}); err != nil {
			return nil, err
		}
	}
	if ui.AvatarURL != req.AvatarURL {
		if err = entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
			return r.updateAvatar(ctx, tx, ui, req.AvatarURL)
		}); err != nil {
			return nil, err
		}
	}
	return ui.Edges.User, nil
}

func (r *UserRepo) OAuthRegister(ctx context.Context, platform consts.UserPlatform, inviteCode string, req *domain.OAuthUserInfo) (*db.User, error) {
	var u *db.User
	err := entx.WithTx(ctx, r.db, func(tx *db.Tx) error {

		_, err := tx.UserIdentity.Query().
			WithUser().
			Where(useridentity.Platform(platform), useridentity.IdentityID(req.ID)).
			First(ctx)
		if err == nil {
			e := fmt.Errorf("user already exists for platform %s and identity ID %s", platform, req.ID)
			return errcode.ErrAccountAlreadyExist.Wrap(e)
		}
		if !db.IsNotFound(err) {
			return err
		}
		user, err := tx.User.Create().
			SetUsername(req.Name).
			SetEmail(req.Email).
			SetAvatarURL(req.AvatarURL).
			SetPlatform(platform).
			SetStatus(consts.UserStatusActive).
			Save(ctx)
		if err != nil {
			return err
		}
		_, err = tx.UserIdentity.Create().
			SetUserID(user.ID).
			SetPlatform(platform).
			SetIdentityID(req.ID).
			SetUnionID(req.UnionID).
			SetNickname(req.Name).
			SetAvatarURL(req.AvatarURL).
			SetEmail(req.Email).
			Save(ctx)
		if err != nil {
			return err
		}
		u = user
		return nil
	})
	return u, err
}
