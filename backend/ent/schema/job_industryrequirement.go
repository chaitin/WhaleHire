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

type JobIndustryRequirement struct{ ent.Schema }

func (JobIndustryRequirement) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Table: "job_industry_requirement"},
	}
}

func (JobIndustryRequirement) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("job_id", uuid.UUID{}),
		field.String("industry").MaxLen(100),
		field.String("company_name").Optional().Nillable().MaxLen(200),
		field.Int("weight").Range(0, 100),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
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
