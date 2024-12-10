package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithMessageTextHandler("/start", bot.MatchTypePrefix, startCommandHandler),
		bot.WithCallbackQueryDataHandler("event", bot.MatchTypePrefix, eventPollResponseHandler),
		bot.WithDefaultHandler(handler),
	}
	// todo read token from env
	token := os.Getenv("TG_BOT_TOKEN")
	b, err := bot.New(token, opts...)
	if err != nil {
		panic(err)
	}

	log.Println("Starting bot")
	b.Start(ctx)
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	if eventStepHandler(ctx, b, update) {
		return
	}
}
