package domain

import (
	"context"
	"time"

	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/pkg/web"
	"github.com/google/uuid"
)

// WeightTemplate 权重模板领域模型
type WeightTemplate struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Weights     *DimensionWeights `json:"weights"`
	CreatedBy   uuid.UUID         `json:"created_by"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// From 方法用于从 db 实体转换
func (wt *WeightTemplate) From(dbWT *db.WeightTemplate) *WeightTemplate {
	if dbWT == nil {
		return nil
	}

	wt.ID = dbWT.ID
	wt.Name = dbWT.Name
	wt.Description = dbWT.Description
	wt.CreatedBy = dbWT.CreatedBy
	wt.CreatedAt = dbWT.CreatedAt
	wt.UpdatedAt = dbWT.UpdatedAt

	// 转换 Weights JSONB 到 DimensionWeights
	if len(dbWT.Weights) > 0 {
		wt.Weights = &DimensionWeights{}
		if v, ok := dbWT.Weights["skill"].(float64); ok {
			wt.Weights.Skill = v
		}
		if v, ok := dbWT.Weights["responsibility"].(float64); ok {
			wt.Weights.Responsibility = v
		}
		if v, ok := dbWT.Weights["experience"].(float64); ok {
			wt.Weights.Experience = v
		}
		if v, ok := dbWT.Weights["education"].(float64); ok {
			wt.Weights.Education = v
		}
		if v, ok := dbWT.Weights["industry"].(float64); ok {
			wt.Weights.Industry = v
		}
		if v, ok := dbWT.Weights["basic"].(float64); ok {
			wt.Weights.Basic = v
		}
	}

	return wt
}

// CreateWeightTemplateReq 创建权重模板请求
type CreateWeightTemplateReq struct {
	// Name 模板名称，必填，最大长度 100
	Name string `json:"name" validate:"required,max=100"`
	// Description 模板描述，可选，最大长度 500
	Description string `json:"description,omitempty" validate:"max=500"`
	// Weights 权重配置，必填，支持的维度包括: skill(技能), responsibility(职责), experience(经验),
	// education(教育), industry(行业), basic(基本信息)
	// 示例: {"skill": 0.35, "responsibility": 0.20, "experience": 0.20, "education": 0.15, "industry": 0.07, "basic": 0.03}
	Weights *DimensionWeights `json:"weights" validate:"required"`
}

// UpdateWeightTemplateReq 更新权重模板请求
type UpdateWeightTemplateReq struct {
	// Name 模板名称，必填，最大长度 100
	Name string `json:"name" validate:"required,max=100"`
	// Description 模板描述，可选，最大长度 500
	Description string `json:"description,omitempty" validate:"max=500"`
	// Weights 权重配置，必填，支持的维度包括: skill(技能), responsibility(职责), experience(经验),
	// education(教育), industry(行业), basic(基本信息)
	// 示例: {"skill": 0.35, "responsibility": 0.20, "experience": 0.20, "education": 0.15, "industry": 0.07, "basic": 0.03}
	Weights *DimensionWeights `json:"weights" validate:"required"`
}

// WeightTemplateResp 权重模板响应
type WeightTemplateResp struct {
	// ID 模板ID
	ID uuid.UUID `json:"id"`
	// Name 模板名称
	Name string `json:"name"`
	// Description 模板描述
	Description string `json:"description,omitempty"`
	// Weights 权重配置
	Weights *DimensionWeights `json:"weights"`
	// CreatedBy 创建者ID
	CreatedBy uuid.UUID `json:"created_by"`
	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// From 方法用于从领域模型转换
func (wtr *WeightTemplateResp) From(wt *WeightTemplate) *WeightTemplateResp {
	if wt == nil {
		return nil
	}

	wtr.ID = wt.ID
	wtr.Name = wt.Name
	wtr.Description = wt.Description
	wtr.Weights = wt.Weights
	wtr.CreatedBy = wt.CreatedBy
	wtr.CreatedAt = wt.CreatedAt
	wtr.UpdatedAt = wt.UpdatedAt

	return wtr
}

// ListWeightTemplatesReq 查询权重模板列表请求
type ListWeightTemplatesReq struct {
	web.Pagination

	// Name 按模板名称搜索（模糊匹配），可选
	Name string `json:"name,omitempty" query:"name"`
}

// ListWeightTemplatesResp 查询权重模板列表响应
type ListWeightTemplatesResp struct {
	// Items 模板列表
	Items []*WeightTemplateResp `json:"items"`
	// PageInfo 分页信息
	PageInfo *db.PageInfo `json:"page_info"`
}

// WeightTemplateRepo 权重模板数据访问接口
type WeightTemplateRepo interface {
	// Create 创建模板
	Create(ctx context.Context, wt *db.WeightTemplate) (*db.WeightTemplate, error)
	// GetByID 根据 ID 获取模板
	GetByID(ctx context.Context, id uuid.UUID) (*db.WeightTemplate, error)
	// List 查询模板列表（支持分页和过滤）
	List(ctx context.Context, filter *WeightTemplateFilter) ([]*db.WeightTemplate, *db.PageInfo, error)
	// Update 更新模板
	Update(ctx context.Context, id uuid.UUID, updates map[string]any) error
	// Delete 删除模板（软删除）
	Delete(ctx context.Context, id uuid.UUID) error
}

// WeightTemplateFilter 权重模板过滤条件
type WeightTemplateFilter struct {
	Name string
	Page int
	Size int
}
