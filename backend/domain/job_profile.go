package domain

import (
	"context"
	"fmt"

	"github.com/GoYoko/web"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
)

// SkillType represents required/bonus skill categories for a job profile.
type SkillType string

const (
	SkillTypeRequired SkillType = "required"
	SkillTypeBonus    SkillType = "bonus"
)

// JobProfileUsecase defines business logic for job profiles.
type JobProfileUsecase interface {
	Create(ctx context.Context, req *CreateJobProfileReq) (*JobProfileDetail, error)
	Update(ctx context.Context, req *UpdateJobProfileReq) (*JobProfileDetail, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*JobProfileDetail, error)
	List(ctx context.Context, req *ListJobProfileReq) (*ListJobProfileResp, error)
	Search(ctx context.Context, req *SearchJobProfileReq) (*SearchJobProfileResp, error)

	CreateSkillMeta(ctx context.Context, req *CreateSkillMetaReq) (*JobSkillMeta, error)
	ListSkillMeta(ctx context.Context, req *ListSkillMetaReq) (*ListSkillMetaResp, error)
	DeleteSkillMeta(ctx context.Context, id string) error
}

// JobProfileRepo encapsulates persistence for job profiles aggregate.
type JobProfileRepo interface {
	Create(ctx context.Context, job *db.JobPosition, related *JobProfileRelatedEntities) (*db.JobPosition, error)
	Update(ctx context.Context, id string, fn func(tx *db.Tx, current *db.JobPosition, updater *db.JobPositionUpdateOne) error) (*db.JobPosition, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*db.JobPosition, error)
	List(ctx context.Context, req *ListJobProfileRepoReq) ([]*db.JobPosition, *db.PageInfo, error)
	Search(ctx context.Context, req *SearchJobProfileRepoReq) ([]*db.JobPosition, *db.PageInfo, error)
}

// JobSkillMetaRepo manages job skill dictionary persistence.
type JobSkillMetaRepo interface {
	Create(ctx context.Context, name string) (*db.JobSkillMeta, error)
	GetByName(ctx context.Context, name string) (*db.JobSkillMeta, error)
	List(ctx context.Context, req *ListSkillMetaReq) ([]*db.JobSkillMeta, *db.PageInfo, error)
	Delete(ctx context.Context, id string) error
}

// JobProfile represents the core job position info returned to clients.
type JobProfile struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	DepartmentID  string   `json:"department_id"`
	Department    string   `json:"department,omitempty"`
	WorkType      *string  `json:"work_type,omitempty"`
	Location      *string  `json:"location,omitempty"`
	SalaryMin     *float64 `json:"salary_min,omitempty"`
	SalaryMax     *float64 `json:"salary_max,omitempty"`
	Description   *string  `json:"description,omitempty"`
	CreatedBy     *string  `json:"created_by,omitempty"`
	CreatorName   *string  `json:"creator_name,omitempty"`
	Status        string   `json:"status"`
	CreatedAtUnix int64    `json:"created_at"`
	UpdatedAtUnix int64    `json:"updated_at"`
}

// From hydrates JobProfile from ent entity.
func (jp *JobProfile) From(entity *db.JobPosition) *JobProfile {
	if entity == nil {
		return jp
	}
	jp.ID = entity.ID.String()
	jp.Name = entity.Name
	jp.DepartmentID = entity.DepartmentID.String()
	if entity.Edges.Department != nil {
		jp.Department = entity.Edges.Department.Name
	}
	if entity.WorkType != "" {
		jp.WorkType = ptrString(string(entity.WorkType))
	}
	if entity.Location != nil {
		jp.Location = ptrString(*entity.Location)
	}
	if entity.SalaryMin != nil {
		jp.SalaryMin = ptrFloat64(*entity.SalaryMin)
	}
	if entity.SalaryMax != nil {
		jp.SalaryMax = ptrFloat64(*entity.SalaryMax)
	}
	if entity.Description != nil {
		jp.Description = ptrString(*entity.Description)
	}
	if entity.CreatedBy != nil {
		jp.CreatedBy = ptrString(entity.CreatedBy.String())
	}
	if entity.Edges.Creator != nil {
		jp.CreatorName = ptrString(entity.Edges.Creator.Username)
	}
	if entity.Edges.Creator != nil {
		jp.CreatorName = ptrString(entity.Edges.Creator.Username)
	}
	jp.Status = string(entity.Status)
	jp.CreatedAtUnix = entity.CreatedAt.Unix()
	jp.UpdatedAtUnix = entity.UpdatedAt.Unix()
	return jp
}

