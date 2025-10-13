package schema

import (
	"time"

	entschema "entgo.io/ent/schema"
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/pkg/entx"
)

type JobEducationRequirement struct{ ent.Schema }

func (JobEducationRequirement) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Table: "job_education_requirement"},
	}
}

func (JobEducationRequirement) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

func (JobEducationRequirement) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("job_id", uuid.UUID{}),
		field.String("education_type").GoType(consts.JobEducationType("")).Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (JobEducationRequirement) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("job", JobPosition.Type).
			Ref("education_requirements").
			Field("job_id").
			Unique().
			Required(),
	}
}
