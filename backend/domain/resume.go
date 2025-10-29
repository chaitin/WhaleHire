package domain

import (
	"context"
	"io"
	"time"

	"github.com/chaitin/WhaleHire/backend/pkg/web"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
)

// ResumeUsecase 简历业务逻辑接口
type ResumeUsecase interface {
	// 简历上传
	Upload(ctx context.Context, req *UploadResumeReq) (*Resume, error)

	// 简历管理
	GetByID(ctx context.Context, id string) (*ResumeDetail, error)
	List(ctx context.Context, req *ListResumeReq) (*ListResumeResp, error)
	Update(ctx context.Context, req *UpdateResumeReq) (*Resume, error)
	Delete(ctx context.Context, id string) error

	// 搜索功能
	Search(ctx context.Context, req *SearchResumeReq) (*SearchResumeResp, error)

	// 状态管理
	UpdateStatus(ctx context.Context, id string, status ResumeStatus) error

	// 重新解析
	Reparse(ctx context.Context, id string) error

	// 获取解析进度
	GetParseProgress(ctx context.Context, id string) (*ResumeParseProgress, error)

	// 批量上传相关接口
	BatchUpload(ctx context.Context, req *BatchUploadResumeReq) (*BatchUploadTask, error)
	GetBatchUploadStatus(ctx context.Context, taskID string) (*BatchUploadTask, error)
	CancelBatchUpload(ctx context.Context, taskID string) error
}

// ResumeRepo 简历数据访问接口
type ResumeRepo interface {
	// 基础CRUD
	Create(ctx context.Context, resume *db.Resume) (*db.Resume, error)
	GetByID(ctx context.Context, id string) (*db.Resume, error)
	Update(ctx context.Context, id string, fn func(*db.Tx, *db.Resume, *db.ResumeUpdateOne) error) (*db.Resume, error)
	Delete(ctx context.Context, id string) error

	// 查询功能
	List(ctx context.Context, req *ListResumeReq) ([]*db.Resume, *db.PageInfo, error)
	Search(ctx context.Context, req *SearchResumeReq) ([]*db.Resume, *db.PageInfo, error)

	// 关联数据操作
	CreateEducation(ctx context.Context, education *db.ResumeEducation) (*db.ResumeEducation, error)
	CreateExperience(ctx context.Context, experience *db.ResumeExperience) (*db.ResumeExperience, error)
	CreateSkill(ctx context.Context, skill *db.ResumeSkill) (*db.ResumeSkill, error)
	CreateProject(ctx context.Context, project *db.ResumeProject) (*db.ResumeProject, error)

	// 获取关联数据
	GetEducationsByResumeID(ctx context.Context, resumeID string) ([]*db.ResumeEducation, error)
	GetExperiencesByResumeID(ctx context.Context, resumeID string) ([]*db.ResumeExperience, error)
	GetSkillsByResumeID(ctx context.Context, resumeID string) ([]*db.ResumeSkill, error)
	GetProjectsByResumeID(ctx context.Context, resumeID string) ([]*db.ResumeProject, error)
	GetLogsByResumeID(ctx context.Context, resumeID string) ([]*db.ResumeLog, error)

	// 日志记录
	CreateLog(ctx context.Context, log *db.ResumeLog) (*db.ResumeLog, error)

	// 文档解析结果
	CreateDocumentParse(ctx context.Context, docParse *db.ResumeDocumentParse) (*db.ResumeDocumentParse, error)
	GetDocumentParseByResumeID(ctx context.Context, resumeID string) (*db.ResumeDocumentParse, error)

	// 状态更新
	UpdateStatus(ctx context.Context, id string, status ResumeStatus) error
	UpdateErrorMessage(ctx context.Context, id string, errorMsg string) error

	// 数据清理
	ClearParsedData(ctx context.Context, resumeID string) error
}

// ParserService LLM解析服务接口
type ParserService interface {
	ParseResume(ctx context.Context, resumeID string, fileURL string, fileType string) (*ParsedResumeData, error)
}

// StorageService 文件存储服务接口
type StorageService interface {
	Upload(ctx context.Context, file io.Reader, filename string) (*FileInfo, error)
	Download(ctx context.Context, url string) (io.Reader, error)
	Delete(ctx context.Context, url string) error
	GetLocalPath(url string) (string, error)
}

// ResumeStatus 简历状态枚举
type ResumeStatus string

