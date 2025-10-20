package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/domain"
)

type ResumeUsecase struct {
	config                *config.Config
	repo                  domain.ResumeRepo
	parserService         domain.ParserService
	storageService        domain.StorageService
	jobApplicationUsecase domain.JobApplicationUsecase
	notificationUsecase   domain.NotificationUsecase
	redis                 *redis.Client
	logger                *slog.Logger
}

// NewResumeUsecase 创建简历用例
func NewResumeUsecase(
	cfg *config.Config,
	repo domain.ResumeRepo,
	parserService domain.ParserService,
	storageService domain.StorageService,
	jobApplicationUsecase domain.JobApplicationUsecase,
	notificationUsecase domain.NotificationUsecase,
	redis *redis.Client,
	logger *slog.Logger,
) domain.ResumeUsecase {
	return &ResumeUsecase{
		config:                cfg,
		repo:                  repo,
		parserService:         parserService,
		storageService:        storageService,
		jobApplicationUsecase: jobApplicationUsecase,
		notificationUsecase:   notificationUsecase,
		redis:                 redis,
		logger:                logger,
	}
}

// Upload 上传简历
func (u *ResumeUsecase) Upload(ctx context.Context, req *domain.UploadResumeReq) (*domain.Resume, error) {
	return u.uploadWithOptions(ctx, req, false)
}

