package schema

import (
	"time"

	entschema "entgo.io/ent/schema"
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/chaitin/WhaleHire/backend/pkg/entx"
)

type JobIndustryRequirement struct{ ent.Schema }

func (JobIndustryRequirement) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Table: "job_industry_requirement"},
	}
}

func (JobIndustryRequirement) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

func (JobIndustryRequirement) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("job_id", uuid.UUID{}),
		field.String("industry").MaxLen(100),
		field.String("company_name").Optional().Nillable().MaxLen(200),
		field.Int("weight").Range(0, 100).Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (JobIndustryRequirement) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("job", JobPosition.Type).
			Ref("industry_requirements").
			Field("job_id").
			Unique().
			Required(),
	}
}
