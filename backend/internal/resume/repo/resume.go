package repo

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/resume"
	"github.com/chaitin/WhaleHire/backend/db/resumedocumentparse"
	"github.com/chaitin/WhaleHire/backend/db/resumeeducation"
	"github.com/chaitin/WhaleHire/backend/db/resumeexperience"
	"github.com/chaitin/WhaleHire/backend/db/resumelog"
	"github.com/chaitin/WhaleHire/backend/db/resumeskill"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/entx"
)

type ResumeRepo struct {
	db *db.Client
}

func NewResumeRepo(db *db.Client) domain.ResumeRepo {
	return &ResumeRepo{db: db}
}

// Create 创建简历
func (r *ResumeRepo) Create(ctx context.Context, resume *db.Resume) (*db.Resume, error) {
	creator := r.db.Resume.Create().
		SetUploaderID(resume.UploaderID).
		SetName(resume.Name).
		SetGender(resume.Gender).
		SetEmail(resume.Email).
		SetPhone(resume.Phone).
		SetCurrentCity(resume.CurrentCity).
		SetHighestEducation(resume.HighestEducation).
		SetYearsExperience(resume.YearsExperience).
		SetResumeFileURL(resume.ResumeFileURL).
		SetStatus(resume.Status)

	// 处理可空字段
	if !resume.Birthday.IsZero() {
		creator = creator.SetBirthday(resume.Birthday)
	}
	if resume.ErrorMessage != "" {
		creator = creator.SetErrorMessage(resume.ErrorMessage)
	}
	if !resume.ParsedAt.IsZero() {
		creator = creator.SetParsedAt(resume.ParsedAt)
	}

	return creator.Save(ctx)
}

// GetByID 根据ID获取简历
func (r *ResumeRepo) GetByID(ctx context.Context, id string) (*db.Resume, error) {
	resumeID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid resume ID: %w", err)
	}

	return r.db.Resume.Query().
		Where(resume.ID(resumeID)).
		WithUser().
		WithEducations().
		WithExperiences().
		WithSkills().
		WithLogs().
		Only(ctx)
}

// Update 更新简历
func (r *ResumeRepo) Update(ctx context.Context, id string, fn func(*db.Tx, *db.Resume, *db.ResumeUpdateOne) error) (*db.Resume, error) {
	resumeID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid resume ID: %w", err)
	}

	var result *db.Resume
	err = entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		resume, err := tx.Resume.Query().
			Where(resume.ID(resumeID)).
			Only(ctx)
		if err != nil {
			return err
		}

		updater := tx.Resume.UpdateOneID(resumeID)
		if err := fn(tx, resume, updater); err != nil {
			return err
		}

		result, err = updater.Save(ctx)
		return err
	})

	return result, err
}

// Delete 删除简历
func (r *ResumeRepo) Delete(ctx context.Context, id string) error {
	resumeID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid resume ID: %w", err)
	}

	// 使用事务确保级联删除的原子性
	return entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		// 删除简历教育经历
		if _, err := tx.ResumeEducation.Delete().Where(
			resumeeducation.ResumeID(resumeID),
		).Exec(ctx); err != nil {
			return fmt.Errorf("failed to delete resume educations: %w", err)
		}

		// 删除简历工作经历
		if _, err := tx.ResumeExperience.Delete().Where(
			resumeexperience.ResumeID(resumeID),
		).Exec(ctx); err != nil {
			return fmt.Errorf("failed to delete resume experiences: %w", err)
		}

		// 删除简历技能
		if _, err := tx.ResumeSkill.Delete().Where(
			resumeskill.ResumeID(resumeID),
		).Exec(ctx); err != nil {
			return fmt.Errorf("failed to delete resume skills: %w", err)
		}

		// 删除简历日志
		if _, err := tx.ResumeLog.Delete().Where(
			resumelog.ResumeID(resumeID),
		).Exec(ctx); err != nil {
			return fmt.Errorf("failed to delete resume logs: %w", err)
		}

		// 删除简历文档解析记录
		if _, err := tx.ResumeDocumentParse.Delete().Where(
			resumedocumentparse.ResumeID(resumeID),
		).Exec(ctx); err != nil {
			return fmt.Errorf("failed to delete resume document parses: %w", err)
		}

		// 最后删除简历主记录
		if err := tx.Resume.DeleteOneID(resumeID).Exec(ctx); err != nil {
			return fmt.Errorf("failed to delete resume: %w", err)
		}

		return nil
	})
}

