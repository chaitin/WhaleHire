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

// ScreeningTask holds the schema definition for the ScreeningTask entity.
type ScreeningTask struct {
	ent.Schema
}

func (ScreeningTask) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "screening_tasks",
		},
	}
}

func (ScreeningTask) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the ScreeningTask.
func (ScreeningTask) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("job_position_id", uuid.UUID{}),
		field.UUID("created_by", uuid.UUID{}),
		field.String("status").Default(string(consts.ScreeningTaskStatusPending)).Comment("pending/running/completed/failed"),
		field.JSON("dimension_weights", map[string]interface{}{}).Optional().Comment("维度权重配置"),
		field.JSON("llm_config", map[string]interface{}{}).Optional().Comment("LLM/Embedding配置"),
		field.Text("notes").Optional().Comment("任务备注"),
		field.Int("resume_total").Default(0).Comment("任务内候选总数"),
		field.Int("resume_processed").Default(0).Comment("已处理数"),
		field.Int("resume_succeeded").Default(0).Comment("成功匹配数"),
		field.Int("resume_failed").Default(0).Comment("失败数"),
		field.String("agent_version").Optional().MaxLen(50).Comment("智能匹配Agent版本号"),
		field.Time("started_at").Optional().Comment("任务开始时间"),
		field.Time("finished_at").Optional().Comment("任务完成时间"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the ScreeningTask.
func (ScreeningTask) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("job_position", JobPosition.Type).
			Ref("screening_tasks").
			Field("job_position_id").
			Unique().
			Required(),
		edge.From("creator", User.Type).
			Ref("created_screening_tasks").
			Field("created_by").
			Unique().
			Required(),
		edge.To("task_resumes", ScreeningTaskResume.Type),
		edge.To("results", ScreeningResult.Type),
		edge.To("run_metrics", ScreeningRunMetric.Type),
		edge.To("node_runs", ScreeningNodeRun.Type),
	}
}

// Indexes of the ScreeningTask.
func (ScreeningTask) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("job_position_id"),
		index.Fields("status"),
		index.Fields("created_by"),
		index.Fields("created_at"),
	}
}
