package domain

import (
	"context"

	"github.com/GoYoko/web"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/db"
)

type DepartmentUsecase interface {
	Create(ctx context.Context, req *CreateDepartmentReq) (*Department, error)
	Update(ctx context.Context, req *UpdateDepartmentReq) (*Department, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*Department, error)
	List(ctx context.Context, req *ListDepartmentReq) (*ListDepartmentResp, error)
}

type DepartmentRepo interface {
	Create(ctx context.Context, req *CreateDepartmentRepoReq) (*db.Department, error)
	Update(ctx context.Context, id string, fn func(tx *db.Tx, current *db.Department, updater *db.DepartmentUpdateOne) error) (*db.Department, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*db.Department, error)
	List(ctx context.Context, req *ListDepartmentRepoReq) ([]*db.Department, *db.PageInfo, error)
	HasChildren(ctx context.Context, id uuid.UUID) (bool, error)
	HasPositions(ctx context.Context, id uuid.UUID) (bool, error)
}

type Department struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	ParentID    *string `json:"parent_id,omitempty"`
	CreatedAt   int64   `json:"created_at"`
	UpdatedAt   int64   `json:"updated_at"`
}

func (d *Department) From(entity *db.Department) *Department {
	if entity == nil {
		return d
	}

	d.ID = entity.ID.String()
	d.Name = entity.Name
	if entity.Description != "" {
		desc := entity.Description
		d.Description = &desc
	} else {
		d.Description = nil
	}
	if entity.ParentID != uuid.Nil {
		pid := entity.ParentID.String()
		d.ParentID = &pid
	} else {
		d.ParentID = nil
	}
	d.CreatedAt = entity.CreatedAt.Unix()
	d.UpdatedAt = entity.UpdatedAt.Unix()
	return d
}

type CreateDepartmentReq struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description,omitempty"`
	ParentID    *string `json:"parent_id,omitempty"`
}

type CreateDepartmentRepoReq struct {
	Name        string
	Description *string
	ParentID    *uuid.UUID
}

type UpdateDepartmentReq struct {
	ID          string  `json:"id" validate:"required"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	ParentID    *string `json:"parent_id,omitempty"`
}

type ListDepartmentReq struct {
	web.Pagination

	Keyword  *string `json:"keyword,omitempty" query:"keyword"`
	ParentID *string `json:"parent_id,omitempty" query:"parent_id"`
}

type ListDepartmentRepoReq struct {
	*ListDepartmentReq
}

type ListDepartmentResp struct {
	Items []*Department `json:"items"`
	*db.PageInfo
}
