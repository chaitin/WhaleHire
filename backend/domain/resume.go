package domain

import (
	"context"
	"io"
	"time"

	"github.com/GoYoko/web"

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

	// 获取关联数据
	GetEducationsByResumeID(ctx context.Context, resumeID string) ([]*db.ResumeEducation, error)
	GetExperiencesByResumeID(ctx context.Context, resumeID string) ([]*db.ResumeExperience, error)
	GetSkillsByResumeID(ctx context.Context, resumeID string) ([]*db.ResumeSkill, error)
	GetLogsByResumeID(ctx context.Context, resumeID string) ([]*db.ResumeLog, error)

	// 日志记录
	CreateLog(ctx context.Context, log *db.ResumeLog) (*db.ResumeLog, error)

	// 文档解析结果
	CreateDocumentParse(ctx context.Context, docParse *db.ResumeDocumentParse) (*db.ResumeDocumentParse, error)
	GetDocumentParseByResumeID(ctx context.Context, resumeID string) (*db.ResumeDocumentParse, error)

	// 状态更新
	UpdateStatus(ctx context.Context, id string, status ResumeStatus) error
	UpdateErrorMessage(ctx context.Context, id string, errorMsg string) error
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
	UserID   string    `json:"user_id" validate:"required"`
	File     io.Reader `json:"-"`
	Filename string    `json:"filename" validate:"required"`
	JobID    *string   `json:"job_id,omitempty"`
	Source   *string   `json:"source,omitempty"`
}

// ListResumeReq 简历列表请求
type ListResumeReq struct {
	web.Pagination

	UserID   *string       `json:"user_id,omitempty" query:"user_id"`
	Status   *ResumeStatus `json:"status,omitempty" query:"status"`
	Search   *string       `json:"search,omitempty" query:"search"`
	JobID    *string       `json:"job_id,omitempty" query:"job_id"`
	Source   *string       `json:"source,omitempty" query:"source"`
	DateFrom *time.Time    `json:"date_from,omitempty" query:"date_from"`
	DateTo   *time.Time    `json:"date_to,omitempty" query:"date_to"`
}

// SearchResumeReq 搜索简历请求
type SearchResumeReq struct {
	web.Pagination

	UserID          *string  `json:"user_id,omitempty" query:"user_id"`
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
	ID               string                    `json:"-"`
	Name             *string                   `json:"name,omitempty"`
	Gender           *string                   `json:"gender,omitempty"`
	Birthday         *time.Time                `json:"birthday,omitempty"`
	Email            *string                   `json:"email,omitempty"`
	Phone            *string                   `json:"phone,omitempty"`
	CurrentCity      *string                   `json:"current_city,omitempty"`
	HighestEducation *string                   `json:"highest_education,omitempty"`
	YearsExperience  *float64                  `json:"years_experience,omitempty"`
	Educations       []*UpdateResumeEducation  `json:"educations,omitempty"`
	Experiences      []*UpdateResumeExperience `json:"experiences,omitempty"`
	Skills           []*UpdateResumeSkill      `json:"skills,omitempty"`
}

// UpdateResumeEducation 更新教育经历请求
type UpdateResumeEducation struct {
	ID        *string    `json:"id,omitempty"`         // 教育经历ID，更新时必填
	Action    string     `json:"action"`               // 操作类型：create, update, delete
	School    *string    `json:"school,omitempty"`     // 学校
	Degree    *string    `json:"degree,omitempty"`     // 学位
	Major     *string    `json:"major,omitempty"`      // 专业
	StartDate *time.Time `json:"start_date,omitempty"` // 开始时间
	EndDate   *time.Time `json:"end_date,omitempty"`   // 结束时间
}

// UpdateResumeExperience 更新工作经历请求
type UpdateResumeExperience struct {
	ID          *string    `json:"id,omitempty"`          // 工作经历ID，更新时必填
	Action      string     `json:"action"`                // 操作类型：create, update, delete
	Company     *string    `json:"company,omitempty"`     // 公司
	Position    *string    `json:"position,omitempty"`    // 职位
	Title       *string    `json:"title,omitempty"`       // 担任岗位
	StartDate   *time.Time `json:"start_date,omitempty"`  // 开始时间
	EndDate     *time.Time `json:"end_date,omitempty"`    // 结束时间
	Description *string    `json:"description,omitempty"` // 工作描述
}

// UpdateResumeSkill 更新技能请求
type UpdateResumeSkill struct {
	ID          *string `json:"id,omitempty"`          // 技能ID，更新时必填
	Action      string  `json:"action"`                // 操作类型：create, update, delete
	SkillName   *string `json:"skill_name,omitempty"`  // 技能名称
	Level       *string `json:"level,omitempty"`       // 技能水平
	Description *string `json:"description,omitempty"` // 技能描述
}

