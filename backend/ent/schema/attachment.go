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

// Attachment holds the schema definition for the Attachment entity.
type Attachment struct {
	ent.Schema
}

func (Attachment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "attachments",
		},
	}
}

func (Attachment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the Attachment.
func (Attachment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("message_id", uuid.UUID{}),
		field.String("type").NotEmpty().Comment("image, audio, video, file"),
		field.String("url").NotEmpty(),
		field.JSON("metadata", map[string]interface{}{}).Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Attachment.
func (Attachment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("message", Message.Type).Ref("attachments").Field("message_id").Unique().Required(),
	}
}
