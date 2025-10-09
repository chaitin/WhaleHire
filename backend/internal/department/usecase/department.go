package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/domain"
)

type DepartmentUsecase struct {
	repo   domain.DepartmentRepo
	logger *slog.Logger
}

func NewDepartmentUsecase(repo domain.DepartmentRepo, logger *slog.Logger) domain.DepartmentUsecase {
	return &DepartmentUsecase{
		repo:   repo,
		logger: logger,
	}
}

func (u *DepartmentUsecase) Create(ctx context.Context, req *domain.CreateDepartmentReq) (*domain.Department, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("department name is required")
	}

	var parentUUID *uuid.UUID
	if req.ParentID != nil {
		parent := strings.TrimSpace(*req.ParentID)
		if parent != "" {
			id, err := uuid.Parse(parent)
			if err != nil {
				return nil, fmt.Errorf("invalid parent id: %w", err)
			}
			if _, err := u.repo.GetByID(ctx, id.String()); err != nil {
				if db.IsNotFound(err) {
					return nil, fmt.Errorf("parent department not found")
				}
				return nil, fmt.Errorf("failed to load parent department: %w", err)
			}
			parentUUID = &id
		}
	}

	var description *string
	if req.Description != nil {
		desc := strings.TrimSpace(*req.Description)
		if desc != "" {
			description = &desc
		}
	}

	created, err := u.repo.Create(ctx, &domain.CreateDepartmentRepoReq{
		Name:        name,
		Description: description,
		ParentID:    parentUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create department: %w", err)
	}

	return (&domain.Department{}).From(created), nil
}

func (u *DepartmentUsecase) Update(ctx context.Context, req *domain.UpdateDepartmentReq) (*domain.Department, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	id := strings.TrimSpace(req.ID)
	if id == "" {
		return nil, fmt.Errorf("department id is required")
	}

	updated, err := u.repo.Update(ctx, id, func(tx *db.Tx, current *db.Department, updater *db.DepartmentUpdateOne) error {
		if req.Name != nil {
			name := strings.TrimSpace(*req.Name)
			if name == "" {
				return fmt.Errorf("department name is required")
			}
			updater.SetName(name)
		}

		if req.Description != nil {
			desc := strings.TrimSpace(*req.Description)
			if desc == "" {
				updater.ClearDescription()
			} else {
				updater.SetDescription(desc)
			}
		}

		if req.ParentID != nil {
			parent := strings.TrimSpace(*req.ParentID)
			if parent == "" {
				updater.ClearParentID()
				return nil
			}

			parentID, err := uuid.Parse(parent)
			if err != nil {
				return fmt.Errorf("invalid parent id: %w", err)
			}
			if parentID == current.ID {
				return fmt.Errorf("department cannot be its own parent")
			}

			parentDept, err := tx.Department.Get(ctx, parentID)
			if err != nil {
				if db.IsNotFound(err) {
					return fmt.Errorf("parent department not found")
				}
				return fmt.Errorf("failed to load parent department: %w", err)
			}

			if err := ensureNoCycle(ctx, tx, current.ID, parentDept); err != nil {
				return err
			}

			updater.SetParentID(parentID)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update department: %w", err)
	}

	return (&domain.Department{}).From(updated), nil
}

func (u *DepartmentUsecase) Delete(ctx context.Context, id string) error {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return fmt.Errorf("department id is required")
	}

	departmentID, err := uuid.Parse(trimmed)
	if err != nil {
		return fmt.Errorf("invalid department id: %w", err)
	}

	hasChildren, err := u.repo.HasChildren(ctx, departmentID)
	if err != nil {
		return fmt.Errorf("failed to check child departments: %w", err)
	}
	if hasChildren {
		return fmt.Errorf("department has child departments")
	}

	hasPositions, err := u.repo.HasPositions(ctx, departmentID)
	if err != nil {
		return fmt.Errorf("failed to check related job positions: %w", err)
	}
	if hasPositions {
		return fmt.Errorf("department is in use by job positions")
	}

	if err := u.repo.Delete(ctx, trimmed); err != nil {
		return fmt.Errorf("failed to delete department: %w", err)
	}
	return nil
}

func (u *DepartmentUsecase) GetByID(ctx context.Context, id string) (*domain.Department, error) {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return nil, fmt.Errorf("department id is required")
	}

	dept, err := u.repo.GetByID(ctx, trimmed)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, fmt.Errorf("department not found")
		}
		return nil, fmt.Errorf("failed to get department: %w", err)
	}

	return (&domain.Department{}).From(dept), nil
}

func (u *DepartmentUsecase) List(ctx context.Context, req *domain.ListDepartmentReq) (*domain.ListDepartmentResp, error) {
	items, pageInfo, err := u.repo.List(ctx, &domain.ListDepartmentRepoReq{ListDepartmentReq: req})
	if err != nil {
		return nil, fmt.Errorf("failed to list departments: %w", err)
	}

	result := make([]*domain.Department, 0, len(items))
	for _, item := range items {
		result = append(result, (&domain.Department{}).From(item))
	}

	return &domain.ListDepartmentResp{
		Items:    result,
		PageInfo: pageInfo,
	}, nil
}

func ensureNoCycle(ctx context.Context, tx *db.Tx, currentID uuid.UUID, parent *db.Department) error {
	cursor := parent
	for cursor != nil {
		if cursor.ID == currentID {
			return fmt.Errorf("department cannot be assigned to its descendant")
		}
		if cursor.ParentID == uuid.Nil {
			return nil
		}
		next, err := tx.Department.Get(ctx, cursor.ParentID)
		if err != nil {
			if db.IsNotFound(err) {
				return fmt.Errorf("parent department hierarchy is invalid")
			}
			return fmt.Errorf("failed to traverse department hierarchy: %w", err)
		}
		cursor = next
	}
	return nil
}
