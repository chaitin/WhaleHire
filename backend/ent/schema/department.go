package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/pkg/entx"
)

type Department struct{ ent.Schema }

func (Department) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Table: "department"},
	}
}

func (Department) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

func (Department) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name").MaxLen(100),
		field.String("description").MaxLen(255).Optional(), // 部门描述，可选
		field.UUID("parent_id", uuid.UUID{}).Optional(),    // 上级部门ID，可选
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Department) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("positions", JobPosition.Type),
	}
}
