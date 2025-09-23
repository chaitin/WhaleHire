package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/ent/types"
)

// Setting holds the schema definition for the Setting entity.
type Setting struct {
	ent.Schema
}

func (Setting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "settings",
		},
	}
}

// Fields of the Setting.
func (Setting) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Bool("enable_sso").Default(false),
		field.Bool("force_two_factor_auth").Default(false),
		field.Bool("disable_password_login").Default(false),
		field.Bool("enable_auto_login").Default(false),
		field.JSON("dingtalk_oauth", &types.DingtalkOAuth{}).Optional(),
		field.JSON("custom_oauth", &types.CustomOAuth{}).Optional(),
		field.String("base_url").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Setting.
func (Setting) Edges() []ent.Edge {
	return nil
}
