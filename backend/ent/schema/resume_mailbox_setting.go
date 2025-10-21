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

// ResumeMailboxSetting holds the schema definition for the ResumeMailboxSetting entity.
type ResumeMailboxSetting struct {
	ent.Schema
}

func (ResumeMailboxSetting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "resume_mailbox_settings",
		},
	}
}

func (ResumeMailboxSetting) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the ResumeMailboxSetting.
func (ResumeMailboxSetting) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name").MaxLen(256).Comment("任务名称，页面展示用"),
		field.String("email_address").MaxLen(256).Comment("邮箱账号"),
		field.Enum("protocol").Values("imap", "pop3").Comment("邮箱协议"),
		field.String("host").MaxLen(256).Comment("邮箱服务器地址"),
		field.Int("port").Comment("邮箱服务器端口"),
		field.Bool("use_ssl").Default(true).Comment("是否使用SSL/TLS"),
		field.String("folder").MaxLen(256).Optional().Comment("IMAP专用文件夹，默认INBOX"),
		field.Enum("auth_type").Values("password", "oauth").Default("password").Comment("认证类型"),
		field.JSON("encrypted_credential", map[string]interface{}{}).Comment("加密后的凭证信息"),
		field.UUID("uploader_id", uuid.UUID{}).Comment("上传人用户ID"),
		field.UUID("job_profile_id", uuid.UUID{}).Optional().Nillable().Comment("岗位画像ID"),
		field.Int("sync_interval_minutes").Optional().Nillable().Comment("自定义同步频率(分钟)，为空则使用平台默认"),
		field.Enum("status").Values("enabled", "disabled").Default("enabled").Comment("状态"),
		field.Time("last_synced_at").Optional().Nillable().Comment("最后同步时间"),
		field.Text("last_error").Optional().Comment("最后一次错误信息"),
		field.Int("retry_count").Default(0).Comment("重试次数"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the ResumeMailboxSetting.
func (ResumeMailboxSetting) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("uploader", User.Type).
			Unique().
			Required().
			Field("uploader_id"),
		edge.To("job_profile", JobPosition.Type).
			Unique().
			Field("job_profile_id"),
		edge.To("cursors", ResumeMailboxCursor.Type),
		edge.To("statistics", ResumeMailboxStatistic.Type),
	}
}

// Indexes of the ResumeMailboxSetting.
func (ResumeMailboxSetting) Indexes() []ent.Index {
	return []ent.Index{
		// 协议 + 邮箱地址唯一约束
		index.Fields("protocol", "email_address").Unique(),
		// 状态索引，用于查询启用的邮箱
		index.Fields("status"),
		// 上传人索引
		index.Fields("uploader_id"),
		// 岗位画像索引
		index.Fields("job_profile_id"),
		// 最后同步时间索引，用于监控
		index.Fields("last_synced_at"),
	}
}
