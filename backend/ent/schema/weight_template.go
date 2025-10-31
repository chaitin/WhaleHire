package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/pkg/entx"
)

// WeightTemplate holds the schema definition for the WeightTemplate entity.
type WeightTemplate struct {
	ent.Schema
}

func (WeightTemplate) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "weight_template",
		},
	}
}

func (WeightTemplate) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the WeightTemplate.
func (WeightTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name").MaxLen(100).Comment("模板名称"),
		field.String("description").MaxLen(500).Optional().Comment("模板描述"),
		field.JSON("weights", map[string]interface{}{}).Comment("权重配置 JSONB"),
		field.UUID("created_by", uuid.UUID{}).Comment("创建者ID"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
	}
}

// Edges of the WeightTemplate.
func (WeightTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("creator", User.Type).
			Ref("created_weight_templates").
			Field("created_by").
			Unique().
			Required(),
	}
}

// Indexes of the WeightTemplate.
func (WeightTemplate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_by"),
		index.Fields("created_at"),
	}
}
