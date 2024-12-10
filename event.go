package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	eventYes = "event_yes"
	eventNo  = "event_no"
)

// Struct to store the event details
type Event struct {
	Description  string
	MinPax       int
	MaxPax       int
	OptedInUsers map[string]struct{}
	ChatID       int64
}

func (e Event) String() string {
	return fmt.Sprintf("Description: %s\nMin Pax: %d\nMax Pax: %d", e.Description, e.MinPax, e.MaxPax)
}

func (e Event) GetPollMessage() string {
	msg := fmt.Sprintf("Please cast your votes\n%s\n", e.Description)
	if e.MinPax > 0 {
		msg += fmt.Sprintf("Min Pax: %d\n", e.MinPax)
	}
	if e.MaxPax > 0 {
		msg += fmt.Sprintf("Max Pax: %d\n", e.MaxPax)
	}
	if len(e.OptedInUsers) == 0 {
		msg += "No user opts in yet!"
		return msg
	}
	msg += "Opted In Users:\n"
	for user := range e.OptedInUsers {
		msg += fmt.Sprintf("• %s\n", user)
	}
	return msg
}

func (e Event) onFirstSent(chatID int64, messageID int) {
	e.ChatID = chatID
	e.OptedInUsers = map[string]struct{}{}
	pollMsgEvents[messageID] = &e
}

// State tracking for each user
type UserState struct {
	Step        int
	CurrentData Event
}

// Map to track state for each user
var userStates = make(map[int64]*UserState)
var pollMsgEvents = make(map[int]*Event)

func startCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	// Initialize user state
	userStates[chatID] = &UserState{Step: 1, CurrentData: Event{}}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Let's start creating the event. First, please enter the description.",
	})
}

// Handler for collecting user responses step-by-step for an event
func eventStepHandler(ctx context.Context, b *bot.Bot, update *models.Update) bool {
	chatID := update.Message.Chat.ID
	userState, exists := userStates[chatID]

	if !exists {
		return false
	}

	switch userState.Step {
	case 1:
		// Collect description
		userState.CurrentData.Description = update.Message.Text
		userState.Step = 2
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Got it! Now, please enter the minimum number of participants (min pax). Enter 0 to skip.",
		})
	case 2:
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
		userState.Step = 3
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Great! Now, please enter the maximum number of participants (max pax). Enter 0 to skip.",
		})
	case 3:
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

		// Event collection complete
		event := userState.CurrentData

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   event.String(),
		})

		// Clean up user state
		delete(userStates, chatID)
		eventMsgID := sendEventPoll(ctx, b, event, chatID)
		log.Println("event message id", eventMsgID)
		event.onFirstSent(chatID, eventMsgID)
		// TODO store event to db
	}
	return true
}

func sendEventPoll(ctx context.Context, b *bot.Bot, event Event, chatID any) int {
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

func eventPollResponseHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	messageID := update.CallbackQuery.Message.Message.ID
	log.Println("callback for message", messageID, "from", update.CallbackQuery.From.FirstName, update.CallbackQuery.From.LastName)
	event, ok := pollMsgEvents[messageID]
	if !ok || event == nil {
		log.Println("unknow messageID")
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       true,
			Text:            "Likely out-dated event. No more modification.",
		})
		return
	}
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})
	user := update.CallbackQuery.From.FirstName + " " + update.CallbackQuery.From.LastName
	// TODO: handle concurrent map update
	if update.CallbackQuery.Data == eventYes {
		event.OptedInUsers[user] = struct{}{}
	} else if update.CallbackQuery.Data == eventNo {
		delete(event.OptedInUsers, user)
	}
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Yes", CallbackData: eventYes},
				{Text: "No", CallbackData: eventNo},
			},
		},
	}
	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      event.ChatID,
		MessageID:   messageID,
		Text:        event.GetPollMessage(),
		ReplyMarkup: kb,
	})
}
