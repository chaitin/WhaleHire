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

// UserIdentity holds the schema definition for the UserIdentity entity.
type UserIdentity struct {
	ent.Schema
}

func (UserIdentity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entx.SoftDeleteMixin{},
	}
}

func (UserIdentity) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "user_identities",
		},
	}
}

// Fields of the UserIdentity.
func (UserIdentity) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}).Optional(),
		field.String("platform").GoType(consts.UserPlatform("")).Default(string(consts.UserPlatformEmail)),
		field.String("identity_id"),
		field.String("union_id").Optional(),
		field.String("nickname").Optional(),
		field.String("email").Optional(),
		field.String("avatar_url").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

// Edges of the UserIdentity.
func (UserIdentity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("identities").Field("user_id").Unique(),
	}
}
