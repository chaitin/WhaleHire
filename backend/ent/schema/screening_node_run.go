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

// ScreeningNodeRun holds the schema definition for the ScreeningNodeRun entity.
type ScreeningNodeRun struct {
	ent.Schema
}

func (ScreeningNodeRun) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "screening_node_runs",
		},
	}
}

func (ScreeningNodeRun) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the ScreeningNodeRun.
func (ScreeningNodeRun) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("task_id", uuid.UUID{}),
		field.UUID("task_resume_id", uuid.UUID{}),
		field.String("node_key").Optional().Comment("节点类型，参考 domain.TaskMetaDataNode 等常量"),
		field.String("status").Default(string(consts.ScreeningNodeRunStatusPending)).Comment("pending/running/completed/failed"),
		field.Int("attempt_no").Default(1).Comment("第几次尝试"),
		field.String("trace_id").Optional().MaxLen(100).Comment("链路追踪ID"),
		field.String("agent_version").Optional().MaxLen(50).Comment("Agent版本号"),
		field.String("model_name").Optional().MaxLen(100).Comment("模型名称"),
		field.String("model_provider").Optional().MaxLen(50).Comment("模型提供商"),
		field.JSON("llm_params", map[string]interface{}{}).Optional().Comment("LLM参数快照"),
		field.JSON("input_payload", map[string]interface{}{}).Optional().Comment("节点输入快照"),
		field.JSON("output_payload", map[string]interface{}{}).Optional().Comment("节点输出快照"),
		field.Text("error_message").Optional().Comment("错误信息"),
		field.Int64("tokens_input").Optional().Comment("输入tokens"),
		field.Int64("tokens_output").Optional().Comment("输出tokens"),
		field.Float("total_cost").Optional().Comment("调用成本"),
		field.Time("started_at").Optional().Comment("开始时间"),
		field.Time("finished_at").Optional().Comment("结束时间"),
		field.Int("duration_ms").Optional().Comment("耗时(毫秒)"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the ScreeningNodeRun.
func (ScreeningNodeRun) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("task", ScreeningTask.Type).
			Ref("node_runs").
			Field("task_id").
			Unique().
			Required(),
		edge.From("task_resume", ScreeningTaskResume.Type).
			Ref("node_runs").
			Field("task_resume_id").
			Unique().
			Required(),
	}
}

// Indexes of the ScreeningNodeRun.
func (ScreeningNodeRun) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("task_id"),
		index.Fields("task_resume_id"),
		index.Fields("node_key"),
		index.Fields("status"),
		index.Fields("created_at"),
		index.Fields("task_resume_id", "node_key", "attempt_no").Unique(),
		index.Fields("task_id", "node_key"),
		index.Fields("node_key", "status"),
	}
}
