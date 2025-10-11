package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/pkg/entx"
)

// ResumeEducation holds the schema definition for the ResumeEducation entity.
type ResumeEducation struct {
	ent.Schema
}

func (ResumeEducation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "resume_educations",
		},
	}
}

func (ResumeEducation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the ResumeEducation.
func (ResumeEducation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("resume_id", uuid.UUID{}),
		field.String("school").Optional(),
		field.String("degree").Optional(),
		field.String("major").Optional(),
		field.String("university_type").GoType(consts.UniversityType("")).Optional(),
		field.Time("start_date").Optional(),
		field.Time("end_date").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the ResumeEducation.
func (ResumeEducation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("resume", Resume.Type).Ref("educations").Field("resume_id").Unique().Required(),
	}
}
