package repo

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/department"
	"github.com/chaitin/WhaleHire/backend/db/jobeducationrequirement"
	"github.com/chaitin/WhaleHire/backend/db/jobexperiencerequirement"
	"github.com/chaitin/WhaleHire/backend/db/jobindustryrequirement"
	"github.com/chaitin/WhaleHire/backend/db/jobposition"
	"github.com/chaitin/WhaleHire/backend/db/jobresponsibility"
	"github.com/chaitin/WhaleHire/backend/db/jobskill"
	"github.com/chaitin/WhaleHire/backend/db/resumejobapplication"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/entx"
)

type JobProfileRepo struct {
	db *db.Client
}

func NewJobProfileRepo(db *db.Client) domain.JobProfileRepo {
	return &JobProfileRepo{db: db}
}

func (r *JobProfileRepo) Create(ctx context.Context, job *db.JobPosition, related *domain.JobProfileRelatedEntities) (*db.JobPosition, error) {
	var workType *consts.JobWorkType
	if job.WorkType != "" {
		workType = &job.WorkType
	}

	var result *db.JobPosition
	err := entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		// 创建主要的JobPosition记录
		created, createErr := tx.JobPosition.Create().
			SetName(job.Name).
			SetDepartmentID(job.DepartmentID).
			SetNillableCreatedBy(job.CreatedBy).
			SetStatus(job.Status).
			SetNillableWorkType(workType).
			SetNillableLocation(job.Location).
			SetNillableSalaryMin(job.SalaryMin).
			SetNillableSalaryMax(job.SalaryMax).
			SetNillableDescription(job.Description).
			Save(ctx)
		if createErr != nil {
			return fmt.Errorf("failed to create job position: %w", createErr)
		}

		jobID := created.ID

		// 创建相关实体
		if related != nil {
			if respErr := r.createResponsibilities(ctx, tx, jobID, related.Responsibilities); respErr != nil {
				return respErr
			}
			if skillErr := r.createSkills(ctx, tx, jobID, related.Skills); skillErr != nil {
				return skillErr
			}
			if eduErr := r.createEducationRequirements(ctx, tx, jobID, related.EducationRequirements); eduErr != nil {
				return eduErr
			}
			if expErr := r.createExperienceRequirements(ctx, tx, jobID, related.ExperienceRequirements); expErr != nil {
				return expErr
			}
			if indErr := r.createIndustryRequirements(ctx, tx, jobID, related.IndustryRequirements); indErr != nil {
				return indErr
			}
		}

		// 查询完整的结果（包含所有关联数据）
		var queryErr error
		result, queryErr = tx.JobPosition.Query().
			Where(jobposition.ID(jobID)).
			WithDepartment().
			WithResponsibilities().
			WithSkills(func(q *db.JobSkillQuery) {
				q.WithSkill()
			}).
			WithEducationRequirements().
			WithExperienceRequirements().
			WithIndustryRequirements().
			Only(ctx)
		return queryErr
	})

	return result, err
}

func (r *JobProfileRepo) Update(ctx context.Context, id string, fn func(tx *db.Tx, current *db.JobPosition, updater *db.JobPositionUpdateOne) error) (*db.JobPosition, error) {
	jobID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid job profile ID: %w", err)
	}

	var result *db.JobPosition
	err = entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		current, queryErr := tx.JobPosition.Query().
			Where(jobposition.ID(jobID)).
			WithResponsibilities().
			WithSkills(func(q *db.JobSkillQuery) {
				q.WithSkill()
			}).
			WithEducationRequirements().
			WithExperienceRequirements().
			WithIndustryRequirements().
			Only(ctx)
		if queryErr != nil {
			return queryErr
		}

		updater := tx.JobPosition.UpdateOneID(jobID)
		if fnErr := fn(tx, current, updater); fnErr != nil {
			return fnErr
		}
		if _, saveErr := updater.Save(ctx); saveErr != nil {
			return saveErr
		}

		var finalErr error
		result, finalErr = tx.JobPosition.Query().
			Where(jobposition.ID(jobID)).
			WithDepartment().
			WithResponsibilities().
			WithSkills(func(q *db.JobSkillQuery) {
				q.WithSkill()
			}).
			WithEducationRequirements().
			WithExperienceRequirements().
			WithIndustryRequirements().
			Only(ctx)
		return finalErr
	})

	return result, err
}

