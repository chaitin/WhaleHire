package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/pkg/entx"
)

// NotificationEvent holds the schema definition for the NotificationEvent entity.
type NotificationEvent struct {
	ent.Schema
}

func (NotificationEvent) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "notification_events",
		},
	}
}

func (NotificationEvent) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the NotificationEvent.
func (NotificationEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("event_type").
			GoType(consts.NotificationEventType("")).
			Comment("事件类型"),
		field.String("channel").
			GoType(consts.NotificationChannel("")).
			Default(string(consts.NotificationChannelDingTalk)).
			Comment("通知渠道"),
		field.String("status").
			GoType(consts.NotificationStatus("")).
			Default(string(consts.NotificationStatusPending)).
			Comment("通知状态"),
		field.JSON("payload", map[string]interface{}{}),
		field.String("template_id"),
		field.String("target").Comment("目标地址，如webhook URL"),
		field.Int("retry_count").Default(0),
		field.Int("max_retry").Default(3),
		field.Int("timeout").Default(300).Comment("超时时间(秒)"),
		field.Text("last_error").Optional(),
		field.String("trace_id").Optional(),
		field.Time("scheduled_at").Optional().Comment("计划投递时间"),
		field.Time("delivered_at").Optional().Comment("实际投递时间"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the NotificationEvent.
func (NotificationEvent) Edges() []ent.Edge {
	return nil
}

// Indexes of the NotificationEvent.
func (NotificationEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("event_type"),
		index.Fields("channel"),
		index.Fields("status"),
		index.Fields("trace_id"),
		index.Fields("scheduled_at"),
		index.Fields("created_at"),
		// 复合索引用于查询待处理的事件
		index.Fields("status", "scheduled_at"),
		index.Fields("event_type", "status"),
	}
}