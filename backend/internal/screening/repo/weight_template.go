package repo

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/weighttemplate"
	"github.com/chaitin/WhaleHire/backend/domain"
)

// WeightTemplateRepo 权重模板仓储实现
type WeightTemplateRepo struct {
	db *db.Client
}

// NewWeightTemplateRepo 构建仓储
func NewWeightTemplateRepo(client *db.Client) domain.WeightTemplateRepo {
	return &WeightTemplateRepo{db: client}
}

// Create 创建模板
func (r *WeightTemplateRepo) Create(ctx context.Context, wt *db.WeightTemplate) (*db.WeightTemplate, error) {
	if wt == nil {
		return nil, fmt.Errorf("weight template is nil")
	}

	entity, err := r.db.WeightTemplate.Create().
		SetID(wt.ID).
		SetName(wt.Name).
		SetDescription(wt.Description).
		SetWeights(wt.Weights).
		SetCreatedBy(wt.CreatedBy).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("create weight template failed: %w", err)
	}

	return entity, nil
}

// GetByID 根据 ID 获取模板
func (r *WeightTemplateRepo) GetByID(ctx context.Context, id uuid.UUID) (*db.WeightTemplate, error) {
	template, err := r.db.WeightTemplate.Query().
		Where(weighttemplate.ID(id)).
		WithCreator(). // 预加载创建者信息
		Only(ctx)

	if err != nil {
		return nil, err
	}

	return template, nil
}

// List 查询模板列表（支持分页和过滤）
func (r *WeightTemplateRepo) List(ctx context.Context, filter *domain.WeightTemplateFilter) ([]*db.WeightTemplate, *db.PageInfo, error) {
	query := r.db.WeightTemplate.Query().
		WithCreator() // 预加载创建者信息

	if filter != nil {
		if filter.Name != "" {
			query = query.Where(weighttemplate.NameContains(filter.Name))
		}
	}

	// 排序：按创建时间倒序
	query = query.Order(weighttemplate.ByCreatedAt(sql.OrderDesc()))

	// 使用 db 层的分页方法
	page := 1
	pageSize := 10
	if filter != nil {
		if filter.Page > 0 {
			page = filter.Page
		}
		if filter.Size > 0 {
			pageSize = filter.Size
		}
	}

	items, pageInfo, err := query.Page(ctx, page, pageSize)
	if err != nil {
		return nil, nil, fmt.Errorf("list weight templates failed: %w", err)
	}

	return items, pageInfo, nil
}

// Update 更新模板
func (r *WeightTemplateRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	builder := r.db.WeightTemplate.UpdateOneID(id)

	// 应用更新字段
	for key, value := range updates {
		switch key {
		case "name":
			if v, ok := value.(string); ok {
				builder = builder.SetName(v)
			}
		case "description":
			if v, ok := value.(string); ok {
				builder = builder.SetDescription(v)
			}
		case "weights":
			if v, ok := value.(map[string]interface{}); ok {
				builder = builder.SetWeights(v)
			}
		}
	}

	if err := builder.Exec(ctx); err != nil {
		return fmt.Errorf("update weight template failed: %w", err)
	}

	return nil
}

// Delete 删除模板（软删除）
func (r *WeightTemplateRepo) Delete(ctx context.Context, id uuid.UUID) error {
	// SoftDeleteMixin 会自动处理软删除
	err := r.db.WeightTemplate.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete weight template failed: %w", err)
	}

	return nil
}
