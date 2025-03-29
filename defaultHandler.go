package main

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type DefaultHandler struct {
	createEventHandler *CreateEventHandler
	activityHandler    *ActivityHandler
}

func NewDefaultHandler(createEventHandler *CreateEventHandler, activityHandler *ActivityHandler) *DefaultHandler {
	return &DefaultHandler{createEventHandler: createEventHandler, activityHandler: activityHandler}
}

func (h *DefaultHandler) handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	msgThreadID := update.Message.MessageThreadID
	userStateKey := getUserStateKey(chatID, msgThreadID, update.Message.From)
	userState, exists := userStates[userStateKey]

	if !exists {
		return
	}

	switch userState.StateType {
	case CREATE_EVENT:
		h.createEventHandler.handleSteps(ctx, b, update)
	case ADD_ACTIVITY:
	case UPDATE_ACTIVITY:
	case DELETE_ACTIVITY:
	}
}
