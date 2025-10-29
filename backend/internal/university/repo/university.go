package repo

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/universityprofile"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/internal/university/service"
	"github.com/chaitin/WhaleHire/backend/pkg/entx"
)

type UniversityRepo struct {
	db     *db.Client
	config *config.Config
}

func NewUniversityRepo(dbClient *db.Client, cfg *config.Config) domain.UniversityRepo {
	return &UniversityRepo{
		db:     dbClient,
		config: cfg,
	}
}

func (r *UniversityRepo) Create(ctx context.Context, req *domain.CreateUniversityReq) (*db.UniversityProfile, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	// 生成向量
	vectorService := service.NewVectorService(r.config)
	vector, err := vectorService.GenerateVector(ctx, req.VectorContent)
	if err != nil {
		return nil, fmt.Errorf("向量生成失败: %w", err)
	}

	// 创建记录并一次性保存所有字段包括向量
	creator := r.db.UniversityProfile.Create().
		SetNameCn(req.NameCn).
		SetVector(&vector)

	if req.NameEn != nil {
		creator.SetNameEn(*req.NameEn)
	}
	if req.Alias != nil {
		creator.SetAlias(*req.Alias)
	}
	if req.Country != nil {
		creator.SetCountry(*req.Country)
	}
	if req.IsDoubleFirstClass != nil {
		creator.SetIsDoubleFirstClass(*req.IsDoubleFirstClass)
	}
	if req.IsProject985 != nil {
		creator.SetIsProject985(*req.IsProject985)
	}
	if req.IsProject211 != nil {
		creator.SetIsProject211(*req.IsProject211)
	}
	if req.IsQsTop100 != nil {
		creator.SetIsQsTop100(*req.IsQsTop100)
	}
	if req.RankQs != nil {
		creator.SetRankQs(*req.RankQs)
	}
	if req.OverallScore != nil {
		creator.SetOverallScore(*req.OverallScore)
	}
	if req.VectorContent != "" {
		creator.SetVectorContent(req.VectorContent)
	}
	if req.Metadata != nil {
		creator.SetMetadata(req.Metadata)
	}

	return creator.Save(ctx)
}

func (r *UniversityRepo) Update(ctx context.Context, id string, fn func(tx *db.Tx, current *db.UniversityProfile, updater *db.UniversityProfileUpdateOne) error) (*db.UniversityProfile, error) {
	universityID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid university ID: %w", err)
	}

	var result *db.UniversityProfile
	err = entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		current, err := tx.UniversityProfile.Get(ctx, universityID)
		if err != nil {
			return fmt.Errorf("failed to get university: %w", err)
		}

		updater := tx.UniversityProfile.UpdateOneID(universityID)
		if err := fn(tx, current, updater); err != nil {
			return err
		}

		result, err = updater.Save(ctx)
		return err
	})

	return result, err
}

func (r *UniversityRepo) Delete(ctx context.Context, id string) error {
	universityID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid university ID: %w", err)
	}

	return r.db.UniversityProfile.DeleteOneID(universityID).Exec(ctx)
}

func (r *UniversityRepo) GetByID(ctx context.Context, id string) (*db.UniversityProfile, error) {
	universityID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid university ID: %w", err)
	}

	return r.db.UniversityProfile.Get(ctx, universityID)
}

func (r *UniversityRepo) List(ctx context.Context, req *domain.ListUniversityRepoReq) ([]*db.UniversityProfile, *db.PageInfo, error) {
	query := r.db.UniversityProfile.Query()

	// 应用过滤条件
	if req.Country != nil {
		query = query.Where(universityprofile.CountryEQ(*req.Country))
	}
	if req.IsDoubleFirstClass != nil {
		query = query.Where(universityprofile.IsDoubleFirstClassEQ(*req.IsDoubleFirstClass))
	}
	if req.IsProject985 != nil {
		query = query.Where(universityprofile.IsProject985EQ(*req.IsProject985))
	}
	if req.IsProject211 != nil {
		query = query.Where(universityprofile.IsProject211EQ(*req.IsProject211))
	}
	if req.IsQsTop100 != nil {
		query = query.Where(universityprofile.IsQsTop100EQ(*req.IsQsTop100))
	}

	// 应用排序
	if req.OrderBy != nil {
		switch *req.OrderBy {
		case "name_cn":
			query = query.Order(db.Asc(universityprofile.FieldNameCn))
		case "name_cn_desc":
			query = query.Order(db.Desc(universityprofile.FieldNameCn))
		case "rank_qs":
			query = query.Order(db.Asc(universityprofile.FieldRankQs))
		case "rank_qs_desc":
			query = query.Order(db.Desc(universityprofile.FieldRankQs))
		case "overall_score":
			query = query.Order(db.Asc(universityprofile.FieldOverallScore))
		case "overall_score_desc":
			query = query.Order(db.Desc(universityprofile.FieldOverallScore))
		default:
			query = query.Order(db.Desc(universityprofile.FieldCreatedAt))
		}
	} else {
		query = query.Order(db.Desc(universityprofile.FieldCreatedAt))
	}

	// 使用Ent生成的分页方法
	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.Size
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}

	return query.Page(ctx, page, size)
}

