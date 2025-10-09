package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/jobeducationrequirement"
	"github.com/chaitin/WhaleHire/backend/db/jobexperiencerequirement"
	"github.com/chaitin/WhaleHire/backend/db/jobindustryrequirement"
	"github.com/chaitin/WhaleHire/backend/db/jobresponsibility"
	"github.com/chaitin/WhaleHire/backend/db/jobskill"
	"github.com/chaitin/WhaleHire/backend/domain"
)

type JobProfileUsecase struct {
	repo          domain.JobProfileRepo
	skillMetaRepo domain.JobSkillMetaRepo
	logger        *slog.Logger
}

func NewJobProfileUsecase(
	repo domain.JobProfileRepo,
	skillMetaRepo domain.JobSkillMetaRepo,
	logger *slog.Logger,
) domain.JobProfileUsecase {
	return &JobProfileUsecase{
		repo:          repo,
		skillMetaRepo: skillMetaRepo,
		logger:        logger,
	}
}

func (u *JobProfileUsecase) Create(ctx context.Context, req *domain.CreateJobProfileReq) (*domain.JobProfileDetail, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("job name is required")
	}

	departmentID, err := uuid.Parse(strings.TrimSpace(req.DepartmentID))
	if err != nil {
		return nil, fmt.Errorf("invalid department id: %w", err)
	}

	job := &db.JobPosition{
		Name:         name,
		DepartmentID: departmentID,
	}

	job.Location = safeTrimPtr(req.Location)
	job.SalaryMin = req.SalaryMin
	job.SalaryMax = req.SalaryMax
	job.Description = safeTrimPtr(req.Description)

	// 设置创建者ID，需要转换为UUID
	if req.CreatedBy != nil && *req.CreatedBy != "" {
		createdByUUID, err := uuid.Parse(*req.CreatedBy)
		if err != nil {
			return nil, fmt.Errorf("invalid created_by id: %w", err)
		}
		job.CreatedBy = &createdByUUID
	}

	// 设置状态，如果未提供则使用默认值
	if req.Status != nil && *req.Status != "" {
		job.Status = consts.JobPositionStatus(*req.Status)
	} else {
		job.Status = consts.JobPositionStatusDraft
	}

	related, err := u.buildRelatedEntities(req.Skills, req.Responsibilities, req.EducationRequirements, req.ExperienceRequirements, req.IndustryRequirements)
	if err != nil {
		return nil, err
	}

	created, err := u.repo.Create(ctx, job, related)
	if err != nil {
		return nil, fmt.Errorf("failed to create job profile: %w", err)
	}

	return toJobProfileDetail(created), nil
}

func (u *JobProfileUsecase) Update(ctx context.Context, req *domain.UpdateJobProfileReq) (*domain.JobProfileDetail, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	updated, err := u.repo.Update(ctx, req.ID, func(tx *db.Tx, current *db.JobPosition, updater *db.JobPositionUpdateOne) error {
		if req.Name != nil {
			updater.SetName(strings.TrimSpace(*req.Name))
		}

		if req.DepartmentID != nil {
			departmentID, err := uuid.Parse(strings.TrimSpace(*req.DepartmentID))
			if err != nil {
				return fmt.Errorf("invalid department id: %w", err)
			}
			updater.SetDepartmentID(departmentID)
		}

		if req.Location != nil {
			if trimmed := safeTrimPtr(req.Location); trimmed != nil {
				updater.SetLocation(*trimmed)
			} else {
				updater.ClearLocation()
			}
		}
		if req.SalaryMin != nil {
			updater.SetNillableSalaryMin(req.SalaryMin)
		}
		if req.SalaryMax != nil {
			updater.SetNillableSalaryMax(req.SalaryMax)
		}
		if req.Description != nil {
			if trimmed := safeTrimPtr(req.Description); trimmed != nil {
				updater.SetDescription(*trimmed)
			} else {
				updater.ClearDescription()
			}
		}

		// 更新创建者ID
		if req.CreatedBy != nil {
			if *req.CreatedBy != "" {
				createdByUUID, err := uuid.Parse(*req.CreatedBy)
				if err != nil {
					return fmt.Errorf("invalid created_by id: %w", err)
				}
				updater.SetCreatedBy(createdByUUID)
			} else {
				updater.ClearCreatedBy()
			}
		}

		// 更新状态
		if req.Status != nil && *req.Status != "" {
			updater.SetStatus(consts.JobPositionStatus(*req.Status))
		}

		updater.SetUpdatedAt(time.Now())

		jobID := current.ID

		if req.Responsibilities != nil {
			if err := rebuildResponsibilities(ctx, tx, jobID, req.Responsibilities); err != nil {
				return err
			}
		}

		if req.Skills != nil {
			if err := rebuildSkills(ctx, tx, jobID, req.Skills); err != nil {
				return err
			}
		}

		if req.EducationRequirements != nil {
			if err := rebuildEducationRequirements(ctx, tx, jobID, req.EducationRequirements); err != nil {
				return err
			}
		}

		if req.ExperienceRequirements != nil {
			if err := rebuildExperienceRequirements(ctx, tx, jobID, req.ExperienceRequirements); err != nil {
				return err
			}
		}

		if req.IndustryRequirements != nil {
			if err := rebuildIndustryRequirements(ctx, tx, jobID, req.IndustryRequirements); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update job profile: %w", err)
	}

	return toJobProfileDetail(updated), nil
}

func (u *JobProfileUsecase) Delete(ctx context.Context, id string) error {
	if err := u.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete job profile: %w", err)
	}
	return nil
}

func (u *JobProfileUsecase) GetByID(ctx context.Context, id string) (*domain.JobProfileDetail, error) {
	profile, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get job profile: %w", err)
	}
	return toJobProfileDetail(profile), nil
}

