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

// ResumeDocumentParse holds the schema definition for the ResumeDocumentParse entity.
type ResumeDocumentParse struct {
	ent.Schema
}

func (ResumeDocumentParse) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "resume_document_parses",
		},
	}
}

func (ResumeDocumentParse) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the ResumeDocumentParse.
func (ResumeDocumentParse) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("resume_id", uuid.UUID{}),
		field.String("file_id").Optional(),        // docparser返回的文件ID
		field.Text("content").Optional(),          // docparser解析的原始文本内容
		field.String("file_type").Optional(),      // docparser识别的文件类型
		field.String("filename").Optional(),       // docparser返回的文件名
		field.String("title").Optional(),          // docparser解析的文档标题
		field.Time("upload_at").Optional(),        // docparser上传时间
		field.String("status").Default("pending"), // pending, success, failed
		field.String("error_message").Optional(),  // 解析错误信息
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the ResumeDocumentParse.
func (ResumeDocumentParse) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("resume", Resume.Type).Ref("document_parse").Field("resume_id").Unique().Required(),
	}
}
