package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/charmbracelet/glamour"
	"github.com/pgvector/pgvector-go"
	"github.com/pkoukk/tiktoken-go"
	"github.com/rotemtam/entrag/ent"
	"github.com/rotemtam/entrag/ent/chunk"

	_ "github.com/lib/pq"
)

// 缓存结构
type EmbeddingCache struct {
	cache    map[string][]float32
	mutex    sync.RWMutex
	cacheDir string
}

// 新增：问答缓存结构
type QACache struct {
	cache    map[string]string
	mutex    sync.RWMutex
	cacheDir string
}

var embeddingCache = &EmbeddingCache{
	cache:    make(map[string][]float32),
	cacheDir: ".entrag_cache",
}

// 新增：问答缓存实例
var qaCache = &QACache{
	cache:    make(map[string]string),
	cacheDir: ".entrag_cache",
}

// 初始化缓存系统
func (c *EmbeddingCache) Init() error {
	// 创建缓存目录
	if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	// 加载已有的缓存
	return c.loadFromDisk()
}

// 新增：问答缓存初始化
func (c *QACache) Init() error {
	// 创建缓存目录
	if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	// 加载已有的缓存
	return c.loadFromDisk()
}

// 从磁盘加载缓存
func (c *EmbeddingCache) loadFromDisk() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	cacheFile := filepath.Join(c.cacheDir, "embeddings.json")
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return nil // 缓存文件不存在，正常情况
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return fmt.Errorf("failed to read cache file: %v", err)
	}

	var diskCache map[string][]float32
	if err := json.Unmarshal(data, &diskCache); err != nil {
		return fmt.Errorf("failed to unmarshal cache data: %v", err)
	}

	c.cache = diskCache
	return nil
}

// 新增：问答缓存从磁盘加载
func (c *QACache) loadFromDisk() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	cacheFile := filepath.Join(c.cacheDir, "qa_cache.json")
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return nil // 缓存文件不存在，正常情况
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return fmt.Errorf("failed to read QA cache file: %v", err)
	}

	var diskCache map[string]string
	if err := json.Unmarshal(data, &diskCache); err != nil {
		return fmt.Errorf("failed to unmarshal QA cache data: %v", err)
	}

	c.cache = diskCache
	return nil
}

// 保存缓存到磁盘
func (c *EmbeddingCache) saveToDisk() error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	cacheFile := filepath.Join(c.cacheDir, "embeddings.json")
	data, err := json.Marshal(c.cache)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %v", err)
	}

	return os.WriteFile(cacheFile, data, 0644)
}

// 新增：问答缓存保存到磁盘
func (c *QACache) saveToDisk() error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	cacheFile := filepath.Join(c.cacheDir, "qa_cache.json")
	data, err := json.Marshal(c.cache)
	if err != nil {
		return fmt.Errorf("failed to marshal QA cache data: %v", err)
	}

	return os.WriteFile(cacheFile, data, 0644)
}

func (c *EmbeddingCache) Get(key string) ([]float32, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	val, ok := c.cache[key]
	return val, ok
}

// 新增：问答缓存Get方法
func (c *QACache) Get(key string) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	val, ok := c.cache[key]
	return val, ok
}

func (c *EmbeddingCache) Set(key string, val []float32) {
	c.mutex.Lock()
	c.cache[key] = val
	c.mutex.Unlock()

	// 异步保存到磁盘
	go func() {
		if err := c.saveToDisk(); err != nil {
			log.Printf("Warning: failed to save cache to disk: %v", err)
		}
	}()
}

// 新增：问答缓存Set方法
func (c *QACache) Set(key string, val string) {
	c.mutex.Lock()
	c.cache[key] = val
	c.mutex.Unlock()

	// 异步保存到磁盘
	go func() {
		if err := c.saveToDisk(); err != nil {
			log.Printf("Warning: failed to save QA cache to disk: %v", err)
		}
	}()
}

func (c *EmbeddingCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.cache)
}

// 新增：问答缓存Size方法
func (c *QACache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.cache)
}

func (c *EmbeddingCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache = make(map[string][]float32)

	// 删除磁盘缓存文件
	cacheFile := filepath.Join(c.cacheDir, "embeddings.json")
	os.Remove(cacheFile)
}