func (u *JobProfileUsecase) List(ctx context.Context, req *domain.ListJobProfileReq) (*domain.ListJobProfileResp, error) {
	items, pageInfo, err := u.repo.List(ctx, &domain.ListJobProfileRepoReq{ListJobProfileReq: req})
	if err != nil {
		return nil, fmt.Errorf("failed to list job profiles: %w", err)
	}

	profiles := make([]*domain.JobProfile, 0, len(items))
	for _, item := range items {
		profiles = append(profiles, (&domain.JobProfile{}).From(item))
	}

	return &domain.ListJobProfileResp{
		Items:    profiles,
		PageInfo: pageInfo,
	}, nil
}

func (u *JobProfileUsecase) Search(ctx context.Context, req *domain.SearchJobProfileReq) (*domain.SearchJobProfileResp, error) {
	items, pageInfo, err := u.repo.Search(ctx, &domain.SearchJobProfileRepoReq{SearchJobProfileReq: req})
	if err != nil {
		return nil, fmt.Errorf("failed to search job profiles: %w", err)
	}

	profiles := make([]*domain.JobProfile, 0, len(items))
	for _, item := range items {
		profiles = append(profiles, (&domain.JobProfile{}).From(item))
	}

	return &domain.SearchJobProfileResp{
		Items:    profiles,
		PageInfo: pageInfo,
	}, nil
}

func (u *JobProfileUsecase) CreateSkillMeta(ctx context.Context, req *domain.CreateSkillMetaReq) (*domain.JobSkillMeta, error) {
	if req == nil || strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("skill name is required")
	}

	name := strings.TrimSpace(req.Name)

	existing, err := u.skillMetaRepo.GetByName(ctx, name)
	if err == nil {
		return (&domain.JobSkillMeta{}).From(existing), nil
	}
	if err != nil && !db.IsNotFound(err) {
		return nil, fmt.Errorf("failed to check existing skill: %w", err)
	}

	created, err := u.skillMetaRepo.Create(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create skill meta: %w", err)
	}

	return (&domain.JobSkillMeta{}).From(created), nil
}

func (u *JobProfileUsecase) ListSkillMeta(ctx context.Context, req *domain.ListSkillMetaReq) (*domain.ListSkillMetaResp, error) {
	items, pageInfo, err := u.skillMetaRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list skill meta: %w", err)
	}

	result := make([]*domain.JobSkillMeta, 0, len(items))
	for _, item := range items {
		result = append(result, (&domain.JobSkillMeta{}).From(item))
	}

	return &domain.ListSkillMetaResp{
		Items:    result,
		PageInfo: pageInfo,
	}, nil
}

func (u *JobProfileUsecase) DeleteSkillMeta(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("skill meta id is required")
	}

	// Check if skill meta is being used by any job positions
	// This is a business rule to prevent deletion of skills that are in use
	// For now, we'll allow deletion but this could be enhanced with usage checks

	err := u.skillMetaRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete skill meta: %w", err)
	}

	return nil
}

