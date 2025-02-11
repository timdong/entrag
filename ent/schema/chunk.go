package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Chunk holds the schema definition for the Chunk entity.
type Chunk struct {
	ent.Schema
}

// Fields of the Chunk.
func (Chunk) Fields() []ent.Field {
	return []ent.Field{
		field.String("path"),
		field.Int("nchunk"),
		field.Text("data"),
	}
}

// Edges of the Chunk.
func (Chunk) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("embedding", Embedding.Type).StorageKey(edge.Column("chunk_id")).Unique(),
	}
}