const (
	ResumeStatusPending    ResumeStatus = "pending"    // 已上传，等待解析
	ResumeStatusProcessing ResumeStatus = "processing" // 正在解析中
	ResumeStatusCompleted  ResumeStatus = "completed"  // 解析完成
	ResumeStatusFailed     ResumeStatus = "failed"     // 解析失败
	ResumeStatusArchived   ResumeStatus = "archived"   // 已归档
)

// UploadResumeReq 上传简历请求
type UploadResumeReq struct {
	UploaderID     string    `json:"uploader_id" validate:"required"`
	File           io.Reader `json:"-"`
	Filename       string    `json:"filename" validate:"required"`
	JobPositionIDs []string  `json:"job_position_ids,omitempty"`  // 关联的岗位ID列表
	Source         *string   `json:"source"  validate:"required"` // 申请来源，可选值：email（邮箱采集）、manual（手动上传）
	Notes          *string   `json:"notes,omitempty"`             // 备注信息
}

// ListResumeReq 简历列表请求
type ListResumeReq struct {
	web.Pagination

	UploaderID    *string       `json:"uploader_id,omitempty" query:"uploader_id"`
	Status        *ResumeStatus `json:"status,omitempty" query:"status"`
	Search        *string       `json:"search,omitempty" query:"search"`
	JobPositionID *string       `json:"job_position_id,omitempty" query:"job_position_id"` // 按岗位ID筛选
	Source        *string       `json:"source,omitempty" query:"source"`
	DateFrom      *time.Time    `json:"date_from,omitempty" query:"date_from"`
	DateTo        *time.Time    `json:"date_to,omitempty" query:"date_to"`
}

// SearchResumeReq 搜索简历请求
type SearchResumeReq struct {
	web.Pagination

	UploaderID      *string  `json:"uploader_id,omitempty" query:"uploader_id"`
	Keywords        *string  `json:"keywords,omitempty" query:"keywords"`
	Skills          []string `json:"skills,omitempty" query:"skills"`
	Education       *string  `json:"education,omitempty" query:"education"`
	YearsExperience *float64 `json:"years_experience,omitempty" query:"years_experience"`
	CurrentCity     *string  `json:"current_city,omitempty" query:"current_city"`
	Gender          *string  `json:"gender,omitempty" query:"gender"`
	AgeFrom         *int     `json:"age_from,omitempty" query:"age_from"`
	AgeTo           *int     `json:"age_to,omitempty" query:"age_to"`
}

// UpdateResumeReq 更新简历请求
type UpdateResumeReq struct {
	ID                 string                    `json:"-"`
	Name               *string                   `json:"name,omitempty"`
	Gender             *string                   `json:"gender,omitempty"`
	Birthday           *time.Time                `json:"birthday,omitempty"`
	Age                *int                      `json:"age,omitempty"` // 年龄
	Email              *string                   `json:"email,omitempty"`
	Phone              *string                   `json:"phone,omitempty"`
	CurrentCity        *string                   `json:"current_city,omitempty"`
	HighestEducation   *string                   `json:"highest_education,omitempty"`
	YearsExperience    *float64                  `json:"years_experience,omitempty"`
	PersonalSummary    *string                   `json:"personal_summary,omitempty"`    // 个人总结
	ExpectedSalary     *string                   `json:"expected_salary,omitempty"`     // 期望薪资
	ExpectedCity       *string                   `json:"expected_city,omitempty"`       // 期望工作城市
	AvailableDate      *time.Time                `json:"available_date,omitempty"`      // 可入职时间
	HonorsCertificates *string                   `json:"honors_certificates,omitempty"` // 荣誉证书
	OtherInfo          *string                   `json:"other_info,omitempty"`          // 其他信息
	Educations         []*UpdateResumeEducation  `json:"educations,omitempty"`
	Experiences        []*UpdateResumeExperience `json:"experiences,omitempty"`
	Skills             []*UpdateResumeSkill      `json:"skills,omitempty"`
	Projects           []*UpdateResumeProject    `json:"projects,omitempty"`
	// 岗位关联更新
	JobPositionIDs  []string                     `json:"job_position_ids,omitempty"` // 关联的岗位ID列表
	JobApplications []*JobApplicationAssociation `json:"job_applications,omitempty"` // 岗位申请详细信息
}