func (u *JobProfileUsecase) buildRelatedEntities(
	skills []*domain.JobSkillInput,
	responsibilities []*domain.JobResponsibilityInput,
	education []*domain.JobEducationRequirementInput,
	experience []*domain.JobExperienceRequirementInput,
	industry []*domain.JobIndustryRequirementInput,
) (*domain.JobProfileRelatedEntities, error) {
	jobSkills, err := buildJobSkills(skills)
	if err != nil {
		return nil, err
	}

	jobResponsibilities, err := buildResponsibilities(responsibilities)
	if err != nil {
		return nil, err
	}

	eduRequirements, err := buildEducationRequirements(education)
	if err != nil {
		return nil, err
	}

	expRequirements, err := buildExperienceRequirements(experience)
	if err != nil {
		return nil, err
	}

	industryRequirements, err := buildIndustryRequirements(industry)
	if err != nil {
		return nil, err
	}

	return &domain.JobProfileRelatedEntities{
		Responsibilities:       jobResponsibilities,
		Skills:                 jobSkills,
		EducationRequirements:  eduRequirements,
		ExperienceRequirements: expRequirements,
		IndustryRequirements:   industryRequirements,
	}, nil
}

func buildResponsibilities(inputs []*domain.JobResponsibilityInput) ([]*db.JobResponsibility, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	items := make([]*db.JobResponsibility, 0, len(inputs))
	for _, input := range inputs {
		if input == nil {
			continue
		}
		resp := strings.TrimSpace(input.Responsibility)
		if resp == "" {
			return nil, fmt.Errorf("responsibility content cannot be empty")
		}
		items = append(items, &db.JobResponsibility{
			Responsibility: resp,
			SortOrder:      input.SortOrder,
		})
	}

	return items, nil
}

func buildJobSkills(inputs []*domain.JobSkillInput) ([]*db.JobSkill, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	items := make([]*db.JobSkill, 0, len(inputs))
	for _, input := range inputs {
		if input == nil {
			continue
		}

		skillID, err := uuid.Parse(strings.TrimSpace(input.SkillID))
		if err != nil {
			return nil, fmt.Errorf("invalid skill id %q: %w", input.SkillID, err)
		}

		typeVal := jobskill.Type(strings.ToLower(string(input.Type)))
		if typeVal != jobskill.TypeRequired && typeVal != jobskill.TypeBonus {
			return nil, fmt.Errorf("invalid skill type %q", input.Type)
		}

		if input.Weight < 0 || input.Weight > 100 {
			return nil, fmt.Errorf("skill weight must be between 0 and 100")
		}

		items = append(items, &db.JobSkill{
			SkillID: skillID,
			Type:    typeVal,
			Weight:  input.Weight,
		})
	}

	return items, nil
}

func buildEducationRequirements(inputs []*domain.JobEducationRequirementInput) ([]*db.JobEducationRequirement, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	items := make([]*db.JobEducationRequirement, 0, len(inputs))
	for _, input := range inputs {
		if input == nil {
			continue
		}
		degree := strings.TrimSpace(input.MinDegree)
		if degree == "" {
			return nil, fmt.Errorf("education degree cannot be empty")
		}
		if input.Weight < 0 || input.Weight > 100 {
			return nil, fmt.Errorf("education weight must be between 0 and 100")
		}
		items = append(items, &db.JobEducationRequirement{
			MinDegree: degree,
			Weight:    input.Weight,
		})
	}
	return items, nil
}

func buildExperienceRequirements(inputs []*domain.JobExperienceRequirementInput) ([]*db.JobExperienceRequirement, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	items := make([]*db.JobExperienceRequirement, 0, len(inputs))
	for _, input := range inputs {
		if input == nil {
			continue
		}
		if input.Weight < 0 || input.Weight > 100 {
			return nil, fmt.Errorf("experience weight must be between 0 and 100")
		}
		items = append(items, &db.JobExperienceRequirement{
			MinYears:   input.MinYears,
			IdealYears: input.IdealYears,
			Weight:     input.Weight,
		})
	}

	return items, nil
}

