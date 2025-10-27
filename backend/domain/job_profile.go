package domain

import (
	"context"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/pkg/web"

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
	GetByIDs(ctx context.Context, ids []string) ([]*JobProfile, error)
	List(ctx context.Context, req *ListJobProfileReq) (*ListJobProfileResp, error)
	Search(ctx context.Context, req *SearchJobProfileReq) (*SearchJobProfileResp, error)

	// ParseJobProfile 解析岗位描述，返回结构化的岗位画像数据
	ParseJobProfile(ctx context.Context, req *ParseJobProfileReq) (*ParseJobProfileResp, error)

	// PolishPrompt AI 润色岗位需求描述
	PolishPrompt(ctx context.Context, req *PolishJobPromptReq) (*PolishJobPromptResp, error)

	// GenerateByPrompt 基于润色后的 Prompt 生成岗位画像
	GenerateByPrompt(ctx context.Context, req *GenerateJobProfileReq) (*GenerateJobProfileResp, error)

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
	GetByIDs(ctx context.Context, ids []string) ([]*db.JobPosition, error)
	List(ctx context.Context, req *ListJobProfileRepoReq) ([]*db.JobPosition, *db.PageInfo, error)
	Search(ctx context.Context, req *SearchJobProfileRepoReq) ([]*db.JobPosition, *db.PageInfo, error)
	HasRelatedResumes(ctx context.Context, id string) (bool, error)
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
	ID        *string `json:"id,omitempty"`
	SkillID   *string `json:"skill_id,omitempty"`                              // 技能ID，如果为空则根据SkillName自动创建新技能
	SkillName string  `json:"skill_name" validate:"required"`                  // 技能名称，用于前端显示
	Type      string  `json:"type" validate:"required" enums:"required,bonus"` // 技能类型: 必需技能/加分技能
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
	Items []*JobProfile `json:"items"`
	*db.PageInfo
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

// ParseJobProfileReq 解析岗位画像请求
// @Description 岗位画像解析请求参数
type ParseJobProfileReq struct {
	Description string `json:"description" validate:"required" example:"我们正在寻找一名资深的Go后端开发工程师，负责微服务架构设计和开发..."` // 岗位描述文本，用于AI解析
}

// ParseJobProfileResp 解析岗位画像响应
// @Description AI解析岗位描述后返回的结构化数据，包含岗位的基本信息和详细要求
type ParseJobProfileResp struct {
	// 基本信息
	Name      string   `json:"name" example:"Go后端开发工程师"`                                                                   // 岗位名称，从描述中提取的职位标题
	WorkType  *string  `json:"work_type,omitempty" example:"full_time" enums:"full_time,part_time,internship,outsourcing"` // 工作性质：全职/兼职/实习/外包
	Location  *string  `json:"location,omitempty" example:"北京"`                                                            // 工作地点，可能包含多个城市
	SalaryMin *float64 `json:"salary_min,omitempty" example:"15000"`                                                       // 最低薪资（月薪，单位：元）
	SalaryMax *float64 `json:"salary_max,omitempty" example:"25000"`                                                       // 最高薪资（月薪，单位：元）

	// 详细信息
	Responsibilities       []*JobResponsibilityInput        `json:"responsibilities"`        // 岗位职责列表，描述具体的工作内容和责任
	Skills                 []*JobSkillInput                 `json:"skills"`                  // 技能要求列表，包含必需技能和加分技能
	EducationRequirements  []*JobEducationRequirementInput  `json:"education_requirements"`  // 学历要求列表，可能包含多个学历层次
	ExperienceRequirements []*JobExperienceRequirementInput `json:"experience_requirements"` // 工作经验要求列表，包含年限和具体经验类型
	IndustryRequirements   []*JobIndustryRequirementInput   `json:"industry_requirements"`   // 行业背景要求列表，可能指定特定行业或公司经验
}

// PolishJobPromptReq AI 润色岗位需求描述请求
// @Description 用户输入的原始岗位需求，用于 AI 润色和结构化
type PolishJobPromptReq struct {
	Idea string `json:"idea" validate:"required" example:"我们需要招聘一名数据分析师"` // 用户原始想法或一句话描述
}

// PolishJobPromptResp AI 润色岗位需求描述响应
// @Description AI 润色后返回的结构化 Prompt 和建议
type PolishJobPromptResp struct {
	PolishedPrompt     string   `json:"polished_prompt" example:"我们正在寻找一名经验丰富的数据分析师..."`          // 完整且可直接用于生成链的 Prompt
	SuggestedTitle     string   `json:"suggested_title" example:"数据分析师"`                          // AI 推荐的岗位标题
	ResponsibilityTips []string `json:"responsibility_tips" example:"[\"负责数据收集和清洗\",\"制作数据报表\"]"` // 岗位职责提示
	RequirementTips    []string `json:"requirement_tips" example:"[\"熟练使用SQL\",\"具备统计学基础\"]"`     // 任职要求提示
	BonusTips          []string `json:"bonus_tips" example:"[\"有机器学习经验\",\"熟悉Python\"]"`          // 加分项提示
}

// GenerateJobProfileReq 基于 Prompt 生成岗位画像请求
// @Description 使用润色后的 Prompt 生成完整的岗位画像
type GenerateJobProfileReq struct {
	Prompt       string  `json:"prompt" validate:"required" example:"我们正在寻找一名经验丰富的数据分析师..."`             // 已润色的 Prompt
	DepartmentID *string `json:"department_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"` // 生成时可提前绑定部门
}

// GenerateJobProfileResp 基于 Prompt 生成岗位画像响应
// @Description AI 生成的完整岗位画像，包含结构化数据和 Markdown 描述
type GenerateJobProfileResp struct {
	Profile             *ParseJobProfileResp `json:"profile"`              // 结构化岗位画像
	DescriptionMarkdown string               `json:"description_markdown"` // Markdown 格式的岗位描述
}
