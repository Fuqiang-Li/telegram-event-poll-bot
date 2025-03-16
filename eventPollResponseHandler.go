package main

import (
	"context"
	"log"
	"strings"

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
	user := getUserFullName(&update.CallbackQuery.From)
	optionInputs := strings.Split(update.CallbackQuery.Data, callbackSeparator)
	if len(optionInputs) < 2 {
		log.Println("invalid option callback", update.CallbackQuery.Data)
		return
	}
	option := optionInputs[1]
	optionType := optionInputs[2]
	eventUser := EventUser{
		EventID: event.ID,
		User:    user,
		Option:  option,
	}
	if optionType == callbackPostFixIn {
		err = h.eventDao.SaveEventUser(&eventUser)
		if err != nil {
			log.Println("error updating event user", err)
			return
		}
	} else if optionType == callbackPostFixOut {
		affectedRows, err := h.eventDao.DeleteEventUser(&eventUser)
		if affectedRows == 0 {
			log.Println("no event user deleted", err)
			return
		}
	}
	users, err := h.eventDao.GetEventUsers(event.ID)
	if err != nil {
		log.Println("error getting event users", err)
	}
	msgText, kb := getPollParams(*event, users)
	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      event.ChatID,
		MessageID:   messageID,
		Text:        msgText,
		ReplyMarkup: kb,
		ParseMode:   "Markdown",
	})
	if err != nil {
		log.Println("error editing message", err)
	}
}