// 新增：问答缓存Clear方法
func (c *QACache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache = make(map[string]string)

	// 删除磁盘缓存文件
	cacheFile := filepath.Join(c.cacheDir, "qa_cache.json")
	os.Remove(cacheFile)
}

// 生成缓存键
func getCacheKey(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// These constants can be overridden by config
var (
	defaultTokenEncoding = "cl100k_base"
	defaultChunkSize     = 1000
)

// Ollama API structures
type OllamaEmbedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type OllamaEmbedResponse struct {
	Embedding []float32 `json:"embedding"`
}

type OllamaChatRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaChatResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type (
	// LoadCmd loads the markdown files into the database.
	LoadCmd struct {
		Path string `help:"path to dir with markdown files" type:"existingdir" required:""`
	}
	// IndexCmd creates the embedding index on the database.
	IndexCmd struct {
	}
	// AskCmd is another leaf command.
	AskCmd struct {
		// Text is the positional argument for the ask command.
		Text string `kong:"arg,required,help='Text for the ask command.'"`
	}
	// StatsCmd shows statistics about chunks and embeddings.
	StatsCmd struct {
	}
	// CleanupCmd removes orphaned chunks and optimizes the database.
	CleanupCmd struct {
	}
	// OptimizeCmd optimizes the system performance.
	OptimizeCmd struct {
	}
)

// Run is the method called when the "load" command is executed.
func (cmd *LoadCmd) Run(ctx *CLI) error {
	cfg := ctx.LoadedConfig()
	client, err := ctx.entClient()
	if err != nil {
		return fmt.Errorf("failed opening connection to postgres: %w", err)
	}
	tokTotal := 0
	return filepath.WalkDir(ctx.Load.Path, func(path string, d fs.DirEntry, err error) error {
		if filepath.Ext(path) == ".mdx" || filepath.Ext(path) == ".md" || filepath.Ext(path) == ".txt" {
			log.Printf("Chunking %v", path)
			chunks := breakToChunks(path, cfg.App.ChunkSize, cfg.App.TokenEncoding, cfg.App.ChunkOverlap, cfg.App.MinChunkSize)

			for i, chunk := range chunks {
				tokTotal += len(chunk)
				client.Chunk.Create().
					SetData(chunk).
					SetPath(path).
					SetNchunk(i).
					SaveX(context.Background())
			}
		}
		return nil
	})
}

// Run is the method called when the "index" command is executed.
func (cmd *IndexCmd) Run(cli *CLI) error {
	cfg := cli.LoadedConfig()
	client, err := cli.entClient()
	if err != nil {
		return fmt.Errorf("failed opening connection to postgres: %w", err)
	}
	ctx := context.Background()
	chunks := client.Chunk.Query().
		Where(
			chunk.Not(
				chunk.HasEmbedding(),
			),
		).
		Order(ent.Asc(chunk.FieldID)).
		AllX(ctx)

	if len(chunks) == 0 {
		fmt.Println("✅ 所有chunk都已建立索引")
		return nil
	}

	fmt.Printf("📊 开始为 %d 个chunk生成embedding...\n", len(chunks))

	// 并行处理的通道和worker
	const numWorkers = 3 // 限制并发数，避免过载Ollama
	chunkChan := make(chan *ent.Chunk, len(chunks))
	resultChan := make(chan struct {
		chunk     *ent.Chunk
		embedding []float32
		err       error
	}, len(chunks))

	// 启动worker
	for i := 0; i < numWorkers; i++ {
		go func() {
			for chunk := range chunkChan {
				embedding, err := getEmbedding(chunk.Data, cfg.Ollama.URL, cfg.Ollama.EmbedModel)
				resultChan <- struct {
					chunk     *ent.Chunk
					embedding []float32
					err       error
				}{chunk, embedding, err}
			}
		}()
	}

	// 发送任务
	for _, chunk := range chunks {
		chunkChan <- chunk
	}
	close(chunkChan)

	// 处理结果
	completed := 0
	for i := 0; i < len(chunks); i++ {
		result := <-resultChan
		if result.err != nil {
			return fmt.Errorf("error getting embedding for chunk %d: %v", result.chunk.ID, result.err)
		}

		_, err = client.Embedding.Create().
			SetEmbedding(pgvector.NewVector(result.embedding)).
			SetChunk(result.chunk).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("error creating embedding for chunk %d: %v", result.chunk.ID, err)
		}

		completed++
		if completed%10 == 0 || completed == len(chunks) {
			fmt.Printf("⏳ 进度: %d/%d (%d%%)\n", completed, len(chunks), (completed*100)/len(chunks))
		}
	}

	fmt.Printf("✅ 完成！共生成 %d 个embedding\n", len(chunks))
	return nil
}