// uploadWithOptions 内部上传函数，支持控制是否等待解析完成
func (u *ResumeUsecase) uploadWithOptions(ctx context.Context, req *domain.UploadResumeReq, waitForParsing bool) (*domain.Resume, error) {
	// 上传文件到存储服务
	fileInfo, err := u.storageService.Upload(ctx, req.File, req.Filename)
	if err != nil {
		u.logger.Error("Failed to upload resume file", "error", err)
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// 解析用户ID
	uploaderID, err := uuid.Parse(req.UploaderID)
	if err != nil {
		return nil, fmt.Errorf("invalid uploader ID: %w", err)
	}

	// 创建简历记录
	resume := &db.Resume{
		UploaderID:    uploaderID,
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

	if waitForParsing {
		// 同步解析简历（用于批量上传）
		u.parseResumeSync(ctx, createdResume.ID.String())

		// 重新获取更新后的简历数据
		updatedResume, err := u.repo.GetByID(ctx, createdResume.ID.String())
		if err != nil {
			u.logger.Error("Failed to get updated resume after parsing", "error", err)
			// 即使获取失败，也返回原始简历数据
		} else {
			createdResume = updatedResume
		}
	} else {
		// 异步解析简历 - 使用独立的context避免HTTP请求context被取消
		go u.parseResumeAsync(context.Background(), createdResume.ID.String())
	}

	// 转换为domain对象
	result := &domain.Resume{}
	return result.From(createdResume), nil
}

// GetByID 根据ID获取简历详情
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

	projects, err := u.getProjectsByResumeID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get projects", "error", err, "resume_id", id)
	}

	logs, err := u.getLogsByResumeID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get logs", "error", err, "resume_id", id)
	}

	// 转换为domain对象
	domainResume := (&domain.Resume{}).From(resume)

	// 转换岗位申请信息
	if resume.Edges.JobApplications != nil {
		jobApplications := make([]*domain.JobApplication, 0, len(resume.Edges.JobApplications))
		for _, app := range resume.Edges.JobApplications {
			jobApp := &domain.JobApplication{}
			jobApplications = append(jobApplications, jobApp.From(app))
		}
		domainResume.JobPositions = jobApplications
	}

	result := &domain.ResumeDetail{
		Resume:      domainResume,
		Educations:  educations,
		Experiences: experiences,
		Skills:      skills,
		Projects:    projects,
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

	// 转换为domain对象，确保返回空数组而不是nil
	result := make([]*domain.Resume, 0, len(resumes))
	for _, resume := range resumes {
		domainResume := &domain.Resume{}
		domainResume = domainResume.From(resume)

		// 转换岗位申请信息
		if resume.Edges.JobApplications != nil {
			jobApplications := make([]*domain.JobApplication, 0, len(resume.Edges.JobApplications))
			for _, app := range resume.Edges.JobApplications {
				jobApp := &domain.JobApplication{}
				jobApplications = append(jobApplications, jobApp.From(app))
			}
			domainResume.JobPositions = jobApplications
		}

		result = append(result, domainResume)
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
	result := make([]*domain.Resume, 0, len(resumes))
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

		// 处理项目经验更新
		if req.Projects != nil {
			if err := u.updateProjects(ctx, tx, resume.ID, req.Projects); err != nil {
				return fmt.Errorf("failed to update projects: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update resume: %w", err)
	}

	// 处理岗位关联更新（在事务外执行，因为JobApplicationUsecase有自己的事务管理）
	if req.JobPositionIDs != nil || req.JobApplications != nil {
		updateJobReq := &domain.UpdateJobApplicationsReq{
			ResumeID:       req.ID,
			JobPositionIDs: req.JobPositionIDs,
			Applications:   req.JobApplications,
		}

		if len(updateJobReq.JobPositionIDs) == 0 {
			updateJobReq.JobPositionIDs = []string{} // 确保至少是空数组而不是nil
		}

		_, err := u.jobApplicationUsecase.UpdateJobApplications(ctx, updateJobReq)
		if err != nil {
			u.logger.Error("failed to update job applications", "error", err, "resume_id", req.ID)
			return nil, fmt.Errorf("failed to update job applications: %w", err)
		}
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

	// 清理已有的解析数据，防止重复记录
	if err := u.repo.ClearParsedData(ctx, id); err != nil {
		return fmt.Errorf("清理已有解析数据失败: %w", err)
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
	if errUpdate := u.repo.UpdateStatus(ctx, resumeID, domain.ResumeStatusProcessing); errUpdate != nil {
		u.logger.Error("Failed to update status to processing", "error", errUpdate, "resume_id", resumeID)
	}

	// 调用LLM解析服务
	parsedData, err := u.parserService.ParseResume(ctx, resumeID, resume.ResumeFileURL, "pdf")
	if err != nil {
		u.logger.Error("Failed to parse resume", "error", err, "resume_id", resumeID)
		u.updateParseError(ctx, resumeID, fmt.Sprintf("解析失败: %v", err))
		return
	}

	u.logger.Debug("Parsed resume data", "resume_id", resumeID, "data", parsedData)

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

// parseResumeSync 同步解析简历，用于批量上传场景
func (u *ResumeUsecase) parseResumeSync(ctx context.Context, resumeID string) error {
	// 获取简历信息
	resume, err := u.repo.GetByID(ctx, resumeID)
	if err != nil {
		u.logger.Error("Failed to get resume for parsing", "error", err, "resume_id", resumeID)
		u.updateParseError(ctx, resumeID, fmt.Sprintf("获取简历失败: %v", err))
		return err
	}

	// 更新状态为处理中
	if errUpdate := u.repo.UpdateStatus(ctx, resumeID, domain.ResumeStatusProcessing); errUpdate != nil {
		u.logger.Error("Failed to update status to processing", "error", errUpdate, "resume_id", resumeID)
	}

	// 调用LLM解析服务
	parsedData, err := u.parserService.ParseResume(ctx, resumeID, resume.ResumeFileURL, "pdf")
	if err != nil {
		u.logger.Error("Failed to parse resume", "error", err, "resume_id", resumeID)
		u.updateParseError(ctx, resumeID, fmt.Sprintf("解析失败: %v", err))
		return err
	}

	u.logger.Debug("Parsed resume data", "resume_id", resumeID, "data", parsedData)

	// 更新解析后的数据
	if err := u.updateParsedData(ctx, resumeID, parsedData); err != nil {
		u.logger.Error("Failed to update parsed data", "error", err, "resume_id", resumeID)
		u.updateParseError(ctx, resumeID, fmt.Sprintf("保存解析数据失败: %v", err))
		return err
	}

	// 更新状态为完成
	if err := u.repo.UpdateStatus(ctx, resumeID, domain.ResumeStatusCompleted); err != nil {
		u.logger.Error("Failed to update status to completed", "error", err, "resume_id", resumeID)
		return err
	}

	u.logger.Info("Resume parsed successfully", "resume_id", resumeID)
	return nil
}

// publishResumeParseCompletedNotification 发布简历解析完成通知
func (u *ResumeUsecase) publishResumeParseCompletedNotification(ctx context.Context, resumeID string, success bool, errorMsg string) {
	// 获取简历信息
	resume, err := u.repo.GetByID(ctx, resumeID)
	if err != nil {
		u.logger.Error("Failed to get resume for notification", "error", err, "resume_id", resumeID)
		return
	}

	// 解析UUID
	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		u.logger.Error("Invalid resume ID for notification", "error", err, "resume_id", resumeID)
		return
	}

	// 创建通知负载
	payload := domain.ResumeParseCompletedPayload{
		ResumeID: resumeUUID,
		UserID:   resume.UploaderID,
		FileName: resume.Name,
		ParsedAt: time.Now(),
		Success:  success,
		ErrorMsg: errorMsg,
	}

	// 发布通知事件
	if err := u.notificationUsecase.PublishEvent(ctx, payload); err != nil {
		u.logger.Error("Failed to publish resume parse completed notification", "error", err, "resume_id", resumeID)
	}
}

// publishBatchResumeParseCompletedNotification 发送批量简历解析完成通知
func (u *ResumeUsecase) publishBatchResumeParseCompletedNotification(ctx context.Context, task *domain.BatchUploadTask) {
	// 解析UUID
	uploaderUUID, err := uuid.Parse(task.UploaderID)
	if err != nil {
		u.logger.Error("Invalid uploader ID for batch notification", "error", err, "uploader_id", task.UploaderID)
		return
	}

	// 获取上传人姓名
	uploaderName := ""
	// 通过任何一个成功的简历获取用户信息
	for _, item := range task.Items {
		if item.ResumeID != nil && item.Status == domain.BatchUploadStatusCompleted {
			resume, err := u.repo.GetByID(ctx, *item.ResumeID)
			if err == nil && resume.Edges.User != nil {
				uploaderName = resume.Edges.User.Username
				break
			}
		}
	}

	// 如果没有成功的简历，使用UploaderID作为默认值
	if uploaderName == "" {
		uploaderName = task.UploaderID
	}

	// 处理Source字段
	source := ""
	if task.Source != nil {
		source = *task.Source
	}

	// 创建通知负载
	payload := domain.BatchResumeParseCompletedPayload{
		TaskID:         task.TaskID,
		UploaderID:     uploaderUUID,
		UploaderName:   uploaderName,
		TotalCount:     task.TotalCount,
		SuccessCount:   task.SuccessCount,
		FailedCount:    task.FailedCount,
		JobPositionIDs: task.JobPositionIDs,
		Source:         source,
		CompletedAt:    time.Now(),
	}

	// 发布通知事件
	if err := u.notificationUsecase.PublishEvent(ctx, payload); err != nil {
		u.logger.Error("Failed to publish batch resume parse completed notification", "error", err, "task_id", task.TaskID)
	}
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
			ResumeID:       resumeUUID,
			School:         edu.School,
			Degree:         edu.Degree,
			Major:          edu.Major,
			UniversityType: edu.UniversityType,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
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

	// 保存项目经验
	for _, project := range data.Projects {
		resumeProject := &db.ResumeProject{
			ResumeID:         resumeUUID,
			Name:             project.Name,
			Role:             project.Role,
			Company:          project.Company,
			Description:      project.Description,
			Responsibilities: project.Responsibilities,
			Achievements:     project.Achievements,
			Technologies:     project.Technologies,
			ProjectURL:       project.ProjectURL,
			ProjectType:      project.ProjectType,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if project.StartDate != nil {
			resumeProject.StartDate = *project.StartDate
		}
		if project.EndDate != nil {
			resumeProject.EndDate = *project.EndDate
		}
		if _, err := u.repo.CreateProject(ctx, resumeProject); err != nil {
			u.logger.Error("Failed to create project", "error", err, "resume_id", resumeID)
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

// getProjectsByResumeID 根据简历ID获取项目经验
func (u *ResumeUsecase) getProjectsByResumeID(ctx context.Context, resumeID string) ([]*domain.ResumeProject, error) {
	// 从数据库获取项目经验
	dbProjects, err := u.repo.GetProjectsByResumeID(ctx, resumeID)
	if err != nil {
		return nil, err
	}

	// 转换为domain对象
	projects := make([]*domain.ResumeProject, 0, len(dbProjects))
	for _, dbProject := range dbProjects {
		project := &domain.ResumeProject{}
		projects = append(projects, project.From(dbProject))
	}

	return projects, nil
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
			if education.UniversityType != nil {
				updater.SetUniversityType(*education.UniversityType)
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
			if education.UniversityType != nil {
				creator.SetUniversityType(*education.UniversityType)
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

// updateProjects 更新项目经验
func (u *ResumeUsecase) updateProjects(ctx context.Context, tx *db.Tx, resumeID uuid.UUID, projects []*domain.UpdateResumeProject) error {
	// 删除需要删除的项目经验
	for _, project := range projects {
		if project.Action == "delete" && project.ID != nil {
			projectID, err := uuid.Parse(*project.ID)
			if err != nil {
				return fmt.Errorf("invalid project ID: %w", err)
			}
			if err := tx.ResumeProject.DeleteOneID(projectID).Exec(ctx); err != nil {
				return fmt.Errorf("failed to delete project: %w", err)
			}
		}
	}

	// 更新或创建项目经验
	for _, project := range projects {
		if project.Action == "delete" {
			continue
		}

		if project.Action == "update" && project.ID != nil {
			// 更新现有项目经验
			projectID, err := uuid.Parse(*project.ID)
			if err != nil {
				return fmt.Errorf("invalid project ID: %w", err)
			}

			updater := tx.ResumeProject.UpdateOneID(projectID)
			if project.Name != nil {
				updater.SetName(*project.Name)
			}
			if project.Role != nil {
				updater.SetRole(*project.Role)
			}
			if project.Company != nil {
				updater.SetCompany(*project.Company)
			}
			if project.StartDate != nil {
				updater.SetStartDate(*project.StartDate)
			}
			if project.EndDate != nil {
				updater.SetEndDate(*project.EndDate)
			}
			if project.Description != nil {
				updater.SetDescription(*project.Description)
			}
			if project.Responsibilities != nil {
				updater.SetResponsibilities(*project.Responsibilities)
			}
			if project.Achievements != nil {
				updater.SetAchievements(*project.Achievements)
			}
			if project.Technologies != nil {
				updater.SetTechnologies(*project.Technologies)
			}
			if project.ProjectURL != nil {
				updater.SetProjectURL(*project.ProjectURL)
			}
			if project.ProjectType != nil {
				updater.SetProjectType(*project.ProjectType)
			}
			updater.SetUpdatedAt(time.Now())

			if _, err := updater.Save(ctx); err != nil {
				return fmt.Errorf("failed to update project: %w", err)
			}
		} else if project.Action == "create" {
			// 创建新项目经验
			creator := tx.ResumeProject.Create().
				SetResumeID(resumeID).
				SetCreatedAt(time.Now()).
				SetUpdatedAt(time.Now())

			if project.Name != nil {
				creator.SetName(*project.Name)
			}
			if project.Role != nil {
				creator.SetRole(*project.Role)
			}
			if project.Company != nil {
				creator.SetCompany(*project.Company)
			}
			if project.StartDate != nil {
				creator.SetStartDate(*project.StartDate)
			}
			if project.EndDate != nil {
				creator.SetEndDate(*project.EndDate)
			}
			if project.Description != nil {
				creator.SetDescription(*project.Description)
			}
			if project.Responsibilities != nil {
				creator.SetResponsibilities(*project.Responsibilities)
			}
			if project.Achievements != nil {
				creator.SetAchievements(*project.Achievements)
			}
			if project.Technologies != nil {
				creator.SetTechnologies(*project.Technologies)
			}
			if project.ProjectURL != nil {
				creator.SetProjectURL(*project.ProjectURL)
			}
			if project.ProjectType != nil {
				creator.SetProjectType(*project.ProjectType)
			}

			if _, err := creator.Save(ctx); err != nil {
				return fmt.Errorf("failed to create project: %w", err)
			}
		}
	}

	return nil
}

// BatchUpload 批量上传简历
func (u *ResumeUsecase) BatchUpload(ctx context.Context, req *domain.BatchUploadResumeReq) (*domain.BatchUploadTask, error) {
	// 生成任务ID
	taskID := uuid.New().String()

	// 创建批量上传任务
	task := &domain.BatchUploadTask{
		TaskID:         taskID,
		UploaderID:     req.UploaderID,
		Status:         domain.BatchUploadStatusPending,
		TotalCount:     len(req.Files),
		CompletedCount: 0,
		SuccessCount:   0,
		FailedCount:    0,
		JobPositionIDs: req.JobPositionIDs,
		Source:         req.Source,
		Notes:          req.Notes,
		Items:          make([]*domain.BatchUploadItem, 0, len(req.Files)),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 为每个文件创建任务项
	for _, fileInfo := range req.Files {
		itemID := uuid.New().String()
		item := &domain.BatchUploadItem{
			ItemID:    itemID,
			Filename:  fileInfo.Filename,
			Status:    domain.BatchUploadStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		task.Items = append(task.Items, item)
	}

	// 将任务保存到Redis
	err := u.saveBatchUploadTaskToRedis(ctx, task)
	if err != nil {
		u.logger.Error("failed to save batch upload task to redis", "task_id", taskID, "error", err)
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	// 异步处理批量上传
	go u.processBatchUploadAsync(context.Background(), taskID, req)

	u.logger.Info("batch upload task created", "task_id", taskID, "total_files", len(req.Files))
	return task, nil
}

// GetBatchUploadStatus 获取批量上传任务状态
func (u *ResumeUsecase) GetBatchUploadStatus(ctx context.Context, taskID string) (*domain.BatchUploadTask, error) {
	// 获取基础任务信息
	task, err := u.getBatchUploadTaskFromRedis(ctx, taskID)
	if err != nil {
		u.logger.Error("failed to get batch upload task from redis", "task_id", taskID, "error", err)
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	// 汇总各个简历项目的状态
	u.recalculateBatchUploadTaskStatus(ctx, task)

	// 根据统计信息确定任务状态
	if task.CompletedCount == task.TotalCount {
		if task.FailedCount > 0 && task.SuccessCount == 0 {
			task.Status = domain.BatchUploadStatusFailed
		} else {
			task.Status = domain.BatchUploadStatusCompleted
		}
	} else if task.FailedCount > 0 || task.SuccessCount > 0 {
		task.Status = domain.BatchUploadStatusProcessing
	} else {
		task.Status = domain.BatchUploadStatusPending
	}

	return task, nil
}

// CancelBatchUpload 取消批量上传任务
func (u *ResumeUsecase) CancelBatchUpload(ctx context.Context, taskID string) error {
	// 从Redis获取任务
	task, err := u.getBatchUploadTaskFromRedis(ctx, taskID)
	if err != nil {
		u.logger.Error("failed to get batch upload task from redis", "task_id", taskID, "error", err)
		return fmt.Errorf("failed to get task: %w", err)
	}

	if task == nil {
		return fmt.Errorf("task not found")
	}

	// 检查任务状态
	if task.Status == domain.BatchUploadStatusCompleted || task.Status == domain.BatchUploadStatusCancelled {
		return fmt.Errorf("task already completed or cancelled")
	}

	// 更新任务状态为已取消
	task.Status = domain.BatchUploadStatusCancelled
	task.UpdatedAt = time.Now()
	now := time.Now()
	task.CompletedAt = &now

	// 更新所有未完成的任务项状态
	for _, item := range task.Items {
		if item.Status == domain.BatchUploadStatusPending || item.Status == domain.BatchUploadStatusProcessing {
			item.Status = domain.BatchUploadStatusCancelled
			item.UpdatedAt = time.Now()
			item.CompletedAt = &now
		}
	}

	// 保存到Redis
	err = u.saveBatchUploadTaskToRedis(ctx, task)
	if err != nil {
		u.logger.Error("failed to save cancelled task to redis", "task_id", taskID, "error", err)
		return fmt.Errorf("failed to save cancelled task: %w", err)
	}

	u.logger.Info("batch upload task cancelled", "task_id", taskID)
	return nil
}

// processBatchUploadAsync 异步处理批量上传
func (u *ResumeUsecase) processBatchUploadAsync(ctx context.Context, taskID string, req *domain.BatchUploadResumeReq) {
	// 获取任务
	task, err := u.getBatchUploadTaskFromRedis(ctx, taskID)
	if err != nil {
		u.logger.Error("failed to get batch upload task", "task_id", taskID, "error", err)
		return
	}

	// 更新任务状态为处理中
	task.Status = domain.BatchUploadStatusProcessing
	task.UpdatedAt = time.Now()
	if err := u.saveBatchUploadTaskToRedis(ctx, task); err != nil {
		u.logger.Error("failed to save batch upload task to redis", "task_id", taskID, "error", err)
		// 继续处理，不因为Redis保存失败而中断批量上传
	}

	// 使用goroutine并发处理文件，但限制并发数
	const maxConcurrency = 3
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for i, fileInfo := range req.Files {
		// 检查任务是否被取消
		currentTask, _ := u.getBatchUploadTaskFromRedis(ctx, taskID)
		if currentTask != nil && currentTask.Status == domain.BatchUploadStatusCancelled {
			u.logger.Info("batch upload task cancelled, stopping processing", "task_id", taskID)
			break
		}

		wg.Add(1)
		go func(index int, file *domain.BatchUploadFileInfo) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			u.processSingleFileInBatch(ctx, taskID, index, file, req)
		}(i, fileInfo)
	}

	// 等待所有文件处理完成
	wg.Wait()

	// 更新最终任务状态
	u.updateFinalBatchUploadStatus(ctx, taskID)
}

// processSingleFileInBatch 处理批量上传中的单个文件
func (u *ResumeUsecase) processSingleFileInBatch(ctx context.Context, taskID string, itemIndex int, fileInfo *domain.BatchUploadFileInfo, req *domain.BatchUploadResumeReq) {
	// 更新任务项状态为处理中
	u.updateBatchUploadItemStatus(ctx, taskID, itemIndex, domain.BatchUploadStatusProcessing, nil, nil, nil)

	// 构建单个文件上传请求
	uploadReq := &domain.UploadResumeReq{
		UploaderID:     req.UploaderID,
		File:           fileInfo.File,
		Filename:       fileInfo.Filename,
		JobPositionIDs: req.JobPositionIDs,
		Source:         req.Source,
		Notes:          req.Notes,
	}

	// 调用单个文件上传逻辑，使用同步解析模式
	resume, err := u.uploadWithOptions(ctx, uploadReq, true) // waitForParsing = true
	now := time.Now()

	if err != nil {
		// 上传失败
		errorMsg := err.Error()
		u.updateBatchUploadItemStatus(ctx, taskID, itemIndex, domain.BatchUploadStatusFailed, &errorMsg, nil, &now)
		u.logger.Error("failed to upload file in batch", "task_id", taskID, "filename", fileInfo.Filename, "error", err)
	} else {
		// 上传成功，创建岗位关联关系
		if len(req.JobPositionIDs) > 0 {
			jobAppReq := &domain.CreateJobApplicationsReq{
				ResumeID:       resume.ID,
				JobPositionIDs: req.JobPositionIDs,
				Source:         req.Source,
				Notes:          req.Notes,
			}

			_, jobAppErr := u.jobApplicationUsecase.CreateJobApplications(ctx, jobAppReq)
			if jobAppErr != nil {
				u.logger.Error("failed to create job applications in batch", "task_id", taskID, "resume_id", resume.ID, "error", jobAppErr)
				// 不将此错误视为致命错误，因为简历已经上传成功
				u.logger.Warn("resume uploaded successfully but job applications creation failed in batch", "task_id", taskID, "resume_id", resume.ID)
			} else {
				u.logger.Info("job applications created successfully in batch", "task_id", taskID, "resume_id", resume.ID, "job_position_count", len(req.JobPositionIDs))
			}
		}

		u.updateBatchUploadItemStatus(ctx, taskID, itemIndex, domain.BatchUploadStatusCompleted, nil, &resume.ID, &now)
		u.logger.Info("successfully uploaded file in batch", "task_id", taskID, "filename", fileInfo.Filename, "resume_id", resume.ID)
	}
}

// updateFinalBatchUploadStatus 更新批量上传任务的最终状态
func (u *ResumeUsecase) updateFinalBatchUploadStatus(ctx context.Context, taskID string) {
	task, err := u.getBatchUploadTaskFromRedis(ctx, taskID)
	if err != nil {
		u.logger.Error("failed to get task for final status update", "task_id", taskID, "error", err)
		return
	}

	if task == nil {
		return
	}

	// 如果任务已被取消，不需要更新状态
	if task.Status == domain.BatchUploadStatusCancelled {
		return
	}

	// 重新计算任务状态和统计信息
	u.recalculateBatchUploadTaskStatus(ctx, task)

	// 更新最终状态
	if task.FailedCount > 0 && task.SuccessCount == 0 {
		task.Status = domain.BatchUploadStatusFailed
	} else {
		task.Status = domain.BatchUploadStatusCompleted
	}

	now := time.Now()
	task.UpdatedAt = now
	task.CompletedAt = &now

	if err := u.saveBatchUploadTaskToRedis(ctx, task); err != nil {
		u.logger.Error("failed to save final batch upload task to redis", "task_id", taskID, "error", err)
		// 即使Redis保存失败，也记录完成日志
	}

	u.publishBatchResumeParseCompletedNotification(ctx, task)

	u.logger.Info("batch upload task completed", "task_id", taskID, "success_count", task.SuccessCount, "failed_count", task.FailedCount)
}

// recalculateBatchUploadTaskStatus 重新计算批量上传任务的状态和统计信息
func (u *ResumeUsecase) recalculateBatchUploadTaskStatus(ctx context.Context, task *domain.BatchUploadTask) {
	// 汇总各个简历项目的状态
	for i := range task.Items {
		itemKey := fmt.Sprintf("batch_upload_item:%s:%d", task.TaskID, i)
		itemData, err := u.redis.Get(ctx, itemKey).Result()
		if err != nil {
			if err == redis.Nil {
				// 项目状态不存在，保持原有状态（pending）
				continue
			}
			u.logger.Error("failed to get item status", "task_id", task.TaskID, "item_index", i, "error", err)
			continue
		}

		// 解析项目状态
		var itemStatus map[string]interface{}
		if err := json.Unmarshal([]byte(itemData), &itemStatus); err != nil {
			u.logger.Error("failed to unmarshal item status", "task_id", task.TaskID, "item_index", i, "error", err)
			continue
		}

		// 更新项目状态
		if status, ok := itemStatus["status"].(string); ok {
			task.Items[i].Status = domain.BatchUploadStatus(status)
		}
		if errorMsg, ok := itemStatus["error_message"].(string); ok {
			task.Items[i].ErrorMessage = &errorMsg
		}
		if resumeID, ok := itemStatus["resume_id"].(string); ok {
			task.Items[i].ResumeID = &resumeID
		}
		if completedAtStr, ok := itemStatus["completed_at"].(string); ok {
			if completedAt, err := time.Parse(time.RFC3339, completedAtStr); err == nil {
				task.Items[i].CompletedAt = &completedAt
			}
		}
		if updatedAtStr, ok := itemStatus["updated_at"].(string); ok {
			if updatedAt, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
				task.Items[i].UpdatedAt = updatedAt
			}
		}
	}

	// 重新计算任务统计信息
	task.CompletedCount = 0
	task.SuccessCount = 0
	task.FailedCount = 0

	for _, item := range task.Items {
		switch item.Status {
		case domain.BatchUploadStatusCompleted:
			task.CompletedCount++
			task.SuccessCount++
		case domain.BatchUploadStatusFailed:
			task.CompletedCount++
			task.FailedCount++
		}
	}
}

// saveBatchUploadTaskToRedis 将批量上传任务保存到Redis
func (u *ResumeUsecase) saveBatchUploadTaskToRedis(ctx context.Context, task *domain.BatchUploadTask) error {
	taskKey := fmt.Sprintf(consts.BatchUploadTaskKeyFmt, task.TaskID)

	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// 设置过期时间为24小时
	expiration := 24 * time.Hour
	err = u.redis.Set(ctx, taskKey, taskData, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to save task to redis: %w", err)
	}

	return nil
}

// updateBatchUploadItemStatus 更新批量上传任务中单个项目的状态
func (u *ResumeUsecase) updateBatchUploadItemStatus(ctx context.Context, taskID string, itemIndex int, status domain.BatchUploadStatus, errorMsg *string, resumeID *string, completedAt *time.Time) {
	// 为每个简历项目创建独立的Redis键
	itemKey := fmt.Sprintf("batch_upload_item:%s:%d", taskID, itemIndex)

	// 创建项目状态数据
	itemData := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if errorMsg != nil {
		itemData["error_message"] = *errorMsg
	}
	if resumeID != nil {
		itemData["resume_id"] = *resumeID
	}
	if completedAt != nil {
		itemData["completed_at"] = *completedAt
	}

	// 序列化数据
	itemJSON, err := json.Marshal(itemData)
	if err != nil {
		u.logger.Error("failed to marshal item data", "task_id", taskID, "item_index", itemIndex, "error", err)
		return
	}

	// 存储到Redis，设置24小时过期时间
	err = u.redis.Set(ctx, itemKey, itemJSON, 24*time.Hour).Err()
	if err != nil {
		u.logger.Error("failed to save item status to redis", "task_id", taskID, "item_index", itemIndex, "error", err)
		return
	}
}

// getBatchUploadTaskFromRedis 从Redis获取批量上传任务
func (u *ResumeUsecase) getBatchUploadTaskFromRedis(ctx context.Context, taskID string) (*domain.BatchUploadTask, error) {
	taskKey := fmt.Sprintf(consts.BatchUploadTaskKeyFmt, taskID)

	taskData, err := u.redis.Get(ctx, taskKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 任务不存在
		}
		return nil, fmt.Errorf("failed to get task from redis: %w", err)
	}

	var task domain.BatchUploadTask
	err = json.Unmarshal([]byte(taskData), &task)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}
