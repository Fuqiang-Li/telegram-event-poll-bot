package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type LogConfig struct {
	Filename   string `json:"filename"`
	MaxSize    int    `json:"maxsize"`
	MaxBackups int    `json:"maxbackups"`
	MaxAge     int    `json:"maxage"`
	Compress   bool   `json:"compress"`
}

type Config struct {
	TelegramToken string         `json:"telegram_token"`
	BotName       string         `json:"bot_name"`
	TimezoneStr   string         `json:"timezone"`
	Logger        LogConfig      `json:"logger"`
	Timezone      *time.Location `json:"-"`
	// Add other config fields as needed
}

var AppConfig *Config

func loadConfig(filename string) (*Config, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}
	tz, err := time.LoadLocation(config.TimezoneStr)
	if err != nil {
		log.Println("err loading time location", config.TimezoneStr, err)
		tz = time.UTC
	}
	config.Timezone = tz
	AppConfig = &config
	return AppConfig, nil
}
