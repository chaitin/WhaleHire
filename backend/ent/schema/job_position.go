package schema

import (
	"time"

	entschema "entgo.io/ent/schema"
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type JobPosition struct{ ent.Schema }

func (JobPosition) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Table: "job_position"},
	}
}

func (JobPosition) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name").MaxLen(100),
		field.UUID("department_id", uuid.UUID{}),
		field.String("location").Optional().Nillable().MaxLen(200),
		field.Float("salary_min").Optional().Nillable(),
		field.Float("salary_max").Optional().Nillable(),
		field.String("description").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (JobPosition) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("department", Department.Type).
			Ref("positions").
			Field("department_id").
			Unique().
			Required(),
		edge.To("responsibilities", JobResponsibility.Type),
		edge.To("skills", JobSkill.Type),
		edge.To("education_requirements", JobEducationRequirement.Type),
		edge.To("experience_requirements", JobExperienceRequirement.Type),
		edge.To("industry_requirements", JobIndustryRequirement.Type),
	}
}
