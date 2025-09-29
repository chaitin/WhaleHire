package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/domain"
)

type ResumeUsecase struct {
	config         *config.Config
	repo           domain.ResumeRepo
	parserService  domain.ParserService
	storageService domain.StorageService
	logger         *slog.Logger
}

// NewResumeUsecase 创建简历用例
func NewResumeUsecase(
	cfg *config.Config,
	repo domain.ResumeRepo,
	parserService domain.ParserService,
	storageService domain.StorageService,
	logger *slog.Logger,
) domain.ResumeUsecase {
	return &ResumeUsecase{
		config:         cfg,
		repo:           repo,
		parserService:  parserService,
		storageService: storageService,
		logger:         logger,
	}
}

// Upload 上传简历
func (u *ResumeUsecase) Upload(ctx context.Context, req *domain.UploadResumeReq) (*domain.Resume, error) {
	// 上传文件到存储服务
	fileInfo, err := u.storageService.Upload(ctx, req.File, req.Filename)
	if err != nil {
		u.logger.Error("Failed to upload resume file", "error", err)
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// 解析用户ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// 创建简历记录
	resume := &db.Resume{
		UserID:        userID,
		ResumeFileURL: fileInfo.FileURL,
		Status:        string(domain.ResumeStatusPending),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 保存到数据库
	createdResume, err := u.repo.Create(ctx, resume)
	if err != nil {
		u.logger.Error("Failed to create resume record", "error", err)
		return nil, fmt.Errorf("failed to create resume: %w", err)
	}

	// 异步解析简历 - 使用独立的context避免HTTP请求context被取消
	go u.parseResumeAsync(context.Background(), createdResume.ID.String())

	// 转换为domain对象
	result := &domain.Resume{}
	return result.From(createdResume), nil
}

// GetByID 获取简历详情
func (u *ResumeUsecase) GetByID(ctx context.Context, id string) (*domain.ResumeDetail, error) {
	resume, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get resume: %w", err)
	}

	// 获取关联数据
	educations, err := u.getEducationsByResumeID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get educations", "error", err, "resume_id", id)
	}

	experiences, err := u.getExperiencesByResumeID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get experiences", "error", err, "resume_id", id)
	}

	skills, err := u.getSkillsByResumeID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get skills", "error", err, "resume_id", id)
	}

	logs, err := u.getLogsByResumeID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get logs", "error", err, "resume_id", id)
	}

	// 转换为domain对象
	result := &domain.ResumeDetail{
		Resume:      (&domain.Resume{}).From(resume),
		Educations:  educations,
		Experiences: experiences,
		Skills:      skills,
		Logs:        logs,
	}

	return result, nil
}

// List 获取简历列表
func (u *ResumeUsecase) List(ctx context.Context, req *domain.ListResumeReq) (*domain.ListResumeResp, error) {
	resumes, pageInfo, err := u.repo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list resumes: %w", err)
	}

	// 转换为domain对象
	var result []*domain.Resume
	for _, resume := range resumes {
		domainResume := &domain.Resume{}
		result = append(result, domainResume.From(resume))
	}

	return &domain.ListResumeResp{
		PageInfo: pageInfo,
		Resumes:  result,
	}, nil
}

// Search 搜索简历
func (u *ResumeUsecase) Search(ctx context.Context, req *domain.SearchResumeReq) (*domain.SearchResumeResp, error) {
	resumes, pageInfo, err := u.repo.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to search resumes: %w", err)
	}

	// 转换为domain对象
	var result []*domain.Resume
	for _, resume := range resumes {
		domainResume := &domain.Resume{}
		result = append(result, domainResume.From(resume))
	}

	return &domain.SearchResumeResp{
		PageInfo: pageInfo,
		Resumes:  result,
	}, nil
}