// List 获取简历列表
func (r *ResumeRepo) List(ctx context.Context, req *domain.ListResumeReq) ([]*db.Resume, *db.PageInfo, error) {
	query := r.db.Resume.Query()

	// 应用过滤条件
	if req.UploaderID != nil {
		uploaderID, err := uuid.Parse(*req.UploaderID)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid uploader ID: %w", err)
		}
		query = query.Where(resume.UploaderID(uploaderID))
	}

	if req.Status != nil {
		query = query.Where(resume.Status(string(*req.Status)))
	}

	if req.Search != nil && *req.Search != "" {
		query = query.Where(
			resume.Or(
				resume.NameContains(*req.Search),
				resume.EmailContains(*req.Search),
				resume.PhoneContains(*req.Search),
				resume.CurrentCityContains(*req.Search),
			),
		)
	}

	if req.DateFrom != nil {
		query = query.Where(resume.CreatedAtGTE(*req.DateFrom))
	}

	if req.DateTo != nil {
		query = query.Where(resume.CreatedAtLTE(*req.DateTo))
	}

	// 排序
	query = query.Order(resume.ByCreatedAt(sql.OrderDesc()))

	// 分页
	return query.Page(ctx, req.Pagination.Page, req.Pagination.Size)
}

// Search 搜索简历
func (r *ResumeRepo) Search(ctx context.Context, req *domain.SearchResumeReq) ([]*db.Resume, *db.PageInfo, error) {
	query := r.db.Resume.Query()

	// 用户ID过滤
	if req.UploaderID != nil {
		uploaderID, err := uuid.Parse(*req.UploaderID)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid uploader ID: %w", err)
		}
		query = query.Where(resume.UploaderID(uploaderID))
	}

	// 关键词搜索
	if req.Keywords != nil && *req.Keywords != "" {
		keywords := strings.Fields(*req.Keywords)
		var predicates []func(*sql.Selector)

		for _, keyword := range keywords {
			keyword = "%" + keyword + "%"
			predicates = append(predicates, func(s *sql.Selector) {
				s.Where(sql.Or(
					sql.Like(resume.FieldName, keyword),
					sql.Like(resume.FieldEmail, keyword),
					sql.Like(resume.FieldPhone, keyword),
					sql.Like(resume.FieldCurrentCity, keyword),
				))
			})
		}

		for _, predicate := range predicates {
			query = query.Where(predicate)
		}
	}

	// 技能搜索
	if len(req.Skills) > 0 {
		query = query.Where(
			resume.HasSkillsWith(
				resumeskill.SkillNameIn(req.Skills...),
			),
		)
	}

	// 教育背景
	if req.Education != nil && *req.Education != "" {
		query = query.Where(resume.HighestEducationContains(*req.Education))
	}

	// 工作年限
	if req.YearsExperience != nil {
		query = query.Where(resume.YearsExperienceGTE(*req.YearsExperience))
	}

	// 城市
	if req.CurrentCity != nil && *req.CurrentCity != "" {
		query = query.Where(resume.CurrentCityContains(*req.CurrentCity))
	}

	// 性别
	if req.Gender != nil && *req.Gender != "" {
		query = query.Where(resume.Gender(*req.Gender))
	}

	// 年龄范围（基于生日计算）
	if req.AgeFrom != nil || req.AgeTo != nil {
		// TODO: 实现基于生日的年龄范围查询
		// 这里需要根据生日计算年龄，暂时跳过复杂的SQL计算
		// 可以在业务层进行过滤
		_ = req.AgeFrom // 避免未使用变量警告
		_ = req.AgeTo   // 避免未使用变量警告
	}

	// 排序
	query = query.Order(resume.ByCreatedAt(sql.OrderDesc()))

	// 分页
	return query.Page(ctx, req.Pagination.Page, req.Pagination.Size)
}

// CreateEducation 创建教育经历
func (r *ResumeRepo) CreateEducation(ctx context.Context, education *db.ResumeEducation) (*db.ResumeEducation, error) {
	creator := r.db.ResumeEducation.Create().
		SetResumeID(education.ResumeID).
		SetSchool(education.School).
		SetDegree(education.Degree).
		SetMajor(education.Major)

	if !education.StartDate.IsZero() {
		creator = creator.SetStartDate(education.StartDate)
	}
	if !education.EndDate.IsZero() {
		creator = creator.SetEndDate(education.EndDate)
	}

	return creator.Save(ctx)
}

// CreateExperience 创建工作经历
func (r *ResumeRepo) CreateExperience(ctx context.Context, experience *db.ResumeExperience) (*db.ResumeExperience, error) {
	creator := r.db.ResumeExperience.Create().
		SetResumeID(experience.ResumeID).
		SetCompany(experience.Company).
		SetPosition(experience.Position).
		SetTitle(experience.Title).
		SetDescription(experience.Description)

	if !experience.StartDate.IsZero() {
		creator = creator.SetStartDate(experience.StartDate)
	}
	if !experience.EndDate.IsZero() {
		creator = creator.SetEndDate(experience.EndDate)
	}

	return creator.Save(ctx)
}

