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

type JobExperienceRequirement struct{ ent.Schema }

func (JobExperienceRequirement) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Table: "job_experience_requirement"},
	}
}

func (JobExperienceRequirement) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

func (JobExperienceRequirement) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("job_id", uuid.UUID{}),
		field.String("experience_type").Optional(),
		// 用于数值计算和筛选的字段（可以根据experience_type自动计算）
		field.Int("min_years").Default(0).NonNegative().Optional(),
		field.Int("ideal_years").Default(0).NonNegative().Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (JobExperienceRequirement) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("job", JobPosition.Type).
			Ref("experience_requirements").
			Field("job_id").
			Unique().
			Required(),
	}
}
