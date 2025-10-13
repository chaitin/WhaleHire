package schema

import (
	"time"

	entschema "entgo.io/ent/schema"
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/pkg/entx"
)

type JobSkill struct{ ent.Schema }

func (JobSkill) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Table: "job_skill"},
	}
}

func (JobSkill) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

func (JobSkill) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("job_id", uuid.UUID{}),
		field.UUID("skill_id", uuid.UUID{}),
		field.String("type").GoType(consts.JobSkillType("")),
		field.Int("weight").Range(0, 100).Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (JobSkill) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("job", JobPosition.Type).
			Ref("skills").
			Field("job_id").
			Unique().
			Required(),
		edge.From("skill", JobSkillMeta.Type).
			Ref("job_links").
			Field("skill_id").
			Unique().
			Required(),
	}
}

func (JobSkill) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("job_id", "skill_id", "type").Unique(),
	}
}