// UpdateResumeEducation 更新教育经历请求
type UpdateResumeEducation struct {
	ID              *string                  `json:"id,omitempty"`               // 教育经历ID，更新时必填
	Action          string                   `json:"action"`                     // 操作类型：create, update, delete
	School          *string                  `json:"school,omitempty"`           // 学校
	Degree          *string                  `json:"degree,omitempty"`           // 学位
	Major           *string                  `json:"major,omitempty"`            // 专业
	GPA             *float64                 `json:"gpa,omitempty"`              // 在校学分绩点
	StartDate       *time.Time               `json:"start_date,omitempty"`       // 开始时间
	EndDate         *time.Time               `json:"end_date,omitempty"`         // 结束时间
	UniversityTypes *[]consts.UniversityType `json:"university_types,omitempty"` // 学校类型
}

// UpdateResumeExperience 更新工作经历请求
type UpdateResumeExperience struct {
	ID             *string                `json:"id,omitempty"`              // 工作经历ID，更新时必填
	Action         string                 `json:"action"`                    // 操作类型：create, update, delete
	Company        *string                `json:"company,omitempty"`         // 公司
	Position       *string                `json:"position,omitempty"`        // 职位
	Title          *string                `json:"title,omitempty"`           // 担任岗位
	StartDate      *time.Time             `json:"start_date,omitempty"`      // 开始时间
	EndDate        *time.Time             `json:"end_date,omitempty"`        // 结束时间
	Description    *string                `json:"description,omitempty"`     // 工作描述
	ExperienceType *consts.ExperienceType `json:"experience_type,omitempty"` // 经历类型：work, organization, volunteer, internship
}

// UpdateResumeSkill 更新技能请求
type UpdateResumeSkill struct {
	ID          *string `json:"id,omitempty"`          // 技能ID，更新时必填
	Action      string  `json:"action"`                // 操作类型：create, update, delete
	SkillName   *string `json:"skill_name,omitempty"`  // 技能名称
	Level       *string `json:"level,omitempty"`       // 技能水平
	Description *string `json:"description,omitempty"` // 技能描述
}

// UpdateResumeProject 更新项目经验请求
type UpdateResumeProject struct {
	ID               *string             `json:"id,omitempty"`               // 项目经验ID，更新时必填
	Action           string              `json:"action"`                     // 操作类型：create, update, delete
	Name             *string             `json:"name,omitempty"`             // 项目名称
	Role             *string             `json:"role,omitempty"`             // 担任角色
	Company          *string             `json:"company,omitempty"`          // 所属公司
	StartDate        *time.Time          `json:"start_date,omitempty"`       // 开始时间
	EndDate          *time.Time          `json:"end_date,omitempty"`         // 结束时间
	Description      *string             `json:"description,omitempty"`      // 项目描述
	Responsibilities *string             `json:"responsibilities,omitempty"` // 主要职责
	Achievements     *string             `json:"achievements,omitempty"`     // 项目成果
	Technologies     *string             `json:"technologies,omitempty"`     // 使用技术
	ProjectURL       *string             `json:"project_url,omitempty"`      // 项目链接
	ProjectType      *consts.ProjectType `json:"project_type,omitempty"`     // 项目类型: personal/team/opensource/paper/other
}

// Resume 简历基本信息
type Resume struct {
	ID                 string            `json:"id"`
	UploaderID         string            `json:"uploader_id"`
	UploaderName       string            `json:"uploader_name"`
	Name               string            `json:"name"`
	Gender             string            `json:"gender"`
	Birthday           *time.Time        `json:"birthday,omitempty"`
	Age                *int              `json:"age,omitempty"` // 年龄
	Email              string            `json:"email"`
	Phone              string            `json:"phone"`
	CurrentCity        string            `json:"current_city"`
	HighestEducation   string            `json:"highest_education"`
	YearsExperience    float64           `json:"years_experience"`
	PersonalSummary    string            `json:"personal_summary,omitempty"`    // 个人总结
	ExpectedSalary     string            `json:"expected_salary,omitempty"`     // 期望薪资
	ExpectedCity       string            `json:"expected_city,omitempty"`       // 期望工作城市
	AvailableDate      *time.Time        `json:"available_date,omitempty"`      // 可入职时间
	HonorsCertificates string            `json:"honors_certificates,omitempty"` // 荣誉证书
	OtherInfo          string            `json:"other_info,omitempty"`          // 其他信息
	ResumeFileURL      string            `json:"resume_file_url"`
	Status             ResumeStatus      `json:"status"`
	ErrorMessage       string            `json:"error_message,omitempty"`
	ParsedAt           *time.Time        `json:"parsed_at,omitempty"`
	JobPositions       []*JobApplication `json:"job_positions,omitempty"` // 关联的岗位信息
	CreatedAt          int64             `json:"created_at"`
	UpdatedAt          int64             `json:"updated_at"`
}

