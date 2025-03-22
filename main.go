package main

import (
	"context"
	"database/sql"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	// load config
	config, err := loadConfig("config.json")
	if err != nil {
		panic(err)
	}
	// setup logging
	setupLogging(config.Logger)
	defer log.Println("Stopping app")

	// Add database connection
	db, err := sql.Open("sqlite", "events.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Initialize the DAO
	eventDAO := NewEventDAO(db)
	err = eventDAO.Initialize()
	if err != nil {
		panic(err)
	}

	createEventHandler := NewCreateEventHandler(eventDAO)
	eventPollResponseHandler := NewEventPollResponseHandler(eventDAO)
	defaultHandler := NewDefaultHandler(createEventHandler)

	opts := []bot.Option{
		bot.WithDefaultHandler(defaultHandler.handle),
		bot.WithMessageTextHandler("/start", bot.MatchTypePrefix, createEventHandler.handleStart),
		bot.WithMessageTextHandler("/send", bot.MatchTypePrefix, createEventHandler.handleSend),
		bot.WithCallbackQueryDataHandler("event", bot.MatchTypePrefix, eventPollResponseHandler.handle),
	}
	b, err := bot.New(config.TelegramToken, opts...)
	if err != nil {
		panic(err)
	}

	log.Println("Starting App")
	b.Start(ctx)
}

func setupLogging(logger LogConfig) {
	// Configure the lumberjack logger
	logFile := &lumberjack.Logger{
		Filename:   logger.Filename,   // Log file path
		MaxSize:    logger.MaxSize,    // Max megabytes before rotation
		MaxBackups: logger.MaxBackups, // Max number of old log files to keep
		MaxAge:     logger.MaxAge,     // Max number of days to retain logs
		Compress:   logger.Compress,   // Compress the old logs
	}

	// Optionally, also log to console
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
}
