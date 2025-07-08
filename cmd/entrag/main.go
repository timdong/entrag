package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alecthomas/kong"
)

// CLI holds global options and subcommands.
type CLI struct {
	// Configuration file path
	Config string `kong:"help='Path to configuration file.',default='config.yaml'"`

	// Subcommands
	Load     *LoadCmd     `kong:"cmd,help='Load command that accepts a path.'"`
	Index    *IndexCmd    `kong:"cmd,help='Create embeddings for any chunks that do not have one.'"`
	Ask      *AskCmd      `kong:"cmd,help='Ask a question about the indexed documents'"`
	Stats    *StatsCmd    `kong:"cmd,help='Show statistics about chunks and embeddings'"`
	Cleanup  *CleanupCmd  `kong:"cmd,help='Remove orphaned chunks and optimize the database'"`
	Optimize *OptimizeCmd `kong:"cmd,help='Optimize system performance and warm up caches'"`

	// Internal config (loaded from file)
	cfg *Config `kong:"-"`
}

// LoadedConfig returns the loaded configuration
func (c *CLI) LoadedConfig() *Config {
	if c.cfg == nil {
		var err error
		c.cfg, err = LoadConfig(c.Config)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
	}
	return c.cfg
}

func main() {
	var cli CLI
	app := kong.Parse(&cli,
		kong.Name("entrag"),
		kong.Description("Ask questions about markdown files using RAG with Ollama."),
		kong.UsageOnError(),
	)

	// Load configuration
	cfg := cli.LoadedConfig()

	// Set the config in CLI for commands to access
	cli.cfg = cfg

	if err := app.Run(&cli); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
