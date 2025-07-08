# entrag
RAG demo with ent and Ollama

一个基于Ent ORM和Ollama的检索增强生成（RAG）演示项目。

## 功能特点

- 使用Ent ORM进行数据库操作
- 支持PostgreSQL + pgvector扩展用于向量存储
- 使用Ollama本地大语言模型替代OpenAI
- 支持Markdown文档的分块处理和向量化
- 提供问答功能

## 依赖要求

- Go 1.23+
- PostgreSQL 15+ (带pgvector扩展)
- Ollama服务器

## 安装和运行

### 1. 启动PostgreSQL数据库

```bash
# 使用Docker启动PostgreSQL容器
docker run -d --name entrag-postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=entrag \
  -p 15432:5432 \
  postgres:15-alpine

# 安装pgvector扩展（需要手动安装）
```

### 2. 初始化数据库

```bash
# 设置数据库连接
export DB_URL="postgres://postgres:password@localhost:15432/entrag?sslmode=disable"

# 运行数据库初始化脚本
PGPASSWORD=password psql -h localhost -p 15432 -U postgres -d entrag -f setup.sql
```

### 3. 启动Ollama服务器

```bash
# 安装并启动Ollama
ollama serve

# 下载所需模型
ollama pull nomic-embed-text  # 嵌入模型
ollama pull llama3.1          # 聊天模型
```

### 4. 构建项目

```bash
go mod tidy
go build -o entrag cmd/entrag/*.go
```

### 5. 使用项目

```bash
# 设置环境变量
export DB_URL="postgres://postgres:password@localhost:15432/entrag?sslmode=disable"
export OLLAMA_URL="http://localhost:11434"
export EMBED_MODEL="nomic-embed-text"
export CHAT_MODEL="llama3.1"

# 加载文档
./entrag load --path=data

# 创建向量索引
./entrag index

# 提问
./entrag ask "What is Ent ORM?"
./entrag ask "How to define relationships in Ent?"
./entrag ask "How to perform database migrations?"
```

## 环境变量

- `DB_URL`: PostgreSQL数据库连接字符串
- `OLLAMA_URL`: Ollama服务器地址 (默认: http://localhost:11434)
- `EMBED_MODEL`: 嵌入模型名称 (默认: nomic-embed-text)
- `CHAT_MODEL`: 聊天模型名称 (默认: llama3.1)

## 支持的命令

- `load --path=<directory>`: 加载指定目录下的Markdown文件
- `index`: 为未创建嵌入的文档块创建向量索引
- `ask <question>`: 基于索引的文档回答问题

## 注意事项

- 确保Ollama服务器正在运行并已下载相应模型
- 首次运行时需要下载嵌入模型，可能需要一些时间
- 建议使用支持中文的模型以获得更好的中文问答体验
