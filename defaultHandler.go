package main

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type DefaultHandler struct {
	createEventHandler *CreateEventHandler
}

func NewDefaultHandler(createEventHandler *CreateEventHandler) *DefaultHandler {
	return &DefaultHandler{createEventHandler: createEventHandler}
}

func (h *DefaultHandler) handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	if h.createEventHandler.handleSteps(ctx, b, update) {
		return
	}
}