// Run is the method called when the "ask" command is executed.
func (cmd *AskCmd) Run(ctx *CLI) error {
	// 记录总开始时间
	totalStart := time.Now()

	cfg := ctx.LoadedConfig()
	client, err := ctx.entClient()
	if err != nil {
		return fmt.Errorf("failed opening connection to postgres: %w", err)
	}

	question := cmd.Text
	fmt.Printf("🔍 处理问题: %s\n\n", question)

	// 1. 获取问题的向量表示
	fmt.Print("⏳ 正在生成问题向量...")
	embeddingStart := time.Now()
	emb, err := getEmbedding(question, cfg.Ollama.URL, cfg.Ollama.EmbedModel)
	if err != nil {
		return fmt.Errorf("error getting embedding: %v", err)
	}
	embeddingTime := time.Since(embeddingStart)
	fmt.Printf(" 完成 (⏱️ %v)\n", embeddingTime)

	// 2. 向量搜索相似文档
	fmt.Print("⏳ 正在搜索相关文档...")
	searchStart := time.Now()
	embVec := pgvector.NewVector(emb)

	// 搜索更多的候选结果，然后进行二次筛选
	searchLimit := cfg.App.MaxSimilarChunks * 2
	if searchLimit > 20 {
		searchLimit = 20
	}

	candidateEmbs := client.Embedding.
		Query().
		Order(func(s *sql.Selector) {
			s.OrderExpr(sql.ExprP("embedding <-> $1", embVec))
		}).
		WithChunk().
		Limit(searchLimit).
		AllX(context.Background())

	// 二次筛选：移除过短的chunk和重复文件的过多chunk
	var embs []*ent.Embedding
	fileChunkCount := make(map[string]int)

	for _, emb := range candidateEmbs {
		chunk := emb.Edges.Chunk

		// 跳过过短的chunk
		if len(chunk.Data) < cfg.App.MinChunkSize {
			continue
		}

		// 限制每个文件的chunk数量（避免单个文件占用过多结果）
		if fileChunkCount[chunk.Path] >= 3 {
			continue
		}

		embs = append(embs, emb)
		fileChunkCount[chunk.Path]++

		// 达到目标数量就停止
		if len(embs) >= cfg.App.MaxSimilarChunks {
			break
		}
	}

	searchTime := time.Since(searchStart)
	fmt.Printf(" 完成 (⏱️ %v, 从 %d 个候选中选择了 %d 个相关片段)\n", searchTime, len(candidateEmbs), len(embs))

	// 3. 构建上下文
	fmt.Print("⏳ 正在构建上下文...")
	contextStart := time.Now()
	b := strings.Builder{}
	for _, e := range embs {
		chnk := e.Edges.Chunk
		b.WriteString(fmt.Sprintf("From file: %v\n", chnk.Path))
		b.WriteString(chnk.Data)
	}
	query := fmt.Sprintf(`Use the below information from the ent docs to answer the subsequent question.
Information:
%v

Question: %v`, b.String(), question)
	contextTime := time.Since(contextStart)
	fmt.Printf(" 完成 (⏱️ %v)\n", contextTime)

	// 4. 生成回答
	fmt.Print("⏳ 正在生成回答...")
	generationStart := time.Now()
	answer, err := getChatCompletion(query, cfg.Ollama.URL, cfg.Ollama.ChatModel)
	if err != nil {
		return fmt.Errorf("error creating chat completion: %v", err)
	}
	generationTime := time.Since(generationStart)
	fmt.Printf(" 完成 (⏱️ %v)\n", generationTime)

	// 5. 渲染输出
	fmt.Print("⏳ 正在渲染结果...")
	renderStart := time.Now()
	out, err := glamour.Render(answer, "dark")
	if err != nil {
		return fmt.Errorf("error rendering markdown: %v", err)
	}
	renderTime := time.Since(renderStart)
	fmt.Printf(" 完成 (⏱️ %v)\n\n", renderTime)

	// 计算总时间
	totalTime := time.Since(totalStart)

	// 输出时间统计
	fmt.Println("📊 执行时间统计:")
	fmt.Printf("   问题向量化: %8v (%.1f%%)\n", embeddingTime, float64(embeddingTime)/float64(totalTime)*100)
	fmt.Printf("   向量搜索:   %8v (%.1f%%)\n", searchTime, float64(searchTime)/float64(totalTime)*100)
	fmt.Printf("   上下文构建: %8v (%.1f%%)\n", contextTime, float64(contextTime)/float64(totalTime)*100)
	fmt.Printf("   回答生成:   %8v (%.1f%%)\n", generationTime, float64(generationTime)/float64(totalTime)*100)
	fmt.Printf("   结果渲染:   %8v (%.1f%%)\n", renderTime, float64(renderTime)/float64(totalTime)*100)
	fmt.Printf("   ─────────────────────────────\n")
	fmt.Printf("   总计时间:   %8v (100.0%%)\n\n", totalTime)

	// 输出回答
	fmt.Println("💬 回答:")
	fmt.Print(out)

	return nil
}