// Delete performs soft delete by setting deleted_at timestamp for job position and related entities
func (r *JobProfileRepo) Delete(ctx context.Context, id string) error {
	jobID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid job profile ID: %w", err)
	}

	return entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		// Soft delete job responsibilities using SoftDeleteMixin
		if _, err := tx.JobResponsibility.Delete().
			Where(jobresponsibility.JobID(jobID)).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to soft delete job responsibilities: %w", err)
		}

		// Soft delete job skills using SoftDeleteMixin
		if _, err := tx.JobSkill.Delete().
			Where(jobskill.JobID(jobID)).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to soft delete job skills: %w", err)
		}

		// Soft delete job education requirements using SoftDeleteMixin
		if _, err := tx.JobEducationRequirement.Delete().
			Where(jobeducationrequirement.JobID(jobID)).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to soft delete job education requirements: %w", err)
		}

		// Soft delete job experience requirements using SoftDeleteMixin
		if _, err := tx.JobExperienceRequirement.Delete().
			Where(jobexperiencerequirement.JobID(jobID)).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to soft delete job experience requirements: %w", err)
		}

		// Soft delete job industry requirements using SoftDeleteMixin
		if _, err := tx.JobIndustryRequirement.Delete().
			Where(jobindustryrequirement.JobID(jobID)).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to soft delete job industry requirements: %w", err)
		}

		// Soft delete job position using SoftDeleteMixin
		if err := tx.JobPosition.DeleteOneID(jobID).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to soft delete job position: %w", err)
		}

		return nil
	})
}

func (r *JobProfileRepo) GetByID(ctx context.Context, id string) (*db.JobPosition, error) {
	jobID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid job profile ID: %w", err)
	}

	return r.db.JobPosition.Query().
		Where(jobposition.ID(jobID), jobposition.DeletedAtIsNil()).
		WithDepartment(func(q *db.DepartmentQuery) {
			q.Where(department.DeletedAtIsNil())
		}).
		WithCreator().
		WithResponsibilities(func(q *db.JobResponsibilityQuery) {
			q.Where(jobresponsibility.DeletedAtIsNil())
		}).
		WithSkills(func(q *db.JobSkillQuery) {
			q.Where(jobskill.DeletedAtIsNil()).WithSkill()
		}).
		WithEducationRequirements(func(q *db.JobEducationRequirementQuery) {
			q.Where(jobeducationrequirement.DeletedAtIsNil())
		}).
		WithExperienceRequirements(func(q *db.JobExperienceRequirementQuery) {
			q.Where(jobexperiencerequirement.DeletedAtIsNil())
		}).
		WithIndustryRequirements(func(q *db.JobIndustryRequirementQuery) {
			q.Where(jobindustryrequirement.DeletedAtIsNil())
		}).
		Only(ctx)
}

func (r *JobProfileRepo) GetByIDs(ctx context.Context, ids []string) ([]*db.JobPosition, error) {
	if len(ids) == 0 {
		return []*db.JobPosition{}, nil
	}

	jobIDs := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		jobID, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("invalid job profile ID %s: %w", id, err)
		}
		jobIDs = append(jobIDs, jobID)
	}

	return r.db.JobPosition.Query().
		Where(jobposition.IDIn(jobIDs...), jobposition.DeletedAtIsNil()).
		WithDepartment(func(q *db.DepartmentQuery) {
			q.Where(department.DeletedAtIsNil())
		}).
		WithCreator().
		All(ctx)
}