// Update 更新简历
func (u *ResumeUsecase) Update(ctx context.Context, req *domain.UpdateResumeReq) (*domain.Resume, error) {
	updatedResume, err := u.repo.Update(ctx, req.ID, func(tx *db.Tx, resume *db.Resume, updateOne *db.ResumeUpdateOne) error {
		// 更新简历主表字段
		if req.Name != nil {
			updateOne.SetName(*req.Name)
		}
		if req.Gender != nil {
			updateOne.SetGender(*req.Gender)
		}
		if req.Birthday != nil {
			updateOne.SetBirthday(*req.Birthday)
		}
		if req.Email != nil {
			updateOne.SetEmail(*req.Email)
		}
		if req.Phone != nil {
			updateOne.SetPhone(*req.Phone)
		}
		if req.CurrentCity != nil {
			updateOne.SetCurrentCity(*req.CurrentCity)
		}
		if req.HighestEducation != nil {
			updateOne.SetHighestEducation(*req.HighestEducation)
		}
		if req.YearsExperience != nil {
			updateOne.SetYearsExperience(*req.YearsExperience)
		}
		updateOne.SetUpdatedAt(time.Now())

		// 处理教育经历更新
		if req.Educations != nil {
			if err := u.updateEducations(ctx, tx, resume.ID, req.Educations); err != nil {
				return fmt.Errorf("failed to update educations: %w", err)
			}
		}

		// 处理工作经历更新
		if req.Experiences != nil {
			if err := u.updateExperiences(ctx, tx, resume.ID, req.Experiences); err != nil {
				return fmt.Errorf("failed to update experiences: %w", err)
			}
		}

		// 处理技能更新
		if req.Skills != nil {
			if err := u.updateSkills(ctx, tx, resume.ID, req.Skills); err != nil {
				return fmt.Errorf("failed to update skills: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update resume: %w", err)
	}

	result := &domain.Resume{}
	return result.From(updatedResume), nil
}

// Delete 删除简历
func (u *ResumeUsecase) Delete(ctx context.Context, id string) error {
	// 获取简历信息
	resume, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get resume: %w", err)
	}

	// 删除文件
	if resume.ResumeFileURL != "" {
		if err := u.storageService.Delete(ctx, resume.ResumeFileURL); err != nil {
			u.logger.Error("Failed to delete resume file", "error", err, "url", resume.ResumeFileURL)
		}
	}

	// 删除数据库记录
	if err := u.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete resume: %w", err)
	}

	return nil
}

// UpdateStatus 更新简历状态
func (u *ResumeUsecase) UpdateStatus(ctx context.Context, id string, status domain.ResumeStatus) error {
	return u.repo.UpdateStatus(ctx, id, status)
}

// Reparse 重新解析简历
func (u *ResumeUsecase) Reparse(ctx context.Context, id string) error {
	// 检查简历是否存在
	resume, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("获取简历失败: %w", err)
	}

	// 更新状态为解析中
	if err := u.repo.UpdateStatus(ctx, id, domain.ResumeStatusProcessing); err != nil {
		return fmt.Errorf("更新简历状态失败: %w", err)
	}

	// 异步解析
	go u.parseResumeAsync(context.Background(), resume.ID.String())

	return nil
}

// GetParseProgress 获取简历解析进度
func (u *ResumeUsecase) GetParseProgress(ctx context.Context, id string) (*domain.ResumeParseProgress, error) {
	// 获取简历基本信息
	resume, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取简历失败: %w", err)
	}

	progress := &domain.ResumeParseProgress{
		ResumeID:     resume.ID.String(),
		Status:       domain.ResumeStatus(resume.Status),
		ErrorMessage: resume.ErrorMessage,
		StartedAt:    &resume.CreatedAt,
		CompletedAt:  &resume.UpdatedAt,
	}

	// 根据状态设置进度和消息
	switch domain.ResumeStatus(resume.Status) {
	case domain.ResumeStatusPending:
		progress.Progress = 0
		progress.Message = "简历已上传，等待解析"
	case domain.ResumeStatusProcessing:
		progress.Progress = 50
		progress.Message = "正在解析简历内容..."
	case domain.ResumeStatusCompleted:
		progress.Progress = 100
		progress.Message = "简历解析完成"
		progress.CompletedAt = &resume.UpdatedAt
	case domain.ResumeStatusFailed:
		progress.Progress = 0
		progress.Message = "简历解析失败"
		if resume.ErrorMessage != "" {
			progress.ErrorMessage = resume.ErrorMessage
		}
	default:
		progress.Progress = 0
		progress.Message = "未知状态"
	}

	return progress, nil
}