// JobProfileDetail aggregates profile with related collections.
type JobProfileDetail struct {
	*JobProfile
	Responsibilities       []*JobResponsibility        `json:"responsibilities"`
	Skills                 []*JobSkill                 `json:"skills"`
	EducationRequirements  []*JobEducationRequirement  `json:"education_requirements"`
	ExperienceRequirements []*JobExperienceRequirement `json:"experience_requirements"`
	IndustryRequirements   []*JobIndustryRequirement   `json:"industry_requirements"`
}

// JobResponsibility represents a responsibility row.
type JobResponsibility struct {
	ID             string `json:"id"`
	JobID          string `json:"job_id"`
	Responsibility string `json:"responsibility"`
}

// JobSkill represents a skill linked to a job.
type JobSkill struct {
	ID      string `json:"id"`
	JobID   string `json:"job_id"`
	SkillID string `json:"skill_id"`
	Skill   string `json:"skill"`
	Type    string `json:"type"`
}

// JobEducationRequirement models education expectations.
type JobEducationRequirement struct {
	ID            string `json:"id"`
	JobID         string `json:"job_id"`
	EducationType string `json:"education_type"`
}

// JobExperienceRequirement models experience expectations.
type JobExperienceRequirement struct {
	ID             string `json:"id"`
	JobID          string `json:"job_id"`
	ExperienceType string `json:"experience_type"`
	MinYears       int    `json:"min_years"`
	IdealYears     int    `json:"ideal_years"`
}

// JobIndustryRequirement models target industries or companies.
type JobIndustryRequirement struct {
	ID          string  `json:"id"`
	JobID       string  `json:"job_id"`
	Industry    string  `json:"industry"`
	CompanyName *string `json:"company_name,omitempty"`
}

// JobSkillMeta represents dictionary entries for skills.
type JobSkillMeta struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// From hydrates JobSkillMeta from ent entity.
func (m *JobSkillMeta) From(entity *db.JobSkillMeta) *JobSkillMeta {
	if entity == nil {
		return m
	}
	m.ID = entity.ID.String()
	m.Name = entity.Name
	m.CreatedAt = entity.CreatedAt.Unix()
	m.UpdatedAt = entity.UpdatedAt.Unix()
	return m
}

// CreateJobProfileReq encapsulates payload for creating a job profile.
type CreateJobProfileReq struct {
	Name                   string                           `json:"name" validate:"required"`
	DepartmentID           string                           `json:"department_id" validate:"required"`
	WorkType               *string                          `json:"work_type,omitempty" enums:"full_time,part_time,internship,outsourcing"` // 工作性质: 全职/兼职/实习/外包
	Location               *string                          `json:"location,omitempty"`
	SalaryMin              *float64                         `json:"salary_min,omitempty"`
	SalaryMax              *float64                         `json:"salary_max,omitempty"`
	Description            *string                          `json:"description,omitempty"`
	CreatedBy              *string                          `json:"created_by,omitempty"`
	Status                 *string                          `json:"status,omitempty" enums:"draft,published"` // 职位状态: draft/published
	Responsibilities       []*JobResponsibilityInput        `json:"responsibilities"`
	Skills                 []*JobSkillInput                 `json:"skills"`
	EducationRequirements  []*JobEducationRequirementInput  `json:"education_requirements"`
	ExperienceRequirements []*JobExperienceRequirementInput `json:"experience_requirements"`
	IndustryRequirements   []*JobIndustryRequirementInput   `json:"industry_requirements"`
}

// Validate validates the CreateJobProfileReq fields against constants
func (req *CreateJobProfileReq) Validate() error {
	// Validate WorkType
	if req.WorkType != nil && *req.WorkType != "" {
		workType := consts.JobWorkType(*req.WorkType)
		if !workType.IsValid() {
			return fmt.Errorf("invalid work_type: %s", *req.WorkType)
		}
	}

	// Validate Status
	if req.Status != nil && *req.Status != "" {
		status := consts.JobPositionStatus(*req.Status)
		if !status.IsValid() {
			return fmt.Errorf("invalid status: %s", *req.Status)
		}
	}

	// Validate Skills
	for i, skill := range req.Skills {
		if err := skill.Validate(); err != nil {
			return fmt.Errorf("invalid skill at index %d: %w", i, err)
		}
	}

	// Validate EducationRequirements
	for i, edu := range req.EducationRequirements {
		if err := edu.Validate(); err != nil {
			return fmt.Errorf("invalid education requirement at index %d: %w", i, err)
		}
	}

	// Validate ExperienceRequirements
	for i, exp := range req.ExperienceRequirements {
		if err := exp.Validate(); err != nil {
			return fmt.Errorf("invalid experience requirement at index %d: %w", i, err)
		}
	}

	return nil
}

