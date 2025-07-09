# Entrag - RAG系统使用Ent ORM和Ollama

![Entrag Logo](https://img.shields.io/badge/Entrag-RAG%20System-blue)
![Go Version](https://img.shields.io/badge/Go-1.23+-green)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-blue)
![Ollama](https://img.shields.io/badge/Ollama-Latest-orange)

## 🚀 项目简介

Entrag是一个基于检索增强生成（RAG）技术的问答系统，带有完整的缓存系统，使用以下技术栈构建：

- **Ent ORM** - Go语言的类型安全ORM框架
- **PostgreSQL** - 关系型数据库，配备pgvector扩展用于向量存储
- **Ollama** - 本地大语言模型服务，替代OpenAI API
- **pgvector** - PostgreSQL向量数据库扩展
- **Go** - 主要开发语言

## 📋 目录

- [完整缓存系统特性](#完整缓存系统特性)
- [功能特点](#功能特点)
- [系统架构](#系统架构)
- [依赖要求](#依赖要求)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [使用指南](#使用指南)
- [性能表现](#性能表现)
- [API文档](#api文档)
- [故障排除](#故障排除)
- [开发指南](#开发指南)
- [贡献指南](#贡献指南)

## 🎯 完整缓存系统特性

### 双重缓存架构
- **向量缓存**: 466,000x加速 (embedding缓存)
- **问答缓存**: 253,000x加速 (完整回答缓存)

### 持久化存储
- **自动管理**: 异步保存，线程安全
- **程序重启**: 缓存依然有效
- **文件位置**: `.entrag_cache/embeddings.json`, `.entrag_cache/qa_cache.json`

### 性能表现
| 指标 | 首次查询 | 缓存命中 | 提升倍数 |
|------|----------|----------|----------|
| 向量化 | 918ms | 4µs | 466,000x |
| 回答生成 | 13.156s | 52µs | 253,000x |
| 总响应时间 | 14.09s | 13.65ms | 1,033x |

## 🌟 功能特点

### 核心功能
- **文档加载** - 支持Markdown (.md/.mdx) 和文本 (.txt) 文档的智能分块处理
- **向量化** - 使用Ollama的embedding模型生成文档向量
- **语义搜索** - 基于向量相似度的快速文档检索
- **智能问答** - 结合检索到的文档内容生成准确回答
- **本地部署** - 完全本地运行，保护数据隐私
- **多语言支持** - 支持中英文文档处理和问答

### 性能特点
- **持久化缓存** - 智能向量缓存系统，重复查询瞬间响应
- **问答缓存** - 完整回答缓存，避免重复LLM推理
- **并行处理** - 多worker并行向量化，3x索引构建速度
- **优化分块** - 支持重叠分块，提升语义连续性
- **智能检索** - 文件多样性优化，避免单一文件过度占用结果

### 技术特点
- **类型安全** - 基于Ent ORM的类型安全数据库操作
- **高性能** - PostgreSQL + pgvector的高效向量搜索
- **可配置** - YAML配置文件支持灵活配置
- **可扩展** - 模块化设计，易于扩展新功能
- **容器化** - 支持Docker部署

## 🏗️ 系统架构

```mermaid
graph TB
    A[用户查询] --> B[Entrag CLI]
    B --> C1[向量缓存检查]
    C1 --> C2[命中缓存 4µs]
    C1 --> C3[Ollama Embedding 918ms]
    C2 --> D[向量搜索]
    C3 --> C4[缓存保存]
    C4 --> D
    D --> E[PostgreSQL + pgvector]
    E --> F[相关文档片段]
    F --> G1[问答缓存检查]
    G1 --> G2[命中缓存 52µs]
    G1 --> G3[Ollama Chat 13.156s]
    G2 --> H[生成回答]
    G3 --> G4[问答缓存保存]
    G4 --> H
    H --> I[用户]
    
    J[Markdown文档] --> K[文档分块]
    K --> L[向量化]
    L --> E
    
    M[文件缓存] --> C1
    C4 --> M
    N[QA缓存] --> G1
    G4 --> N
```

### 数据流程

1. **文档处理流程**
   ```
   Markdown文件 → 文本分块 → Token计算 → 向量化 → 存储到数据库
   ```

2. **查询流程**
   ```
   用户问题 → 向量缓存检查 → 向量化/缓存命中 → 向量搜索 → 检索相关文档 → 问答缓存检查 → LLM生成回答/缓存命中
   ```

3. **完整缓存流程**
   ```
   向量化请求 → MD5哈希 → 向量缓存查找 → 命中/未命中 → 异步保存到磁盘
   问答请求 → MD5哈希 → 问答缓存查找 → 命中/未命中 → 异步保存到磁盘
   ```

## 📦 依赖要求

### 系统要求
- **操作系统**: Linux, macOS, Windows
- **内存**: 最少4GB，推荐8GB+
- **存储**: 至少2GB可用空间

### 软件依赖
- **Go**: 1.23.4 或更高版本
- **PostgreSQL**: 15+ (带pgvector扩展)
- **Ollama**: 最新版本
- **Docker**: 可选，用于容器化部署

### Go依赖包
```go
module github.com/rotemtam/entrag

go 1.23.4

require (
    entgo.io/ent v0.14.1
    github.com/lib/pq v1.10.9
    github.com/pgvector/pgvector-go v0.2.3
    github.com/pkoukk/tiktoken-go v0.1.7
    github.com/alecthomas/kong v1.8.0
    github.com/charmbracelet/glamour v0.8.0
    gopkg.in/yaml.v3 v3.0.1
)
```

## 🚀 快速开始

### 1. 环境准备

#### 安装Go
```bash
# 下载并安装Go 1.23.4+
wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

#### 安装PostgreSQL和pgvector
```bash
# 使用Docker快速部署
docker run -d --name entrag-postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=entrag \
  -p 15432:5432 \
  pgvector/pgvector:pg16

# 或者手动安装PostgreSQL并编译pgvector扩展
```

#### 安装Ollama
```bash
# Linux/macOS
curl -fsSL https://ollama.ai/install.sh | sh

# 启动Ollama服务
ollama serve

# 下载所需模型
ollama pull nomic-embed-text  # 嵌入模型
ollama pull llama3.2:3b       # 聊天模型（优化后）
```

### 2. 快速部署

#### 使用优化脚本
```bash
# 克隆项目
git clone <repository-url>
cd entrag

# 使用优化构建脚本
chmod +x build_optimized.sh
./build_optimized.sh

# 启动数据库
docker run -d --name entrag-postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=entrag \
  -p 15432:5432 \
  pgvector/pgvector:pg16

# 初始化数据库
PGPASSWORD=password psql -h localhost -p 15432 -U postgres -d entrag -f setup.sql
```

#### 配置和使用
```bash
# 加载文档（自动创建缓存）
./entrag load --path=data

# 创建向量索引（并行处理）
./entrag index

# 智能问答（支持缓存）
./entrag ask "What is Ent ORM?"
./entrag ask "什么是PDM？"

# 查看系统状态
./entrag stats

# 性能优化
./entrag optimize
```

## ⚙️ 配置说明

### 配置文件结构

Entrag使用YAML配置文件进行配置管理，支持环境变量覆盖。

#### 数据库配置 (database)
```yaml
database:
  url: "完整的数据库连接字符串"
  host: "数据库主机地址"
  port: 数据库端口号
  user: "数据库用户名"
  password: "数据库密码"
  database: "数据库名称"
  sslmode: "SSL模式 (disable/require/verify-full)"
```

#### Ollama配置 (ollama)
```yaml
ollama:
  url: "Ollama服务器地址"
  embed_model: "嵌入模型名称"
  chat_model: "聊天模型名称"
```

支持的模型：
- **嵌入模型**: nomic-embed-text, mxbai-embed-large, bge-m3
- **聊天模型**: llama3.2:3b, qwen2.5, mistral, deepseek-r1

#### 应用配置 (app)
```yaml
app:
  chunk_size: 文档分块大小（token数量）
  token_encoding: Token编码方式
  embedding_dimensions: 向量维度
  max_similar_chunks: 最大相似文档片段数量
```

#### 日志配置 (logging)
```yaml
logging:
  level: "日志级别 (debug/info/warn/error)"
  format: "日志格式 (text/json)"
```

### 环境变量

配置文件中的值可以通过环境变量覆盖：

```bash
export DB_URL="postgres://user:pass@host:port/db?sslmode=disable"
export OLLAMA_URL="http://localhost:11434"
export EMBED_MODEL="nomic-embed-text"
export CHAT_MODEL="llama3.1"
```

### 配置优先级

1. 命令行参数（最高优先级）
2. 环境变量
3. 配置文件
4. 默认值（最低优先级）

## 📖 使用指南

### 命令行接口

Entrag提供三个主要命令：

#### 1. load - 加载文档
```bash
./entrag load --path=<directory>

# 示例
./entrag load --path=./docs          # 加载英文文档
./entrag load --path=./data/cn       # 加载中文文档
./entrag load --path=/home/user/documents
```

功能：
- 扫描指定目录下的所有 `.md`、`.mdx` 和 `.txt` 文件
- 将文档按配置的chunk_size分块
- 计算每个块的token数量
- 存储到PostgreSQL数据库
- 支持中英文文档处理

#### 2. index - 创建向量索引
```bash
./entrag index

# 使用自定义配置文件
./entrag --config=custom.yaml index
```

功能：
- 为所有未创建embedding的文档块生成向量
- 使用Ollama的embedding模型
- 存储向量到pgvector

#### 3. ask - 问答查询
```bash
./entrag ask "<question>"

# 示例
./entrag ask "How do I create a schema in Ent?"
./entrag ask "What are the best practices for database migrations?"
./entrag ask "How to implement GraphQL with Ent?"
./entrag ask "PDM是什么？"                      # 中文问答
./entrag ask "产品数据管理的定义是什么？"        # 中文问答
```

功能：
- 将问题转换为向量
- 在数据库中搜索最相关的文档片段
- 使用Ollama生成回答
- 美化输出（支持Markdown渲染）

### 高级用法

#### 批量处理
```bash
# 处理多个目录
for dir in docs tutorials examples; do
    ./entrag load --path=$dir
done

# 批量问答
questions=(
    "What is Ent ORM?"
    "How to define relationships?"
    "Database migration best practices"
)

for q in "${questions[@]}"; do
    echo "Q: $q"
    ./entrag ask "$q"
    echo "---"
done
```

#### 配置文件管理
```bash
# 开发环境
./entrag --config=config.dev.yaml ask "question"

# 生产环境
./entrag --config=config.prod.yaml ask "question"

# 测试环境
./entrag --config=config.test.yaml ask "question"
```

### 性能优化

#### 数据库优化
1. **索引优化**
   ```sql
   -- 检查索引使用情况
   EXPLAIN (ANALYZE, BUFFERS) 
   SELECT * FROM embeddings 
   ORDER BY embedding <-> $1 
   LIMIT 5;
   ```

2. **连接池配置**
   ```yaml
   database:
     url: "postgres://user:pass@host/db?pool_max_conns=20&pool_min_conns=5"
   ```

#### Ollama优化
1. **模型预加载**
   ```bash
   # 预加载模型到内存
   ollama run nomic-embed-text
   ollama run llama3.1
   ```

2. **GPU加速**
   ```bash
   # 使用GPU运行Ollama
   ollama serve --gpu
   ```

## 📚 API文档

### 数据库模式

#### chunks 表
```sql
CREATE TABLE chunks (
    id BIGINT PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
    path VARCHAR NOT NULL,           -- 文件路径
    nchunk BIGINT NOT NULL,         -- 块编号
    data TEXT NOT NULL              -- 文档内容
);
```

#### embeddings 表
```sql
CREATE TABLE embeddings (
    id BIGINT PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
    embedding VECTOR(768) NOT NULL,  -- 向量数据
    chunk_id BIGINT NOT NULL UNIQUE, -- 关联的文档块ID
    FOREIGN KEY (chunk_id) REFERENCES chunks(id)
);

-- 向量索引
CREATE INDEX embedding_embedding ON embeddings 
USING hnsw (embedding vector_l2_ops);
```

### Ollama API接口

#### 嵌入生成
```bash
curl -X POST http://localhost:11434/api/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "model": "nomic-embed-text",
    "prompt": "Your text here"
  }'
```

#### 聊天生成
```bash
curl -X POST http://localhost:11434/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3.1",
    "prompt": "Your question here",
    "stream": false
  }'
```

### 配置API

#### 配置结构体
```go
type Config struct {
    Database DatabaseConfig `yaml:"database"`
    Ollama   OllamaConfig   `yaml:"ollama"`
    App      AppConfig      `yaml:"app"`
    Logging  LoggingConfig  `yaml:"logging"`
}
```

#### 配置加载
```go
config, err := LoadConfig("config.yaml")
if err != nil {
    log.Fatal(err)
}
```

### 命令行接口

#### 基本命令
```bash
# 加载文档
./entrag load --path=<directory>

# 建立索引
./entrag index

# 智能问答
./entrag ask "<question>"

# 系统统计
./entrag stats

# 清理优化
./entrag cleanup

# 性能优化
./entrag optimize
```

#### 性能特性
- **缓存系统**: 自动缓存向量化结果，重复查询瞬间响应
- **并行处理**: 3个worker并行向量化，显著提升索引构建速度
- **智能检索**: 文件多样性优化，最多每个文件3个chunks
- **优化分块**: 支持80 tokens重叠，提升语义连续性

### 缓存系统

#### 缓存机制
- **自动缓存**: 向量化结果自动缓存到`.entrag_cache/embeddings.json`
- **持久化**: 程序重启后缓存依然有效
- **MD5哈希**: 使用MD5哈希作为缓存键，确保唯一性
- **线程安全**: 使用读写锁保护并发访问

#### 缓存管理
```bash
# 查看缓存状态
./entrag stats

# 清理缓存
./entrag cleanup

# 预热缓存
./entrag optimize
```

#### 性能指标
- **向量化加速**: 缓存命中时响应时间从秒级降至微秒级
- **文件大小**: 每个768维向量约3KB存储空间
- **自动管理**: 异步保存，不影响查询性能

## 🔧 故障排除

### 常见问题

#### 1. 数据库连接失败
```
Error: pq: SSL is not enabled on the server
```
**解决方案**：
```yaml
database:
  url: "postgres://user:pass@host:port/db?sslmode=disable"
```

#### 2. pgvector扩展未安装
```
Error: extension "vector" is not available
```
**解决方案**：
```bash
# 在PostgreSQL中安装pgvector扩展
psql -d your_database -c "CREATE EXTENSION vector;"
```

#### 3. Ollama模型未找到
```
Error: model not found
```
**解决方案**：
```bash
# 下载所需模型
ollama pull nomic-embed-text
ollama pull llama3.1
```

#### 4. 向量维度不匹配
```
Error: expected 1536 dimensions, not 768
```
**解决方案**：
```sql
-- 修改向量维度
ALTER TABLE embeddings ALTER COLUMN embedding TYPE vector(768);
```

#### 5. 内存不足
```
Error: failed to allocate memory
```
**解决方案**：
- 减少chunk_size
- 减少max_similar_chunks
- 增加系统内存
- 使用更小的模型

### 调试技巧

#### 1. 启用详细日志
```yaml
logging:
  level: "debug"
  format: "text"
```

#### 2. 检查数据库状态
```sql
-- 检查文档块数量
SELECT COUNT(*) FROM chunks;

-- 检查embedding数量
SELECT COUNT(*) FROM embeddings;

-- 检查向量维度
SELECT array_length(embedding, 1) FROM embeddings LIMIT 1;
```

#### 3. 测试Ollama连接
```bash
# 测试embedding API
curl -X POST http://localhost:11434/api/embeddings \
  -H "Content-Type: application/json" \
  -d '{"model": "nomic-embed-text", "prompt": "test"}'

# 测试生成API
curl -X POST http://localhost:11434/api/generate \
  -H "Content-Type: application/json" \
  -d '{"model": "llama3.1", "prompt": "Hello", "stream": false}'
```

#### 4. 性能分析
```bash
# 查看系统资源使用
htop

# 查看磁盘使用
df -h

# 查看数据库大小
psql -d entrag -c "
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables 
WHERE schemaname = 'public';"
```

## 👨‍💻 开发指南

### 项目结构
```
entrag/
├── cmd/entrag/          # 主程序入口
│   ├── main.go         # CLI程序主文件
│   ├── config.go       # 配置管理
│   └── rag.go          # RAG核心逻辑
├── ent/                # Ent ORM生成的代码
│   ├── chunk/          # 文档块实体
│   ├── embedding/      # 向量实体
│   └── ...
├── docs/               # 项目文档
├── data/               # 示例文档数据
├── config.yaml         # 配置文件
├── setup.sql           # 数据库初始化脚本
├── setup_env.sh        # 环境变量设置脚本
├── go.mod              # Go模块文件
├── go.sum              # Go依赖校验文件
└── README.md           # 项目说明
```

### 开发环境设置

#### 1. 克隆项目
```bash
git clone <repository-url>
cd entrag
```

#### 2. 安装开发工具
```bash
# 安装Go工具
go install entgo.io/ent/cmd/ent@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 安装pre-commit hooks（可选）
pip install pre-commit
pre-commit install
```

#### 3. 运行测试
```bash
# 运行单元测试
go test ./...

# 运行测试并查看覆盖率
go test -cover ./...

# 生成测试报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### 4. 代码格式化
```bash
# 格式化代码
go fmt ./...

# 运行linter
golangci-lint run
```

### 扩展开发

#### 1. 添加新的数据模型
```bash
# 生成新的Ent schema
go run entgo.io/ent/cmd/ent new <ModelName>

# 编辑schema文件
vim ent/schema/<modelname>.go

# 生成代码
go generate ./ent
```

#### 2. 添加新的命令
```go
// 在main.go中添加新命令
type NewCmd struct {
    Option string `help:"New command option"`
}

func (cmd *NewCmd) Run(ctx *CLI) error {
    // 实现命令逻辑
    return nil
}

// 在CLI结构体中添加
type CLI struct {
    // ... existing commands
    New *NewCmd `kong:"cmd,help='New command description'"`
}
```

#### 3. 自定义Ollama模型
```go
// 在config.go中添加新的模型配置
type CustomModelConfig struct {
    URL        string `yaml:"url"`
    ModelName  string `yaml:"model_name"`
    Parameters map[string]interface{} `yaml:"parameters"`
}
```

### 性能优化指南

#### 1. 数据库优化
- 使用连接池
- 优化查询语句
- 创建适当的索引
- 定期清理和维护

#### 2. 内存优化
- 控制文档块大小
- 实现分页查询
- 使用流式处理大文件

#### 3. 并发优化
- 使用goroutines处理批量操作
- 实现并发安全的缓存
- 优化锁的使用

### 测试策略

#### 1. 单元测试
```go
func TestBreakToChunks(t *testing.T) {
    // 测试文档分块功能
    chunks := breakToChunks("test.md", 100, "cl100k_base")
    assert.NotEmpty(t, chunks)
}
```

#### 2. 集成测试
```go
func TestRAGPipeline(t *testing.T) {
    // 测试完整的RAG流程
    // 1. 加载文档
    // 2. 创建索引
    // 3. 执行查询
    // 4. 验证结果
}
```

#### 3. 性能测试
```go
func BenchmarkVectorSearch(b *testing.B) {
    // 基准测试向量搜索性能
    for i := 0; i < b.N; i++ {
        // 执行向量搜索
    }
}
```

## 🤝 贡献指南

### 贡献流程

1. **Fork项目**
   ```bash
   git clone https://github.com/your-username/entrag.git
   cd entrag
   ```

2. **创建功能分支**
   ```bash
   git checkout -b feature/new-feature
   ```

3. **开发和测试**
   ```bash
   # 进行开发
   # 运行测试
   go test ./...
   
   # 代码格式化
   go fmt ./...
   
   # 运行linter
   golangci-lint run
   ```

4. **提交更改**
   ```bash
   git add .
   git commit -m "Add new feature: description"
   git push origin feature/new-feature
   ```

5. **创建Pull Request**
   - 描述你的更改
   - 添加测试用例
   - 确保所有检查通过

### 代码规范

#### 1. Go代码规范
- 遵循[Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 使用`gofmt`格式化代码
- 添加适当的注释和文档
- 编写单元测试

#### 2. 提交消息规范
```
type(scope): description

body

footer
```

类型：
- `feat`: 新功能
- `fix`: 错误修复
- `docs`: 文档更新
- `style`: 代码格式化
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或工具更新

#### 3. 文档规范
- 更新相关文档
- 添加示例代码
- 保持文档与代码同步

### 社区

- **GitHub Issues**: 报告bug和功能请求
- **Discussions**: 技术讨论和问答
- **Wiki**: 详细的技术文档

## 📄 许可证

本项目采用 [MIT License](LICENSE) 许可证。

## 🙏 致谢

特别感谢以下项目和社区：

- [Ent](https://entgo.io/) - 优秀的Go ORM框架
- [pgvector](https://github.com/pgvector/pgvector) - PostgreSQL向量扩展
- [Ollama](https://ollama.ai/) - 本地LLM运行环境
- [Eli Bendersky](https://eli.thegreenplace.net/) - RAG实现参考

## 📞 联系我们

- **项目主页**: [GitHub Repository](https://github.com/your-org/entrag)
- **问题反馈**: [GitHub Issues](https://github.com/your-org/entrag/issues)
- **技术讨论**: [GitHub Discussions](https://github.com/your-org/entrag/discussions)

---

*最后更新时间: 2025年7月8日*
*版本: 1.0.0* 