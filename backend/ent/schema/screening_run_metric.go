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

// ScreeningRunMetric holds the schema definition for the ScreeningRunMetric entity.
type ScreeningRunMetric struct {
	ent.Schema
}

func (ScreeningRunMetric) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "screening_run_metrics",
		},
	}
}

func (ScreeningRunMetric) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the ScreeningRunMetric.
func (ScreeningRunMetric) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("task_id", uuid.UUID{}),
		field.Float("avg_score").Optional().Comment("平均分数"),
		field.JSON("histogram", map[string]interface{}{}).Optional().Comment("分数段分布"),
		field.Int64("tokens_input").Optional().Comment("任务总输入tokens"),
		field.Int64("tokens_output").Optional().Comment("任务总输出tokens"),
		field.Float("total_cost").Optional().Comment("模型调用成本统计"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the ScreeningRunMetric.
func (ScreeningRunMetric) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("task", ScreeningTask.Type).
			Ref("run_metrics").
			Field("task_id").
			Unique().
			Required(),
	}
}

// Indexes of the ScreeningRunMetric.
func (ScreeningRunMetric) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("task_id"),
		index.Fields("avg_score"),
		index.Fields("created_at"),
	}
}