// UpdateJobProfileReq updates base info and related collections.
type UpdateJobProfileReq struct {
	ID                     string                           `json:"id" validate:"required"`
	Name                   *string                          `json:"name,omitempty"`
	DepartmentID           *string                          `json:"department_id,omitempty"`
	WorkType               *string                          `json:"work_type,omitempty" enums:"full_time,part_time,internship,outsourcing"` // 工作性质: 全职/兼职/实习/外包
	Location               *string                          `json:"location,omitempty"`
	SalaryMin              *float64                         `json:"salary_min,omitempty"`
	SalaryMax              *float64                         `json:"salary_max,omitempty"`
	Description            *string                          `json:"description,omitempty"`
	CreatedBy              *string                          `json:"created_by,omitempty"`
	Status                 *string                          `json:"status,omitempty" enums:"draft,published"` // 职位状态: draft/published
	Responsibilities       []*JobResponsibilityInput        `json:"responsibilities"`
	Skills                 []*JobSkillInput                 `json:"skills"`
	EducationRequirements  []*JobEducationRequirementInput  `json:"education_requirements"`
	ExperienceRequirements []*JobExperienceRequirementInput `json:"experience_requirements"`
	IndustryRequirements   []*JobIndustryRequirementInput   `json:"industry_requirements"`
}

// Validate validates the UpdateJobProfileReq fields against constants
func (req *UpdateJobProfileReq) Validate() error {
	// Validate WorkType
	if req.WorkType != nil && *req.WorkType != "" {
		workType := consts.JobWorkType(*req.WorkType)
		if !workType.IsValid() {
			return fmt.Errorf("invalid work_type: %s", *req.WorkType)
		}
	}

	// Validate Status
	if req.Status != nil && *req.Status != "" {
		status := consts.JobPositionStatus(*req.Status)
		if !status.IsValid() {
			return fmt.Errorf("invalid status: %s", *req.Status)
		}
	}

	// Validate Skills
	for i, skill := range req.Skills {
		if err := skill.Validate(); err != nil {
			return fmt.Errorf("invalid skill at index %d: %w", i, err)
		}
	}

	// Validate EducationRequirements
	for i, edu := range req.EducationRequirements {
		if err := edu.Validate(); err != nil {
			return fmt.Errorf("invalid education requirement at index %d: %w", i, err)
		}
	}

	// Validate ExperienceRequirements
	for i, exp := range req.ExperienceRequirements {
		if err := exp.Validate(); err != nil {
			return fmt.Errorf("invalid experience requirement at index %d: %w", i, err)
		}
	}

	return nil
}

// JobResponsibilityInput is used for upserting responsibilities.
type JobResponsibilityInput struct {
	ID             *string `json:"id,omitempty"`
	Responsibility string  `json:"responsibility" validate:"required"`
}

// JobSkillInput is used for upserting skills.
type JobSkillInput struct {
	ID      *string `json:"id,omitempty"`
	SkillID string  `json:"skill_id" validate:"required"`
	Type    string  `json:"type" validate:"required" enums:"required,bonus"` // 技能类型: 必需技能/加分技能
}

// Validate validates the JobSkillInput fields against constants
func (input *JobSkillInput) Validate() error {
	if input.Type != "" {
		skillType := consts.JobSkillType(input.Type)
		if !skillType.IsValid() {
			return fmt.Errorf("invalid skill type: %s", input.Type)
		}
	}
	return nil
}

// JobEducationRequirementInput upserts education expectations.
type JobEducationRequirementInput struct {
	ID            *string `json:"id,omitempty"`
	EducationType string  `json:"education_type" validate:"required" enums:"unlimited,junior_college,bachelor,master,doctor"` // 学历要求: 不限/大专/本科/硕士/博士
}

