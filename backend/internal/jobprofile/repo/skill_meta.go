package repo

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/jobskillmeta"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/entx"
)

type JobSkillMetaRepo struct {
	db *db.Client
}

func NewJobSkillMetaRepo(db *db.Client) domain.JobSkillMetaRepo {
	return &JobSkillMetaRepo{db: db}
}

func (r *JobSkillMetaRepo) Create(ctx context.Context, name string) (*db.JobSkillMeta, error) {
	if name == "" {
		return nil, fmt.Errorf("skill name is required")
	}

	return r.db.JobSkillMeta.Create().
		SetName(name).
		Save(ctx)
}

func (r *JobSkillMetaRepo) GetByName(ctx context.Context, name string) (*db.JobSkillMeta, error) {
	if name == "" {
		return nil, fmt.Errorf("skill name is required")
	}

	return r.db.JobSkillMeta.Query().
		Where(jobskillmeta.NameEqualFold(name)).
		Where(jobskillmeta.DeletedAtIsNil()).
		Only(ctx)
}

func (r *JobSkillMetaRepo) List(ctx context.Context, req *domain.ListSkillMetaReq) ([]*db.JobSkillMeta, *db.PageInfo, error) {
	query := r.db.JobSkillMeta.Query().
		Where(jobskillmeta.DeletedAtIsNil()).
		Order(jobskillmeta.ByUpdatedAt(sql.OrderDesc()))

	if req != nil {
		if req.Keyword != nil && *req.Keyword != "" {
			query = query.Where(jobskillmeta.NameContainsFold(*req.Keyword))
		}

	}

	return query.Page(ctx, req.Pagination.Page, req.Pagination.Size)
}

// Delete performs soft delete using SoftDeleteMixin
func (r *JobSkillMetaRepo) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("skill meta id is required")
	}

	return r.db.JobSkillMeta.DeleteOneID(uuid.MustParse(id)).Exec(ctx)
}

// HardDelete performs physical deletion from database
func (r *JobSkillMetaRepo) HardDelete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("skill meta id is required")
	}

	// 使用 SkipSoftDelete 上下文跳过软删除逻辑，执行真正的物理删除
	ctx = entx.SkipSoftDelete(ctx)
	return r.db.JobSkillMeta.DeleteOneID(uuid.MustParse(id)).Exec(ctx)
}
