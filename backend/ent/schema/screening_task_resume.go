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

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/pkg/entx"
)

// ScreeningTaskResume holds the schema definition for the ScreeningTaskResume entity.
type ScreeningTaskResume struct {
	ent.Schema
}

func (ScreeningTaskResume) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "screening_task_resumes",
		},
	}
}

func (ScreeningTaskResume) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the ScreeningTaskResume.
func (ScreeningTaskResume) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("task_id", uuid.UUID{}),
		field.UUID("resume_id", uuid.UUID{}),
		field.String("status").Default(string(consts.ScreeningTaskResumeStatusPending)).Comment("pending/running/completed/failed"),
		field.Text("error_message").Optional().Comment("错误信息"),
		field.Int("ranking").Optional().Comment("排名（任务内）"),
		field.Float("score").Optional().Comment("综合分快照（任务维度）"),
		field.Time("processed_at").Optional().Comment("处理完成时间"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the ScreeningTaskResume.
func (ScreeningTaskResume) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("task", ScreeningTask.Type).
			Ref("task_resumes").
			Field("task_id").
			Unique().
			Required(),
		edge.From("resume", Resume.Type).
			Ref("screening_task_resumes").
			Field("resume_id").
			Unique().
			Required(),
	}
}

// Indexes of the ScreeningTaskResume.
func (ScreeningTaskResume) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("task_id"),
		index.Fields("status"),
		index.Fields("score").Annotations(entsql.Desc()),
		index.Fields("task_id", "resume_id").Unique(),
		index.Fields("task_id", "ranking"),
	}
}
