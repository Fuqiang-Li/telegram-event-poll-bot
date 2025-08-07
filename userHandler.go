package main

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// UserHandler handles user-related operations
type UserHandler struct {
	eventDAO *EventDAO
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(eventDAO *EventDAO) *UserHandler {
	return &UserHandler{
		eventDAO: eventDAO,
	}
}

// sendMyVotedEvents send events that a user has voted for
// only contains events starting from 2 months ago
func (h *UserHandler) sendMyVotedEvents(ctx context.Context, b *bot.Bot, update *models.Update) {
	// startTime := getCurrentMonthInUTC().AddDate(0, -2, 0)
	chatID := update.Message.Chat.ID
	msgThreadID := update.Message.MessageThreadID

	user := getUserFullName(update.Message.From)
	userID := update.Message.From.ID
	events, eventUsers, err := h.eventDAO.GetEventsVotedByUser(user, userID)
	if err != nil {
		log.Println("failed to get events voted", user, err)
		return
	}
	filteredEvents := make([]*Event, 0, len(events))
	startTime := getCurrentMonthInUTC().AddDate(0, -2, 0)
	for _, event := range events {
		if event.StartedAt == nil || event.StartedAt.After(startTime) {
			filteredEvents = append(filteredEvents, event)
		}
	}
	sort.Slice(filteredEvents, func(i, j int) bool {
		// the one with no startedAt is less - show first
		if filteredEvents[i].StartedAt == nil {
			return true
		}
		if filteredEvents[j].StartedAt == nil {
			return false
		}
		return filteredEvents[i].StartedAt.Before(*filteredEvents[j].StartedAt)
	})

	// For each event, collect all options the user voted for (could be multiple)
	userOptions := make(map[int64][]string)
	for _, eu := range eventUsers {
		userOptions[eu.EventID] = append(userOptions[eu.EventID], eu.Option)
	}

	text := fmt.Sprintf("You Voted Events: %d\n", len(filteredEvents))
	for i, e := range filteredEvents {
		text += fmt.Sprintf("*%d. Description:* %s", i+1, e.Description)
		if e.StartedAt != nil {
			text += fmt.Sprintf("\n*Starts at:* %s", e.StartedAt.Format(displayTimeFormat))
		} else {
			text += "\n*Starts at:* Not set"
		}
		text += "\n*Voted Option(s):*\n"
		opts := userOptions[e.ID]
		if len(opts) == 0 {
			text += "None\n"
		} else {
			for _, opt := range opts {
				text += fmt.Sprintf("• %s\n", opt)
			}
		}
		text += "\n"
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          chatID,
		MessageThreadID: msgThreadID,
		Text:            text,
		ParseMode:       "Markdown",
	})
}