// Validate validates the JobEducationRequirementInput fields against constants
func (input *JobEducationRequirementInput) Validate() error {
	if input.EducationType != "" {
		educationType := consts.JobEducationType(input.EducationType)
		if !educationType.IsValid() {
			return fmt.Errorf("invalid education type: %s", input.EducationType)
		}
	}
	return nil
}

// JobExperienceRequirementInput upserts experienc expectations.
type JobExperienceRequirementInput struct {
	ID             *string `json:"id,omitempty"`
	ExperienceType string  `json:"experience_type" validate:"required" enums:"unlimited,fresh_graduate,under_one_year,one_to_three_years,three_to_five_years,five_to_ten_years,over_ten_years"` // 工作经验: 不限/应届毕业生/1年以下/1-3年/3-5年/5-10年/10年以上
	MinYears       int     `json:"min_years"`
	IdealYears     int     `json:"ideal_years"`
}

// Validate validates the JobExperienceRequirementInput fields against constants
func (input *JobExperienceRequirementInput) Validate() error {
	if input.ExperienceType != "" {
		experienceType := consts.JobExperienceType(input.ExperienceType)
		if !experienceType.IsValid() {
			return fmt.Errorf("invalid experience type: %s", input.ExperienceType)
		}
	}
	return nil
}

// JobIndustryRequirementInput upserts industry expectations.
type JobIndustryRequirementInput struct {
	ID          *string `json:"id,omitempty"`
	Industry    string  `json:"industry" validate:"required"`
	CompanyName *string `json:"company_name,omitempty"`
}

// ListJobProfileReq filters job profile listing.
type ListJobProfileReq struct {
	web.Pagination

	Keyword      *string  `json:"keyword,omitempty" query:"keyword"`
	DepartmentID *string  `json:"department_id,omitempty" query:"department_id"`
	SkillIDs     []string `json:"skill_ids,omitempty" query:"skill_ids"`
	Locations    []string `json:"locations,omitempty" query:"locations"`
}

// ListJobProfileRepoReq is used by repository to translate filters.
type ListJobProfileRepoReq struct {
	*ListJobProfileReq
}

// ListJobProfileResp contains paginated job profiles.
type ListJobProfileResp struct {
	Items    []*JobProfile `json:"items"`
	PageInfo *db.PageInfo  `json:"page_info"`
}

// SearchJobProfileReq defines search filters for job profiles.
type SearchJobProfileReq struct {
	web.Pagination

	Keyword      *string  `json:"keyword,omitempty" query:"keyword"`
	DepartmentID *string  `json:"department_id,omitempty" query:"department_id"`
	SkillIDs     []string `json:"skill_ids,omitempty" query:"skill_ids"`
	Locations    []string `json:"locations,omitempty" query:"locations"`
	SalaryMin    *float64 `json:"salary_min,omitempty" query:"salary_min"`
	SalaryMax    *float64 `json:"salary_max,omitempty" query:"salary_max"`
}

// SearchJobProfileRepoReq is used by repository to translate search filters.
type SearchJobProfileRepoReq struct {
	*SearchJobProfileReq
}

// SearchJobProfileResp contains search results for job profiles.
type SearchJobProfileResp struct {
	Items    []*JobProfile `json:"items"`
	PageInfo *db.PageInfo  `json:"page_info"`
}

// CreateSkillMetaReq defines payload for skill meta creation.
type CreateSkillMetaReq struct {
	Name string `json:"name" validate:"required"`
}

// ListSkillMetaReq defines pagination and filters for skill dictionary.
type ListSkillMetaReq struct {
	web.Pagination
	Keyword *string `json:"keyword,omitempty" query:"keyword"`
}

// ListSkillMetaResp contains skill meta results.
type ListSkillMetaResp struct {
	Items    []*JobSkillMeta `json:"items"`
	PageInfo *db.PageInfo    `json:"page_info"`
}

// JobProfileRelatedEntities groups related slices for create flow.
type JobProfileRelatedEntities struct {
	Responsibilities       []*db.JobResponsibility
	Skills                 []*db.JobSkill
	EducationRequirements  []*db.JobEducationRequirement
	ExperienceRequirements []*db.JobExperienceRequirement
	IndustryRequirements   []*db.JobIndustryRequirement
}

func ptrString(v string) *string { return &v }

func ptrFloat64(v float64) *float64 { return &v }
