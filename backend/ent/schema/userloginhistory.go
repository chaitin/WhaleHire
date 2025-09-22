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
)

// UserLoginHistory holds the schema definition for the UserLoginHistory entity.
type UserLoginHistory struct {
	ent.Schema
}

func (UserLoginHistory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "user_login_histories",
		},
	}
}

// Fields of the UserLoginHistory.
func (UserLoginHistory) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}).Optional(),
		field.String("ip"),
		field.String("country"),
		field.String("province"),
		field.String("city"),
		field.String("isp").Optional(),
		field.String("asn").Optional(),
		field.String("client_version").Optional(),
		field.String("os_type").GoType(consts.OSType("")).Optional(),
		field.String("os_release").GoType(consts.OSRelease("")).Optional(),
		field.String("hostname").Optional(),
		field.String("client_id").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

// Edges of the UserLoginHistory.
func (UserLoginHistory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).Field("user_id").Ref("login_histories").Unique(),
	}
}
