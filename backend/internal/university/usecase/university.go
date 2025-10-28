package usecase

import (
	"context"
	"fmt"
	"io"

	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/errcode"
	"github.com/chaitin/WhaleHire/backend/pkg/csvparser"
	"github.com/pgvector/pgvector-go"
)

type UniversityUsecase struct {
	repo domain.UniversityRepo
}

func NewUniversityUsecase(repo domain.UniversityRepo) domain.UniversityUsecase {
	return &UniversityUsecase{
		repo: repo,
	}
}

func (u *UniversityUsecase) Create(ctx context.Context, req *domain.CreateUniversityReq) (*domain.University, error) {
	// 构建元数据，将请求数据转换为 metadata
	metadata := map[string]interface{}{
		"name_cn": req.NameCn,
	}

	if req.NameEn != nil {
		metadata["name_en"] = *req.NameEn
	}
	if req.Alias != nil {
		metadata["alias"] = *req.Alias
	}
	if req.Country != nil {
		metadata["country"] = *req.Country
	}
	if req.IsDoubleFirstClass != nil {
		metadata["is_double_first_class"] = *req.IsDoubleFirstClass
	}
	if req.IsProject985 != nil {
		metadata["is_project_985"] = *req.IsProject985
	}
	if req.IsProject211 != nil {
		metadata["is_project_211"] = *req.IsProject211
	}
	if req.IsQsTop100 != nil {
		metadata["is_qs_top100"] = *req.IsQsTop100
	}
	if req.RankQs != nil {
		metadata["rank_qs"] = *req.RankQs
	}
	if req.OverallScore != nil {
		metadata["overall_score"] = *req.OverallScore
	}
	// 如果用户提供了额外的 metadata，合并进去
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			metadata[k] = v
		}
	}
	req.Metadata = metadata

	VectorContent := req.NameCn
	if req.NameEn != nil {
		VectorContent += " " + *req.NameEn
	}
	if req.Alias != nil {
		VectorContent += " " + *req.Alias
	}
	if req.Country != nil {
		VectorContent += " " + *req.Country
	}

	req.VectorContent = VectorContent

	entity, err := u.repo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create university: %w", err)
	}

	return (&domain.University{}).From(entity), nil
}

