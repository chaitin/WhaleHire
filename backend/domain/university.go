package domain

import (
	"context"
	"io"

	"github.com/chaitin/WhaleHire/backend/pkg/web"
	"github.com/cloudwego/eino/schema"
	"github.com/pgvector/pgvector-go"

	"github.com/chaitin/WhaleHire/backend/db"
)

// UniversityUsecase 高校管理用例接口
type UniversityUsecase interface {
	Create(ctx context.Context, req *CreateUniversityReq) (*University, error)
	Update(ctx context.Context, req *UpdateUniversityReq) (*University, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*University, error)
	List(ctx context.Context, req *ListUniversityReq) (*ListUniversityResp, error)
	SearchByName(ctx context.Context, name string, limit int) ([]*University, error)
	SearchByVector(ctx context.Context, vector pgvector.Vector, limit int) ([]*University, error)
	ImportFromCSV(ctx context.Context, reader io.Reader) (*ImportResult, error)
	ImportFromData(ctx context.Context, universities []*CreateUniversityReq) (*ImportResult, error)
	GenerateVectors(ctx context.Context, ids []string) error
	BatchMatch(ctx context.Context, names []string) ([]*UniversityMatch, error)
}

// UniversityRepo 高校管理仓储接口
type UniversityRepo interface {
	Create(ctx context.Context, req *CreateUniversityReq) (*db.UniversityProfile, error)
	Update(ctx context.Context, id string, fn func(tx *db.Tx, current *db.UniversityProfile, updater *db.UniversityProfileUpdateOne) error) (*db.UniversityProfile, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*db.UniversityProfile, error)
	List(ctx context.Context, req *ListUniversityRepoReq) ([]*db.UniversityProfile, *db.PageInfo, error)
	SearchByName(ctx context.Context, name string, limit int) ([]*db.UniversityProfile, error)
	SearchByVector(ctx context.Context, vector pgvector.Vector, limit int) ([]*db.UniversityProfile, error)
	BatchCreate(ctx context.Context, universities []*CreateUniversityReq) ([]*db.UniversityProfile, error)
	GetByName(ctx context.Context, name string) (*db.UniversityProfile, error)
	UpdateVector(ctx context.Context, id string, vector pgvector.Vector) error
	BatchUpdateVectors(ctx context.Context, updates map[string]pgvector.Vector) error
}

// University 高校信息
type University struct {
	ID                 string                 `json:"id"`
	NameCn             string                 `json:"name_cn"`
	NameEn             *string                `json:"name_en,omitempty"`
	Alias              *string                `json:"alias,omitempty"`
	Country            *string                `json:"country,omitempty"`
	IsDoubleFirstClass bool                   `json:"is_double_first_class"`
	IsProject985       bool                   `json:"is_project_985"`
	IsProject211       bool                   `json:"is_project_211"`
	IsQsTop100         bool                   `json:"is_qs_top100"`
	RankQs             *int                   `json:"rank_qs,omitempty"`
	OverallScore       *float64               `json:"overall_score,omitempty"`
	VectorContent      string                 `json:"vector_content,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty" swaggerignore:"true"`
	CreatedAt          int64                  `json:"created_at"`
	UpdatedAt          int64                  `json:"updated_at"`
}

func (u *University) From(entity *db.UniversityProfile) *University {
	if entity == nil {
		return u
	}

	u.ID = entity.ID.String()
	u.NameCn = entity.NameCn
	if entity.NameEn != "" {
		u.NameEn = &entity.NameEn
	}
	if entity.Alias != "" {
		u.Alias = &entity.Alias
	}
	if entity.Country != "" {
		u.Country = &entity.Country
	}
	u.IsDoubleFirstClass = entity.IsDoubleFirstClass
	u.IsProject985 = entity.IsProject985
	u.IsProject211 = entity.IsProject211
	u.IsQsTop100 = entity.IsQsTop100
	if entity.RankQs > 0 {
		u.RankQs = &entity.RankQs
	}
	if entity.OverallScore > 0 {
		u.OverallScore = &entity.OverallScore
	}
	if entity.VectorContent != "" {
		u.VectorContent = entity.VectorContent
	}
	if entity.Metadata != nil {
		u.Metadata = entity.Metadata
	}
	u.CreatedAt = entity.CreatedAt.Unix()
	u.UpdatedAt = entity.UpdatedAt.Unix()
	return u
}

