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
)

type JobSkillMeta struct{ ent.Schema }

func (JobSkillMeta) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Table: "job_skillmeta"},
	}
}

func (JobSkillMeta) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name").MaxLen(100).Unique(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (JobSkillMeta) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("job_links", JobSkill.Type),
	}
}

func (JobSkillMeta) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Unique(),
	}
}
