package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/charmbracelet/glamour"
	"github.com/pgvector/pgvector-go"
	"github.com/pkoukk/tiktoken-go"
	"github.com/rotemtam/entrag/ent"
	"github.com/rotemtam/entrag/ent/chunk"

	_ "github.com/lib/pq"
)

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
			chunks := breakToChunks(path, cfg.App.ChunkSize, cfg.App.TokenEncoding)

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
	for _, ch := range chunks {
		log.Println("Created embedding for chunk", ch.Path, ch.Nchunk)
		embedding, err := getEmbedding(ch.Data, cfg.Ollama.URL, cfg.Ollama.EmbedModel)
		if err != nil {
			return fmt.Errorf("error getting embedding: %v", err)
		}
		_, err = client.Embedding.Create().
			SetEmbedding(pgvector.NewVector(embedding)).
			SetChunk(ch).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("error creating embedding: %v", err)
		}
	}
	return nil
}

// Run is the method called when the "ask" command is executed.
func (cmd *AskCmd) Run(ctx *CLI) error {
	cfg := ctx.LoadedConfig()
	client, err := ctx.entClient()
	if err != nil {
		return fmt.Errorf("failed opening connection to postgres: %w", err)
	}
	question := cmd.Text
	emb, err := getEmbedding(question, cfg.Ollama.URL, cfg.Ollama.EmbedModel)
	if err != nil {
		return fmt.Errorf("error getting embedding: %v", err)
	}
	embVec := pgvector.NewVector(emb)
	embs := client.Embedding.
		Query().
		Order(func(s *sql.Selector) {
			s.OrderExpr(sql.ExprP("embedding <-> $1", embVec))
		}).
		WithChunk().
		Limit(cfg.App.MaxSimilarChunks).
		AllX(context.Background())
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

	answer, err := getChatCompletion(query, cfg.Ollama.URL, cfg.Ollama.ChatModel)
	if err != nil {
		return fmt.Errorf("error creating chat completion: %v", err)
	}

	out, err := glamour.Render(answer, "dark")
	if err != nil {
		return fmt.Errorf("error rendering markdown: %v", err)
	}
	fmt.Print(out)
	return nil
}

func (c *CLI) entClient() (*ent.Client, error) {
	cfg := c.LoadedConfig()
	return ent.Open("postgres", cfg.Database.URL)
}

// breakToChunks reads the file in `path` and breaks it into chunks of
// approximately chunkSize tokens each, returning the chunks.
// This method  as well as `splitByParagraph` and `getEmbedding` were taken almost verbatim from Eli
// Bendersky's great blog post on RAGs with Go: https://eli.thegreenplace.net/2023/retrieval-augmented-generation-in-go
func breakToChunks(path string, chunkSize int, tokenEncoding string) []string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer f.Close()

	tke, err := tiktoken.GetEncoding(tokenEncoding)
	if err != nil {
		log.Fatalf("Error getting token encoding: %v", err)
	}

	chunks := []string{""}

	scanner := bufio.NewScanner(f)
	scanner.Split(splitByParagraph)

	for scanner.Scan() {
		chunks[len(chunks)-1] = chunks[len(chunks)-1] + scanner.Text() + "\n"
		toks := tke.Encode(chunks[len(chunks)-1], nil, nil)
		if len(toks) > chunkSize {
			chunks = append(chunks, "")
		}
	}

	// If we added a new empty chunk but there weren't any paragraphs to add to
	// it, make sure to remove it.
	if len(chunks[len(chunks)-1]) == 0 {
		chunks = chunks[:len(chunks)-1]
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

	return embedResp.Embedding, nil
}

// getChatCompletion invokes the Ollama chat API to generate a response
func getChatCompletion(prompt string, ollamaURL string, model string) (string, error) {
	reqBody := OllamaChatRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	resp, err := http.Post(ollamaURL+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s", string(body))
	}

	var chatResp OllamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	return chatResp.Response, nil
}
