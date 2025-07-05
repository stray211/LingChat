package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Character holds the schema definition for the Character entity.
type Character struct {
	ent.Schema
}

// Fields of the Character.
func (Character) Fields() []ent.Field {
	return []ent.Field{
		field.String("character_id").
			NotEmpty().
			Unique().
			Comment("The unique identifier of the character"),
		field.String("title").
			NotEmpty().
			Comment("The title/name of the character"),
		field.String("info").
			Default("").
			Comment("The description/info of the character"),
		field.String("resource_path").
			Default("").
			Comment("The resource path for character assets"),
		field.JSON("display_params", map[string]interface{}{}).
			Optional().
			Comment("JSON parameters for character display"),
		field.Text("system_prompt").
			Default("").
			Comment("The system prompt for the character"),
	}
}

// Edges of the Character.
func (Character) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the Character.
func (Character) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("character_id"),
	}
}

// Mixin of the Character.
func (Character) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimestampMixin{},
	}
}
