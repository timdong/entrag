package main

import (
	"bufio"
	"bytes"
	"context"
	"entgo.io/ent/dialect/sql"
	"fmt"
	"github.com/charmbracelet/glamour"
	"github.com/pgvector/pgvector-go"
	"github.com/pkoukk/tiktoken-go"
	"github.com/rotemtam/entrag/ent"
	"github.com/rotemtam/entrag/ent/chunk"
	"github.com/sashabaranov/go-openai"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
)

const (
	tokenEncoding = "cl100k_base"
	chunkSize     = 1000
)

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
	client, err := ctx.entClient()
	if err != nil {
		return fmt.Errorf("failed opening connection to postgres: %w", err)
	}
	tokTotal := 0
	return filepath.WalkDir(ctx.Load.Path, func(path string, d fs.DirEntry, err error) error {
		if filepath.Ext(path) == ".mdx" || filepath.Ext(path) == ".md" {
			log.Printf("Chunking %v", path)
			chunks := breakToChunks(path)

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
		embedding := getEmbedding(ch.Data)
		_, err := client.Embedding.Create().
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
	client, err := ctx.entClient()
	if err != nil {
		return fmt.Errorf("failed opening connection to postgres: %w", err)
	}
	question := cmd.Text
	emb := getEmbedding(question)
	embVec := pgvector.NewVector(emb)
	embs := client.Embedding.
		Query().
		Order(func(s *sql.Selector) {
			s.OrderExpr(sql.ExprP("embedding <-> $1", embVec))
		}).
		WithChunk().
		Limit(5).
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

	oac := openai.NewClient(ctx.OpenAIKey)
	resp, err := oac.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: query,
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("error creating chat completion: %v", err)
	}
	choice := resp.Choices[0]
	out, err := glamour.Render(choice.Message.Content, "dark")
	fmt.Print(out)
	return nil
}

func (c *CLI) entClient() (*ent.Client, error) {
	return ent.Open("postgres", c.DBURL)
}

// breakToChunks reads the file in `path` and breaks it into chunks of
// approximately chunkSize tokens each, returning the chunks.
func breakToChunks(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

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

// getEmbedding invokes the OpenAI embedding API to calculate the embedding
// for the given string. It returns the embedding.
func getEmbedding(data string) []float32 {
	client := openai.NewClient(os.Getenv("OPENAI_KEY"))

	queryReq := openai.EmbeddingRequest{
		Input: []string{data},
		Model: openai.AdaEmbeddingV2,
	}

	queryResponse, err := client.CreateEmbeddings(context.Background(), queryReq)
	if err != nil {
		log.Fatalf("Error getting embedding: %v", err)
	}
	return queryResponse.Data[0].Embedding
}
