package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"

	"github.com/chaitin/WhaleHire/backend/pkg/entx"
)

// UniversityProfile 定义高校画像实体的结构。
type UniversityProfile struct {
	ent.Schema
}

func (UniversityProfile) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "university_profiles",
		},
	}
}

func (UniversityProfile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields 定义高校画像实体的字段。
func (UniversityProfile) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name_cn").NotEmpty().Comment("中文名称"),
		field.String("name_en").Optional().Comment("英文名称"),
		field.String("alias").Optional().Comment("别名"),
		field.String("country").Optional().Comment("国家/地区"),
		field.Bool("is_double_first_class").Default(false).Comment("是否为双一流"),
		field.Bool("is_project_985").Default(false).Comment("是否为985工程"),
		field.Bool("is_project_211").Default(false).Comment("是否为211工程"),
		field.Bool("is_qs_top100").Default(false).Comment("是否为QS Top 100"),
		field.Int("rank_qs").Optional().Comment("QS最新排名"),
		field.Float("overall_score").Optional().Comment("综合评分"),
		field.JSON("metadata", map[string]interface{}{}).Optional().Comment("元数据，包含数据源、维护时间等"),
		field.Text("vector_content").Optional().Comment("用于向量化的文档内容"),
		field.Other("vector", &pgvector.Vector{}).Optional().SchemaType(map[string]string{
			"postgres": "vector",
		}).Comment("向量嵌入（需要在迁移中执行 CREATE EXTENSION IF NOT EXISTS vector;）"),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges 定义高校画像实体的边。
func (UniversityProfile) Edges() []ent.Edge {
	return nil
}

// Indexes 定义高校画像实体的索引。
func (UniversityProfile) Indexes() []ent.Index {
	return []ent.Index{
		// 中文名称唯一索引
		index.Fields("name_cn").Unique(),
		// 元数据 GIN 索引
		index.Fields("metadata").Annotations(
			entsql.IndexType("GIN"),
		),
		// 向量相似度索引 (IVFFLAT)，迁移脚本中需额外指定 WITH (lists = 100)
		index.Fields("vector").Annotations(
			entsql.IndexType("ivfflat"),
			entsql.OpClass("vector_cosine_ops"),
		),
		// 标签组合索引，用于快速筛选
		index.Fields("is_double_first_class", "is_project_985", "is_project_211", "is_qs_top100"),
		// 国家索引
		index.Fields("country"),
	}
}
