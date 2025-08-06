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

const (
	updatePollCallbackPrefix       = "updatePoll"
	updatePollCallbackkDesc        = "desc"
	updatePollCallbackStartedAt    = "startedAt"
	updatePollCallbackAddOption    = "addOption"
	updatePollCallbackDeleteOption = "deleteOption"
)

type UpdatePollResponse struct {
	MsgText string
	Step    int
}

var (
	updatePollCallbackResponses = map[string]UpdatePollResponse{
		updatePollCallbackkDesc:        {MsgText: "Please enter the new description for the event.", Step: 1},
		updatePollCallbackStartedAt:    {MsgText: "Please enter the new start time in the format YYYY-MM-DD HH:MM.", Step: 2},
		updatePollCallbackAddOption:    {MsgText: "Please enter the new option to add.", Step: 3},
		updatePollCallbackDeleteOption: {MsgText: "Please enter the option to delete.", Step: 4},
	}
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
	if !isSameUser(update.Message.From, event.CreatedBy, event.CreatedByID) {
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
	event.updateDetails(chatID, eventMsgID, event.CreatedBy, event.CreatedByID)
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
			Options: []string{"Available"},
		},
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          chatID,
		MessageThreadID: msgThreadID,
		Text:            "Let's start creating the event. First, please enter the description.",
	})
}

// Handler for creating an event from description
func (h *CreateEventHandler) handleCreateEvent(ctx context.Context, b *bot.Bot, update *models.Update, userStateKey string) {
	chatID := update.Message.Chat.ID
	msgThreadID := update.Message.MessageThreadID

	event := Event{
		Options:     []string{"Available"},
		Description: update.Message.Text,
	}

	event.updateDetails(chatID, 0, getUserFullName(update.Message.From), update.Message.From.ID)
	eventID, err := h.eventDao.SaveEvent(&event)
	if err != nil {
		log.Println("error saving event", err)
	}
	event.ID = eventID
	h.sendEvent(b, chatID, msgThreadID, &event, true)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          chatID,
		MessageThreadID: msgThreadID,
		Text:            fmt.Sprintf("/send@%s %d", h.botName, eventID),
	})
	// Clean up user state
	delete(userStates, userStateKey)
}

func (h *CreateEventHandler) handleUpdatePollCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	msgThreadID := update.CallbackQuery.Message.Message.MessageThreadID
	userStateKey := getUserStateKey(chatID, msgThreadID, &update.CallbackQuery.From)
	optionInputs := strings.Split(update.CallbackQuery.Data, callbackSeparator)
	if len(optionInputs) < 3 {
		log.Println("invalid callback data", update.CallbackQuery.Data)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       true,
			Text:            "Invalid callback data",
		})
		return
	}
	option := optionInputs[1]
	eventID, err := strconv.ParseInt(optionInputs[2], 10, 64)
	if err != nil {
		log.Println("error parsing event ID", err)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       true,
			Text:            "Invalid event ID",
		})
		return
	}
	event, err := h.eventDao.GetEventByID(eventID)
	if err != nil {
		log.Println("error getting event", err)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       true,
			Text:            "Event not found",
		})
		return
	}

	if !isSameUser(&update.CallbackQuery.From, event.CreatedBy, event.CreatedByID) {
		log.Println("event not created by user", getUserFullName(update.Message.From))
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       true,
			Text:            "You are not authorized to update this event",
		})
		return
	}

	// Handle all 4 update options: description, start time, add option, delete option
	response, ok := updatePollCallbackResponses[option]
	if !ok {
		log.Println("invalid option callback", option)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       true,
			Text:            "Invalid option callback",
		})
		return
	}
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          chatID,
		MessageThreadID: msgThreadID,
		Text:            response.MsgText,
	})

	userStates[userStateKey] = &UserState{
		StateType: UPDATE_EVENT,
		Event:     *event,
		Step:      response.Step,
	}
}

func (h *CreateEventHandler) handleUpdatePollInput(ctx context.Context, b *bot.Bot, update *models.Update, userStateKey string, userState *UserState) {
	chatID := update.Message.Chat.ID
	msgThreadID := update.Message.MessageThreadID

	switch userState.Step {
	case 1:
		// Collect description
		userState.Event.Description = update.Message.Text
	case 2:
		// Collect start time
		startTime, err := time.Parse(timeFormat, update.Message.Text)
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:          chatID,
				MessageThreadID: msgThreadID,
				Text:            "Invalid input. Please enter a valid start time in the format YYYY-MM-DD HH:MM. For example, " + timeFormat,
			})
			return
		}
		userState.Event.StartedAt = &startTime
	case 3:
		// Collect option to add
		option := strings.TrimSpace(update.Message.Text)
		if option == "" {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:          chatID,
				MessageThreadID: msgThreadID,
				Text:            "Empty input. Please enter the option to add.",
			})
			return
		}
		userState.Event.Options = append(userState.Event.Options, option)
	case 4:
		// Collect option to delete
		option := strings.TrimSpace(update.Message.Text)
		if option == "" {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:          chatID,
				MessageThreadID: msgThreadID,
				Text:            "Empty input. Please enter the option to delete.",
			})
			return
		}
		userState.Event.Options = deleteElementFromStrSlice(userState.Event.Options, option)
	}

	err := h.eventDao.UpdateEvent(&userState.Event)
	if err != nil {
		log.Println("error updating poll", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          chatID,
			MessageThreadID: msgThreadID,
			Text:            "Error updating poll",
		})
		return
	}
	h.sendEvent(b, chatID, msgThreadID, &userState.Event, false)
	delete(userStates, userStateKey)
}

func (h *CreateEventHandler) sendEvent(b *bot.Bot, chatID int64, msgThreadID int, event *Event, isNew bool) error {
	// Construct the inline keyboard for poll options
	eventIDStr := strconv.FormatInt(event.ID, 10)

	// Create inline keyboard for update options
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Description", CallbackData: strings.Join([]string{updatePollCallbackPrefix, updatePollCallbackkDesc, eventIDStr}, callbackSeparator)},
				{Text: "Start Time", CallbackData: strings.Join([]string{updatePollCallbackPrefix, updatePollCallbackStartedAt, eventIDStr}, callbackSeparator)},
			},
			{
				{Text: "Add Option", CallbackData: strings.Join([]string{updatePollCallbackPrefix, updatePollCallbackAddOption, eventIDStr}, callbackSeparator)},
				{Text: "Delete Option", CallbackData: strings.Join([]string{updatePollCallbackPrefix, updatePollCallbackDeleteOption, eventIDStr}, callbackSeparator)},
			},
		},
	}

	text := event.String() + "\n\n" + "You can update the poll by clicking the buttons below."
	if isNew {
		text += "\nYou can now send it to the group by copy pasting the following command sent as a separate message, in the format: " + fmt.Sprintf("/send@%s <EventID>", h.botName)
	}

	// Send the event as a message
	_, err := b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID:          chatID,
		MessageThreadID: msgThreadID,
		Text:            text,
		ParseMode:       "Markdown",
		ReplyMarkup:     keyboard,
	})

	return err
}
