package usecase

import (
	"context"
	"time"

	"github.com/chaitin/WhaleHire/backend/domain"
)

type jobApplicationUsecase struct {
	jobApplicationRepo domain.JobApplicationRepo
	jobProfileRepo     domain.JobProfileRepo
}

// NewJobApplicationUsecase 创建岗位申请用例实例
func NewJobApplicationUsecase(
	jobApplicationRepo domain.JobApplicationRepo,
	jobProfileRepo domain.JobProfileRepo,
) domain.JobApplicationUsecase {
	return &jobApplicationUsecase{
		jobApplicationRepo: jobApplicationRepo,
		jobProfileRepo:     jobProfileRepo,
	}
}

// CreateJobApplications 创建岗位申请关联
func (u *jobApplicationUsecase) CreateJobApplications(ctx context.Context, req *domain.CreateJobApplicationsReq) ([]*domain.JobApplication, error) {
	if len(req.JobPositionIDs) == 0 {
		return []*domain.JobApplication{}, nil
	}

	// 构建domain对象
	var domainApplications []*domain.JobApplication
	now := time.Now()

	for _, jobPositionID := range req.JobPositionIDs {
		domainApp := &domain.JobApplication{
			ResumeID:      req.ResumeID,
			JobPositionID: jobPositionID,
			Status:        "applied", // 默认状态
			AppliedAt:     &now,
		}

		if req.Source != nil {
			domainApp.Source = *req.Source
		}

		if req.Notes != nil {
			domainApp.Notes = *req.Notes
		}

		domainApplications = append(domainApplications, domainApp)
	}

	// 调用仓储层创建
	createdApplications, err := u.jobApplicationRepo.CreateJobApplications(ctx, domainApplications)
	if err != nil {
		return nil, err
	}

	// 获取岗位信息并填充到结果中
	for _, app := range createdApplications {
		jobPosition, err := u.jobProfileRepo.GetByID(ctx, app.JobPositionID)
		if err != nil {
			continue
		}

		app.JobTitle = jobPosition.Name
		if jobPosition.Edges.Department != nil {
			app.JobDepartment = jobPosition.Edges.Department.Name
		}
	}

	return createdApplications, nil
}

// UpdateJobApplications 更新岗位申请关联
func (u *jobApplicationUsecase) UpdateJobApplications(ctx context.Context, req *domain.UpdateJobApplicationsReq) ([]*domain.JobApplication, error) {
	if len(req.JobPositionIDs) == 0 {
		return []*domain.JobApplication{}, nil
	}

	// 构建domain对象
	var domainApplications []*domain.JobApplication
	now := time.Now()

	// 创建岗位ID到应用信息的映射
	appMap := make(map[string]*domain.JobApplicationAssociation)
	for _, app := range req.Applications {
		appMap[app.JobPositionID] = app
	}

	for _, jobPositionID := range req.JobPositionIDs {
		domainApp := &domain.JobApplication{
			ResumeID:      req.ResumeID,
			JobPositionID: jobPositionID,
			Status:        "applied", // 默认状态
			AppliedAt:     &now,
		}

		// 如果有特定的应用信息，使用它
		if appInfo, exists := appMap[jobPositionID]; exists {
			if appInfo.Source != nil {
				domainApp.Source = *appInfo.Source
			}
			if appInfo.Notes != nil {
				domainApp.Notes = *appInfo.Notes
			}
		}

		domainApplications = append(domainApplications, domainApp)
	}

	// 调用仓储层更新
	updatedApplications, err := u.jobApplicationRepo.UpdateJobApplications(ctx, req.ResumeID, domainApplications)
	if err != nil {
		return nil, err
	}

	// 获取岗位信息并填充到结果中
	for _, app := range updatedApplications {
		jobPosition, err := u.jobProfileRepo.GetByID(ctx, app.JobPositionID)
		if err != nil {
			continue
		}

		app.JobTitle = jobPosition.Name
		if jobPosition.Edges.Department != nil {
			app.JobDepartment = jobPosition.Edges.Department.Name
		}
	}

	return updatedApplications, nil
}

// GetJobPositionsByResumeID 根据简历ID获取关联的岗位信息
func (u *jobApplicationUsecase) GetJobPositionsByResumeID(ctx context.Context, resumeID string) ([]*domain.JobApplication, error) {
	return u.jobApplicationRepo.GetJobPositionsByResumeID(ctx, resumeID)
}

// GetResumesByJobPositionID 根据岗位ID获取关联的简历信息
func (u *jobApplicationUsecase) GetResumesByJobPositionID(ctx context.Context, jobPositionID string, req *domain.GetResumesByJobPositionIDReq) ([]*domain.JobApplication, int64, error) {
	return u.jobApplicationRepo.GetResumesByJobPositionID(ctx, jobPositionID, req)
}
