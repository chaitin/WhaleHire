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

// Resume holds the schema definition for the Resume entity.
type Resume struct {
	ent.Schema
}

func (Resume) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "resumes",
		},
	}
}

func (Resume) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the Resume.
func (Resume) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.String("name").Optional(),
		field.String("gender").Optional(),
		field.Time("birthday").Optional(), // 生t
		field.String("email").Optional(),
		field.String("phone").Optional(),
		field.String("current_city").Optional(),                 //当前城市
		field.String("highest_education").Optional().MaxLen(50), // 最高教育等级
		field.Float("years_experience").Optional(),              // 工作年限
		field.String("resume_file_url").Optional(),
		field.String("status").Default("pending"), // pending, processing, completed, failed, archived
		field.String("error_message").Optional(),
		field.Time("parsed_at").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Resume.
func (Resume) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("resumes").Field("user_id").Unique().Required(),
		edge.To("educations", ResumeEducation.Type),
		edge.To("experiences", ResumeExperience.Type),
		edge.To("skills", ResumeSkill.Type),
		edge.To("logs", ResumeLog.Type),
	}
}
