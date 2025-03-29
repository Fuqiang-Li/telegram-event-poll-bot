package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type CreateEventHandler struct {
	eventDao *EventDAO
	botName  string
}

func NewCreateEventHandler(eventDao *EventDAO, botName string) *CreateEventHandler {
	return &CreateEventHandler{eventDao: eventDao, botName: botName}
}

func (h *CreateEventHandler) handleSend(ctx context.Context, b *bot.Bot, update *models.Update) {
	eventID, err := strconv.ParseInt(getCommandArgument(update), 10, 64)
	if err != nil {
		log.Println("error parsing event ID", err)
		return
	}
	event, err := h.eventDao.GetEventByID(eventID)
	if err != nil {
		log.Println("error getting event", err)
		return
	}
	chatID := update.Message.Chat.ID
	msgThreadID := update.Message.MessageThreadID
	if event.CreatedBy != getUserFullName(update.Message.From) {
		log.Println("event not created by user", getUserFullName(update.Message.From))
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          chatID,
			MessageThreadID: msgThreadID,
			Text:            "You are not authorized to send this event",
		})
		return
	}
	users, err := h.eventDao.GetEventUsers(eventID)
	if err != nil {
		log.Println("error getting event users", err)
	}
	eventMsgID := sendEventPoll(ctx, b, chatID, msgThreadID, *event, users)
	event.updateDetails(chatID, eventMsgID, event.CreatedBy)
	err = h.eventDao.UpdateEvent(event)
	if err != nil {
		log.Println("error saving event", err)
	}
	log.Println("event id", eventID)
}

func (h *CreateEventHandler) handleStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	msgThreadID := update.Message.MessageThreadID
	userStateKey := getUserStateKey(chatID, msgThreadID, update.Message.From)
	// Initialize user state
	userStates[userStateKey] = &UserState{
		Step:      1,
		StateType: CREATE_EVENT,
		Event: Event{
			Options: []string{"Support"},
		},
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          chatID,
		MessageThreadID: msgThreadID,
		Text:            "Let's start creating the event. First, please enter the description.",
	})
}

// Handler for collecting user responses step-by-step for an event
func (h *CreateEventHandler) handleSteps(ctx context.Context, b *bot.Bot, update *models.Update) bool {
	chatID := update.Message.Chat.ID
	msgThreadID := update.Message.MessageThreadID
	userStateKey := getUserStateKey(chatID, msgThreadID, update.Message.From)
	userState, exists := userStates[userStateKey]

	if !exists {
		return false
	}

	if strings.ToUpper(update.Message.Text) == "S" {
		userState.Step = -1
	}

	switch userState.Step {
	case 1:
		// Collect description
		userState.Event.Description = update.Message.Text
		userState.Step = 2
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          chatID,
			MessageThreadID: msgThreadID,
			Text: `Got it! You will be able to set the following settings in sequence. 
			1. Start Time
			2. Options (default: Support)
			At any step, you can enter 'S' to skip all the remaining settings.

			Now, please enter the start time. For example, ` + timeFormat + ". Enter 0 to skip.",
		})
	case 2:
		// Collect start time
		if update.Message.Text != "0" {
			startTime, err := time.Parse(timeFormat, update.Message.Text)
			if err != nil {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID:          chatID,
					MessageThreadID: msgThreadID,
					Text:            "Invalid input. Please enter a valid start time in the format YYYY-MM-DD HH:MM. For example, " + timeFormat,
				})
				return true
			}
			userState.Event.StartedAt = &startTime
		}
		userState.Step = 3
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          chatID,
			MessageThreadID: msgThreadID,
			Text:            "Now, please enter the options, separated by semicolon and should not contain underscore (e.g. Option 1;Option 2). Enter 0 to skip (default: Support).",
		})
	case 3:
		// Collect options
		if update.Message.Text != "0" {
			if strings.Contains(update.Message.Text, "_") {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID:          chatID,
					MessageThreadID: msgThreadID,
					Text:            "Invalid input. Please enter the options, separated by semicolon and should not contain underscore (e.g. Option 1;Option 2). Enter 0 to skip (default: Support).",
				})
				return true
			}
			options := strings.Split(update.Message.Text, ";")
			// Remove empty options
			var validOptions []string
			for _, option := range options {
				if opt := strings.TrimSpace(option); opt != "" {
					validOptions = append(validOptions, opt)
				}
			}
			userState.Event.Options = validOptions
		}

		userState.Step = -1
	}

	if userState.Step > 0 {
		return true
	}
	// Event collection complete
	event := userState.Event

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          chatID,
		Text:            event.String() + "\n\n" + "You can now send it to the group by copy pasting the following command sent as a separate message.",
		MessageThreadID: msgThreadID,
		ParseMode:       "Markdown",
	})

	event.updateDetails(chatID, 0, getUserFullName(update.Message.From))
	eventID, err := h.eventDao.SaveEvent(&event)
	if err != nil {
		log.Println("error saving event", err)
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          chatID,
		MessageThreadID: msgThreadID,
		Text:            fmt.Sprintf("/send@%s %d", h.botName, eventID),
	})
	// Clean up user state
	delete(userStates, userStateKey)
	return true
}