func (r *Resume) From(e *db.Resume) *Resume {
	if e == nil {
		return r
	}
	r.ID = e.ID.String()
	r.UploaderID = e.UploaderID.String()
	// 填充上传人姓名
	if e.Edges.User != nil {
		r.UploaderName = e.Edges.User.Username
	}
	r.Name = e.Name
	r.Gender = e.Gender
	if !e.Birthday.IsZero() {
		r.Birthday = &e.Birthday
	}
	r.Age = &e.Age
	r.Email = e.Email
	r.Phone = e.Phone
	r.CurrentCity = e.CurrentCity
	r.HighestEducation = e.HighestEducation
	r.YearsExperience = e.YearsExperience
	r.PersonalSummary = e.PersonalSummary
	r.ExpectedSalary = e.ExpectedSalary
	r.ExpectedCity = e.ExpectedCity
	if !e.AvailableDate.IsZero() {
		r.AvailableDate = &e.AvailableDate
	}
	r.HonorsCertificates = e.HonorsCertificates
	r.OtherInfo = e.OtherInfo
	r.ResumeFileURL = e.ResumeFileURL
	r.Status = ResumeStatus(e.Status)
	r.ErrorMessage = e.ErrorMessage
	if !e.ParsedAt.IsZero() {
		r.ParsedAt = &e.ParsedAt
	}
	r.CreatedAt = e.CreatedAt.Unix()
	r.UpdatedAt = e.UpdatedAt.Unix()
	return r
}

// ResumeDetail 简历详情（包含关联数据）
type ResumeDetail struct {
	*Resume
	Educations  []*ResumeEducation  `json:"educations"`
	Experiences []*ResumeExperience `json:"experiences"`
	Skills      []*ResumeSkill      `json:"skills"`
	Projects    []*ResumeProject    `json:"projects"`
	Logs        []*ResumeLog        `json:"logs"`
}

// ResumeEducation 教育经历
type ResumeEducation struct {
	ID              string                  `json:"id"`
	ResumeID        string                  `json:"resume_id"`
	School          string                  `json:"school"`
	Degree          string                  `json:"degree"`
	Major           string                  `json:"major"`
	GPA             *float64                `json:"gpa,omitempty"` // 在校学分绩点
	StartDate       *time.Time              `json:"start_date,omitempty"`
	EndDate         *time.Time              `json:"end_date,omitempty"`
	UniversityTypes []consts.UniversityType `json:"university_types"`
	CreatedAt       int64                   `json:"created_at"`
	UpdatedAt       int64                   `json:"updated_at"`
}

func (e *ResumeEducation) From(entity *db.ResumeEducation) *ResumeEducation {
	if entity == nil {
		return nil
	}

	var gpa *float64
	if entity.Gpa != 0 {
		gpa = &entity.Gpa
	}

	var startDate, endDate *time.Time
	if !entity.StartDate.IsZero() {
		startDate = &entity.StartDate
	}
	if !entity.EndDate.IsZero() {
		endDate = &entity.EndDate
	}

	return &ResumeEducation{
		ID:              entity.ID.String(),
		ResumeID:        entity.ResumeID.String(),
		School:          entity.School,
		Degree:          entity.Degree,
		Major:           entity.Major,
		GPA:             gpa,
		StartDate:       startDate,
		EndDate:         endDate,
		UniversityTypes: entity.UniversityTypes,
		CreatedAt:       entity.CreatedAt.Unix(),
		UpdatedAt:       entity.UpdatedAt.Unix(),
	}
}

// ResumeExperience 工作经历
type ResumeExperience struct {
	ID             string                `json:"id"`
	ResumeID       string                `json:"resume_id"`
	Company        string                `json:"company"`
	Position       string                `json:"position"`
	Title          string                `json:"title"`
	StartDate      *time.Time            `json:"start_date,omitempty"`
	EndDate        *time.Time            `json:"end_date,omitempty"`
	Description    string                `json:"description"`
	ExperienceType consts.ExperienceType `json:"experience_type,omitempty"` // work, organization, volunteer, internship
	CreatedAt      int64                 `json:"created_at"`
	UpdatedAt      int64                 `json:"updated_at"`
}

