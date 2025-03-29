package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	workplanCallbackPrefix         = "workplan"
	workplanOptionViewCurrentMonth = "view_current_month"
	workplanOptionViewCalendar     = "view_calendar"
	workplanOptionAddEvent         = "add_event"
	workplanOptionUpdateEvent      = "update_event"
	workplanOptionDeleteEvent      = "delete_event"
)

type ActivityHandler struct {
	activityDAO *ActivityDAO
}

func NewActivityHandler(activityDao *ActivityDAO) *ActivityHandler {
	return &ActivityHandler{activityDAO: activityDao}
}

func (h *ActivityHandler) handleWorkplan(ctx context.Context, b *bot.Bot, update *models.Update) {
	inlineKeyboard := [][]models.InlineKeyboardButton{
		{
			{Text: "View Current Month", CallbackData: strings.Join([]string{workplanCallbackPrefix, workplanOptionViewCurrentMonth}, callbackSeparator)},
			{Text: "View Calendar", CallbackData: strings.Join([]string{workplanCallbackPrefix, workplanOptionViewCalendar}, callbackSeparator)},
		},
		{
			{Text: "Add", CallbackData: strings.Join([]string{workplanCallbackPrefix, workplanOptionAddEvent}, callbackSeparator)},
			//{Text: "Update", CallbackData: strings.Join([]string{workplanCallbackPrefix, workplanOptionUpdateEvent}, callbackSeparator)},
			{Text: "Delete", CallbackData: strings.Join([]string{workplanCallbackPrefix, workplanOptionDeleteEvent}, callbackSeparator)},
		},
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: inlineKeyboard,
	}

	messageText := "Please choose an option:"
	chatID := update.Message.Chat.ID
	msgThreadID := update.Message.MessageThreadID
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          chatID,
		MessageThreadID: msgThreadID,
		Text:            messageText,
		ReplyMarkup:     kb,
		ParseMode:       "Markdown",
	})
}

func (h *ActivityHandler) handleWorkplanCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	messageID := update.CallbackQuery.Message.Message.ID
	log.Println("workpan callback for message", messageID, "from", getUserFullName(&update.CallbackQuery.From), "Data", update.CallbackQuery.Data)

	options := strings.Split(update.CallbackQuery.Data, callbackSeparator)
	if len(options) < 2 {
		log.Println("invalid option callback", update.CallbackQuery.Data)
		return
	}

	chatID := update.CallbackQuery.Message.Message.Chat.ID
	msgThreadID := update.CallbackQuery.Message.Message.MessageThreadID
	userStateKey := getUserStateKey(chatID, msgThreadID, &update.CallbackQuery.From)

	switch options[1] {
	case workplanOptionViewCurrentMonth:
		// Logic to view current month's activities
		startTime := time.Now().Truncate(24*time.Hour).AddDate(0, 0, -time.Now().Day()+1) // First day of the current month
		endTime := startTime.AddDate(0, 1, 0).Add(-time.Nanosecond)
		activities, err := h.activityDAO.GetByDuration(startTime, endTime)
		if err != nil {
			log.Println("error retrieving current month activities", err)
			return
		}
		messageText := "Current Month Activities:\n"
		for _, activity := range activities {
			messageText += activity.string() + "\n"
		}
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          chatID,
			MessageThreadID: msgThreadID,
			Text:            messageText,
			ParseMode:       "Markdown",
		})

	case workplanOptionViewCalendar:
		// Logic to view calendar of activities from past 2 months to the next 18 months
		startTime := time.Now().Truncate(24*time.Hour).AddDate(0, -2, -time.Now().Day()+1)
		endTime := startTime.AddDate(0, 18, 0).Add(-time.Nanosecond)
		activities, err := h.activityDAO.GetByDuration(startTime, endTime)
		if err != nil {
			log.Println("error retrieving activities for calendar", err)
			return
		}
		calendarText := "Activity Calendar:\n"
		for _, activity := range activities {
			calendarText += activity.string() + "\n"
		}
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          chatID,
			MessageThreadID: msgThreadID,
			Text:            calendarText,
			ParseMode:       "Markdown",
		})

	case workplanOptionAddEvent:
		// Logic to add a new event
		// Initialize user state
		userStates[userStateKey] = &UserState{Step: 1, StateType: ADD_ACTIVITY}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          chatID,
			MessageThreadID: msgThreadID,
			Text:            "Please provide the name for the new activity.",
		})

	case workplanOptionUpdateEvent:
		// Logic to update an existing event
		userStates[userStateKey] = &UserState{Step: 1, StateType: UPDATE_ACTIVITY}
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          chatID,
			MessageThreadID: msgThreadID,
			Text:            "Please provide the ID of the activity you want to update.",
		})

	case workplanOptionDeleteEvent:
		// Logic to delete an event
		userStates[userStateKey] = &UserState{Step: 1, StateType: DELETE_ACTIVITY}
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          chatID,
			MessageThreadID: msgThreadID,
			Text:            "Please provide the ID of the activity you want to delete.",
		})
	}
}

func (h *ActivityHandler) handleAddActivitySteps(ctx context.Context, b *bot.Bot, update *models.Update, userStateKey string, userState *UserState) {

}