func (r *JobProfileRepo) List(ctx context.Context, req *domain.ListJobProfileRepoReq) ([]*db.JobPosition, *db.PageInfo, error) {
	query := r.db.JobPosition.Query().
		Where(jobposition.DeletedAtIsNil()).
		WithDepartment(func(q *db.DepartmentQuery) {
			q.Where(department.DeletedAtIsNil())
		}).
		WithCreator()

	if req != nil && req.ListJobProfileReq != nil {
		filters := req.ListJobProfileReq

		if filters.DepartmentID != nil && *filters.DepartmentID != "" {
			departmentID, err := uuid.Parse(*filters.DepartmentID)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid department ID: %w", err)
			}
			query = query.Where(jobposition.DepartmentID(departmentID))
		}

		if filters.Keyword != nil && *filters.Keyword != "" {
			terms := strings.Fields(*filters.Keyword)
			for _, term := range terms {
				query = query.Where(jobposition.Or(
					jobposition.NameContainsFold(term),
					jobposition.DescriptionContainsFold(term),
					jobposition.LocationContainsFold(term),
					jobposition.HasDepartmentWith(department.NameContainsFold(term)),
				))
			}
		}

		if len(filters.SkillIDs) > 0 {
			skillUUIDs := make([]uuid.UUID, 0, len(filters.SkillIDs))
			for _, id := range filters.SkillIDs {
				if id == "" {
					continue
				}
				skillID, err := uuid.Parse(id)
				if err != nil {
					return nil, nil, fmt.Errorf("invalid skill ID: %w", err)
				}
				skillUUIDs = append(skillUUIDs, skillID)
			}
			if len(skillUUIDs) > 0 {
				query = query.Where(jobposition.HasSkillsWith(jobskill.SkillIDIn(skillUUIDs...)))
			}
		}

		if len(filters.Locations) > 0 {
			var locations []string
			for _, loc := range filters.Locations {
				trimmed := strings.TrimSpace(loc)
				if trimmed == "" {
					continue
				}
				locations = append(locations, trimmed)
			}
			if len(locations) > 0 {
				query = query.Where(jobposition.LocationIn(locations...))
			}
		}
	}

	query = query.Order(jobposition.ByUpdatedAt(sql.OrderDesc()))

	return query.Page(ctx, req.ListJobProfileReq.Pagination.Page, req.ListJobProfileReq.Pagination.Size)
}

func (r *JobProfileRepo) Search(ctx context.Context, req *domain.SearchJobProfileRepoReq) ([]*db.JobPosition, *db.PageInfo, error) {
	query := r.db.JobPosition.Query().
		Where(jobposition.DeletedAtIsNil()).
		WithDepartment(func(q *db.DepartmentQuery) {
			q.Where(department.DeletedAtIsNil())
		}).
		WithCreator()

	if req != nil && req.SearchJobProfileReq != nil {
		filters := req.SearchJobProfileReq

		// 部门过滤
		if filters.DepartmentID != nil && *filters.DepartmentID != "" {
			departmentID, err := uuid.Parse(*filters.DepartmentID)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid department ID: %w", err)
			}
			query = query.Where(jobposition.DepartmentID(departmentID))
		}

		// 关键词模糊搜索 - 支持多个关键词
		if filters.Keyword != nil && *filters.Keyword != "" {
			terms := strings.Fields(*filters.Keyword)
			for _, term := range terms {
				query = query.Where(jobposition.Or(
					jobposition.NameContainsFold(term),
					jobposition.DescriptionContainsFold(term),
					jobposition.LocationContainsFold(term),
					jobposition.HasDepartmentWith(department.NameContainsFold(term)),
				))
			}
		}

		// 技能过滤
		if len(filters.SkillIDs) > 0 {
			skillUUIDs := make([]uuid.UUID, 0, len(filters.SkillIDs))
			for _, id := range filters.SkillIDs {
				if id == "" {
					continue
				}
				skillID, err := uuid.Parse(id)
				if err != nil {
					return nil, nil, fmt.Errorf("invalid skill ID: %w", err)
				}
				skillUUIDs = append(skillUUIDs, skillID)
			}
			if len(skillUUIDs) > 0 {
				query = query.Where(jobposition.HasSkillsWith(jobskill.SkillIDIn(skillUUIDs...)))
			}
		}

		// 地点过滤
		if len(filters.Locations) > 0 {
			var locations []string
			for _, loc := range filters.Locations {
				trimmed := strings.TrimSpace(loc)
				if trimmed == "" {
					continue
				}
				locations = append(locations, trimmed)
			}
			if len(locations) > 0 {
				query = query.Where(jobposition.LocationIn(locations...))
			}
		}

		// 薪资范围过滤
		if filters.SalaryMin != nil && *filters.SalaryMin > 0 {
			query = query.Where(jobposition.SalaryMaxGTE(*filters.SalaryMin))
		}
		if filters.SalaryMax != nil && *filters.SalaryMax > 0 {
			query = query.Where(jobposition.SalaryMinLTE(*filters.SalaryMax))
		}
	}

	// 按更新时间倒序排列
	query = query.Order(jobposition.ByUpdatedAt(sql.OrderDesc()))

	return query.Page(ctx, req.SearchJobProfileReq.Pagination.Page, req.SearchJobProfileReq.Pagination.Size)
}

