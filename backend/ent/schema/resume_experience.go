package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/pkg/entx"
	"github.com/google/uuid"
)

// ResumeExperience holds the schema definition for the ResumeExperience entity.
type ResumeExperience struct {
	ent.Schema
}

func (ResumeExperience) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "resume_experiences",
		},
	}
}

func (ResumeExperience) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the ResumeExperience.
func (ResumeExperience) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("resume_id", uuid.UUID{}),
		field.String("company").Optional(),  // 公司名称 / 社团组织名称
		field.String("position").Optional(), // 职位 / 在社团组织中的角色
		field.String("title").Optional(),    // 担任岗位
		field.Time("start_date").Optional(),
		field.Time("end_date").Optional(),
		field.Text("description").Optional(),
		field.String("experience_type").GoType(consts.ExperienceType("")).Optional(), // work, organization, volunteer, internship
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the ResumeExperience.
func (ResumeExperience) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("resume", Resume.Type).Ref("experiences").Field("resume_id").Unique().Required(),
	}
}