// CreateSkill 创建技能
func (r *ResumeRepo) CreateSkill(ctx context.Context, skill *db.ResumeSkill) (*db.ResumeSkill, error) {
	return r.db.ResumeSkill.Create().
		SetResumeID(skill.ResumeID).
		SetSkillName(skill.SkillName).
		SetLevel(skill.Level).
		SetDescription(skill.Description).
		Save(ctx)
}

// CreateLog 创建操作日志
func (r *ResumeRepo) CreateLog(ctx context.Context, log *db.ResumeLog) (*db.ResumeLog, error) {
	return r.db.ResumeLog.Create().
		SetResumeID(log.ResumeID).
		SetAction(log.Action).
		SetMessage(log.Message).
		Save(ctx)
}

// UpdateStatus 更新简历状态
func (r *ResumeRepo) UpdateStatus(ctx context.Context, id string, status domain.ResumeStatus) error {
	resumeID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid resume ID: %w", err)
	}

	return r.db.Resume.UpdateOneID(resumeID).
		SetStatus(string(status)).
		Exec(ctx)
}

// UpdateErrorMessage 更新错误信息
func (r *ResumeRepo) UpdateErrorMessage(ctx context.Context, id string, errorMsg string) error {
	resumeID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid resume ID: %w", err)
	}

	return r.db.Resume.UpdateOneID(resumeID).
		SetErrorMessage(errorMsg).
		Exec(ctx)
}

// GetEducationsByResumeID 根据简历ID获取教育经历
func (r *ResumeRepo) GetEducationsByResumeID(ctx context.Context, resumeID string) ([]*db.ResumeEducation, error) {
	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		return nil, fmt.Errorf("invalid resume ID: %w", err)
	}

	return r.db.ResumeEducation.Query().
		Where(resumeeducation.ResumeIDEQ(resumeUUID)).
		Order(resumeeducation.ByStartDate(sql.OrderDesc())).
		All(ctx)
}

// GetExperiencesByResumeID 根据简历ID获取工作经验
func (r *ResumeRepo) GetExperiencesByResumeID(ctx context.Context, resumeID string) ([]*db.ResumeExperience, error) {
	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		return nil, fmt.Errorf("invalid resume ID: %w", err)
	}

	return r.db.ResumeExperience.Query().
		Where(resumeexperience.ResumeIDEQ(resumeUUID)).
		Order(resumeexperience.ByStartDate(sql.OrderDesc())).
		All(ctx)
}

// GetSkillsByResumeID 根据简历ID获取技能
func (r *ResumeRepo) GetSkillsByResumeID(ctx context.Context, resumeID string) ([]*db.ResumeSkill, error) {
	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		return nil, fmt.Errorf("invalid resume ID: %w", err)
	}

	return r.db.ResumeSkill.Query().
		Where(resumeskill.ResumeIDEQ(resumeUUID)).
		Order(resumeskill.ByCreatedAt(sql.OrderDesc())).
		All(ctx)
}

// GetLogsByResumeID 根据简历ID获取操作日志
func (r *ResumeRepo) GetLogsByResumeID(ctx context.Context, resumeID string) ([]*db.ResumeLog, error) {
	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		return nil, fmt.Errorf("invalid resume ID: %w", err)
	}

	return r.db.ResumeLog.Query().
		Where(resumelog.ResumeIDEQ(resumeUUID)).
		Order(resumelog.ByCreatedAt(sql.OrderDesc())).
		All(ctx)
}

// CreateDocumentParse 创建文档解析记录
func (r *ResumeRepo) CreateDocumentParse(ctx context.Context, documentParse *db.ResumeDocumentParse) (*db.ResumeDocumentParse, error) {
	creator := r.db.ResumeDocumentParse.Create().
		SetResumeID(documentParse.ResumeID).
		SetFileID(documentParse.FileID).
		SetContent(documentParse.Content).
		SetFileType(documentParse.FileType).
		SetFilename(documentParse.Filename).
		SetTitle(documentParse.Title).
		SetStatus(documentParse.Status)

	if !documentParse.UploadAt.IsZero() {
		creator = creator.SetUploadAt(documentParse.UploadAt)
	}
	if documentParse.ErrorMessage != "" {
		creator = creator.SetErrorMessage(documentParse.ErrorMessage)
	}

	return creator.Save(ctx)
}

// GetDocumentParseByResumeID 根据简历ID获取文档解析记录
func (r *ResumeRepo) GetDocumentParseByResumeID(ctx context.Context, resumeID string) (*db.ResumeDocumentParse, error) {
	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		return nil, fmt.Errorf("invalid resume ID: %w", err)
	}

	return r.db.ResumeDocumentParse.Query().
		Where(resumedocumentparse.ResumeIDEQ(resumeUUID)).
		Order(resumedocumentparse.ByCreatedAt(sql.OrderDesc())).
		First(ctx)
}
