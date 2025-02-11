-- Create index "embedding_embedding" to table: "embeddings"
CREATE INDEX "embedding_embedding" ON "public"."embeddings" USING hnsw ("embedding" vector_l2_ops);
