package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// AdminLoginHistory holds the schema definition for the AdminLoginHistory entity.
type AdminLoginHistory struct {
	ent.Schema
}

func (AdminLoginHistory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "admin_login_histories",
		},
	}
}

// Fields of the AdminLoginHistory.
func (AdminLoginHistory) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}),
		field.UUID("admin_id", uuid.UUID{}).Optional(),
		field.String("ip"),
		field.String("country"),
		field.String("province"),
		field.String("city"),
		field.String("isp").Optional(),
		field.String("asn").Optional(),
		field.String("client_version").Optional(),
		field.String("device").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

// Edges of the AdminLoginHistory.
func (AdminLoginHistory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", Admin.Type).Field("admin_id").Ref("login_histories").Unique(),
	}
}
