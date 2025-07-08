package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Ollama   OllamaConfig   `yaml:"ollama"`
	App      AppConfig      `yaml:"app"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	URL      string `yaml:"url"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"sslmode"`
}

// OllamaConfig represents Ollama configuration
type OllamaConfig struct {
	URL        string `yaml:"url"`
	EmbedModel string `yaml:"embed_model"`
	ChatModel  string `yaml:"chat_model"`
}

// AppConfig represents application configuration
type AppConfig struct {
	ChunkSize           int    `yaml:"chunk_size"`
	TokenEncoding       string `yaml:"token_encoding"`
	EmbeddingDimensions int    `yaml:"embedding_dimensions"`
	MaxSimilarChunks    int    `yaml:"max_similar_chunks"`
	ChunkOverlap        int    `yaml:"chunk_overlap"`
	MinChunkSize        int    `yaml:"min_chunk_size"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	// Default config path
	if configPath == "" {
		configPath = "config.yaml"
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Look for config file in the current directory
		if _, err := os.Stat("./config.yaml"); err == nil {
			configPath = "./config.yaml"
		} else {
			return nil, fmt.Errorf("config file not found: %s", configPath)
		}
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables if they exist
	if dbURL := os.Getenv("DB_URL"); dbURL != "" {
		config.Database.URL = dbURL
	}
	if ollamaURL := os.Getenv("OLLAMA_URL"); ollamaURL != "" {
		config.Ollama.URL = ollamaURL
	}
	if embedModel := os.Getenv("EMBED_MODEL"); embedModel != "" {
		config.Ollama.EmbedModel = embedModel
	}
	if chatModel := os.Getenv("CHAT_MODEL"); chatModel != "" {
		config.Ollama.ChatModel = chatModel
	}

	return &config, nil
}

// GetDefaultConfigPath returns the default config file path
func GetDefaultConfigPath() string {
	// Check current directory first
	if _, err := os.Stat("./config.yaml"); err == nil {
		return "./config.yaml"
	}

	// Check executable directory
	execPath, err := os.Executable()
	if err == nil {
		configPath := filepath.Join(filepath.Dir(execPath), "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// Return default
	return "config.yaml"
}
