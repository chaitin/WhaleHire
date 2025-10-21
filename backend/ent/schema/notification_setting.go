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

// NotificationSetting holds the schema definition for the NotificationSetting entity.
type NotificationSetting struct {
	ent.Schema
}

func (NotificationSetting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "notification_settings",
		},
	}
}

func (NotificationSetting) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the NotificationSetting.
func (NotificationSetting) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name").
			NotEmpty().
			Comment("配置名称，用于区分同类型的不同配置"),
		field.String("channel").
			GoType(consts.NotificationChannel("")).
			Comment("通知渠道"),
		field.Bool("enabled").
			Default(true).
			Comment("是否启用"),
		field.JSON("dingtalk_config", map[string]interface{}{}).
			Optional().
			Comment("钉钉通知配置"),
		field.Int("max_retry").
			Default(3).
			Comment("最大重试次数"),
		field.Int("timeout").
			Default(300).
			Comment("超时时间(秒)"),
		field.String("description").
			Optional().
			Comment("通知设置描述"),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the NotificationSetting.
func (NotificationSetting) Edges() []ent.Edge {
	return nil
}

// Indexes of the NotificationSetting.
func (NotificationSetting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel"),
		index.Fields("enabled"),
		index.Fields("channel", "enabled"),
		index.Fields("name", "channel").Unique(),
	}
}