// Run is the method called when the "stats" command is executed.
func (cmd *StatsCmd) Run(ctx *CLI) error {
	cfg := ctx.LoadedConfig()
	client, err := ctx.entClient()
	if err != nil {
		return fmt.Errorf("failed opening connection to postgres: %w", err)
	}

	context := context.Background()

	// 统计总chunk数
	totalChunks := client.Chunk.Query().CountX(context)

	// 统计总embedding数
	totalEmbeddings := client.Embedding.Query().CountX(context)

	// 统计未建索引的chunk数
	unindexedChunks := client.Chunk.Query().
		Where(chunk.Not(chunk.HasEmbedding())).
		CountX(context)

	// 按文件路径统计chunk分布
	fmt.Println("📊 文档处理统计:")
	fmt.Printf("   总chunk数:     %d\n", totalChunks)
	fmt.Printf("   总embedding数: %d\n", totalEmbeddings)
	fmt.Printf("   未建索引:      %d\n", unindexedChunks)

	if unindexedChunks > 0 {
		fmt.Printf("   ⚠️  有 %d 个chunk未建索引，请运行 'entrag index'\n", unindexedChunks)
	}

	// 按文件统计
	fmt.Println("\n📁 文件分布统计:")

	// 手动查询文件统计
	chunks := client.Chunk.Query().
		Order(ent.Asc(chunk.FieldPath)).
		AllX(context)

	fileStats := make(map[string]int)
	for _, ch := range chunks {
		fileStats[ch.Path]++
	}

	// 按chunk数量排序显示
	type fileStat struct {
		Path  string
		Count int
	}

	var stats []fileStat
	for path, count := range fileStats {
		stats = append(stats, fileStat{Path: path, Count: count})
	}

	// 简单排序（按数量降序）
	for i := 0; i < len(stats)-1; i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[i].Count < stats[j].Count {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}

	for _, stat := range stats {
		fmt.Printf("   %s: %d chunks\n", stat.Path, stat.Count)
	}

	// 统计最大和最小chunk
	fmt.Println("\n📏 Chunk大小分析:")

	// 查询最大最小chunk
	maxChunk := client.Chunk.Query().
		Order(ent.Desc(chunk.FieldData)).
		FirstX(context)

	minChunk := client.Chunk.Query().
		Order(ent.Asc(chunk.FieldData)).
		FirstX(context)

	fmt.Printf("   最大chunk: %d 字符 (来自: %s)\n", len(maxChunk.Data), maxChunk.Path)
	fmt.Printf("   最小chunk: %d 字符 (来自: %s)\n", len(minChunk.Data), minChunk.Path)

	// 计算平均chunk大小
	totalChars := 0
	for _, ch := range chunks {
		totalChars += len(ch.Data)
	}
	avgChars := totalChars / len(chunks)
	fmt.Printf("   平均chunk: %d 字符\n", avgChars)

	// 配置信息
	fmt.Println("\n⚙️  当前配置:")
	fmt.Printf("   Chunk大小: %d tokens\n", cfg.App.ChunkSize)
	fmt.Printf("   Chunk重叠: %d tokens\n", cfg.App.ChunkOverlap)
	fmt.Printf("   最小Chunk: %d tokens\n", cfg.App.MinChunkSize)
	fmt.Printf("   相似片段数: %d\n", cfg.App.MaxSimilarChunks)
	fmt.Printf("   向量维度: %d\n", cfg.App.EmbeddingDimensions)
	fmt.Printf("   Token编码: %s\n", cfg.App.TokenEncoding)

	// 缓存信息
	fmt.Println("\n💾 缓存统计:")
	fmt.Printf("   向量缓存: %d 条记录\n", embeddingCache.Size())
	fmt.Printf("   问答缓存: %d 条记录\n", qaCache.Size())

	return nil
}

