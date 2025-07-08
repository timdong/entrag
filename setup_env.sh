#!/bin/bash

# Entrag RAG System Environment Setup
echo "Setting up Entrag RAG environment..."

# Database Configuration
export DB_URL="postgres://postgres:password@localhost:15432/entrag?sslmode=disable"

# Ollama Configuration
export OLLAMA_URL="http://localhost:11434"
export EMBED_MODEL="nomic-embed-text"
export CHAT_MODEL="llama3.1"

echo "Environment variables set:"
echo "  DB_URL: $DB_URL"
echo "  OLLAMA_URL: $OLLAMA_URL"
echo "  EMBED_MODEL: $EMBED_MODEL"
echo "  CHAT_MODEL: $CHAT_MODEL"
echo ""
echo "Usage examples:"
echo "  ./entrag load --path=data"
echo "  ./entrag index"
echo "  ./entrag ask \"What is Ent ORM?\""
echo "  ./entrag ask \"How do I define relationships?\""
echo ""
echo "Or use explicit parameters:"
echo "  ./entrag --dburl=\"$DB_URL\" ask \"Your question here\"" 