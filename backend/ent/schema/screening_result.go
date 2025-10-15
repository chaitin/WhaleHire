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

// ScreeningResult holds the schema definition for the ScreeningResult entity.
type ScreeningResult struct {
	ent.Schema
}

func (ScreeningResult) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "screening_results",
		},
	}
}

func (ScreeningResult) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the ScreeningResult.
func (ScreeningResult) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("task_id", uuid.UUID{}),
		field.UUID("job_position_id", uuid.UUID{}),
		field.UUID("resume_id", uuid.UUID{}),
		field.Float("overall_score").Comment("聚合总分 0-100"),
		field.Enum("match_level").Values(
			string(consts.MatchLevelExcellent),
			string(consts.MatchLevelGood),
			string(consts.MatchLevelFair),
			string(consts.MatchLevelPoor),
		).Optional().Comment("匹配等级"),
		field.JSON("dimension_scores", map[string]interface{}{}).Optional().Comment("各维度分数"),
		field.JSON("skill_detail", map[string]interface{}{}).Optional().Comment("技能匹配详情"),
		field.JSON("responsibility_detail", map[string]interface{}{}).Optional().Comment("职责匹配详情"),
		field.JSON("experience_detail", map[string]interface{}{}).Optional().Comment("经验匹配详情"),
		field.JSON("education_detail", map[string]interface{}{}).Optional().Comment("教育匹配详情"),
		field.JSON("industry_detail", map[string]interface{}{}).Optional().Comment("行业匹配详情"),
		field.JSON("basic_detail", map[string]interface{}{}).Optional().Comment("基本信息匹配详情"),
		field.JSON("recommendations", []string{}).Optional().Comment("匹配建议"),
		field.String("trace_id").Optional().MaxLen(100).Comment("链路追踪ID"),
		field.JSON("runtime_metadata", map[string]interface{}{}).Optional().Comment("运行时元数据"),
		field.JSON("sub_agent_versions", map[string]interface{}{}).Optional().Comment("各Agent版本快照"),
		field.Time("matched_at").Default(time.Now).Comment("匹配时间"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the ScreeningResult.
func (ScreeningResult) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("task", ScreeningTask.Type).
			Ref("results").
			Field("task_id").
			Unique().
			Required(),
		edge.From("job_position", JobPosition.Type).
			Ref("screening_results").
			Field("job_position_id").
			Unique().
			Required(),
		edge.From("resume", Resume.Type).
			Ref("screening_results").
			Field("resume_id").
			Unique().
			Required(),
	}
}

// Indexes of the ScreeningResult.
func (ScreeningResult) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("task_id"),
		index.Fields("job_position_id", "resume_id"),
		index.Fields("overall_score").Annotations(entsql.Desc()),
		index.Fields("task_id", "resume_id").Unique(),
		index.Fields("match_level"),
		index.Fields("matched_at"),
	}
}
