package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	config, err := loadConfig("config.json")
	if err != nil {
		panic(err)
	}

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

	log.Println("Starting bot")
	b.Start(ctx)
}
