package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/pkg/entx"
)

// AuditLog holds the schema definition for the AuditLog entity.
type AuditLog struct {
	ent.Schema
}

func (AuditLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "audit_logs",
		},
	}
}

func (AuditLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

// Fields of the AuditLog.
func (AuditLog) Fields() []ent.Field {
	return []ent.Field{
		// 主键ID
		field.UUID("id", uuid.UUID{}).Default(uuid.New),

		// 操作者信息
		field.String("operator_type").GoType(consts.OperatorType("")).Comment("操作者类型：user/admin"),
		field.UUID("operator_id", uuid.UUID{}).Optional().Comment("操作者ID"),
		field.String("operator_name").Optional().Comment("操作者名称"),

		// 操作信息
		field.String("operation_type").GoType(consts.OperationType("")).Comment("操作类型：create/update/delete/view/login/logout"),
		field.String("resource_type").GoType(consts.ResourceType("")).Comment("资源类型"),
		field.String("resource_id").Optional().Comment("资源ID"),
		field.String("resource_name").Optional().Comment("资源名称"),

		// 请求信息
		field.String("method").Comment("HTTP方法"),
		field.String("path").Comment("请求路径"),
		field.String("query_params").Optional().Comment("查询参数"),
		field.Text("request_body").Optional().Comment("请求体（敏感信息已脱敏）"),
		field.String("user_agent").Optional().Comment("用户代理"),

		// 响应信息
		field.Int("status_code").Comment("HTTP状态码"),
		field.String("status").GoType(consts.AuditLogStatus("")).Default(string(consts.AuditLogStatusSuccess)).Comment("操作状态：success/failed"),
		field.Text("response_body").Optional().Comment("响应体（敏感信息已脱敏）"),
		field.String("error_message").Optional().Comment("错误信息"),

		// 网络信息
		field.String("ip").Comment("客户端IP"),
		field.String("country").Optional().Comment("国家"),
		field.String("province").Optional().Comment("省份"),
		field.String("city").Optional().Comment("城市"),
		field.String("isp").Optional().Comment("运营商"),

		// 会话信息
		field.String("session_id").Optional().Comment("会话ID"),
		field.String("trace_id").Optional().Comment("链路追踪ID"),

		// 业务信息
		field.Text("business_data").Optional().Comment("业务相关数据（JSON格式）"),
		field.Text("changes").Optional().Comment("变更内容（JSON格式，记录before/after）"),
		field.String("description").Optional().Comment("操作描述"),

		// 时间信息
		field.Time("created_at").Default(time.Now).Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
		field.Int64("duration_ms").Optional().Comment("请求处理时长（毫秒）"),
	}
}

// Edges of the AuditLog.
func (AuditLog) Edges() []ent.Edge {
	return []ent.Edge{
		// 暂不定义边关系，避免循环依赖
		// 可以通过operator_id字段关联到具体的用户或管理员
	}
}

// Indexes of the AuditLog.
func (AuditLog) Indexes() []ent.Index {
	return []ent.Index{
		// 操作者相关索引
		index.Fields("operator_type", "operator_id"),
		index.Fields("operator_type", "created_at"),

		// 操作类型相关索引
		index.Fields("operation_type", "created_at"),
		index.Fields("resource_type", "created_at"),
		index.Fields("resource_type", "resource_id"),

		// 状态相关索引
		index.Fields("status", "created_at"),
		index.Fields("status_code", "created_at"),

		// IP相关索引
		index.Fields("ip", "created_at"),

		// 时间索引（用于日志清理和查询优化）
		index.Fields("created_at"),

		// 会话相关索引
		index.Fields("session_id"),
		index.Fields("trace_id"),

		// 复合索引用于常见查询场景
		index.Fields("operator_type", "operation_type", "created_at"),
		index.Fields("resource_type", "operation_type", "created_at"),
	}
}
