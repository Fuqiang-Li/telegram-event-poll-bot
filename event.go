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
	eventYes   = "event_yes"
	eventNo    = "event_no"
	timeFormat = "2006-01-02 15:04"
)

func (e *Event) String() string {
	msg := fmt.Sprintf("Description: %s\nDesired Pax: %d\nMax Pax: %d", e.Description, e.DesiredPax, e.MaxPax)
	if e.StartedAt != nil {
		msg += fmt.Sprintf("\nStarts at: %s", e.StartedAt.Format(timeFormat))
	} else {
		msg += "\nStarts at: Not set"
	}
	return msg
}

func (e *Event) updateDetails(chatID int64, messageID int, createdBy string) {
	e.ChatID = chatID
	e.MessageID = messageID
	e.CreatedBy = createdBy
}

type EventAndUsers struct {
	Event
	Users []string
}

func (e *EventAndUsers) GetPollMessage() string {
	msg := fmt.Sprintf("Please cast your votes\n%s\n", e.Description)
	if e.StartedAt != nil {
		msg += fmt.Sprintf("Start Time: %s\n", e.StartedAt.Format(timeFormat))
	}
	if e.DesiredPax > 0 {
		msg += fmt.Sprintf("Desired Pax: %d\n", e.DesiredPax)
	}
	if e.MaxPax > 0 {
		msg += fmt.Sprintf("Max Pax: %d\n", e.MaxPax)
	}
	if len(e.Users) == 0 {
		msg += "No user opts in yet!"
		return msg
	}
	msg += "Opted In Users:\n"
	msg += "• " + strings.Join(e.Users, "\n• ") + "\n"
	return msg
}

func sendEventPoll(ctx context.Context, b *bot.Bot, chatID any, event Event, users []string) int {
	msgText, kb := getPollParams(event, users)
	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        msgText,
		ReplyMarkup: kb,
	})
	if err != nil {
		log.Println("Error sending event poll to", chatID, err)
		return 0
	}
	return msg.ID
}

func getPollParams(event Event, users []string) (string, *models.InlineKeyboardMarkup) {
	// TODO: accept options from user input
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Yes", CallbackData: eventYes},
				{Text: "No", CallbackData: eventNo},
			},
		},
	}
	eventAndUsers := EventAndUsers{
		Event: event,
		Users: users,
	}
	return eventAndUsers.GetPollMessage(), kb
}
