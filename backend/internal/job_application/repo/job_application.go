package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/resumejobapplication"
	"github.com/chaitin/WhaleHire/backend/domain"
)

type jobApplicationRepo struct {
	client *db.Client
}

// NewJobApplicationRepo 创建岗位申请仓储实例
func NewJobApplicationRepo(client *db.Client) domain.JobApplicationRepo {
	return &jobApplicationRepo{
		client: client,
	}
}

// CreateJobApplications 批量创建简历与岗位的关联关系
func (r *jobApplicationRepo) CreateJobApplications(ctx context.Context, applications []*domain.JobApplication) ([]*domain.JobApplication, error) {
	if len(applications) == 0 {
		return []*domain.JobApplication{}, nil
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = fmt.Errorf("rollback failed: %v, original error: %w", rollbackErr, err)
			}
		}
	}()

	var results []*domain.JobApplication
	for _, app := range applications {
		resumeUUID, parseResumeErr := uuid.Parse(app.ResumeID)
		if parseResumeErr != nil {
			return nil, fmt.Errorf("invalid resume ID: %w", parseResumeErr)
		}

		jobPositionUUID, parseJobErr := uuid.Parse(app.JobPositionID)
		if parseJobErr != nil {
			return nil, fmt.Errorf("invalid job position ID: %w", parseJobErr)
		}

		created, createErr := tx.ResumeJobApplication.Create().
			SetResumeID(resumeUUID).
			SetJobPositionID(jobPositionUUID).
			SetStatus(app.Status).
			SetSource(app.Source).
			SetNotes(app.Notes).
			SetAppliedAt(*app.AppliedAt).
			Save(ctx)
		if createErr != nil {
			return nil, fmt.Errorf("failed to create job application: %w", createErr)
		}

		result := &domain.JobApplication{
			ID:            created.ID.String(),
			ResumeID:      created.ResumeID.String(),
			JobPositionID: created.JobPositionID.String(),
			Status:        created.Status,
			Source:        created.Source,
			Notes:         created.Notes,
			AppliedAt:     &created.AppliedAt,
			CreatedAt:     created.CreatedAt.Unix(),
			UpdatedAt:     created.UpdatedAt.Unix(),
		}
		results = append(results, result)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return results, nil
}

