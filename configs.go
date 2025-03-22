package main

import (
	"encoding/json"
	"os"
)

type LogConfig struct {
	Filename   string `json:"filename"`
	MaxSize    int    `json:"maxsize"`
	MaxBackups int    `json:"maxbackups"`
	MaxAge     int    `json:"maxage"`
	Compress   bool   `json:"compress"`
}

type Config struct {
	TelegramToken string    `json:"telegram_token"`
	Logger        LogConfig `json:"logger"`
	// Add other config fields as needed
}

func loadConfig(filename string) (*Config, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
