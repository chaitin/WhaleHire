package domain

import (
	"context"
	"time"

	"github.com/chaitin/WhaleHire/backend/db"
)

// JobApplication 简历岗位申请关联信息
type JobApplication struct {
	ID            string     `json:"id"`
	ResumeID      string     `json:"resume_id"`
	JobPositionID string     `json:"job_position_id"`
	JobTitle      string     `json:"job_title"`      // 岗位标题
	JobDepartment string     `json:"job_department"` // 岗位部门
	Status        string     `json:"status"`         // 申请状态
	Source        string     `json:"source"`         // 申请来源
	Notes         string     `json:"notes"`          // 备注信息
	AppliedAt     *time.Time `json:"applied_at"`     // 申请时间
	CreatedAt     int64      `json:"created_at"`
	UpdatedAt     int64      `json:"updated_at"`
}

// From 从数据库实体转换为domain对象
func (ja *JobApplication) From(entity *db.ResumeJobApplication) *JobApplication {
	if entity == nil {
		return ja
	}
	ja.ID = entity.ID.String()
	ja.ResumeID = entity.ResumeID.String()
	ja.JobPositionID = entity.JobPositionID.String()
	ja.Status = entity.Status
	ja.Source = entity.Source
	ja.Notes = entity.Notes
	if !entity.AppliedAt.IsZero() {
		ja.AppliedAt = &entity.AppliedAt
	}
	ja.CreatedAt = entity.CreatedAt.Unix()
	ja.UpdatedAt = entity.UpdatedAt.Unix()

	// 填充岗位信息
	if entity.Edges.JobPosition != nil {
		ja.JobTitle = entity.Edges.JobPosition.Name
		if entity.Edges.JobPosition.Edges.Department != nil {
			ja.JobDepartment = entity.Edges.JobPosition.Edges.Department.Name
		}
	}

	return ja
}

// JobApplicationUsecase 岗位申请用例接口
type JobApplicationUsecase interface {
	// CreateJobApplications 创建简历与岗位的关联关系
	CreateJobApplications(ctx context.Context, req *CreateJobApplicationsReq) ([]*JobApplication, error)
	// UpdateJobApplications 更新简历与岗位的关联关系
	UpdateJobApplications(ctx context.Context, req *UpdateJobApplicationsReq) ([]*JobApplication, error)
	// GetJobPositionsByResumeID 根据简历ID获取关联的岗位信息
	GetJobPositionsByResumeID(ctx context.Context, resumeID string) ([]*JobApplication, error)
	// GetResumesByJobPositionID 根据岗位ID获取关联的简历信息
	GetResumesByJobPositionID(ctx context.Context, jobPositionID string, req *GetResumesByJobPositionIDReq) ([]*JobApplication, int64, error)
}

// JobApplicationRepo 岗位申请仓储接口
type JobApplicationRepo interface {
	// CreateJobApplications 批量创建简历与岗位的关联关系
	CreateJobApplications(ctx context.Context, applications []*JobApplication) ([]*JobApplication, error)
	// UpdateJobApplications 更新简历与岗位的关联关系
	UpdateJobApplications(ctx context.Context, resumeID string, applications []*JobApplication) ([]*JobApplication, error)
	// GetJobPositionsByResumeID 根据简历ID获取关联的岗位信息
	GetJobPositionsByResumeID(ctx context.Context, resumeID string) ([]*JobApplication, error)
	// GetResumesByJobPositionID 根据岗位ID获取关联的简历信息
	GetResumesByJobPositionID(ctx context.Context, jobPositionID string, req *GetResumesByJobPositionIDReq) ([]*JobApplication, int64, error)
}

// CreateJobApplicationsReq 创建岗位申请关联请求
type CreateJobApplicationsReq struct {
	ResumeID       string   `json:"resume_id" validate:"required"`
	JobPositionIDs []string `json:"job_position_ids" validate:"required,min=1"`
	Source         *string  `json:"source,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
}

// UpdateJobApplicationsReq 更新岗位申请关联请求
type UpdateJobApplicationsReq struct {
	ResumeID       string                       `json:"resume_id" validate:"required"`
	JobPositionIDs []string                     `json:"job_position_ids" validate:"required,min=1"`
	Applications   []*JobApplicationAssociation `json:"applications,omitempty"`
}

// JobApplicationAssociation 岗位申请关联信息
type JobApplicationAssociation struct {
	JobPositionID string  `json:"job_position_id" validate:"required"`
	Source        *string `json:"source,omitempty"`
	Notes         *string `json:"notes,omitempty"`
}
