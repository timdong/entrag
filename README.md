# entrag
RAG demo with ent and Ollama

一个基于Ent ORM和Ollama的检索增强生成（RAG）演示项目，带有完整的缓存系统和智能检索功能。

## 🚀 完整缓存系统特性

- **向量缓存**: 466,000x加速 (embedding缓存)
- **问答缓存**: 253,000x加速 (完整回答缓存)
- **持久化存储**: 程序重启后依然有效
- **自动管理**: 异步保存, 线程安全

## 🧠 智能检索系统特性

- **查询分类**: 自动识别概念性/操作性/比较性/列举性/通用查询
- **智能过滤**: 基于查询类型的上下文优化
- **文件多样性**: 防止单一文件过度引用
- **质量保证**: 3倍候选扩展+智能选择
- **兜底机制**: 确保总是有相关结果返回

## 🎯 核心功能

- 使用Ent ORM进行类型安全的数据库操作
- 支持PostgreSQL + pgvector扩展用于向量存储
- 使用Ollama本地大语言模型替代OpenAI
- 支持Markdown和文本文档的智能分块处理
- 提供高性能的问答功能
- 智能检索系统自动优化搜索结果

## 📦 依赖要求

- Go 1.23+
- PostgreSQL 15+ (带pgvector扩展)
- Ollama服务器

## 🛠️ 安装和运行

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
ollama pull llama3.2:3b       # 聊天模型（优化版）
```

### 4. 构建项目

```bash
# 快速构建
go build -o entrag cmd/entrag/*.go

# 或使用优化构建脚本
./build_optimized.sh
```

### 5. 使用项目

```bash
# 设置环境变量
source setup_env.sh

# 加载文档
./entrag load --path=data

# 创建向量索引
./entrag index

# 智能问答
./entrag ask "What is Ent ORM?"
./entrag ask "How to define relationships in Ent?"
./entrag ask "什么是产品数据管理？"
```

## 🔧 命令详解

### 核心命令

```bash
./entrag load --path=<directory>  # 加载文档
./entrag index                    # 建立向量索引
./entrag ask "<question>"         # 智能问答
./entrag stats                    # 统计信息
./entrag cleanup                  # 清理优化
./entrag optimize                 # 性能优化
```

### 缓存文件位置

```bash
.entrag_cache/
├── embeddings.json    # 向量缓存
└── qa_cache.json      # 问答缓存
```

## 🎯 性能表现

### 缓存性能
| 指标 | 首次查询 | 缓存命中 | 提升倍数 |
|------|----------|----------|----------|
| 向量化 | 918ms | 4µs | 466,000x |
| 回答生成 | 13.156s | 52µs | 253,000x |
| 总响应时间 | 14.09s | 13.65ms | 1,033x |

### 智能检索表现
| 查询类型 | 检索片段数 | 上下文长度 | 示例 |
|----------|------------|------------|------|
| 概念性 | 3-4个 | 10K-50K字符 | "什么是Ent ORM？" |
| 操作性 | 4-5个 | 5K-20K字符 | "如何定义关系？" |
| 比较性 | 2-3个 | 15K-30K字符 | "PDM和PLM的区别？" |
| 列举性 | 4个 | 10K-15K字符 | "列举Ent ORM的优点" |
| 通用 | 3个 | 3K-10K字符 | "hello" |

## 🛠️ 配置选项

使用 `config.yaml` 文件或环境变量：

- `DB_URL`: PostgreSQL数据库连接字符串
- `OLLAMA_URL`: Ollama服务器地址 (默认: http://localhost:11434)
- `EMBED_MODEL`: 嵌入模型名称 (默认: nomic-embed-text)
- `CHAT_MODEL`: 聊天模型名称 (默认: llama3.2:3b)

## 📊 测试工具

```bash
# 性能测试
./performance_test.sh

# 模型预加载
./preload_model.sh
```

## 🔍 故障排除

如果性能仍然较慢，建议：

1. 运行 `./entrag optimize` 预热缓存
2. 检查 `ollama ps` 确认模型已加载
3. 考虑使用更小的模型如 `gemma2:2b`
4. 调整 `max_similar_chunks` 到 2-3

## ⚖️ 智能度权衡

智能检索系统带来了权衡：
- ✅ **优点**: 更智能的文档检索，更好的回答质量
- ⚠️ **缺点**: 检索结果的随机性可能降低问答缓存命中率
- 🎯 **平衡**: 首次查询更精准，重复查询依然快速

## 📁 项目结构

```
entrag/
├── cmd/entrag/          # 主程序
├── ent/                 # 数据库模型
├── data/                # 测试文档
├── docs/                # 项目文档
├── .entrag_cache/       # 缓存文件
├── config.yaml          # 配置文件
└── *.sh                 # 构建和测试脚本
```

## �� 贡献

欢迎提交问题和拉取请求。