func buildIndustryRequirements(inputs []*domain.JobIndustryRequirementInput) ([]*db.JobIndustryRequirement, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	items := make([]*db.JobIndustryRequirement, 0, len(inputs))
	for _, input := range inputs {
		if input == nil {
			continue
		}
		industry := strings.TrimSpace(input.Industry)
		if industry == "" {
			return nil, fmt.Errorf("industry cannot be empty")
		}
		if input.Weight < 0 || input.Weight > 100 {
			return nil, fmt.Errorf("industry requirement weight must be between 0 and 100")
		}
		items = append(items, &db.JobIndustryRequirement{
			Industry: industry,
			Weight:   input.Weight,
		})
		if trimmed := safeTrimPtr(input.CompanyName); trimmed != nil {
			items[len(items)-1].CompanyName = trimmed
		}
	}

	return items, nil
}

func rebuildResponsibilities(ctx context.Context, tx *db.Tx, jobID uuid.UUID, inputs []*domain.JobResponsibilityInput) error {
	if _, err := tx.JobResponsibility.Delete().Where(jobresponsibility.JobID(jobID)).Exec(ctx); err != nil {
		return fmt.Errorf("failed to clear responsibilities: %w", err)
	}

	items, err := buildResponsibilities(inputs)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return nil
	}

	builders := make([]*db.JobResponsibilityCreate, 0, len(items))
	for _, item := range items {
		builder := tx.JobResponsibility.Create().
			SetJobID(jobID).
			SetResponsibility(item.Responsibility).
			SetSortOrder(item.SortOrder)
		builders = append(builders, builder)
	}

	if _, err := tx.JobResponsibility.CreateBulk(builders...).Save(ctx); err != nil {
		return fmt.Errorf("failed to recreate responsibilities: %w", err)
	}

	return nil
}

func rebuildSkills(ctx context.Context, tx *db.Tx, jobID uuid.UUID, inputs []*domain.JobSkillInput) error {
	if _, err := tx.JobSkill.Delete().Where(jobskill.JobID(jobID)).Exec(ctx); err != nil {
		return fmt.Errorf("failed to clear job skills: %w", err)
	}

	items, err := buildJobSkills(inputs)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return nil
	}

	builders := make([]*db.JobSkillCreate, 0, len(items))
	for _, item := range items {
		builder := tx.JobSkill.Create().
			SetJobID(jobID).
			SetSkillID(item.SkillID).
			SetType(item.Type).
			SetWeight(item.Weight)
		builders = append(builders, builder)
	}

	if _, err := tx.JobSkill.CreateBulk(builders...).Save(ctx); err != nil {
		return fmt.Errorf("failed to recreate job skills: %w", err)
	}

	return nil
}

func rebuildEducationRequirements(ctx context.Context, tx *db.Tx, jobID uuid.UUID, inputs []*domain.JobEducationRequirementInput) error {
	if _, err := tx.JobEducationRequirement.Delete().Where(jobeducationrequirement.JobID(jobID)).Exec(ctx); err != nil {
		return fmt.Errorf("failed to clear education requirements: %w", err)
	}

	items, err := buildEducationRequirements(inputs)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return nil
	}

	builders := make([]*db.JobEducationRequirementCreate, 0, len(items))
	for _, item := range items {
		builder := tx.JobEducationRequirement.Create().
			SetJobID(jobID).
			SetMinDegree(item.MinDegree).
			SetWeight(item.Weight)
		builders = append(builders, builder)
	}

	if _, err := tx.JobEducationRequirement.CreateBulk(builders...).Save(ctx); err != nil {
		return fmt.Errorf("failed to recreate education requirements: %w", err)
	}

	return nil
}

func rebuildExperienceRequirements(ctx context.Context, tx *db.Tx, jobID uuid.UUID, inputs []*domain.JobExperienceRequirementInput) error {
	if _, err := tx.JobExperienceRequirement.Delete().Where(jobexperiencerequirement.JobID(jobID)).Exec(ctx); err != nil {
		return fmt.Errorf("failed to clear experience requirements: %w", err)
	}

	items, err := buildExperienceRequirements(inputs)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return nil
	}

	builders := make([]*db.JobExperienceRequirementCreate, 0, len(items))
	for _, item := range items {
		builder := tx.JobExperienceRequirement.Create().
			SetJobID(jobID).
			SetMinYears(item.MinYears).
			SetIdealYears(item.IdealYears).
			SetWeight(item.Weight)
		builders = append(builders, builder)
	}

	if _, err := tx.JobExperienceRequirement.CreateBulk(builders...).Save(ctx); err != nil {
		return fmt.Errorf("failed to recreate experience requirements: %w", err)
	}

	return nil
}