func (r *JobProfileRepo) createResponsibilities(ctx context.Context, tx *db.Tx, jobID uuid.UUID, responsibilities []*db.JobResponsibility) error {
	if len(responsibilities) == 0 {
		return nil
	}

	bulk := make([]*db.JobResponsibilityCreate, len(responsibilities))
	for i, resp := range responsibilities {
		bulk[i] = tx.JobResponsibility.Create().
			SetJobID(jobID).
			SetResponsibility(resp.Responsibility)
	}

	_, err := tx.JobResponsibility.CreateBulk(bulk...).Save(ctx)
	return err
}

func (r *JobProfileRepo) createSkills(ctx context.Context, tx *db.Tx, jobID uuid.UUID, skills []*db.JobSkill) error {
	if len(skills) == 0 {
		return nil
	}

	bulk := make([]*db.JobSkillCreate, len(skills))
	for i, skill := range skills {
		bulk[i] = tx.JobSkill.Create().
			SetJobID(jobID).
			SetSkillID(skill.SkillID).
			SetType(consts.JobSkillType(skill.Type))
	}

	_, err := tx.JobSkill.CreateBulk(bulk...).Save(ctx)
	return err
}

func (r *JobProfileRepo) createEducationRequirements(ctx context.Context, tx *db.Tx, jobID uuid.UUID, requirements []*db.JobEducationRequirement) error {
	if len(requirements) == 0 {
		return nil
	}

	bulk := make([]*db.JobEducationRequirementCreate, len(requirements))
	for i, req := range requirements {
		bulk[i] = tx.JobEducationRequirement.Create().
			SetJobID(jobID).
			SetEducationType(consts.JobEducationType(req.EducationType))
	}

	_, err := tx.JobEducationRequirement.CreateBulk(bulk...).Save(ctx)
	return err
}

func (r *JobProfileRepo) createExperienceRequirements(ctx context.Context, tx *db.Tx, jobID uuid.UUID, requirements []*db.JobExperienceRequirement) error {
	if len(requirements) == 0 {
		return nil
	}

	bulk := make([]*db.JobExperienceRequirementCreate, len(requirements))
	for i, req := range requirements {
		bulk[i] = tx.JobExperienceRequirement.Create().
			SetJobID(jobID).
			SetExperienceType(req.ExperienceType).
			SetMinYears(req.MinYears).
			SetIdealYears(req.IdealYears)
	}

	_, err := tx.JobExperienceRequirement.CreateBulk(bulk...).Save(ctx)
	return err
}

func (r *JobProfileRepo) createIndustryRequirements(ctx context.Context, tx *db.Tx, jobID uuid.UUID, requirements []*db.JobIndustryRequirement) error {
	if len(requirements) == 0 {
		return nil
	}

	bulk := make([]*db.JobIndustryRequirementCreate, len(requirements))
	for i, req := range requirements {
		bulk[i] = tx.JobIndustryRequirement.Create().
			SetJobID(jobID).
			SetNillableIndustry(req.Industry).
			SetNillableCompanyName(req.CompanyName)
	}

	_, err := tx.JobIndustryRequirement.CreateBulk(bulk...).Save(ctx)
	return err
}

// HasRelatedResumes 检查岗位是否有关联的简历申请
func (r *JobProfileRepo) HasRelatedResumes(ctx context.Context, id string) (bool, error) {
	jobID, err := uuid.Parse(id)
	if err != nil {
		return false, fmt.Errorf("invalid job profile ID: %w", err)
	}

	count, err := r.db.ResumeJobApplication.Query().
		Where(
			resumejobapplication.JobPositionID(jobID),
			resumejobapplication.DeletedAtIsNil(),
		).
		Count(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check related resumes: %w", err)
	}

	return count > 0, nil
}
