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

// ResumeMailboxCursor holds the schema definition for the ResumeMailboxCursor entity.
type ResumeMailboxCursor struct {
	ent.Schema
}

func (ResumeMailboxCursor) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "resume_mailbox_cursors",
		},
	}
}

func (ResumeMailboxCursor) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the ResumeMailboxCursor.
func (ResumeMailboxCursor) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("mailbox_id", uuid.UUID{}).Comment("邮箱配置ID"),
		field.Text("protocol_cursor").Comment("协议游标信息(IMAP UIDVALIDITY+UID 或 POP3 UIDL列表摘要)"),
		field.String("last_message_id").Optional().MaxLen(512).Comment("最后处理的邮件ID，用于排查"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the ResumeMailboxCursor.
func (ResumeMailboxCursor) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("mailbox", ResumeMailboxSetting.Type).
			Ref("cursors").
			Unique().
			Required().
			Field("mailbox_id"),
	}
}

// Indexes of the ResumeMailboxCursor.
func (ResumeMailboxCursor) Indexes() []ent.Index {
	return []ent.Index{
		// 邮箱ID索引，用于查询特定邮箱的游标
		index.Fields("mailbox_id"),
		// 最后更新时间索引
		index.Fields("updated_at"),
	}
}
