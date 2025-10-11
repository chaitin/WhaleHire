package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/pkg/entx"
)

// ResumeProject holds the schema definition for the ResumeProject entity.
type ResumeProject struct {
	ent.Schema
}

func (ResumeProject) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "resume_projects",
		},
	}
}

func (ResumeProject) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the ResumeProject.
func (ResumeProject) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("resume_id", uuid.UUID{}),
		field.String("name").Optional(),           // 项目名称
		field.String("role").Optional(),           // 在项目中的角色
		field.String("company").Optional(),        // 项目所属公司/组织
		field.Time("start_date").Optional(),       // 项目开始时间
		field.Time("end_date").Optional(),         // 项目结束时间
		field.Text("description").Optional(),      // 项目描述
		field.Text("responsibilities").Optional(), // 项目职责
		field.Text("achievements").Optional(),     // 项目成果/成就
		field.String("technologies").Optional(),   // 使用的技术栈
		field.String("project_url").Optional(),    // 项目链接
		field.String("project_type").Optional(),   // 项目类型（个人项目、团队项目、开源项目等）
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the ResumeProject.
func (ResumeProject) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("resume", Resume.Type).Ref("projects").Field("resume_id").Unique().Required(),
	}
}
