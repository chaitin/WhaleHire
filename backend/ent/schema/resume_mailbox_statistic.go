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

// ResumeMailboxStatistic holds the schema definition for the ResumeMailboxStatistic entity.
type ResumeMailboxStatistic struct {
	ent.Schema
}

func (ResumeMailboxStatistic) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "resume_mailbox_statistics",
		},
	}
}

func (ResumeMailboxStatistic) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the ResumeMailboxStatistic.
func (ResumeMailboxStatistic) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("mailbox_id", uuid.UUID{}).Comment("邮箱配置ID"),
		field.Time("date").Comment("统计日期(按日分区，零点对齐)"),
		field.Int("synced_emails").Default(0).Comment("当日成功同步邮件数"),
		field.Int("parsed_resumes").Default(0).Comment("成功解析入库的简历附件数量"),
		field.Int("failed_resumes").Default(0).Comment("解析失败数量"),
		field.Int("skipped_attachments").Default(0).Comment("被去重或过滤的附件数量"),
		field.Int("last_sync_duration_ms").Default(0).Comment("最近一次同步耗时(毫秒)"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the ResumeMailboxStatistic.
func (ResumeMailboxStatistic) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("mailbox", ResumeMailboxSetting.Type).
			Ref("statistics").
			Unique().
			Required().
			Field("mailbox_id"),
	}
}

// Indexes of the ResumeMailboxStatistic.
func (ResumeMailboxStatistic) Indexes() []ent.Index {
	return []ent.Index{
		// 邮箱ID + 日期唯一约束，确保每个邮箱每天只有一条统计记录
		index.Fields("mailbox_id", "date").Unique(),
		// 邮箱ID索引，用于查询特定邮箱的统计
		index.Fields("mailbox_id"),
		// 日期索引，用于按时间范围查询
		index.Fields("date"),
		// 最后更新时间索引
		index.Fields("updated_at"),
	}
}