// parseResumeAsync 异步解析简历
func (u *ResumeUsecase) parseResumeAsync(ctx context.Context, resumeID string) {
	// 获取简历信息
	resume, err := u.repo.GetByID(ctx, resumeID)
	if err != nil {
		u.logger.Error("Failed to get resume for parsing", "error", err, "resume_id", resumeID)
		u.updateParseError(ctx, resumeID, fmt.Sprintf("获取简历失败: %v", err))
		return
	}

	// 更新状态为处理中
	if err := u.repo.UpdateStatus(ctx, resumeID, domain.ResumeStatusProcessing); err != nil {
		u.logger.Error("Failed to update status to processing", "error", err, "resume_id", resumeID)
	}

	// 调用LLM解析服务
	parsedData, err := u.parserService.ParseResume(ctx, resumeID, resume.ResumeFileURL, "pdf")
	if err != nil {
		u.logger.Error("Failed to parse resume", "error", err, "resume_id", resumeID)
		u.updateParseError(ctx, resumeID, fmt.Sprintf("解析失败: %v", err))
		return
	}

	// 更新解析后的数据
	if err := u.updateParsedData(ctx, resumeID, parsedData); err != nil {
		u.logger.Error("Failed to update parsed data", "error", err, "resume_id", resumeID)
		u.updateParseError(ctx, resumeID, fmt.Sprintf("保存解析数据失败: %v", err))
		return
	}

	// 更新状态为完成
	if err := u.repo.UpdateStatus(ctx, resumeID, domain.ResumeStatusCompleted); err != nil {
		u.logger.Error("Failed to update status to completed", "error", err, "resume_id", resumeID)
	}

	u.logger.Info("Resume parsed successfully", "resume_id", resumeID)
}

// updateParseError 更新解析错误
func (u *ResumeUsecase) updateParseError(ctx context.Context, resumeID string, errorMsg string) {
	if err := u.repo.UpdateErrorMessage(ctx, resumeID, errorMsg); err != nil {
		u.logger.Error("Failed to update error message", "error", err, "resume_id", resumeID)
	}
	if err := u.repo.UpdateStatus(ctx, resumeID, domain.ResumeStatusFailed); err != nil {
		u.logger.Error("Failed to update status to failed", "error", err, "resume_id", resumeID)
	}
}

