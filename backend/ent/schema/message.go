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

// Message holds the schema definition for the Message entity.
type Message struct {
	ent.Schema
}

func (Message) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "messages",
		},
	}
}

func (Message) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the Message.
func (Message) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("conversation_id", uuid.UUID{}),
		field.String("role").NotEmpty().Comment("user, assistant, system, agent"),
		field.String("agent_name").Optional(),
		field.String("type").Default("text").Comment("text, image, audio, video, file"),
		field.Text("content").Optional(),
		field.String("media_url").Optional(),
		field.Int("sequence").Default(0).Comment("消息在对话中的顺序"),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Message.
func (Message) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("conversation", Conversation.Type).Ref("messages").Field("conversation_id").Unique().Required(),
		edge.To("attachments", Attachment.Type),
	}
}