func (u *UniversityUsecase) Update(ctx context.Context, req *domain.UpdateUniversityReq) (*domain.University, error) {
	var result *db.UniversityProfile
	var err error

	result, err = u.repo.Update(ctx, req.ID, func(tx *db.Tx, current *db.UniversityProfile, updater *db.UniversityProfileUpdateOne) error {
		if req.NameCn != nil {
			updater.SetNameCn(*req.NameCn)
		}
		if req.NameEn != nil {
			updater.SetNameEn(*req.NameEn)
		}
		if req.Alias != nil {
			updater.SetAlias(*req.Alias)
		}
		if req.Country != nil {
			updater.SetCountry(*req.Country)
		}
		if req.IsDoubleFirstClass != nil {
			updater.SetIsDoubleFirstClass(*req.IsDoubleFirstClass)
		}
		if req.IsProject985 != nil {
			updater.SetIsProject985(*req.IsProject985)
		}
		if req.IsProject211 != nil {
			updater.SetIsProject211(*req.IsProject211)
		}
		if req.IsQsTop100 != nil {
			updater.SetIsQsTop100(*req.IsQsTop100)
		}
		if req.RankQs != nil {
			updater.SetRankQs(*req.RankQs)
		}
		if req.OverallScore != nil {
			updater.SetOverallScore(*req.OverallScore)
		}
		if req.Metadata != nil {
			updater.SetMetadata(req.Metadata)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update university: %w", err)
	}

	return (&domain.University{}).From(result), nil
}

func (u *UniversityUsecase) Delete(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}

func (u *UniversityUsecase) GetByID(ctx context.Context, id string) (*domain.University, error) {
	entity, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get university: %w", err)
	}

	return (&domain.University{}).From(entity), nil
}

func (u *UniversityUsecase) List(ctx context.Context, req *domain.ListUniversityReq) (*domain.ListUniversityResp, error) {
	repoReq := &domain.ListUniversityRepoReq{
		ListUniversityReq: req,
	}

	entities, pageInfo, err := u.repo.List(ctx, repoReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list universities: %w", err)
	}

	universities := make([]*domain.University, len(entities))
	for i, entity := range entities {
		universities[i] = (&domain.University{}).From(entity)
	}

	return &domain.ListUniversityResp{
		Items:    universities,
		PageInfo: pageInfo,
	}, nil
}

func (u *UniversityUsecase) SearchByName(ctx context.Context, name string, limit int) ([]*domain.University, error) {
	entities, err := u.repo.SearchByName(ctx, name, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search universities by name: %w", err)
	}

	universities := make([]*domain.University, len(entities))
	for i, entity := range entities {
		universities[i] = (&domain.University{}).From(entity)
	}

	return universities, nil
}

func (u *UniversityUsecase) SearchByVector(ctx context.Context, vector pgvector.Vector, limit int) ([]*domain.University, error) {
	entities, err := u.repo.SearchByVector(ctx, vector, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search universities by vector: %w", err)
	}

	universities := make([]*domain.University, len(entities))
	for i, entity := range entities {
		universities[i] = (&domain.University{}).From(entity)
	}

	return universities, nil
}

// ImportFromCSV 从CSV数据导入大学信息
func (u *UniversityUsecase) ImportFromCSV(ctx context.Context, reader io.Reader) (*domain.ImportResult, error) {
	parser := csvparser.NewUniversityCSVParser()

	universities, err := parser.ParseCSV(reader)
	if err != nil {
		return nil, errcode.ErrCSVParseError.Wrap(err)
	}

	if len(universities) == 0 {
		return &domain.ImportResult{
			Total:   0,
			Success: 0,
			Failed:  0,
		}, nil
	}

	// 调用现有的ImportFromData方法
	return u.ImportFromData(ctx, universities)
}

func (u *UniversityUsecase) ImportFromData(ctx context.Context, universities []*domain.CreateUniversityReq) (*domain.ImportResult, error) {
	if len(universities) == 0 {
		return &domain.ImportResult{
			Total:   0,
			Success: 0,
			Failed:  0,
		}, nil
	}

	result := &domain.ImportResult{
		Total:  len(universities),
		Errors: make([]string, 0),
	}

	repoReqs := make([]*domain.CreateUniversityReq, len(universities))
	for i, req := range universities {

		metadata := map[string]interface{}{
			"name_cn": req.NameCn,
		}

		if req.NameEn != nil {
			metadata["name_en"] = *req.NameEn
		}
		if req.Alias != nil {
			metadata["alias"] = *req.Alias
		}
		if req.Country != nil {
			metadata["country"] = *req.Country
		}
		if req.IsDoubleFirstClass != nil {
			metadata["is_double_first_class"] = *req.IsDoubleFirstClass
		}
		if req.IsProject985 != nil {
			metadata["is_project_985"] = *req.IsProject985
		}
		if req.IsProject211 != nil {
			metadata["is_project_211"] = *req.IsProject211
		}
		if req.IsQsTop100 != nil {
			metadata["is_qs_top100"] = *req.IsQsTop100
		}
		if req.RankQs != nil {
			metadata["rank_qs"] = *req.RankQs
		}
		if req.OverallScore != nil {
			metadata["overall_score"] = *req.OverallScore
		}
		// 如果用户提供了额外的 metadata，合并进去
		if req.Metadata != nil {
			for k, v := range req.Metadata {
				metadata[k] = v
			}
		}
		req.Metadata = metadata

		VectorContent := req.NameCn
		if req.NameEn != nil {
			VectorContent += " " + *req.NameEn
		}
		if req.Alias != nil {
			VectorContent += " " + *req.Alias
		}
		if req.Country != nil {
			VectorContent += " " + *req.Country
		}

		repoReqs[i] = &domain.CreateUniversityReq{
			NameCn:             req.NameCn,
			NameEn:             req.NameEn,
			Alias:              req.Alias,
			Country:            req.Country,
			IsDoubleFirstClass: req.IsDoubleFirstClass,
			IsProject985:       req.IsProject985,
			IsProject211:       req.IsProject211,
			IsQsTop100:         req.IsQsTop100,
			RankQs:             req.RankQs,
			OverallScore:       req.OverallScore,
			VectorContent:      VectorContent,
			Metadata:           req.Metadata,
		}
	}

	_, err := u.repo.BatchCreate(ctx, repoReqs)
	if err != nil {
		result.Failed = result.Total
		result.Errors = append(result.Errors, err.Error())
	} else {
		result.Success = result.Total
	}

	return result, nil
}

func (u *UniversityUsecase) GenerateVectors(ctx context.Context, ids []string) error {
	// TODO: 实现向量生成逻辑
	return fmt.Errorf("vector generation not implemented yet")
}

func (u *UniversityUsecase) BatchMatch(ctx context.Context, names []string) ([]*domain.UniversityMatch, error) {
	matches := make([]*domain.UniversityMatch, len(names))

	for i, name := range names {
		// 首先尝试精确匹配
		entity, err := u.repo.GetByName(ctx, name)
		if err == nil && entity != nil {
			matches[i] = &domain.UniversityMatch{
				Query:      name,
				University: (&domain.University{}).From(entity),
				Score:      1.0,
				MatchType:  "exact",
			}
			continue
		}

		// 如果精确匹配失败，尝试模糊搜索
		entities, err := u.repo.SearchByName(ctx, name, 1)
		if err == nil && len(entities) > 0 {
			matches[i] = &domain.UniversityMatch{
				Query:      name,
				University: (&domain.University{}).From(entities[0]),
				Score:      0.8, // 模糊匹配分数
				MatchType:  "fuzzy",
			}
			continue
		}

		// 如果都没有匹配到，返回空匹配
		matches[i] = &domain.UniversityMatch{
			Query:     name,
			Score:     0.0,
			MatchType: "none",
		}
	}

	return matches, nil
}