func (r *UniversityRepo) SearchByName(ctx context.Context, name string, limit int) ([]*db.UniversityProfile, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("search name is required")
	}

	name = strings.TrimSpace(name)
	query := r.db.UniversityProfile.Query().Where(
		universityprofile.Or(
			universityprofile.NameCnContains(name),
			universityprofile.NameEnContains(name),
			universityprofile.AliasContains(name),
		),
	).Order(universityprofile.ByNameCn())

	if limit > 0 {
		query = query.Limit(limit)
	}

	return query.All(ctx)
}

func (r *UniversityRepo) SearchByVector(ctx context.Context, vector pgvector.Vector, limit int) ([]*db.UniversityProfile, error) {
	if limit <= 0 {
		limit = 10
	}

	// 使用原生SQL进行向量相似度搜索
	var results []*db.UniversityProfile
	err := r.db.UniversityProfile.Query().
		Modify(func(s *sql.Selector) {
			s.Select("*").
				OrderBy("vector <-> $1").
				Limit(limit)
		}).
		Scan(ctx, &results)

	return results, err
}

func (r *UniversityRepo) BatchCreate(ctx context.Context, universities []*domain.CreateUniversityReq) ([]*db.UniversityProfile, error) {
	if len(universities) == 0 {
		return nil, fmt.Errorf("universities list is empty")
	}

	// 批量生成向量并创建记录
	var results []*db.UniversityProfile
	vectorService := service.NewVectorService(r.config)

	err := entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		for _, req := range universities {

			// 生成向量
			vector, err := vectorService.GenerateVector(ctx, req.VectorContent)
			if err != nil {
				return fmt.Errorf("向量生成失败 for %s: %w", req.NameCn, err)
			}

			// 创建记录
			creator := tx.UniversityProfile.Create().
				SetNameCn(req.NameCn).
				SetVector(&vector)

			if req.NameEn != nil {
				creator.SetNameEn(*req.NameEn)
			}
			if req.Alias != nil {
				creator.SetAlias(*req.Alias)
			}
			if req.Country != nil {
				creator.SetCountry(*req.Country)
			}
			if req.IsDoubleFirstClass != nil {
				creator.SetIsDoubleFirstClass(*req.IsDoubleFirstClass)
			}
			if req.IsProject985 != nil {
				creator.SetIsProject985(*req.IsProject985)
			}
			if req.IsProject211 != nil {
				creator.SetIsProject211(*req.IsProject211)
			}
			if req.IsQsTop100 != nil {
				creator.SetIsQsTop100(*req.IsQsTop100)
			}
			if req.RankQs != nil {
				creator.SetRankQs(*req.RankQs)
			}
			if req.OverallScore != nil {
				creator.SetOverallScore(*req.OverallScore)
			}
			if req.VectorContent != "" {
				creator.SetVectorContent(req.VectorContent)
			}
			if req.Metadata != nil {
				creator.SetMetadata(req.Metadata)
			}

			result, err := creator.Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create university %s: %w", req.NameCn, err)
			}
			results = append(results, result)
		}
		return nil
	})

	return results, err
}

func (r *UniversityRepo) GetByName(ctx context.Context, name string) (*db.UniversityProfile, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("university name is required")
	}

	name = strings.TrimSpace(name)
	return r.db.UniversityProfile.Query().Where(
		universityprofile.Or(
			universityprofile.NameCnEQ(name),
			universityprofile.NameEnEQ(name),
			universityprofile.AliasEQ(name),
		),
	).First(ctx)
}

func (r *UniversityRepo) UpdateVector(ctx context.Context, id string, vector pgvector.Vector) error {
	universityID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid university ID: %w", err)
	}

	return r.db.UniversityProfile.UpdateOneID(universityID).
		SetVector(&vector).
		Exec(ctx)
}

func (r *UniversityRepo) BatchUpdateVectors(ctx context.Context, updates map[string]pgvector.Vector) error {
	if len(updates) == 0 {
		return nil
	}

	return entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		for idStr, vector := range updates {
			universityID, err := uuid.Parse(idStr)
			if err != nil {
				return fmt.Errorf("invalid university ID %s: %w", idStr, err)
			}

			err = tx.UniversityProfile.UpdateOneID(universityID).
				SetVector(&vector).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to update vector for university %s: %w", idStr, err)
			}
		}
		return nil
	})
}
