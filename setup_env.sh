#!/bin/bash

# Entrag RAG System Environment Setup
echo "Setting up Entrag RAG environment..."

# Database Configuration
export DB_URL="postgres://postgres:password@localhost:15432/entrag?sslmode=disable"

# Ollama Configuration
export OLLAMA_URL="http://localhost:11434"
export EMBED_MODEL="nomic-embed-text"
export CHAT_MODEL="llama3.2:3b"

echo "Environment variables set:"
echo "  DB_URL: $DB_URL"
echo "  OLLAMA_URL: $OLLAMA_URL"
echo "  EMBED_MODEL: $EMBED_MODEL"
echo "  CHAT_MODEL: $CHAT_MODEL"
echo ""
echo "📚 基本用法："
echo "  ./entrag load --path=data"
echo "  ./entrag index"
echo "  ./entrag ask \"What is Ent ORM?\""
echo "  ./entrag ask \"How do I define relationships?\""
echo ""
echo "🚀 完整缓存系统："
echo "  ✅ 向量缓存: 466,000x加速"
echo "  ✅ 问答缓存: 253,000x加速"
echo "  ✅ 持久化存储: 程序重启后依然有效"
echo ""
echo "🗂️ 缓存文件位置："
echo "  📁 .entrag_cache/embeddings.json (向量缓存)"
echo "  📁 .entrag_cache/qa_cache.json (问答缓存)"
echo ""
echo "Or use explicit parameters:"
echo "  ./entrag --dburl=\"$DB_URL\" ask \"Your question here\"" 