// UpdateJobApplications 更新简历与岗位的关联关系（差异更新）
func (r *jobApplicationRepo) UpdateJobApplications(ctx context.Context, resumeID string, applications []*domain.JobApplication) ([]*domain.JobApplication, error) {
	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		return nil, fmt.Errorf("invalid resume ID: %w", err)
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = fmt.Errorf("rollback failed: %v, original error: %w", rollbackErr, err)
			}
		}
	}()

	// 1. 获取现有的关联关系
	existing, err := tx.ResumeJobApplication.Query().
		Where(resumejobapplication.ResumeID(resumeUUID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query existing job applications: %w", err)
	}

	// 2. 构建现有关联的映射表（jobPositionID -> record）
	existingMap := make(map[string]*db.ResumeJobApplication)
	for _, record := range existing {
		existingMap[record.JobPositionID.String()] = record
	}

	// 3. 构建新关联的映射表
	newMap := make(map[string]*domain.JobApplication)
	for _, app := range applications {
		newMap[app.JobPositionID] = app
	}

	var results []*domain.JobApplication

	// 4. 处理需要更新或保留的记录
	for jobPositionID, newApp := range newMap {
		if existingRecord, exists := existingMap[jobPositionID]; exists {
			// 更新现有记录
			updated, updateErr := tx.ResumeJobApplication.UpdateOne(existingRecord).
				SetStatus(newApp.Status).
				SetSource(newApp.Source).
				SetNotes(newApp.Notes).
				Save(ctx)
			if updateErr != nil {
				return nil, fmt.Errorf("failed to update job application: %w", updateErr)
			}

			result := &domain.JobApplication{
				ID:            updated.ID.String(),
				ResumeID:      updated.ResumeID.String(),
				JobPositionID: updated.JobPositionID.String(),
				Status:        updated.Status,
				Source:        updated.Source,
				Notes:         updated.Notes,
				AppliedAt:     &updated.AppliedAt,
				CreatedAt:     updated.CreatedAt.Unix(),
				UpdatedAt:     updated.UpdatedAt.Unix(),
			}
			results = append(results, result)
		} else {
			// 创建新记录
			jobPositionUUID, parseErr := uuid.Parse(newApp.JobPositionID)
			if parseErr != nil {
				return nil, fmt.Errorf("invalid job position ID: %w", parseErr)
			}

			created, createErr := tx.ResumeJobApplication.Create().
				SetResumeID(resumeUUID).
				SetJobPositionID(jobPositionUUID).
				SetStatus(newApp.Status).
				SetSource(newApp.Source).
				SetNotes(newApp.Notes).
				SetAppliedAt(*newApp.AppliedAt).
				Save(ctx)
			if createErr != nil {
				return nil, fmt.Errorf("failed to create job application: %w", createErr)
			}

			result := &domain.JobApplication{
				ID:            created.ID.String(),
				ResumeID:      created.ResumeID.String(),
				JobPositionID: created.JobPositionID.String(),
				Status:        created.Status,
				Source:        created.Source,
				Notes:         created.Notes,
				AppliedAt:     &created.AppliedAt,
				CreatedAt:     created.CreatedAt.Unix(),
				UpdatedAt:     created.UpdatedAt.Unix(),
			}
			results = append(results, result)
		}
	}

	// 5. 删除不再需要的记录
	for jobPositionID, existingRecord := range existingMap {
		if _, stillNeeded := newMap[jobPositionID]; !stillNeeded {
			deleteErr := tx.ResumeJobApplication.DeleteOne(existingRecord).Exec(ctx)
			if deleteErr != nil {
				return nil, fmt.Errorf("failed to delete job application: %w", deleteErr)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return results, nil
}

// GetJobPositionsByResumeID 根据简历ID获取关联的岗位信息
func (r *jobApplicationRepo) GetJobPositionsByResumeID(ctx context.Context, resumeID string) ([]*domain.JobApplication, error) {
	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		return nil, fmt.Errorf("invalid resume ID: %w", err)
	}

	applications, err := r.client.ResumeJobApplication.Query().
		Where(resumejobapplication.ResumeID(resumeUUID)).
		WithJobPosition(func(jpq *db.JobPositionQuery) {
			jpq.WithDepartment()
		}).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get job applications: %w", err)
	}

	var results []*domain.JobApplication
	for _, app := range applications {
		result := &domain.JobApplication{
			ID:            app.ID.String(),
			ResumeID:      app.ResumeID.String(),
			JobPositionID: app.JobPositionID.String(),
			Status:        app.Status,
			Source:        app.Source,
			Notes:         app.Notes,
			AppliedAt:     &app.AppliedAt,
			CreatedAt:     app.CreatedAt.Unix(),
			UpdatedAt:     app.UpdatedAt.Unix(),
		}

		// 设置岗位信息
		if app.Edges.JobPosition != nil {
			result.JobTitle = app.Edges.JobPosition.Name
			if app.Edges.JobPosition.Edges.Department != nil {
				result.JobDepartment = app.Edges.JobPosition.Edges.Department.Name
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// GetResumesByJobPositionID 根据岗位ID获取关联的简历信息
func (r *jobApplicationRepo) GetResumesByJobPositionID(ctx context.Context, jobPositionID string, req *domain.GetResumesByJobPositionIDReq) ([]*domain.JobApplication, int64, error) {
	jobPositionUUID, err := uuid.Parse(jobPositionID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid job position ID: %w", err)
	}

	query := r.client.ResumeJobApplication.Query().
		Where(resumejobapplication.JobPositionID(jobPositionUUID)).
		WithJobPosition(func(jpq *db.JobPositionQuery) {
			jpq.WithDepartment()
		}).
		WithResume()

	// 添加筛选条件
	if req.Status != nil && *req.Status != "" {
		query = query.Where(resumejobapplication.Status(*req.Status))
	}
	if req.Source != nil && *req.Source != "" {
		query = query.Where(resumejobapplication.Source(*req.Source))
	}
	if req.DateFrom != nil {
		query = query.Where(resumejobapplication.AppliedAtGTE(*req.DateFrom))
	}
	if req.DateTo != nil {
		query = query.Where(resumejobapplication.AppliedAtLTE(*req.DateTo))
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count job applications: %w", err)
	}

	// 分页查询
	applications, err := query.
		Offset((req.Page - 1) * req.Size).
		Limit(req.Size).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get job applications: %w", err)
	}

	var results []*domain.JobApplication
	for _, app := range applications {
		result := &domain.JobApplication{
			ID:            app.ID.String(),
			ResumeID:      app.ResumeID.String(),
			JobPositionID: app.JobPositionID.String(),
			Status:        app.Status,
			Source:        app.Source,
			Notes:         app.Notes,
			AppliedAt:     &app.AppliedAt,
			CreatedAt:     app.CreatedAt.Unix(),
			UpdatedAt:     app.UpdatedAt.Unix(),
		}

		// 设置岗位信息
		if app.Edges.JobPosition != nil {
			result.JobTitle = app.Edges.JobPosition.Name
			if app.Edges.JobPosition.Edges.Department != nil {
				result.JobDepartment = app.Edges.JobPosition.Edges.Department.Name
			}
		}

		results = append(results, result)
	}

	return results, int64(total), nil
}
