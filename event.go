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
	timeFormat         = "2006-01-02 15:04"
	callbackPrefix     = "event"
	callbackPostFixIn  = "IN"
	callbackPostFixOut = "OUT"
	callbackSeparator  = "_"
)

func (e *Event) String() string {
	msg := fmt.Sprintf("*Description:* %s\n*Desired Pax:* %d\n*Max Pax:* %d", e.Description, e.DesiredPax, e.MaxPax)
	if e.StartedAt != nil {
		msg += fmt.Sprintf("\n*Starts at:* %s", e.StartedAt.Format(timeFormat))
	} else {
		msg += "\n*Starts at:* Not set"
	}
	msg += "\n*Options:*\n"
	msg += "• " + strings.Join(e.Options, "\n• ")
	return msg
}

func (e *Event) updateDetails(chatID int64, messageID int, createdBy string) {
	e.ChatID = chatID
	e.MessageID = messageID
	e.CreatedBy = createdBy
}

type EventAndUsers struct {
	Event
	OptionUsers map[string][]string
}

func (e *EventAndUsers) GetPollMessage() (string, *models.InlineKeyboardMarkup) {
	inlineKeyboard := make([][]models.InlineKeyboardButton, 0)
	for _, option := range e.Options {
		inlineKeyboard = append(inlineKeyboard, []models.InlineKeyboardButton{
			{Text: option + callbackSeparator + callbackPostFixIn, CallbackData: strings.Join([]string{callbackPrefix, option, callbackPostFixIn}, callbackSeparator)},
			{Text: option + callbackSeparator + callbackPostFixOut, CallbackData: strings.Join([]string{callbackPrefix, option, callbackPostFixOut}, callbackSeparator)},
		})
	}
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: inlineKeyboard,
	}

	msg := fmt.Sprintf("*Please cast your votes*\n%s\n", e.Description)
	if e.StartedAt != nil {
		msg += fmt.Sprintf("*Start Time:* %s\n", e.StartedAt.Format(timeFormat))
	}
	if e.DesiredPax > 0 {
		msg += fmt.Sprintf("*Desired Pax:* %d\n", e.DesiredPax)
	}
	if e.MaxPax > 0 {
		msg += fmt.Sprintf("*Max Pax:* %d\n", e.MaxPax)
	}
	for _, option := range e.Options {
		msg += fmt.Sprintf("*%s*:\n", option)
		msg += "• " + strings.Join(e.OptionUsers[option], "\n• ") + "\n"
	}
	return msg, kb
}

func sendEventPoll(ctx context.Context, b *bot.Bot, chatID any, event Event, users []EventUser) int {
	msgText, kb := getPollParams(event, users)
	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        msgText,
		ReplyMarkup: kb,
		ParseMode:   "Markdown",
	})
	if err != nil {
		log.Println("Error sending event poll to", chatID, err)
		return 0
	}
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