func (e *ResumeExperience) From(entity *db.ResumeExperience) *ResumeExperience {
	if entity == nil {
		return nil
	}

	var startDate, endDate *time.Time
	if !entity.StartDate.IsZero() {
		startDate = &entity.StartDate
	}
	if !entity.EndDate.IsZero() {
		endDate = &entity.EndDate
	}

	return &ResumeExperience{
		ID:             entity.ID.String(),
		ResumeID:       entity.ResumeID.String(),
		Company:        entity.Company,
		Position:       entity.Position,
		Title:          entity.Title,
		StartDate:      startDate,
		EndDate:        endDate,
		Description:    entity.Description,
		ExperienceType: entity.ExperienceType,
		CreatedAt:      entity.CreatedAt.Unix(),
		UpdatedAt:      entity.UpdatedAt.Unix(),
	}
}

// ResumeSkill 技能
type ResumeSkill struct {
	ID          string `json:"id"`
	ResumeID    string `json:"resume_id"`
	SkillName   string `json:"skill_name"`
	Level       string `json:"level"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

func (s *ResumeSkill) From(entity *db.ResumeSkill) *ResumeSkill {
	if entity == nil {
		return s
	}
	s.ID = entity.ID.String()
	s.ResumeID = entity.ResumeID.String()
	s.SkillName = entity.SkillName
	s.Level = entity.Level
	s.Description = entity.Description
	s.CreatedAt = entity.CreatedAt.Unix()
	s.UpdatedAt = entity.UpdatedAt.Unix()
	return s
}

// ResumeProject 项目经验
type ResumeProject struct {
	ID               string             `json:"id"`
	ResumeID         string             `json:"resume_id"`
	Name             string             `json:"name"`
	Role             string             `json:"role"`
	Company          string             `json:"company"`
	StartDate        *time.Time         `json:"start_date,omitempty"`
	EndDate          *time.Time         `json:"end_date,omitempty"`
	Description      string             `json:"description"`
	Responsibilities string             `json:"responsibilities"`
	Achievements     string             `json:"achievements"`
	Technologies     string             `json:"technologies"`
	ProjectURL       string             `json:"project_url"`
	ProjectType      consts.ProjectType `json:"project_type"` // 项目类型: personal/team/opensource/paper/other
	CreatedAt        int64              `json:"created_at"`
	UpdatedAt        int64              `json:"updated_at"`
}

func (p *ResumeProject) From(entity *db.ResumeProject) *ResumeProject {
	if entity == nil {
		return p
	}
	p.ID = entity.ID.String()
	p.ResumeID = entity.ResumeID.String()
	p.Name = entity.Name
	p.Role = entity.Role
	p.Company = entity.Company
	if !entity.StartDate.IsZero() {
		p.StartDate = &entity.StartDate
	}
	if !entity.EndDate.IsZero() {
		p.EndDate = &entity.EndDate
	}
	p.Description = entity.Description
	p.Responsibilities = entity.Responsibilities
	p.Achievements = entity.Achievements
	p.Technologies = entity.Technologies
	p.ProjectURL = entity.ProjectURL
	p.ProjectType = entity.ProjectType
	p.CreatedAt = entity.CreatedAt.Unix()
	p.UpdatedAt = entity.UpdatedAt.Unix()
	return p
}

// ResumeLog 简历日志操作日志
type ResumeLog struct {
	ID        string `json:"id"`
	ResumeID  string `json:"resume_id"`
	Action    string `json:"action"`
	Message   string `json:"message"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

func (l *ResumeLog) From(entity *db.ResumeLog) *ResumeLog {
	if entity == nil {
		return l
	}
	l.ID = entity.ID.String()
	l.ResumeID = entity.ResumeID.String()
	l.Action = entity.Action
	l.Message = entity.Message
	l.CreatedAt = entity.CreatedAt.Unix()
	l.UpdatedAt = entity.UpdatedAt.Unix()
	return l
}

// ListResumeResp 简历列表响应
type ListResumeResp struct {
	*db.PageInfo
	Resumes []*Resume `json:"resumes"`
}

// SearchResumeResp 搜索简历响应
type SearchResumeResp struct {
	*db.PageInfo
	Resumes []*Resume `json:"resumes"`
}

// GetResumesByJobPositionIDReq 根据岗位ID获取简历请求
type GetResumesByJobPositionIDReq struct {
	web.Pagination

	JobPositionID string     `json:"job_position_id,omitempty" query:"job_position_id"` // 岗位ID
	Status        *string    `json:"status,omitempty" query:"status"`                   // 申请状态筛选
	Source        *string    `json:"source,omitempty" query:"source"`                   // 申请来源筛选
	DateFrom      *time.Time `json:"date_from,omitempty" query:"date_from"`             // 申请开始时间
	DateTo        *time.Time `json:"date_to,omitempty" query:"date_to"`                 // 申请结束时间
}

// GetResumesByJobPositionIDResp 根据岗位ID获取简历信息响应
type GetResumesByJobPositionIDResp struct {
	*db.PageInfo
	Resumes []*Resume `json:"resumes"`
}

// ParsedResumeData LLM解析后的简历数据
type ParsedResumeData struct {
	BasicInfo   *ParsedBasicInfo    `json:"basic_info"`
	Educations  []*ParsedEducation  `json:"educations"`
	Experiences []*ParsedExperience `json:"experiences"`
	Skills      []*ParsedSkill      `json:"skills"`
	Projects    []*ParsedProject    `json:"projects"`
}

// ParsedBasicInfo 解析的基本信息
type ParsedBasicInfo struct {
	Name               string     `json:"name"`
	Phone              string     `json:"phone"`
	Email              string     `json:"email"`
	Gender             string     `json:"gender"`
	Birthday           *time.Time `json:"birthday,omitempty"`
	Age                int        `json:"age,omitempty"` // 年龄
	CurrentCity        string     `json:"current_city"`
	HighestEducation   string     `json:"highest_education"`
	YearsExperience    float64    `json:"years_experience"`
	PersonalSummary    string     `json:"personal_summary,omitempty"`    // 个人总结
	ExpectedSalary     string     `json:"expected_salary,omitempty"`     // 期望薪资
	ExpectedCity       string     `json:"expected_city,omitempty"`       // 期望工作城市
	AvailableDate      *time.Time `json:"available_date,omitempty"`      // 可入职时间
	HonorsCertificates string     `json:"honors_certificates,omitempty"` // 荣誉证书
	OtherInfo          string     `json:"other_info,omitempty"`          // 其他信息
}

// ParsedEducation 解析的教育经历
type ParsedEducation struct {
	School                string                       `json:"school"`
	Major                 string                       `json:"major"`
	Degree                string                       `json:"degree"`
	StartDate             *time.Time                   `json:"start_date,omitempty"`
	EndDate               *time.Time                   `json:"end_date,omitempty"`
	GPA                   *float64                     `json:"gpa,omitempty"`
	UniversityTags        []consts.UniversityType      `json:"university_tags,omitempty"`
	UniversityMatchSource consts.UniversityMatchSource `json:"university_match_source,omitempty"`
	UniversityMatchScore  *float64                     `json:"university_match_score,omitempty"`
}

// ParsedExperience 解析的工作经历
type ParsedExperience struct {
	Company        string                `json:"company"`
	Position       string                `json:"position"`
	Title          string                `json:"title"`
	StartDate      *time.Time            `json:"start_date,omitempty"`
	EndDate        *time.Time            `json:"end_date,omitempty"`
	Description    string                `json:"description"`
	Achievements   string                `json:"achievements,omitempty"`
	ExperienceType consts.ExperienceType `json:"experience_type,omitempty"` // work, organization, volunteer, internship
}

// ParsedSkill 解析的技能信息
type ParsedSkill struct {
	Name        string `json:"name"`
	Level       string `json:"level"`
	Description string `json:"description,omitempty"`
}

// ParsedProject 解析的项目经验信息
type ParsedProject struct {
	Name             string             `json:"name"`
	Role             string             `json:"role"`
	Company          string             `json:"company,omitempty"`
	StartDate        *time.Time         `json:"start_date,omitempty"`
	EndDate          *time.Time         `json:"end_date,omitempty"`
	Description      string             `json:"description"`
	Responsibilities string             `json:"responsibilities,omitempty"`
	Achievements     string             `json:"achievements,omitempty"`
	Technologies     string             `json:"technologies,omitempty"`
	ProjectURL       string             `json:"project_url,omitempty"`
	ProjectType      consts.ProjectType `json:"project_type,omitempty"`
}

// FileInfo 文件信息
type FileInfo struct {
	OriginalName string `json:"original_name"`
	StoredName   string `json:"stored_name"`
	FilePath     string `json:"file_path"`
	FileURL      string `json:"file_url"`
	Size         int64  `json:"size"`
	ContentType  string `json:"content_type"`
}

// ResumeParseProgress 简历解析进度信息
type ResumeParseProgress struct {
	ResumeID     string       `json:"resume_id"`
	Status       ResumeStatus `json:"status"`
	Progress     int          `json:"progress"`      // 解析进度百分比 0-100
	Message      string       `json:"message"`       // 当前状态描述
	ErrorMessage string       `json:"error_message"` // 错误信息（如果有）
	StartedAt    *time.Time   `json:"started_at"`    // 开始解析时间
	CompletedAt  *time.Time   `json:"completed_at"`  // 完成解析时间
}

// 批量上传相关数据结构
type BatchUploadResumeReq struct {
	UploaderID     string                 `json:"uploader_id" validate:"required"`
	Files          []*BatchUploadFileInfo `json:"files" validate:"required,min=1"`
	JobPositionIDs []string               `json:"job_position_ids,omitempty"`  // 关联的岗位ID列表
	Source         *string                `json:"source"  validate:"required"` // 申请来源，可选值：email（邮箱采集）、manual（手动上传）
	Notes          *string                `json:"notes,omitempty"`             // 备注信息
}

type BatchUploadFileInfo struct {
	File     io.Reader `json:"-"`
	Filename string    `json:"filename" validate:"required"`
}

type BatchUploadTask struct {
	TaskID         string             `json:"task_id"`
	UploaderID     string             `json:"uploader_id"`
	Status         BatchUploadStatus  `json:"status"`
	TotalCount     int                `json:"total_count"`
	CompletedCount int                `json:"completed_count"`
	SuccessCount   int                `json:"success_count"`
	FailedCount    int                `json:"failed_count"`
	JobPositionIDs []string           `json:"job_position_ids,omitempty"`
	Source         *string            `json:"source,omitempty"`
	Notes          *string            `json:"notes,omitempty"`
	Items          []*BatchUploadItem `json:"items"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	CompletedAt    *time.Time         `json:"completed_at,omitempty"`
}

type BatchUploadItem struct {
	ItemID       string            `json:"item_id"`
	Filename     string            `json:"filename"`
	Status       BatchUploadStatus `json:"status"`
	ResumeID     *string           `json:"resume_id,omitempty"`
	ErrorMessage *string           `json:"error_message,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	CompletedAt  *time.Time        `json:"completed_at,omitempty"`
}

type BatchUploadStatus string

const (
	BatchUploadStatusPending    BatchUploadStatus = "pending"    // 等待处理
	BatchUploadStatusProcessing BatchUploadStatus = "processing" // 正在处理
	BatchUploadStatusCompleted  BatchUploadStatus = "completed"  // 已完成
	BatchUploadStatusFailed     BatchUploadStatus = "failed"     // 失败
	BatchUploadStatusCancelled  BatchUploadStatus = "cancelled"  // 已取消
)

// ResumeDocumentParse 文档解析结果
type ResumeDocumentParse struct {
	ID        string    `json:"id"`
	ResumeID  string    `json:"resume_id"`
	FileID    string    `json:"file_id"`
	Content   string    `json:"content"`
	FileType  string    `json:"file_type"`
	Filename  string    `json:"filename"`
	Title     string    `json:"title"`
	UploadAt  time.Time `json:"upload_at"`
	Status    string    `json:"status"`
	CreatedAt int64     `json:"created_at"`
	UpdatedAt int64     `json:"updated_at"`
}

func (d *ResumeDocumentParse) From(entity *db.ResumeDocumentParse) *ResumeDocumentParse {
	if entity == nil {
		return d
	}
	d.ID = entity.ID.String()
	d.ResumeID = entity.ResumeID.String()
	d.FileID = entity.FileID
	d.Content = entity.Content
	d.FileType = entity.FileType
	d.Filename = entity.Filename
	d.Title = entity.Title
	d.UploadAt = entity.UploadAt
	d.Status = entity.Status
	d.CreatedAt = entity.CreatedAt.Unix()
	d.UpdatedAt = entity.UpdatedAt.Unix()
	return d
}
