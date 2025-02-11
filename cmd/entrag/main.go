package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

// CLI holds global options and subcommands.
type CLI struct {
	// DBURL is read from the environment variable DB_URL.
	DBURL     string `kong:"env='DB_URL',help='Database URL for the application.'"`
	OpenAIKey string `kong:"env='OPENAI_KEY',help='OpenAI API key for the application.'"`

	// Subcommands
	Load  *LoadCmd  `kong:"cmd,help='Load command that accepts a path.'"`
	Index *IndexCmd `kong:"cmd,help='Create embeddings for any chunks that do not have one.'"`
	Ask   *AskCmd   `kong:"cmd,help='Ask a question about the indexed documents'"`
}

func main() {
	var cli CLI
	app := kong.Parse(&cli,
		kong.Name("entrag"),
		kong.Description("Ask questions about markdown files."),
		kong.UsageOnError(),
	)
	if err := app.Run(&cli); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
