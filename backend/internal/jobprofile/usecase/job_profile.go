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
	"github.com/chaitin/WhaleHire/backend/errcode"
	"github.com/chaitin/WhaleHire/backend/internal/jobprofile/service"
)

type JobProfileUsecase struct {
	repo          domain.JobProfileRepo
	skillMetaRepo domain.JobSkillMetaRepo
	parserService *service.JobProfileParserService
	logger        *slog.Logger
}

func NewJobProfileUsecase(
	repo domain.JobProfileRepo,
	skillMetaRepo domain.JobSkillMetaRepo,
	parserService *service.JobProfileParserService,
	logger *slog.Logger,
) domain.JobProfileUsecase {
	return &JobProfileUsecase{
		repo:          repo,
		skillMetaRepo: skillMetaRepo,
		parserService: parserService,
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
		createdByUUID, errParse := uuid.Parse(*req.CreatedBy)
		if errParse != nil {
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

	// 设置工作类型
	if req.WorkType != nil && *req.WorkType != "" {
		job.WorkType = consts.JobWorkType(*req.WorkType)
	}

	// 处理技能输入，自动创建不存在的技能
	processedSkills, err := u.processSkillInputs(ctx, req.Skills)
	if err != nil {
		return nil, fmt.Errorf("failed to process skills: %w", err)
	}

	related, err := u.buildRelatedEntities(processedSkills, req.Responsibilities, req.EducationRequirements, req.ExperienceRequirements, req.IndustryRequirements)
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
			if err := upsertResponsibilities(ctx, tx, jobID, req.Responsibilities); err != nil {
				return err
			}
		}

		if req.Skills != nil {
			// 处理技能输入，自动创建不存在的技能
			processedSkills, err := u.processSkillInputs(ctx, req.Skills)
			if err != nil {
				return fmt.Errorf("failed to process skills: %w", err)
			}

			if err := upsertSkills(ctx, tx, jobID, processedSkills); err != nil {
				return err
			}
		}

		if req.EducationRequirements != nil {
			if err := upsertEducationRequirements(ctx, tx, jobID, req.EducationRequirements); err != nil {
				return err
			}
		}

		if req.ExperienceRequirements != nil {
			if err := upsertExperienceRequirements(ctx, tx, jobID, req.ExperienceRequirements); err != nil {
				return err
			}
		}

		if req.IndustryRequirements != nil {
			if err := upsertIndustryRequirements(ctx, tx, jobID, req.IndustryRequirements); err != nil {
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
	// 检查岗位是否有关联的简历申请
	hasRelated, err := u.repo.HasRelatedResumes(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check related resumes: %w", err)
	}

	if hasRelated {
		return errcode.ErrJobProfileHasResumes
	}

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

func (u *JobProfileUsecase) GetByIDs(ctx context.Context, ids []string) ([]*domain.JobProfile, error) {
	if len(ids) == 0 {
		return []*domain.JobProfile{}, nil
	}

	profiles, err := u.repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get job profiles by IDs: %w", err)
	}

	result := make([]*domain.JobProfile, 0, len(profiles))
	for _, profile := range profiles {
		result = append(result, (&domain.JobProfile{}).From(profile))
	}

	return result, nil
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

		// SkillID现在是可选的，如果为空则跳过验证，后续在服务层处理
		if input.SkillID == nil || strings.TrimSpace(*input.SkillID) == "" {
			// 对于没有SkillID的情况，我们暂时跳过，在上层处理
			continue
		}

		skillID, err := uuid.Parse(strings.TrimSpace(*input.SkillID))
		if err != nil {
			return nil, fmt.Errorf("invalid skill id %q: %w", *input.SkillID, err)
		}

		skillType := strings.TrimSpace(input.Type)
		if skillType == "" {
			return nil, fmt.Errorf("skill type cannot be empty")
		}

		items = append(items, &db.JobSkill{
			SkillID: skillID,
			Type:    consts.JobSkillType(skillType),
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
		educationType := strings.TrimSpace(input.EducationType)
		if educationType == "" {
			return nil, fmt.Errorf("education type cannot be empty")
		}

		items = append(items, &db.JobEducationRequirement{
			EducationType: consts.JobEducationType(educationType),
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

		experienceType := strings.TrimSpace(input.ExperienceType)
		if experienceType == "" {
			return nil, fmt.Errorf("experience type cannot be empty")
		}

		items = append(items, &db.JobExperienceRequirement{
			ExperienceType: experienceType,
			MinYears:       input.MinYears,
			IdealYears:     input.IdealYears,
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

		var industry *string
		if trimmed := strings.TrimSpace(input.Industry); trimmed != "" {
			industry = &trimmed
		}

		items = append(items, &db.JobIndustryRequirement{
			Industry:    industry,
			CompanyName: safeTrimPtr(input.CompanyName),
		})
	}

	return items, nil
}

func upsertResponsibilities(ctx context.Context, tx *db.Tx, jobID uuid.UUID, inputs []*domain.JobResponsibilityInput) error {
	// 获取现有记录
	existing, err := tx.JobResponsibility.Query().Where(jobresponsibility.JobID(jobID)).All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query existing responsibilities: %w", err)
	}

	// 创建现有记录的ID映射
	existingMap := make(map[string]*db.JobResponsibility)
	for _, item := range existing {
		existingMap[item.ID.String()] = item
	}

	// 分类处理输入数据
	var toUpdate []*domain.JobResponsibilityInput
	var toCreate []*domain.JobResponsibilityInput
	inputIDSet := make(map[string]bool)

	for _, input := range inputs {
		if input.ID != nil && *input.ID != "" {
			// 有ID的记录，准备更新
			toUpdate = append(toUpdate, input)
			inputIDSet[*input.ID] = true
		} else {
			// 无ID的记录，准备创建
			toCreate = append(toCreate, input)
		}
	}

	// 找出需要删除的记录（现有记录中不在输入列表中的）
	var toDeleteIDs []uuid.UUID
	for id := range existingMap {
		if !inputIDSet[id] {
			if existingID, err := uuid.Parse(id); err == nil {
				toDeleteIDs = append(toDeleteIDs, existingID)
			}
		}
	}

	// 批量删除不需要的记录
	if len(toDeleteIDs) > 0 {
		if _, err := tx.JobResponsibility.Delete().Where(jobresponsibility.IDIn(toDeleteIDs...)).Exec(ctx); err != nil {
			return fmt.Errorf("failed to delete responsibilities: %w", err)
		}
	}

	// 批量更新现有记录
	for _, input := range toUpdate {
		if input.ID == nil || *input.ID == "" {
			continue
		}

		id, err := uuid.Parse(*input.ID)
		if err != nil {
			return fmt.Errorf("invalid responsibility ID: %w", err)
		}

		updater := tx.JobResponsibility.UpdateOneID(id).
			SetResponsibility(input.Responsibility)

		if _, err := updater.Save(ctx); err != nil {
			return fmt.Errorf("failed to update responsibility: %w", err)
		}
	}

	// 批量创建新记录
	if len(toCreate) > 0 {
		builders := make([]*db.JobResponsibilityCreate, 0, len(toCreate))
		for _, input := range toCreate {
			builder := tx.JobResponsibility.Create().
				SetJobID(jobID).
				SetResponsibility(input.Responsibility)

			builders = append(builders, builder)
		}

		if _, err := tx.JobResponsibility.CreateBulk(builders...).Save(ctx); err != nil {
			return fmt.Errorf("failed to create responsibilities: %w", err)
		}
	}

	return nil
}

// ensureSkillExists 确保技能存在，如果不存在则创建
func (u *JobProfileUsecase) ensureSkillExists(ctx context.Context, skillName string) (string, error) {
	if skillName == "" {
		return "", fmt.Errorf("skill name cannot be empty")
	}

	// 先尝试根据名称查找现有技能
	existing, err := u.skillMetaRepo.GetByName(ctx, skillName)
	if err == nil {
		return existing.ID.String(), nil
	}
	if err != nil && !db.IsNotFound(err) {
		return "", fmt.Errorf("failed to check existing skill: %w", err)
	}

	// 如果不存在，则创建新技能
	created, err := u.skillMetaRepo.Create(ctx, skillName)
	if err != nil {
		return "", fmt.Errorf("failed to create skill meta: %w", err)
	}

	return created.ID.String(), nil
}

// processSkillInputs 处理技能输入，自动创建不存在的技能
func (u *JobProfileUsecase) processSkillInputs(ctx context.Context, inputs []*domain.JobSkillInput) ([]*domain.JobSkillInput, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	processed := make([]*domain.JobSkillInput, 0, len(inputs))
	for _, input := range inputs {
		if input == nil {
			continue
		}

		// 如果没有SkillID，根据SkillName自动创建或查找
		if input.SkillID == nil || strings.TrimSpace(*input.SkillID) == "" {
			skillID, err := u.ensureSkillExists(ctx, strings.TrimSpace(input.SkillName))
			if err != nil {
				return nil, fmt.Errorf("failed to ensure skill exists for %q: %w", input.SkillName, err)
			}

			// 创建新的输入对象，设置SkillID
			processedInput := &domain.JobSkillInput{
				ID:        input.ID,
				SkillID:   &skillID,
				SkillName: input.SkillName,
				Type:      input.Type,
			}
			processed = append(processed, processedInput)
		} else {
			// 已有SkillID，直接使用
			processed = append(processed, input)
		}
	}

	return processed, nil
}

func upsertSkills(ctx context.Context, tx *db.Tx, jobID uuid.UUID, inputs []*domain.JobSkillInput) error {
	// 获取现有记录
	existing, err := tx.JobSkill.Query().Where(jobskill.JobID(jobID)).All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query existing skills: %w", err)
	}

	// 创建现有记录的ID映射
	existingMap := make(map[string]*db.JobSkill)
	for _, item := range existing {
		existingMap[item.ID.String()] = item
	}

	// 分类处理输入数据
	var toUpdate []*domain.JobSkillInput
	var toCreate []*domain.JobSkillInput
	inputIDSet := make(map[string]bool)

	for _, input := range inputs {
		if input.ID != nil && *input.ID != "" {
			// 有ID的记录，准备更新
			toUpdate = append(toUpdate, input)
			inputIDSet[*input.ID] = true
		} else {
			// 无ID的记录，准备创建
			toCreate = append(toCreate, input)
		}
	}

	// 找出需要删除的记录（现有记录中不在输入列表中的）
	var toDeleteIDs []uuid.UUID
	for id := range existingMap {
		if !inputIDSet[id] {
			if existingID, err := uuid.Parse(id); err == nil {
				toDeleteIDs = append(toDeleteIDs, existingID)
			}
		}
	}

	// 批量删除不需要的记录
	if len(toDeleteIDs) > 0 {
		if _, err := tx.JobSkill.Delete().Where(jobskill.IDIn(toDeleteIDs...)).Exec(ctx); err != nil {
			return fmt.Errorf("failed to delete skills: %w", err)
		}
	}

	// 批量更新现有记录
	for _, input := range toUpdate {
		if input.ID == nil || *input.ID == "" {
			continue
		}

		id, err := uuid.Parse(*input.ID)
		if err != nil {
			return fmt.Errorf("invalid skill ID: %w", err)
		}

		// 处理SkillID为可选的情况
		if input.SkillID == nil || strings.TrimSpace(*input.SkillID) == "" {
			return fmt.Errorf("skill_id is required for update operation")
		}

		skillID, err := uuid.Parse(strings.TrimSpace(*input.SkillID))
		if err != nil {
			return fmt.Errorf("invalid skill meta ID: %w", err)
		}

		updater := tx.JobSkill.UpdateOneID(id).
			SetSkillID(skillID).
			SetType(consts.JobSkillType(input.Type))

		if _, err := updater.Save(ctx); err != nil {
			return fmt.Errorf("failed to update skill: %w", err)
		}
	}

	// 批量创建新记录
	if len(toCreate) > 0 {
		builders := make([]*db.JobSkillCreate, 0, len(toCreate))
		for _, input := range toCreate {
			// 处理SkillID为可选的情况
			if input.SkillID == nil || strings.TrimSpace(*input.SkillID) == "" {
				return fmt.Errorf("skill_id is required for create operation")
			}

			skillID, err := uuid.Parse(strings.TrimSpace(*input.SkillID))
			if err != nil {
				return fmt.Errorf("invalid skill meta ID: %w", err)
			}

			builder := tx.JobSkill.Create().
				SetJobID(jobID).
				SetSkillID(skillID).
				SetType(consts.JobSkillType(input.Type))

			builders = append(builders, builder)
		}

		if _, err := tx.JobSkill.CreateBulk(builders...).Save(ctx); err != nil {
			return fmt.Errorf("failed to create skills: %w", err)
		}
	}

	return nil
}

func upsertEducationRequirements(ctx context.Context, tx *db.Tx, jobID uuid.UUID, inputs []*domain.JobEducationRequirementInput) error {
	// 获取现有记录
	existing, err := tx.JobEducationRequirement.Query().Where(jobeducationrequirement.JobID(jobID)).All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query existing education requirements: %w", err)
	}

	// 创建现有记录的ID映射
	existingMap := make(map[string]*db.JobEducationRequirement)
	for _, item := range existing {
		existingMap[item.ID.String()] = item
	}

	// 分类处理输入数据
	var toUpdate []*domain.JobEducationRequirementInput
	var toCreate []*domain.JobEducationRequirementInput
	inputIDSet := make(map[string]bool)

	for _, input := range inputs {
		if input.ID != nil && *input.ID != "" {
			// 有ID的记录，准备更新
			toUpdate = append(toUpdate, input)
			inputIDSet[*input.ID] = true
		} else {
			// 无ID的记录，准备创建
			toCreate = append(toCreate, input)
		}
	}

	// 找出需要删除的记录（现有记录中不在输入列表中的）
	var toDeleteIDs []uuid.UUID
	for id := range existingMap {
		if !inputIDSet[id] {
			if existingID, err := uuid.Parse(id); err == nil {
				toDeleteIDs = append(toDeleteIDs, existingID)
			}
		}
	}

	// 批量删除不需要的记录
	if len(toDeleteIDs) > 0 {
		if _, err := tx.JobEducationRequirement.Delete().Where(jobeducationrequirement.IDIn(toDeleteIDs...)).Exec(ctx); err != nil {
			return fmt.Errorf("failed to delete education requirements: %w", err)
		}
	}

	// 批量更新现有记录
	for _, input := range toUpdate {
		if input.ID == nil || *input.ID == "" {
			continue
		}

		id, err := uuid.Parse(*input.ID)
		if err != nil {
			return fmt.Errorf("invalid education requirement ID: %w", err)
		}

		updater := tx.JobEducationRequirement.UpdateOneID(id).
			SetEducationType(consts.JobEducationType(input.EducationType))

		if _, err := updater.Save(ctx); err != nil {
			return fmt.Errorf("failed to update education requirement: %w", err)
		}
	}

	// 批量创建新记录
	if len(toCreate) > 0 {
		builders := make([]*db.JobEducationRequirementCreate, 0, len(toCreate))
		for _, input := range toCreate {
			builder := tx.JobEducationRequirement.Create().
				SetJobID(jobID).
				SetEducationType(consts.JobEducationType(input.EducationType))

			builders = append(builders, builder)
		}

		if _, err := tx.JobEducationRequirement.CreateBulk(builders...).Save(ctx); err != nil {
			return fmt.Errorf("failed to create education requirements: %w", err)
		}
	}

	return nil
}

func upsertExperienceRequirements(ctx context.Context, tx *db.Tx, jobID uuid.UUID, inputs []*domain.JobExperienceRequirementInput) error {
	// 获取现有记录
	existing, err := tx.JobExperienceRequirement.Query().Where(jobexperiencerequirement.JobID(jobID)).All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query existing experience requirements: %w", err)
	}

	// 创建现有记录的ID映射
	existingMap := make(map[string]*db.JobExperienceRequirement)
	for _, item := range existing {
		existingMap[item.ID.String()] = item
	}

	// 分类处理输入数据
	var toUpdate []*domain.JobExperienceRequirementInput
	var toCreate []*domain.JobExperienceRequirementInput
	inputIDSet := make(map[string]bool)

	for _, input := range inputs {
		if input.ID != nil && *input.ID != "" {
			// 有ID的记录，准备更新
			toUpdate = append(toUpdate, input)
			inputIDSet[*input.ID] = true
		} else {
			// 无ID的记录，准备创建
			toCreate = append(toCreate, input)
		}
	}

	// 找出需要删除的记录（现有记录中不在输入列表中的）
	var toDeleteIDs []uuid.UUID
	for id := range existingMap {
		if !inputIDSet[id] {
			if existingID, err := uuid.Parse(id); err == nil {
				toDeleteIDs = append(toDeleteIDs, existingID)
			}
		}
	}

	// 批量删除不需要的记录
	if len(toDeleteIDs) > 0 {
		if _, err := tx.JobExperienceRequirement.Delete().Where(jobexperiencerequirement.IDIn(toDeleteIDs...)).Exec(ctx); err != nil {
			return fmt.Errorf("failed to delete experience requirements: %w", err)
		}
	}

	// 批量更新现有记录
	for _, input := range toUpdate {
		if input.ID == nil || *input.ID == "" {
			continue
		}

		id, err := uuid.Parse(*input.ID)
		if err != nil {
			return fmt.Errorf("invalid experience requirement ID: %w", err)
		}

		updater := tx.JobExperienceRequirement.UpdateOneID(id).
			SetExperienceType(input.ExperienceType).
			SetMinYears(input.MinYears).
			SetIdealYears(input.IdealYears)

		if _, err := updater.Save(ctx); err != nil {
			return fmt.Errorf("failed to update experience requirement: %w", err)
		}
	}

	// 批量创建新记录
	if len(toCreate) > 0 {
		builders := make([]*db.JobExperienceRequirementCreate, 0, len(toCreate))
		for _, input := range toCreate {
			builder := tx.JobExperienceRequirement.Create().
				SetJobID(jobID).
				SetExperienceType(input.ExperienceType).
				SetMinYears(input.MinYears).
				SetIdealYears(input.IdealYears)

			builders = append(builders, builder)
		}

		if _, err := tx.JobExperienceRequirement.CreateBulk(builders...).Save(ctx); err != nil {
			return fmt.Errorf("failed to create experience requirements: %w", err)
		}
	}

	return nil
}

func upsertIndustryRequirements(ctx context.Context, tx *db.Tx, jobID uuid.UUID, inputs []*domain.JobIndustryRequirementInput) error {
	// 获取现有记录
	existing, err := tx.JobIndustryRequirement.Query().Where(jobindustryrequirement.JobID(jobID)).All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query existing industry requirements: %w", err)
	}

	// 创建现有记录的ID映射
	existingMap := make(map[string]*db.JobIndustryRequirement)
	for _, item := range existing {
		existingMap[item.ID.String()] = item
	}

	// 分类处理输入数据
	var toUpdate []*domain.JobIndustryRequirementInput
	var toCreate []*domain.JobIndustryRequirementInput
	inputIDSet := make(map[string]bool)

	for _, input := range inputs {
		if input.ID != nil && *input.ID != "" {
			// 有ID的记录，准备更新
			toUpdate = append(toUpdate, input)
			inputIDSet[*input.ID] = true
		} else {
			// 无ID的记录，准备创建
			toCreate = append(toCreate, input)
		}
	}

	// 找出需要删除的记录（现有记录中不在输入列表中的）
	var toDeleteIDs []uuid.UUID
	for id := range existingMap {
		if !inputIDSet[id] {
			if existingID, err := uuid.Parse(id); err == nil {
				toDeleteIDs = append(toDeleteIDs, existingID)
			}
		}
	}

	// 批量删除不需要的记录
	if len(toDeleteIDs) > 0 {
		if _, err := tx.JobIndustryRequirement.Delete().Where(jobindustryrequirement.IDIn(toDeleteIDs...)).Exec(ctx); err != nil {
			return fmt.Errorf("failed to delete industry requirements: %w", err)
		}
	}

	// 批量更新现有记录
	for _, input := range toUpdate {
		if input.ID == nil || *input.ID == "" {
			continue
		}

		id, err := uuid.Parse(*input.ID)
		if err != nil {
			return fmt.Errorf("invalid industry requirement ID: %w", err)
		}

		updater := tx.JobIndustryRequirement.UpdateOneID(id)

		// 处理industry字段 - 支持可选
		if trimmed := strings.TrimSpace(input.Industry); trimmed != "" {
			updater.SetIndustry(trimmed)
		} else {
			updater.ClearIndustry()
		}

		if input.CompanyName != nil && *input.CompanyName != "" {
			updater.SetCompanyName(*input.CompanyName)
		} else {
			updater.ClearCompanyName()
		}

		if _, err := updater.Save(ctx); err != nil {
			return fmt.Errorf("failed to update industry requirement: %w", err)
		}
	}

	// 批量创建新记录
	if len(toCreate) > 0 {
		builders := make([]*db.JobIndustryRequirementCreate, 0, len(toCreate))
		for _, input := range toCreate {
			builder := tx.JobIndustryRequirement.Create().
				SetJobID(jobID)

			// 处理industry字段 - 支持可选
			if trimmed := strings.TrimSpace(input.Industry); trimmed != "" {
				builder.SetIndustry(trimmed)
			}

			if input.CompanyName != nil && *input.CompanyName != "" {
				builder.SetCompanyName(*input.CompanyName)
			}

			builders = append(builders, builder)
		}

		if _, err := tx.JobIndustryRequirement.CreateBulk(builders...).Save(ctx); err != nil {
			return fmt.Errorf("failed to create industry requirements: %w", err)
		}
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
				Type:    string(item.Type),
			})
		}
	}

	education := make([]*domain.JobEducationRequirement, 0)
	if list := entity.Edges.EducationRequirements; len(list) > 0 {
		education = make([]*domain.JobEducationRequirement, 0, len(list))
		for _, item := range list {
			education = append(education, &domain.JobEducationRequirement{
				ID:            item.ID.String(),
				JobID:         item.JobID.String(),
				EducationType: string(item.EducationType),
			})
		}
	}

	experience := make([]*domain.JobExperienceRequirement, 0)
	if list := entity.Edges.ExperienceRequirements; len(list) > 0 {
		experience = make([]*domain.JobExperienceRequirement, 0, len(list))
		for _, item := range list {
			experience = append(experience, &domain.JobExperienceRequirement{
				ID:             item.ID.String(),
				JobID:          item.JobID.String(),
				ExperienceType: item.ExperienceType,
				MinYears:       item.MinYears,
				IdealYears:     item.IdealYears,
			})
		}
	}

	industry := make([]*domain.JobIndustryRequirement, 0)
	if list := entity.Edges.IndustryRequirements; len(list) > 0 {
		industry = make([]*domain.JobIndustryRequirement, 0, len(list))
		for _, item := range list {
			industryValue := ""
			if item.Industry != nil {
				industryValue = *item.Industry
			}
			industry = append(industry, &domain.JobIndustryRequirement{
				ID:          item.ID.String(),
				JobID:       item.JobID.String(),
				Industry:    industryValue,
				CompanyName: copyStringPtr(item.CompanyName),
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

// ParseJobProfile 解析岗位描述，返回结构化的岗位画像数据
func (u *JobProfileUsecase) ParseJobProfile(ctx context.Context, req *domain.ParseJobProfileReq) (*domain.ParseJobProfileResp, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	description := strings.TrimSpace(req.Description)
	if description == "" {
		return nil, fmt.Errorf("description is required")
	}

	// 使用 parser service 进行解析
	return u.parserService.ParseJobProfile(ctx, req)
}
