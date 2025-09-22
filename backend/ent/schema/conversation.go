package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/pkg/entx"
)

// Conversation holds the schema definition for the Conversation entity.
type Conversation struct {
	ent.Schema
}

func (Conversation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "conversations",
		},
	}
}

func (Conversation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the Conversation.
func (Conversation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.String("title").NotEmpty(),
		field.String("agent_name").Optional(),
		field.JSON("metadata", map[string]interface{}{}).Optional(),
		field.String("status").Default("active"),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Conversation.
func (Conversation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("conversations").Field("user_id").Unique().Required(),
		edge.To("messages", Message.Type),
	}
}
