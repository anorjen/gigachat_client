package config

import (
	"gigachat_client/internal/app/logger"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type AppConfig struct {
	Client struct {
		ClientId      string `yaml:"client_id"`
		ClientSecret  string `yaml:"client_secret"`
		SaveToken     bool   `yaml:"save_token" default:"true"`
		TokenFilePath string `yaml:"token_file_path" default:"token.json"`
	} `yaml:"client"`
	Chat struct {
		Model             string  `yaml:"model" default:"GigaChat:latest"`
		Temperature       float64 `yaml:"temperature" default:"0.87"`
		N                 int64   `yaml:"n" default:"1"`
		MaxTokens         int64   `yaml:"max_tokens" default:"512"`
		RepetitionPenalty float64 `yaml:"repetition_penalty" default:"1.07"`
		SaveHistory       bool    `yaml:"save_history" default:"true"`
		HistoryFilePath   string  `yaml:"history_file_path" default:"history.json"`
	} `yaml:"chat"`
	Log struct {
		Level string `yaml:"level" default:"none"`
		File  string `yaml:"file" default:""`
	} `yaml:"log"`
}

var (
	log   logger.Logger
	Flags struct {
		IsContinue bool
	}
)

func (c *AppConfig) GetConfig() *AppConfig {
	f, err := os.ReadFile("configs" + string(filepath.Separator) + "application.yml")

	if err != nil {
		log.LogFatal("Error reading config file", err)
	}

	err = yaml.Unmarshal(f, c)
	if err != nil {
		log.LogFatal("Error config file format", err)
	}

	if len(c.Client.ClientId) == 0 || len(c.Client.ClientSecret) == 0 {
		log.LogFatal("Empty client data", nil)
	}
	return c
}
