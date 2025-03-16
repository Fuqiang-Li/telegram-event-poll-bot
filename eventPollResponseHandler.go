package main

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type EventPollResponseHandler struct {
	eventDao *EventDAO
}

func NewEventPollResponseHandler(eventDao *EventDAO) *EventPollResponseHandler {
	return &EventPollResponseHandler{eventDao: eventDao}
}

func (h *EventPollResponseHandler) handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	messageID := update.CallbackQuery.Message.Message.ID
	log.Println("callback for message", messageID, "from", update.CallbackQuery.From.FirstName, update.CallbackQuery.From.LastName)
	event, err := h.eventDao.GetEventByMessageID(messageID)
	if err != nil || event == nil {
		log.Println("unknow messageID", messageID, err)
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
	if update.CallbackQuery.Data == eventYes {
		err = h.eventDao.SaveEventUser(&EventUser{
			EventID: event.ID,
			User:    user,
		})
		if err != nil {
			log.Println("error updating event user", err)
			return
		}
	} else if update.CallbackQuery.Data == eventNo {
		affectedRows, err := h.eventDao.DeleteEventUser(&EventUser{
			EventID: event.ID,
			User:    user,
		})
		if affectedRows == 0 {
			log.Println("no event user deleted", err)
			return
		}
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Yes", CallbackData: eventYes},
				{Text: "No", CallbackData: eventNo},
			},
		},
	}
	eventUsers, err := h.eventDao.GetEventUsers(event.ID)
	if err != nil {
		log.Println("error getting event users", err)
	}
	users := make([]string, len(eventUsers))
	for i, user := range eventUsers {
		users[i] = user.User
	}
	eventAndUsers := EventAndUsers{
		Event: *event,
		Users: users,
	}
	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      event.ChatID,
		MessageID:   messageID,
		Text:        eventAndUsers.GetPollMessage(),
		ReplyMarkup: kb,
	})
	if err != nil {
		log.Println("error editing message", err)
	}
}