// CreateUniversityReq 创建高校请求
type CreateUniversityReq struct {
	NameCn             string                 `json:"name_cn" validate:"required"`
	NameEn             *string                `json:"name_en,omitempty"`
	Alias              *string                `json:"alias,omitempty"`
	Country            *string                `json:"country,omitempty"`
	IsDoubleFirstClass *bool                  `json:"is_double_first_class,omitempty"`
	IsProject985       *bool                  `json:"is_project_985,omitempty"`
	IsProject211       *bool                  `json:"is_project_211,omitempty"`
	IsQsTop100         *bool                  `json:"is_qs_top100,omitempty"`
	RankQs             *int                   `json:"rank_qs,omitempty"`
	OverallScore       *float64               `json:"overall_score,omitempty"`
	VectorContent      string                 `json:"vector_content,omitempty" swaggerignore:"true"`
	Metadata           map[string]interface{} `json:"metadata,omitempty" swaggerignore:"true"`
}

// UpdateUniversityReq 更新高校请求
type UpdateUniversityReq struct {
	ID                 string                 `json:"id" validate:"required"`
	NameCn             *string                `json:"name_cn,omitempty"`
	NameEn             *string                `json:"name_en,omitempty"`
	Alias              *string                `json:"alias,omitempty"`
	Country            *string                `json:"country,omitempty"`
	IsDoubleFirstClass *bool                  `json:"is_double_first_class,omitempty"`
	IsProject985       *bool                  `json:"is_project_985,omitempty"`
	IsProject211       *bool                  `json:"is_project_211,omitempty"`
	IsQsTop100         *bool                  `json:"is_qs_top100,omitempty"`
	RankQs             *int                   `json:"rank_qs,omitempty"`
	OverallScore       *float64               `json:"overall_score,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty" swaggerignore:"true"`
}

// ListUniversityReq 查询高校列表请求
type ListUniversityReq struct {
	web.Pagination

	Keyword            *string  `json:"keyword,omitempty" query:"keyword"`
	Country            *string  `json:"country,omitempty" query:"country"`
	IsDoubleFirstClass *bool    `json:"is_double_first_class,omitempty" query:"is_double_first_class"`
	IsProject985       *bool    `json:"is_project_985,omitempty" query:"is_project_985"`
	IsProject211       *bool    `json:"is_project_211,omitempty" query:"is_project_211"`
	IsQsTop100         *bool    `json:"is_qs_top100,omitempty" query:"is_qs_top100"`
	RankQsMin          *int     `json:"rank_qs_min,omitempty" query:"rank_qs_min"`
	RankQsMax          *int     `json:"rank_qs_max,omitempty" query:"rank_qs_max"`
	OverallScoreMin    *float64 `json:"overall_score_min,omitempty" query:"overall_score_min"`
	OverallScoreMax    *float64 `json:"overall_score_max,omitempty" query:"overall_score_max"`
	OrderBy            *string  `json:"order_by,omitempty" query:"order_by"`
}

// ListUniversityRepoReq 查询高校列表仓储请求
type ListUniversityRepoReq struct {
	*ListUniversityReq
}

// ListUniversityResp 查询高校列表响应
type ListUniversityResp struct {
	Items []*University `json:"items"`
	*db.PageInfo
}

// UniversityMatch 高校匹配结果
type UniversityMatch struct {
	Query      string      `json:"query"`
	University *University `json:"university,omitempty"`
	Score      float64     `json:"score"`
	MatchType  string      `json:"match_type"` // exact, fuzzy, vector
}

// ImportResult 导入结果
type ImportResult struct {
	Total   int      `json:"total"`
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors,omitempty"`
}

// ToDocument 将 University 转换为 schema.Document 格式，用于向量化处理
func (u *University) ToDocument() (*schema.Document, error) {
	// 构建用于向量化的内容
	content := u.NameCn
	if u.NameEn != nil && *u.NameEn != "" {
		content += " " + *u.NameEn
	}
	if u.Alias != nil && *u.Alias != "" {
		content += " " + *u.Alias
	}
	if u.Country != nil && *u.Country != "" {
		content += " " + *u.Country
	}

	// 构建元数据
	metadata := map[string]interface{}{
		"name_cn":               u.NameCn,
		"is_double_first_class": u.IsDoubleFirstClass,
		"is_project_985":        u.IsProject985,
		"is_project_211":        u.IsProject211,
		"is_qs_top100":          u.IsQsTop100,
	}

	if u.NameEn != nil {
		metadata["name_en"] = *u.NameEn
	}
	if u.Alias != nil {
		metadata["alias"] = *u.Alias
	}
	if u.Country != nil {
		metadata["country"] = *u.Country
	}
	if u.RankQs != nil {
		metadata["rank_qs"] = *u.RankQs
	}
	if u.OverallScore != nil {
		metadata["overall_score"] = *u.OverallScore
	}
	if u.Metadata != nil {
		for k, v := range u.Metadata {
			metadata[k] = v
		}
	}

	return &schema.Document{
		ID:       u.ID,
		Content:  content,
		MetaData: metadata,
	}, nil
}
