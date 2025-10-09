package repo

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/department"
	"github.com/chaitin/WhaleHire/backend/db/jobposition"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/entx"
)

type DepartmentRepo struct {
	db *db.Client
}

func NewDepartmentRepo(dbClient *db.Client) domain.DepartmentRepo {
	return &DepartmentRepo{db: dbClient}
}

func (r *DepartmentRepo) Create(ctx context.Context, req *domain.CreateDepartmentRepoReq) (*db.Department, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	creator := r.db.Department.Create().
		SetName(req.Name)

	if req.Description != nil {
		desc := strings.TrimSpace(*req.Description)
		creator.SetDescription(desc)
	}
	if req.ParentID != nil {
		creator.SetParentID(*req.ParentID)
	}

	return creator.Save(ctx)
}

func (r *DepartmentRepo) Update(ctx context.Context, id string, fn func(tx *db.Tx, current *db.Department, updater *db.DepartmentUpdateOne) error) (*db.Department, error) {
	departmentID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid department ID: %w", err)
	}

	var result *db.Department
	err = entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		current, err := tx.Department.Query().
			Where(department.ID(departmentID), department.DeletedAtIsNil()).
			Only(ctx)
		if err != nil {
			return err
		}

		updater := tx.Department.UpdateOneID(departmentID)
		if err := fn(tx, current, updater); err != nil {
			return err
		}
		if _, err := updater.Save(ctx); err != nil {
			return err
		}

		result, err = tx.Department.Query().
			Where(department.ID(departmentID), department.DeletedAtIsNil()).
			Only(ctx)
		return err
	})

	return result, err
}

// Delete 软删除部门
func (r *DepartmentRepo) Delete(ctx context.Context, id string) error {
	departmentID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid department ID: %w", err)
	}

	return r.db.Department.DeleteOneID(departmentID).Exec(ctx)
}

// HardDelete 硬删除部门（物理删除）
func (r *DepartmentRepo) HardDelete(ctx context.Context, id string) error {
	departmentID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid department ID: %w", err)
	}

	// 使用 SkipSoftDelete 上下文跳过软删除逻辑，执行真正的物理删除
	ctx = entx.SkipSoftDelete(ctx)
	return r.db.Department.DeleteOneID(departmentID).Exec(ctx)
}

func (r *DepartmentRepo) GetByID(ctx context.Context, id string) (*db.Department, error) {
	departmentID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid department ID: %w", err)
	}

	return r.db.Department.Query().
		Where(department.ID(departmentID), department.DeletedAtIsNil()).
		Only(ctx)
}

func (r *DepartmentRepo) List(ctx context.Context, req *domain.ListDepartmentRepoReq) ([]*db.Department, *db.PageInfo, error) {
	query := r.db.Department.Query().
		Where(department.DeletedAtIsNil()) // 排除已软删除的记录

	if req != nil && req.ListDepartmentReq != nil {
		filters := req.ListDepartmentReq

		if filters.ParentID != nil {
			parent := strings.TrimSpace(*filters.ParentID)
			switch parent {
			case "":
				query = query.Where(department.ParentIDIsNil())
			default:
				parentID, err := uuid.Parse(parent)
				if err != nil {
					return nil, nil, fmt.Errorf("invalid parent ID: %w", err)
				}
				query = query.Where(department.ParentID(parentID))
			}
		}

		if filters.Keyword != nil {
			keyword := strings.TrimSpace(*filters.Keyword)
			if keyword != "" {
				terms := strings.Fields(keyword)
				for _, term := range terms {
					query = query.Where(department.Or(
						department.NameContainsFold(term),
						department.DescriptionContainsFold(term),
					))
				}
			}
		}
	}

	query = query.Order(department.ByUpdatedAt(sql.OrderDesc()))

	return query.Page(ctx, req.ListDepartmentReq.Pagination.Page, req.ListDepartmentReq.Pagination.Size)
}

func (r *DepartmentRepo) HasChildren(ctx context.Context, id uuid.UUID) (bool, error) {
	return r.db.Department.Query().
		Where(department.ParentID(id)).
		Exist(ctx)
}

func (r *DepartmentRepo) HasPositions(ctx context.Context, id uuid.UUID) (bool, error) {
	return r.db.JobPosition.Query().Where(
		jobposition.DepartmentID(id),
		jobposition.DeletedAtIsNil(),
	).Exist(ctx)
}