// Resume 简历信息
type Resume struct {
	ID               string       `json:"id"`
	UserID           string       `json:"user_id"`
	Name             string       `json:"name"`
	Gender           string       `json:"gender"`
	Birthday         *time.Time   `json:"birthday,omitempty"`
	Email            string       `json:"email"`
	Phone            string       `json:"phone"`
	CurrentCity      string       `json:"current_city"`
	HighestEducation string       `json:"highest_education"`
	YearsExperience  float64      `json:"years_experience"`
	ResumeFileURL    string       `json:"resume_file_url"`
	Status           ResumeStatus `json:"status"`
	ErrorMessage     string       `json:"error_message,omitempty"`
	ParsedAt         *time.Time   `json:"parsed_at,omitempty"`
	CreatedAt        int64        `json:"created_at"`
	UpdatedAt        int64        `json:"updated_at"`
}

func (r *Resume) From(e *db.Resume) *Resume {
	if e == nil {
		return r
	}
	r.ID = e.ID.String()
	r.UserID = e.UserID.String()
	r.Name = e.Name
	r.Gender = e.Gender
	if !e.Birthday.IsZero() {
		r.Birthday = &e.Birthday
	}
	r.Email = e.Email
	r.Phone = e.Phone
	r.CurrentCity = e.CurrentCity
	r.HighestEducation = e.HighestEducation
	r.YearsExperience = e.YearsExperience
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
	Logs        []*ResumeLog        `json:"logs"`
}

// ResumeEducation 教育经历
type ResumeEducation struct {
	ID        string     `json:"id"`
	ResumeID  string     `json:"resume_id"`
	School    string     `json:"school"`
	Degree    string     `json:"degree"`
	Major     string     `json:"major"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	CreatedAt int64      `json:"created_at"`
	UpdatedAt int64      `json:"updated_at"`
}

func (e *ResumeEducation) From(entity *db.ResumeEducation) *ResumeEducation {
	if entity == nil {
		return e
	}
	e.ID = entity.ID.String()
	e.ResumeID = entity.ResumeID.String()
	e.School = entity.School
	e.Degree = entity.Degree
	e.Major = entity.Major
	if !entity.StartDate.IsZero() {
		e.StartDate = &entity.StartDate
	}
	if !entity.EndDate.IsZero() {
		e.EndDate = &entity.EndDate
	}
	e.CreatedAt = entity.CreatedAt.Unix()
	e.UpdatedAt = entity.UpdatedAt.Unix()
	return e
}

// ResumeExperience 工作经历
type ResumeExperience struct {
	ID          string     `json:"id"`
	ResumeID    string     `json:"resume_id"`
	Company     string     `json:"company"`
	Position    string     `json:"position"`
	Title       string     `json:"title"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Description string     `json:"description"`
	CreatedAt   int64      `json:"created_at"`
	UpdatedAt   int64      `json:"updated_at"`
}

func (e *ResumeExperience) From(entity *db.ResumeExperience) *ResumeExperience {
	if entity == nil {
		return e
	}
	e.ID = entity.ID.String()
	e.ResumeID = entity.ResumeID.String()
	e.Company = entity.Company
	e.Position = entity.Position
	e.Title = entity.Title
	if !entity.StartDate.IsZero() {
		e.StartDate = &entity.StartDate
	}
	if !entity.EndDate.IsZero() {
		e.EndDate = &entity.EndDate
	}
	e.Description = entity.Description
	e.CreatedAt = entity.CreatedAt.Unix()
	e.UpdatedAt = entity.UpdatedAt.Unix()
	return e
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

// ResumeLog 操作日志
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

// ParsedResumeData LLM解析后的简历数据
type ParsedResumeData struct {
	BasicInfo   *ParsedBasicInfo    `json:"basic_info"`
	Educations  []*ParsedEducation  `json:"educations"`
	Experiences []*ParsedExperience `json:"experiences"`
	Skills      []*ParsedSkill      `json:"skills"`
}

// ParsedBasicInfo 解析的基本信息
type ParsedBasicInfo struct {
	Name             string     `json:"name"`
	Phone            string     `json:"phone"`
	Email            string     `json:"email"`
	Gender           string     `json:"gender"`
	Birthday         *time.Time `json:"birthday,omitempty"`
	CurrentCity      string     `json:"current_city"`
	HighestEducation string     `json:"highest_education"`
	YearsExperience  float64    `json:"years_experience"`
}

// ParsedEducation 解析的教育经历
type ParsedEducation struct {
	School    string     `json:"school"`
	Major     string     `json:"major"`
	Degree    string     `json:"degree"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	GPA       string     `json:"gpa,omitempty"`
}

// ParsedExperience 解析的工作经历
type ParsedExperience struct {
	Company      string     `json:"company"`
	Position     string     `json:"position"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Description  string     `json:"description"`
	Achievements string     `json:"achievements,omitempty"`
}

// ParsedSkill 解析的技能
type ParsedSkill struct {
	Name        string `json:"name"`
	Level       string `json:"level"`
	Description string `json:"description,omitempty"`
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
