package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/pgvector/pgvector-go"
)

// Embedding holds the schema definition for the Embedding entity.
type Embedding struct {
	ent.Schema
}

// Fields of the Embedding.
func (Embedding) Fields() []ent.Field {
	return []ent.Field{
		field.Other("embedding", pgvector.Vector{}).
			SchemaType(map[string]string{
				dialect.Postgres: "vector(1536)",
			}),
	}
}

// Edges of the Embedding.
func (Embedding) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("chunk", Chunk.Type).Ref("embedding").Unique().Required(),
	}
}

func (Embedding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("embedding").
			Annotations(
				entsql.IndexType("hnsw"),
				entsql.OpClass("vector_l2_ops"),
			),
	}
}