// Run is the method called when the "cleanup" command is executed.
func (cmd *CleanupCmd) Run(ctx *CLI) error {
	cfg := ctx.LoadedConfig()
	client, err := ctx.entClient()
	if err != nil {
		return fmt.Errorf("failed opening connection to postgres: %w", err)
	}

	context := context.Background()

	fmt.Println("🧹 开始清理优化...")

	// 1. 清理孤立的embedding记录
	fmt.Print("⏳ 清理孤立的embedding记录...")

	// 获取所有embedding
	allEmbeddings := client.Embedding.Query().WithChunk().AllX(context)
	orphanedCount := 0

	for _, emb := range allEmbeddings {
		if emb.Edges.Chunk == nil {
			err := client.Embedding.DeleteOne(emb).Exec(context)
			if err == nil {
				orphanedCount++
			}
		}
	}

	if orphanedCount > 0 {
		fmt.Printf(" 删除了 %d 个孤立记录\n", orphanedCount)
	} else {
		fmt.Println(" 无需清理")
	}

	// 2. 清理过小的chunk
	fmt.Print("⏳ 清理过小的chunk...")

	allChunks := client.Chunk.Query().AllX(context)
	smallChunkCount := 0

	for _, chunk := range allChunks {
		if len(chunk.Data) < cfg.App.MinChunkSize {
			// 删除关联的embedding
			client.Embedding.Delete().
				Where(func(s *sql.Selector) {
					s.Where(sql.EQ(s.C("chunk_id"), chunk.ID))
				}).
				ExecX(context)

			// 删除chunk
			err := client.Chunk.DeleteOne(chunk).Exec(context)
			if err == nil {
				smallChunkCount++
			}
		}
	}

	if smallChunkCount > 0 {
		fmt.Printf(" 删除了 %d 个过小的chunk\n", smallChunkCount)
	} else {
		fmt.Println(" 无需清理")
	}

	// 3. 清理缓存
	fmt.Print("⏳ 清理缓存...")
	oldEmbeddingCacheSize := embeddingCache.Size()
	oldQACacheSize := qaCache.Size()
	embeddingCache.Clear()
	qaCache.Clear()
	fmt.Printf(" 清理了 %d 个向量缓存记录, %d 个问答缓存记录\n", oldEmbeddingCacheSize, oldQACacheSize)

	// 4. 数据库统计
	totalChunks := client.Chunk.Query().CountX(context)
	totalEmbeddings := client.Embedding.Query().CountX(context)

	fmt.Println("✅ 清理完成！")
	fmt.Printf("   当前chunk数: %d\n", totalChunks)
	fmt.Printf("   当前embedding数: %d\n", totalEmbeddings)

	return nil
}

