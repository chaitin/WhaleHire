package repo

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/department"
	"github.com/chaitin/WhaleHire/backend/db/jobeducationrequirement"
	"github.com/chaitin/WhaleHire/backend/db/jobexperiencerequirement"
	"github.com/chaitin/WhaleHire/backend/db/jobindustryrequirement"
	"github.com/chaitin/WhaleHire/backend/db/jobposition"
	"github.com/chaitin/WhaleHire/backend/db/jobresponsibility"
	"github.com/chaitin/WhaleHire/backend/db/jobskill"
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
	if job == nil {
		return nil, fmt.Errorf("job profile payload is required")
	}

	var result *db.JobPosition
	err := entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		creator := tx.JobPosition.Create().
			SetName(job.Name).
			SetDepartmentID(job.DepartmentID).
			SetNillableLocation(job.Location).
			SetNillableSalaryMin(job.SalaryMin).
			SetNillableSalaryMax(job.SalaryMax).
			SetNillableDescription(job.Description)

		created, err := creator.Save(ctx)
		if err != nil {
			return err
		}

		jobID := created.ID

		if related != nil {
			if err := createResponsibilities(ctx, tx, jobID, related.Responsibilities); err != nil {
				return err
			}
			if err := createSkills(ctx, tx, jobID, related.Skills); err != nil {
				return err
			}
			if err := createEducationRequirements(ctx, tx, jobID, related.EducationRequirements); err != nil {
				return err
			}
			if err := createExperienceRequirements(ctx, tx, jobID, related.ExperienceRequirements); err != nil {
				return err
			}
			if err := createIndustryRequirements(ctx, tx, jobID, related.IndustryRequirements); err != nil {
				return err
			}
		}

		result, err = tx.JobPosition.Query().
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
		return err
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
		current, err := tx.JobPosition.Query().
			Where(jobposition.ID(jobID)).
			WithResponsibilities().
			WithSkills(func(q *db.JobSkillQuery) {
				q.WithSkill()
			}).
			WithEducationRequirements().
			WithExperienceRequirements().
			WithIndustryRequirements().
			Only(ctx)
		if err != nil {
			return err
		}

		updater := tx.JobPosition.UpdateOneID(jobID)
		if err := fn(tx, current, updater); err != nil {
			return err
		}
		if _, err := updater.Save(ctx); err != nil {
			return err
		}

		result, err = tx.JobPosition.Query().
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
		return err
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

func (r *JobProfileRepo) List(ctx context.Context, req *domain.ListJobProfileRepoReq) ([]*db.JobPosition, *db.PageInfo, error) {
	query := r.db.JobPosition.Query().
		Where(jobposition.DeletedAtIsNil()).
		WithDepartment(func(q *db.DepartmentQuery) {
			q.Where(department.DeletedAtIsNil())
		})

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

func createResponsibilities(ctx context.Context, tx *db.Tx, jobID uuid.UUID, items []*db.JobResponsibility) error {
	for _, item := range items {
		if item == nil {
			continue
		}
		creator := tx.JobResponsibility.Create().
			SetJobID(jobID).
			SetResponsibility(item.Responsibility).
			SetSortOrder(item.SortOrder)
		if _, err := creator.Save(ctx); err != nil {
			return fmt.Errorf("create responsibility: %w", err)
		}
	}
	return nil
}

func createSkills(ctx context.Context, tx *db.Tx, jobID uuid.UUID, items []*db.JobSkill) error {
	for _, item := range items {
		if item == nil {
			continue
		}
		creator := tx.JobSkill.Create().
			SetJobID(jobID).
			SetSkillID(item.SkillID).
			SetType(item.Type).
			SetWeight(item.Weight)
		if _, err := creator.Save(ctx); err != nil {
			return fmt.Errorf("create job skill: %w", err)
		}
	}
	return nil
}

func createEducationRequirements(ctx context.Context, tx *db.Tx, jobID uuid.UUID, items []*db.JobEducationRequirement) error {
	for _, item := range items {
		if item == nil {
			continue
		}
		creator := tx.JobEducationRequirement.Create().
			SetJobID(jobID).
			SetMinDegree(item.MinDegree).
			SetWeight(item.Weight)
		if _, err := creator.Save(ctx); err != nil {
			return fmt.Errorf("create education requirement: %w", err)
		}
	}
	return nil
}

func createExperienceRequirements(ctx context.Context, tx *db.Tx, jobID uuid.UUID, items []*db.JobExperienceRequirement) error {
	for _, item := range items {
		if item == nil {
			continue
		}
		creator := tx.JobExperienceRequirement.Create().
			SetJobID(jobID).
			SetMinYears(item.MinYears).
			SetIdealYears(item.IdealYears).
			SetWeight(item.Weight)
		if _, err := creator.Save(ctx); err != nil {
			return fmt.Errorf("create experience requirement: %w", err)
		}
	}
	return nil
}

func createIndustryRequirements(ctx context.Context, tx *db.Tx, jobID uuid.UUID, items []*db.JobIndustryRequirement) error {
	for _, item := range items {
		if item == nil {
			continue
		}
		creator := tx.JobIndustryRequirement.Create().
			SetJobID(jobID).
			SetIndustry(item.Industry).
			SetWeight(item.Weight).
			SetNillableCompanyName(item.CompanyName)
		if _, err := creator.Save(ctx); err != nil {
			return fmt.Errorf("create industry requirement: %w", err)
		}
	}
	return nil
}