func rebuildIndustryRequirements(ctx context.Context, tx *db.Tx, jobID uuid.UUID, inputs []*domain.JobIndustryRequirementInput) error {
	if _, err := tx.JobIndustryRequirement.Delete().Where(jobindustryrequirement.JobID(jobID)).Exec(ctx); err != nil {
		return fmt.Errorf("failed to clear industry requirements: %w", err)
	}

	items, err := buildIndustryRequirements(inputs)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return nil
	}

	builders := make([]*db.JobIndustryRequirementCreate, 0, len(items))
	for _, item := range items {
		builder := tx.JobIndustryRequirement.Create().
			SetJobID(jobID).
			SetIndustry(item.Industry).
			SetWeight(item.Weight).
			SetNillableCompanyName(item.CompanyName)
		builders = append(builders, builder)
	}

	if _, err := tx.JobIndustryRequirement.CreateBulk(builders...).Save(ctx); err != nil {
		return fmt.Errorf("failed to recreate industry requirements: %w", err)
	}

	return nil
}

func toJobProfileDetail(entity *db.JobPosition) *domain.JobProfileDetail {
	if entity == nil {
		return nil
	}

	profile := (&domain.JobProfile{}).From(entity)

	responsibilities := make([]*domain.JobResponsibility, 0)
	if list := entity.Edges.Responsibilities; len(list) > 0 {
		responsibilities = make([]*domain.JobResponsibility, 0, len(list))
		for _, item := range list {
			responsibilities = append(responsibilities, &domain.JobResponsibility{
				ID:             item.ID.String(),
				JobID:          item.JobID.String(),
				Responsibility: item.Responsibility,
				SortOrder:      item.SortOrder,
			})
		}
	}

	skills := make([]*domain.JobSkill, 0)
	if list := entity.Edges.Skills; len(list) > 0 {
		skills = make([]*domain.JobSkill, 0, len(list))
		for _, item := range list {
			name := ""
			if item.Edges.Skill != nil {
				name = item.Edges.Skill.Name
			}
			skills = append(skills, &domain.JobSkill{
				ID:      item.ID.String(),
				JobID:   item.JobID.String(),
				SkillID: item.SkillID.String(),
				Skill:   name,
				Type:    domain.SkillType(item.Type),
				Weight:  item.Weight,
			})
		}
	}

	education := make([]*domain.JobEducationRequirement, 0)
	if list := entity.Edges.EducationRequirements; len(list) > 0 {
		education = make([]*domain.JobEducationRequirement, 0, len(list))
		for _, item := range list {
			education = append(education, &domain.JobEducationRequirement{
				ID:        item.ID.String(),
				JobID:     item.JobID.String(),
				MinDegree: item.MinDegree,
				Weight:    item.Weight,
			})
		}
	}

	experience := make([]*domain.JobExperienceRequirement, 0)
	if list := entity.Edges.ExperienceRequirements; len(list) > 0 {
		experience = make([]*domain.JobExperienceRequirement, 0, len(list))
		for _, item := range list {
			experience = append(experience, &domain.JobExperienceRequirement{
				ID:         item.ID.String(),
				JobID:      item.JobID.String(),
				MinYears:   item.MinYears,
				IdealYears: item.IdealYears,
				Weight:     item.Weight,
			})
		}
	}

	industry := make([]*domain.JobIndustryRequirement, 0)
	if list := entity.Edges.IndustryRequirements; len(list) > 0 {
		industry = make([]*domain.JobIndustryRequirement, 0, len(list))
		for _, item := range list {
			industry = append(industry, &domain.JobIndustryRequirement{
				ID:          item.ID.String(),
				JobID:       item.JobID.String(),
				Industry:    item.Industry,
				CompanyName: copyStringPtr(item.CompanyName),
				Weight:      item.Weight,
			})
		}
	}

	return &domain.JobProfileDetail{
		JobProfile:             profile,
		Responsibilities:       responsibilities,
		Skills:                 skills,
		EducationRequirements:  education,
		ExperienceRequirements: experience,
		IndustryRequirements:   industry,
	}
}

func safeTrimPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	copy := trimmed
	return &copy
}

func copyStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