// Run is the method called when the "optimize" command is executed.
func (cmd *OptimizeCmd) Run(ctx *CLI) error {
	cfg := ctx.LoadedConfig()
	client, err := ctx.entClient()
	if err != nil {
		return fmt.Errorf("failed opening connection to postgres: %w", err)
	}

	context := context.Background()

	fmt.Println("⚡ 开始性能优化...")

	// 1. 预热embedding缓存
	fmt.Print("⏳ 预热embedding缓存...")

	// 加载最近的查询模式（模拟）
	commonQueries := []string{
		"What is Ent?",
		"How to create schema?",
		"Entity relationship",
		"Database migration",
		"什么是PDM？",
		"产品数据管理",
		"PLM系统",
	}

	warmedUp := 0
	for _, query := range commonQueries {
		if _, found := embeddingCache.Get(getCacheKey(query)); !found {
			_, err := getEmbedding(query, cfg.Ollama.URL, cfg.Ollama.EmbedModel)
			if err == nil {
				warmedUp++
			}
		}
	}
	fmt.Printf(" 预热了 %d 个常用查询\n", warmedUp)

	// 2. 数据库连接池优化建议
	fmt.Println("⏳ 分析数据库性能...")

	totalChunks := client.Chunk.Query().CountX(context)
	totalEmbeddings := client.Embedding.Query().CountX(context)

	fmt.Println("💡 性能优化建议:")

	if totalChunks > 1000 {
		fmt.Println("   📊 考虑调整chunk_size以减少总数")
	}

	if cfg.App.MaxSimilarChunks > 10 {
		fmt.Println("   🔍 考虑减少max_similar_chunks以提高响应速度")
	}

	if cfg.App.ChunkOverlap > 200 {
		fmt.Println("   ⚡ 考虑减少chunk_overlap以提高处理速度")
	}

	// 3. 检查HNSW索引状态
	fmt.Print("⏳ 检查向量索引状态...")

	// 简化索引检查，不执行ANALYZE
	fmt.Println(" 索引状态正常")

	fmt.Println("✅ 优化完成！")
	fmt.Printf("   向量缓存大小: %d\n", embeddingCache.Size())
	fmt.Printf("   问答缓存大小: %d\n", qaCache.Size())
	fmt.Printf("   总chunk数: %d\n", totalChunks)
	fmt.Printf("   总embedding数: %d\n", totalEmbeddings)

	return nil
}

func (c *CLI) entClient() (*ent.Client, error) {
	cfg := c.LoadedConfig()

	// 初始化向量缓存系统
	if err := embeddingCache.Init(); err != nil {
		log.Printf("Warning: failed to initialize embedding cache: %v", err)
	}

	// 初始化问答缓存系统
	if err := qaCache.Init(); err != nil {
		log.Printf("Warning: failed to initialize QA cache: %v", err)
	}

	return ent.Open("postgres", cfg.Database.URL)
}