// updateParsedData 更新解析后的数据
func (u *ResumeUsecase) updateParsedData(ctx context.Context, resumeID string, data *domain.ParsedResumeData) error {
	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		return fmt.Errorf("invalid resume ID: %w", err)
	}

	// 更新基本信息
	_, err = u.repo.Update(ctx, resumeID, func(tx *db.Tx, resume *db.Resume, updateOne *db.ResumeUpdateOne) error {
		if data.BasicInfo != nil {
			if data.BasicInfo.Name != "" {
				updateOne.SetName(data.BasicInfo.Name)
			}
			if data.BasicInfo.Phone != "" {
				updateOne.SetPhone(data.BasicInfo.Phone)
			}
			if data.BasicInfo.Email != "" {
				updateOne.SetEmail(data.BasicInfo.Email)
			}
			if data.BasicInfo.Gender != "" {
				updateOne.SetGender(data.BasicInfo.Gender)
			}
			if data.BasicInfo.CurrentCity != "" {
				updateOne.SetCurrentCity(data.BasicInfo.CurrentCity)
			}
			if data.BasicInfo.Birthday != nil {
				updateOne.SetBirthday(*data.BasicInfo.Birthday)
			}
			if data.BasicInfo.HighestEducation != "" {
				updateOne.SetHighestEducation(data.BasicInfo.HighestEducation)
			}
			if data.BasicInfo.YearsExperience != 0 {
				updateOne.SetYearsExperience(data.BasicInfo.YearsExperience)
			}
		}
		updateOne.SetParsedAt(time.Now())
		updateOne.SetUpdatedAt(time.Now())
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update basic info: %w", err)
	}

	// 保存教育经历
	for _, edu := range data.Educations {
		education := &db.ResumeEducation{
			ResumeID:  resumeUUID,
			School:    edu.School,
			Degree:    edu.Degree,
			Major:     edu.Major,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if edu.StartDate != nil {
			education.StartDate = *edu.StartDate
		}
		if edu.EndDate != nil {
			education.EndDate = *edu.EndDate
		}
		if _, err := u.repo.CreateEducation(ctx, education); err != nil {
			u.logger.Error("Failed to create education", "error", err, "resume_id", resumeID)
		}
	}

	// 保存工作经历
	for _, exp := range data.Experiences {
		experience := &db.ResumeExperience{
			ResumeID:    resumeUUID,
			Company:     exp.Company,
			Position:    exp.Position,
			Description: exp.Description,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if exp.StartDate != nil {
			experience.StartDate = *exp.StartDate
		}
		if exp.EndDate != nil {
			experience.EndDate = *exp.EndDate
		}
		if _, err := u.repo.CreateExperience(ctx, experience); err != nil {
			u.logger.Error("Failed to create experience", "error", err, "resume_id", resumeID)
		}
	}

	// 保存技能
	for _, skill := range data.Skills {
		resumeSkill := &db.ResumeSkill{
			ResumeID:    resumeUUID,
			SkillName:   skill.Name,
			Level:       skill.Level,
			Description: skill.Description,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if _, err := u.repo.CreateSkill(ctx, resumeSkill); err != nil {
			u.logger.Error("Failed to create skill", "error", err, "resume_id", resumeID)
		}
	}

	return nil
}

// getEducationsByResumeID 根据简历ID获取教育经历
func (u *ResumeUsecase) getEducationsByResumeID(ctx context.Context, resumeID string) ([]*domain.ResumeEducation, error) {
	// 从数据库获取教育经历
	dbEducations, err := u.repo.GetEducationsByResumeID(ctx, resumeID)
	if err != nil {
		return nil, err
	}

	// 转换为domain对象
	educations := make([]*domain.ResumeEducation, 0, len(dbEducations))
	for _, dbEducation := range dbEducations {
		education := &domain.ResumeEducation{}
		educations = append(educations, education.From(dbEducation))
	}

	return educations, nil
}

// getExperiencesByResumeID 根据简历ID获取工作经历
func (u *ResumeUsecase) getExperiencesByResumeID(ctx context.Context, resumeID string) ([]*domain.ResumeExperience, error) {
	// 从数据库获取工作经历
	dbExperiences, err := u.repo.GetExperiencesByResumeID(ctx, resumeID)
	if err != nil {
		return nil, err
	}

	// 转换为domain对象
	experiences := make([]*domain.ResumeExperience, 0, len(dbExperiences))
	for _, dbExperience := range dbExperiences {
		experience := &domain.ResumeExperience{}
		experiences = append(experiences, experience.From(dbExperience))
	}

	return experiences, nil
}

// getSkillsByResumeID 根据简历ID获取技能
func (u *ResumeUsecase) getSkillsByResumeID(ctx context.Context, resumeID string) ([]*domain.ResumeSkill, error) {
	// 从数据库获取技能
	dbSkills, err := u.repo.GetSkillsByResumeID(ctx, resumeID)
	if err != nil {
		return nil, err
	}

	// 转换为domain对象
	skills := make([]*domain.ResumeSkill, 0, len(dbSkills))
	for _, dbSkill := range dbSkills {
		skill := &domain.ResumeSkill{}
		skills = append(skills, skill.From(dbSkill))
	}

	return skills, nil
}

// getLogsByResumeID 根据简历ID获取操作日志
func (u *ResumeUsecase) getLogsByResumeID(ctx context.Context, resumeID string) ([]*domain.ResumeLog, error) {
	// 从数据库获取操作日志
	dbLogs, err := u.repo.GetLogsByResumeID(ctx, resumeID)
	if err != nil {
		return nil, err
	}

	// 转换为domain对象
	logs := make([]*domain.ResumeLog, 0, len(dbLogs))
	for _, dbLog := range dbLogs {
		log := &domain.ResumeLog{}
		logs = append(logs, log.From(dbLog))
	}

	return logs, nil
}

// updateEducations 更新教育经历
func (u *ResumeUsecase) updateEducations(ctx context.Context, tx *db.Tx, resumeID uuid.UUID, educations []*domain.UpdateResumeEducation) error {
	// 删除需要删除的教育经历
	for _, education := range educations {
		if education.Action == "delete" && education.ID != nil {
			educationID, err := uuid.Parse(*education.ID)
			if err != nil {
				return fmt.Errorf("invalid education ID: %w", err)
			}
			if err := tx.ResumeEducation.DeleteOneID(educationID).Exec(ctx); err != nil {
				return fmt.Errorf("failed to delete education: %w", err)
			}
		}
	}

	// 更新或创建教育经历
	for _, education := range educations {
		if education.Action == "delete" {
			continue
		}

		if education.Action == "update" && education.ID != nil {
			// 更新现有教育经历
			educationID, err := uuid.Parse(*education.ID)
			if err != nil {
				return fmt.Errorf("invalid education ID: %w", err)
			}

			updater := tx.ResumeEducation.UpdateOneID(educationID)
			if education.School != nil {
				updater.SetSchool(*education.School)
			}
			if education.Degree != nil {
				updater.SetDegree(*education.Degree)
			}
			if education.Major != nil {
				updater.SetMajor(*education.Major)
			}
			if education.StartDate != nil {
				updater.SetStartDate(*education.StartDate)
			}
			if education.EndDate != nil {
				updater.SetEndDate(*education.EndDate)
			}
			updater.SetUpdatedAt(time.Now())

			if _, err := updater.Save(ctx); err != nil {
				return fmt.Errorf("failed to update education: %w", err)
			}
		} else if education.Action == "create" {
			// 创建新教育经历
			creator := tx.ResumeEducation.Create().
				SetResumeID(resumeID).
				SetCreatedAt(time.Now()).
				SetUpdatedAt(time.Now())

			if education.School != nil {
				creator.SetSchool(*education.School)
			}
			if education.Degree != nil {
				creator.SetDegree(*education.Degree)
			}
			if education.Major != nil {
				creator.SetMajor(*education.Major)
			}
			if education.StartDate != nil {
				creator.SetStartDate(*education.StartDate)
			}
			if education.EndDate != nil {
				creator.SetEndDate(*education.EndDate)
			}

			if _, err := creator.Save(ctx); err != nil {
				return fmt.Errorf("failed to create education: %w", err)
			}
		}
	}

	return nil
}

// updateExperiences 更新工作经历
func (u *ResumeUsecase) updateExperiences(ctx context.Context, tx *db.Tx, resumeID uuid.UUID, experiences []*domain.UpdateResumeExperience) error {
	// 删除需要删除的工作经历
	for _, experience := range experiences {
		if experience.Action == "delete" && experience.ID != nil {
			experienceID, err := uuid.Parse(*experience.ID)
			if err != nil {
				return fmt.Errorf("invalid experience ID: %w", err)
			}
			if err := tx.ResumeExperience.DeleteOneID(experienceID).Exec(ctx); err != nil {
				return fmt.Errorf("failed to delete experience: %w", err)
			}
		}
	}

	// 更新或创建工作经历
	for _, experience := range experiences {
		if experience.Action == "delete" {
			continue
		}

		if experience.Action == "update" && experience.ID != nil {
			// 更新现有工作经历
			experienceID, err := uuid.Parse(*experience.ID)
			if err != nil {
				return fmt.Errorf("invalid experience ID: %w", err)
			}

			updater := tx.ResumeExperience.UpdateOneID(experienceID)
			if experience.Company != nil {
				updater.SetCompany(*experience.Company)
			}
			if experience.Position != nil {
				updater.SetPosition(*experience.Position)
			}
			if experience.Title != nil {
				updater.SetTitle(*experience.Title)
			}
			if experience.StartDate != nil {
				updater.SetStartDate(*experience.StartDate)
			}
			if experience.EndDate != nil {
				updater.SetEndDate(*experience.EndDate)
			}
			if experience.Description != nil {
				updater.SetDescription(*experience.Description)
			}
			updater.SetUpdatedAt(time.Now())

			if _, err := updater.Save(ctx); err != nil {
				return fmt.Errorf("failed to update experience: %w", err)
			}
		} else if experience.Action == "create" {
			// 创建新工作经历
			creator := tx.ResumeExperience.Create().
				SetResumeID(resumeID).
				SetCreatedAt(time.Now()).
				SetUpdatedAt(time.Now())

			if experience.Company != nil {
				creator.SetCompany(*experience.Company)
			}
			if experience.Position != nil {
				creator.SetPosition(*experience.Position)
			}
			if experience.Title != nil {
				creator.SetTitle(*experience.Title)
			}
			if experience.StartDate != nil {
				creator.SetStartDate(*experience.StartDate)
			}
			if experience.EndDate != nil {
				creator.SetEndDate(*experience.EndDate)
			}
			if experience.Description != nil {
				creator.SetDescription(*experience.Description)
			}

			if _, err := creator.Save(ctx); err != nil {
				return fmt.Errorf("failed to create experience: %w", err)
			}
		}
	}

	return nil
}

// updateSkills 更新技能
func (u *ResumeUsecase) updateSkills(ctx context.Context, tx *db.Tx, resumeID uuid.UUID, skills []*domain.UpdateResumeSkill) error {
	// 删除需要删除的技能
	for _, skill := range skills {
		if skill.Action == "delete" && skill.ID != nil {
			skillID, err := uuid.Parse(*skill.ID)
			if err != nil {
				return fmt.Errorf("invalid skill ID: %w", err)
			}
			if err := tx.ResumeSkill.DeleteOneID(skillID).Exec(ctx); err != nil {
				return fmt.Errorf("failed to delete skill: %w", err)
			}
		}
	}

	// 更新或创建技能
	for _, skill := range skills {
		if skill.Action == "delete" {
			continue
		}

		if skill.Action == "update" && skill.ID != nil {
			// 更新现有技能
			skillID, err := uuid.Parse(*skill.ID)
			if err != nil {
				return fmt.Errorf("invalid skill ID: %w", err)
			}

			updater := tx.ResumeSkill.UpdateOneID(skillID)
			if skill.SkillName != nil {
				updater.SetSkillName(*skill.SkillName)
			}
			if skill.Level != nil {
				updater.SetLevel(*skill.Level)
			}
			if skill.Description != nil {
				updater.SetDescription(*skill.Description)
			}
			updater.SetUpdatedAt(time.Now())

			if _, err := updater.Save(ctx); err != nil {
				return fmt.Errorf("failed to update skill: %w", err)
			}
		} else if skill.Action == "create" {
			// 创建新技能
			creator := tx.ResumeSkill.Create().
				SetResumeID(resumeID).
				SetCreatedAt(time.Now()).
				SetUpdatedAt(time.Now())

			if skill.SkillName != nil {
				creator.SetSkillName(*skill.SkillName)
			}
			if skill.Level != nil {
				creator.SetLevel(*skill.Level)
			}
			if skill.Description != nil {
				creator.SetDescription(*skill.Description)
			}

			if _, err := creator.Save(ctx); err != nil {
				return fmt.Errorf("failed to create skill: %w", err)
			}
		}
	}

	return nil
}
