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

type EventAndUsers struct {
	Event
	Users []string
}

func (e *EventAndUsers) String() string {
	msg := fmt.Sprintf("Description: %s\nMin Pax: %d\nMax Pax: %d", e.Description, e.MinPax, e.MaxPax)
	if e.StartedAt != nil {
		msg += fmt.Sprintf("\nStarts at: %s", e.StartedAt.Format(timeFormat))
	}
	return msg
}

func (e *EventAndUsers) GetPollMessage() string {
	msg := fmt.Sprintf("Please cast your votes\n%s\n", e.Description)
	if e.StartedAt != nil {
		msg += fmt.Sprintf("Start Time: %s\n", e.StartedAt.Format(timeFormat))
	}
	if e.MinPax > 0 {
		msg += fmt.Sprintf("Min Pax: %d\n", e.MinPax)
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

func (e *EventAndUsers) onFirstSent(chatID int64, messageID int, createdBy string) {
	e.ChatID = chatID
	e.Users = []string{}
	e.MessageID = messageID
	e.CreatedBy = createdBy
}

// State tracking for each user
type UserState struct {
	Step        int
	CurrentData EventAndUsers
}

func sendEventPoll(ctx context.Context, b *bot.Bot, event EventAndUsers, chatID any) int {
	// TODO: accept options from user input
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Yes", CallbackData: eventYes},
				{Text: "No", CallbackData: eventNo},
			},
		},
	}

	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        event.GetPollMessage(),
		ReplyMarkup: kb,
	})
	if err != nil {
		log.Println("Error sending event poll to", chatID, err)
		return 0
	}
	return msg.ID
}
