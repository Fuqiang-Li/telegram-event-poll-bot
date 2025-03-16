package main

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Map to track state for each user
var userStates = make(map[int64]*UserState)

type CreateEventHandler struct {
	eventDao *EventDAO
}

func NewCreateEventHandler(eventDao *EventDAO) *CreateEventHandler {
	return &CreateEventHandler{eventDao: eventDao}
}

func (h *CreateEventHandler) handleStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	// Initialize user state
	userStates[chatID] = &UserState{Step: 1, CurrentData: EventAndUsers{}}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Let's start creating the event. First, please enter the description.",
	})
}

// Handler for collecting user responses step-by-step for an event
func (h *CreateEventHandler) handleSteps(ctx context.Context, b *bot.Bot, update *models.Update) bool {
	chatID := update.Message.Chat.ID
	userState, exists := userStates[chatID]

	if !exists {
		return false
	}

	if strings.ToUpper(update.Message.Text) == "S" {
		userState.Step = -1
	}

	switch userState.Step {
	case 1:
		// Collect description
		userState.CurrentData.Description = update.Message.Text
		userState.Step = 2
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text: `Got it! You will be able to set the following settings in sequence. 
			1. Start Time
			2. Min Pax
			3. Max Pax
			At any step, you can enter 'S' to skip all the remaining settings.
			Now, please enter the start time. For example, ` + timeFormat + ". Enter 0 to skip.",
		})
	case 2:
		// Collect start time
		if update.Message.Text != "0" {
			startTime, err := time.Parse(timeFormat, update.Message.Text)
			if err != nil {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID,
					Text:   "Invalid input. Please enter a valid start time in the format YYYY-MM-DD HH:MM. For example, " + timeFormat,
				})
				return true
			}
			userState.CurrentData.StartedAt = &startTime
		}
		userState.Step = 3
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Great! Now, please enter the minimum number of participants (min pax). Enter 0 to skip.",
		})
	case 3:
		// Collect min pax
		minPax, err := strconv.Atoi(update.Message.Text)
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Invalid input. Please enter a valid number for min pax. Enter 0 to skip.",
			})
			return true
		}
		userState.CurrentData.MinPax = minPax
		userState.Step = 4
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Great! Now, please enter the maximum number of participants (max pax). Enter 0 to skip.",
		})
	case 4:
		// Collect max pax
		maxPax, err := strconv.Atoi(update.Message.Text)
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Invalid input. Please enter a valid number for max pax. Enter 0 to skip.",
			})
			return true
		}
		userState.CurrentData.MaxPax = maxPax
		userState.Step = -1
	}

	if userState.Step > 0 {
		return true
	}
	// Event collection complete
	event := userState.CurrentData

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   event.String(),
	})

	eventMsgID := sendEventPoll(ctx, b, event, chatID)
	log.Println("event message id", eventMsgID)
	event.onFirstSent(chatID, eventMsgID, update.Message.From.FirstName+" "+update.Message.From.LastName)
	err := h.eventDao.SaveEvent(&event.Event)
	if err != nil {
		log.Println("error saving event", err)
	}
	// Clean up user state
	delete(userStates, chatID)
	return true
}
