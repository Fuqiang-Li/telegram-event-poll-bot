package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	eventCallbackPrefix = "event"
	callbackPostFixIn   = "IN"
	callbackPostFixOut  = "OUT"
)

func (e *Event) String() string {
	msg := fmt.Sprintf("*Description:* %s", e.Description)
	if e.StartedAt != nil {
		msg += fmt.Sprintf("\n*Starts at:* %s", e.StartedAt.Format(displayTimeFormat))
	} else {
		msg += "\n*Starts at:* Not set"
	}
	msg += "\n*Options:*\n"
	msg += "• " + strings.Join(e.Options, "\n• ")
	return msg
}

func (e *Event) updateDetails(chatID int64, messageID int, createdBy string, createdByID int64) {
	e.ChatID = chatID
	e.MessageID = messageID
	e.CreatedBy = createdBy
	e.CreatedByID = createdByID
}

type EventAndUsers struct {
	Event
	OptionUsers map[string][]string
}

func (e *EventAndUsers) GetPollMessage() (string, *models.InlineKeyboardMarkup) {
	inlineKeyboard := make([][]models.InlineKeyboardButton, 0)
	for _, option := range e.Options {
		inlineKeyboard = append(inlineKeyboard, []models.InlineKeyboardButton{
			{Text: option, CallbackData: strings.Join([]string{eventCallbackPrefix, option}, callbackSeparator)},
		})
	}
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: inlineKeyboard,
	}

	msg := fmt.Sprintf("*Please cast your votes*\n%s\n", e.Description)
	if e.StartedAt != nil {
		msg += fmt.Sprintf("*Start Time:* %s\n", e.StartedAt.Format(displayTimeFormat))
	}
	for _, option := range e.Options {
		msg += fmt.Sprintf("*%s*:\n", option)
		msg += "• " + strings.Join(e.OptionUsers[option], "\n• ") + "\n"
	}
	return msg, kb
}

func sendEventPoll(ctx context.Context, b *bot.Bot, chatID any, messageThreadID int, event Event, users []EventUser) int {
	msgText, kb := getPollParams(event, users)
	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          chatID,
		MessageThreadID: messageThreadID,
		Text:            msgText,
		ReplyMarkup:     kb,
		ParseMode:       "Markdown",
	})
	if err != nil {
		log.Println("Error sending event poll to", chatID, err)
		return 0
	}
	log.Println("event poll", event.ID, "has been sent to chatID", chatID, "messageThreadID", messageThreadID)
	return msg.ID
}

func getPollParams(event Event, users []EventUser) (string, *models.InlineKeyboardMarkup) {
	eventAndUsers := EventAndUsers{
		Event:       event,
		OptionUsers: make(map[string][]string),
	}
	for _, user := range users {
		eventAndUsers.OptionUsers[user.Option] = append(eventAndUsers.OptionUsers[user.Option], user.User)
	}
	return eventAndUsers.GetPollMessage()
}
