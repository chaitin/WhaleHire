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

// ResumeJobApplication holds the schema definition for the ResumeJobApplication entity.
type ResumeJobApplication struct {
	ent.Schema
}

func (ResumeJobApplication) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "resume_job_applications",
		},
	}
}

func (ResumeJobApplication) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the ResumeJobApplication.
func (ResumeJobApplication) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("resume_id", uuid.UUID{}),
		field.UUID("job_position_id", uuid.UUID{}),
		field.String("status").Default("applied"), // applied, reviewing, interviewed, rejected, accepted
		field.String("source").Optional(),         // 申请来源：direct, referral, headhunter
		field.Text("notes").Optional(),            // 备注信息
		field.Time("applied_at").Default(time.Now),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the ResumeJobApplication.
func (ResumeJobApplication) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("resume", Resume.Type).
			Ref("job_applications").
			Field("resume_id").
			Unique().
			Required(),
		edge.From("job_position", JobPosition.Type).
			Ref("resume_applications").
			Field("job_position_id").
			Unique().
			Required(),
	}
}

// Indexes of the ResumeJobApplication.
func (ResumeJobApplication) Indexes() []ent.Index {
	return []ent.Index{
		// 确保同一简历不能重复申请同一岗位
		index.Fields("resume_id", "job_position_id").Unique(),
		// 提高查询性能
		index.Fields("resume_id"),
		index.Fields("job_position_id"),
		index.Fields("status"),
		index.Fields("applied_at"),
	}
}