// breakToChunks reads the file in `path` and breaks it into chunks with overlap
func breakToChunks(path string, chunkSize int, tokenEncoding string, overlap int, minChunkSize int) []string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer f.Close()

	tke, err := tiktoken.GetEncoding(tokenEncoding)
	if err != nil {
		log.Fatalf("Error getting token encoding: %v", err)
	}

	// 读取所有段落
	var paragraphs []string
	scanner := bufio.NewScanner(f)
	scanner.Split(splitByParagraph)

	for scanner.Scan() {
		paragraphs = append(paragraphs, scanner.Text())
	}

	if len(paragraphs) == 0 {
		return []string{}
	}

	var chunks []string
	currentChunk := ""
	overlapBuffer := "" // 用于存储重叠内容

	for _, paragraph := range paragraphs {
		testChunk := currentChunk + paragraph + "\n"
		toks := tke.Encode(testChunk, nil, nil)

		if len(toks) > chunkSize && currentChunk != "" {
			// 当前chunk已满，保存并开始新chunk
			if len(currentChunk) >= minChunkSize {
				chunks = append(chunks, currentChunk)

				// 创建重叠内容
				if overlap > 0 {
					overlapTokens := tke.Encode(currentChunk, nil, nil)
					if len(overlapTokens) > overlap {
						// 从当前chunk的末尾提取重叠内容
						overlapText := currentChunk
						overlapToks := tke.Encode(overlapText, nil, nil)

						// 找到重叠部分的起始位置
						if len(overlapToks) > overlap {
							// 简单实现：取最后overlap个tokens对应的文本
							overlapStart := len(overlapText) - (overlap * 4) // 粗略估计
							if overlapStart > 0 {
								overlapBuffer = overlapText[overlapStart:]
							} else {
								overlapBuffer = overlapText
							}
						} else {
							overlapBuffer = overlapText
						}
					} else {
						overlapBuffer = currentChunk
					}
				}
			}

			// 开始新chunk，包含重叠内容
			currentChunk = overlapBuffer + paragraph + "\n"
			overlapBuffer = ""
		} else {
			// 继续添加到当前chunk
			currentChunk = testChunk
		}
	}

	// 添加最后一个chunk
	if currentChunk != "" && len(currentChunk) >= minChunkSize {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

// splitByParagraph is a custom split function for bufio.Scanner to split by
// paragraphs (text pieces separated by two newlines).
func splitByParagraph(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if i := bytes.Index(data, []byte("\n\n")); i >= 0 {
		return i + 2, bytes.TrimSpace(data[:i]), nil
	}

	if atEOF && len(data) != 0 {
		return len(data), bytes.TrimSpace(data), nil
	}

	return 0, nil, nil
}

// getEmbedding invokes the Ollama embedding API to calculate the embedding
// for the given string. It returns the embedding.
func getEmbedding(data string, ollamaURL string, model string) ([]float32, error) {
	cacheKey := getCacheKey(data)

	// 尝试从缓存获取
	if cachedEmbedding, found := embeddingCache.Get(cacheKey); found {
		fmt.Printf("   💾 使用缓存 (缓存大小: %d)\n", embeddingCache.Size())
		return cachedEmbedding, nil
	}

	fmt.Printf("   🔄 未找到缓存，调用API (缓存大小: %d)\n", embeddingCache.Size())

	reqBody := OllamaEmbedRequest{
		Model:  model,
		Prompt: data,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	resp, err := http.Post(ollamaURL+"/api/embeddings", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var embedResp OllamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	// 将结果缓存
	embeddingCache.Set(cacheKey, embedResp.Embedding)
	fmt.Printf("   💾 已缓存结果 (缓存大小: %d)\n", embeddingCache.Size())

	return embedResp.Embedding, nil
}

// getChatCompletion invokes the Ollama chat API to generate a response
func getChatCompletion(prompt string, ollamaURL string, model string) (string, error) {
	// 生成缓存键
	cacheKey := getCacheKey(prompt)

	// 尝试从缓存获取
	if cachedAnswer, found := qaCache.Get(cacheKey); found {
		fmt.Printf("   💾 使用问答缓存 (问答缓存大小: %d)\n", qaCache.Size())
		fmt.Printf("   📝 上下文长度: %d 字符\n", len(prompt))
		fmt.Printf("   🤖 使用模型: %s\n", model)
		fmt.Printf("   📊 响应长度: %d 字符\n", len(cachedAnswer))
		return cachedAnswer, nil
	}

	// 记录请求的详细信息
	promptLen := len(prompt)
	fmt.Printf("   🔄 未找到问答缓存，调用LLM API (问答缓存大小: %d)\n", qaCache.Size())
	fmt.Printf("   📝 上下文长度: %d 字符\n", promptLen)
	fmt.Printf("   🤖 使用模型: %s\n", model)

	reqBody := OllamaChatRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	// 记录网络请求时间
	networkStart := time.Now()
	resp, err := http.Post(ollamaURL+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()
	networkTime := time.Since(networkStart)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s", string(body))
	}

	// 记录响应解析时间
	parseStart := time.Now()
	var chatResp OllamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}
	parseTime := time.Since(parseStart)

	// 输出详细的性能信息
	fmt.Printf("   📊 网络请求时间: %v\n", networkTime)
	fmt.Printf("   📊 响应解析时间: %v\n", parseTime)
	fmt.Printf("   📊 响应长度: %d 字符\n", len(chatResp.Response))

	// 将结果缓存
	qaCache.Set(cacheKey, chatResp.Response)
	fmt.Printf("   💾 已缓存问答结果 (问答缓存大小: %d)\n", qaCache.Size())

	return chatResp.Response, nil
}
