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
	workplanCallbackPrefix                = "workplan"
	workplanOptionViewCurrentMonth        = "viewCurrentMonth"
	workplanOptionViewByMonth             = "viewByMonth"
	workplanOptionViewCalendar            = "viewCalendar"
	workplanOptionAddEvent                = "addEvent"
	workplanOptionUpdateEvent             = "updateEvent"
	workplanOptionDeleteEvent             = "deleteEvent"
	workplanViewByMonthCallbackPrefix     = "wpViewByMonth"
	workplanViewByMonthCallbackOptionAll  = "all"
	workplanViewByMonthCallbackOptionBack = "back"
)

type ActivityHandler struct {
	activityDAO *ActivityDAO
}

func NewActivityHandler(activityDao *ActivityDAO) *ActivityHandler {
	return &ActivityHandler{activityDAO: activityDao}
}

func (h *ActivityHandler) handleWorkplan(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	msgThreadID := update.Message.MessageThreadID
	kb, msg := h.getWorkplanMenu()
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          chatID,
		MessageThreadID: msgThreadID,
		Text:            msg,
		ReplyMarkup:     kb,
	})
}

func (h *ActivityHandler) getWorkplanMenu() (models.InlineKeyboardMarkup, string) {
	inlineKeyboard := [][]models.InlineKeyboardButton{
		{
			{Text: "View This Mo", CallbackData: strings.Join([]string{workplanCallbackPrefix, workplanOptionViewCurrentMonth}, callbackSeparator)},
			{Text: "View By Mo", CallbackData: strings.Join([]string{workplanCallbackPrefix, workplanOptionViewByMonth}, callbackSeparator)},
			{Text: "View All", CallbackData: strings.Join([]string{workplanCallbackPrefix, workplanOptionViewCalendar}, callbackSeparator)},
		},
		{
			{Text: "Add Event", CallbackData: strings.Join([]string{workplanCallbackPrefix, workplanOptionAddEvent}, callbackSeparator)},
			//{Text: "Update Event", CallbackData: strings.Join([]string{workplanCallbackPrefix, workplanOptionUpdateEvent}, callbackSeparator)},
			{Text: "Delete Event", CallbackData: strings.Join([]string{workplanCallbackPrefix, workplanOptionDeleteEvent}, callbackSeparator)},
		},
	}

	kb := models.InlineKeyboardMarkup{
		InlineKeyboard: inlineKeyboard,
	}
	messageText := "Please choose an option:"
	return kb, messageText
}

func (h *ActivityHandler) handleWorkplanCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	messageID := update.CallbackQuery.Message.Message.ID
	log.Println("workpan callback for message", messageID, "from", getUserFullName(&update.CallbackQuery.From), "Data", update.CallbackQuery.Data)

	options := strings.Split(update.CallbackQuery.Data, callbackSeparator)
	if len(options) < 2 {
		log.Println("invalid option callback", update.CallbackQuery.Data)
		return
	}

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	chatID := update.CallbackQuery.Message.Message.Chat.ID
	msgThreadID := update.CallbackQuery.Message.Message.MessageThreadID
	userStateKey := getUserStateKey(chatID, msgThreadID, &update.CallbackQuery.From)

	switch options[1] {
	case workplanOptionViewCurrentMonth:
		// Logic to view current month's activities
		startTime := getCurrentMonthStartInUTC()
		endTime := startTime.AddDate(0, 1, 0).Add(-time.Nanosecond)
		h.sendActivitiesForPeriod(ctx, b, chatID, msgThreadID, startTime, endTime)

	case workplanOptionViewByMonth:
		// Logic to view activities by month
		startTime := getCurrentMonthStartInUTC()
		var inlineButtons [][]models.InlineKeyboardButton
		for i := -2; i < 16; i++ {
			month := startTime.AddDate(0, i, 0)
			button := models.InlineKeyboardButton{
				Text:         month.Format(monthFormat),
				CallbackData: strings.Join([]string{workplanViewByMonthCallbackPrefix, month.Format(monthFormat)}, callbackSeparator),
			}
			if len(inlineButtons) == 0 || len(inlineButtons[len(inlineButtons)-1]) == 4 {
				inlineButtons = append(inlineButtons, []models.InlineKeyboardButton{button})
			} else {
				inlineButtons[len(inlineButtons)-1] = append(inlineButtons[len(inlineButtons)-1], button)
			}
		}
		// add special buttons
		button := models.InlineKeyboardButton{
			Text:         "All",
			CallbackData: strings.Join([]string{workplanViewByMonthCallbackPrefix, workplanViewByMonthCallbackOptionAll}, callbackSeparator),
		}
		inlineButtons[len(inlineButtons)-1] = append(inlineButtons[len(inlineButtons)-1], button)
		button = models.InlineKeyboardButton{
			Text:         "<< back",
			CallbackData: strings.Join([]string{workplanViewByMonthCallbackPrefix, workplanViewByMonthCallbackOptionBack}, callbackSeparator),
		}
		inlineButtons[len(inlineButtons)-1] = append(inlineButtons[len(inlineButtons)-1], button)

		kb := &models.InlineKeyboardMarkup{
			InlineKeyboard: inlineButtons,
		}

		messageText := "Select a month to view activities:"
		_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      chatID,
			MessageID:   messageID,
			Text:        messageText,
			ReplyMarkup: kb,
		})
		if err != nil {
			log.Println("error editing message", err)
		}

	case workplanOptionViewCalendar:
		h.sendAllActivities(ctx, b, chatID, msgThreadID)

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

func (h *ActivityHandler) handleViewByMonth(ctx context.Context, b *bot.Bot, update *models.Update) {
	messageID := update.CallbackQuery.Message.Message.ID
	log.Println("workpan callback for message", messageID, "from", getUserFullName(&update.CallbackQuery.From), "Data", update.CallbackQuery.Data)

	options := strings.Split(update.CallbackQuery.Data, callbackSeparator)
	if len(options) < 2 {
		log.Println("invalid option callback", update.CallbackQuery.Data)
		return
	}

	chatID := update.CallbackQuery.Message.Message.Chat.ID
	msgThreadID := update.CallbackQuery.Message.Message.MessageThreadID

	switch options[1] {
	case workplanViewByMonthCallbackOptionAll:
		h.sendAllActivities(ctx, b, chatID, msgThreadID)
		return
	case workplanViewByMonthCallbackOptionBack:
		kb, msg := h.getWorkplanMenu()
		_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      chatID,
			MessageID:   messageID,
			Text:        msg,
			ReplyMarkup: kb,
		})
		if err != nil {
			log.Println("error editing message", err)
		}
		return
	}

	month, err := time.Parse(monthFormat, options[1])
	if err != nil {
		log.Println("error parsing month", options[1], err)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "cannot get the month",
			ShowAlert:       true,
		})
		return
	}

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	endTime := month.AddDate(0, 1, 0).Add(-time.Nanosecond)
	h.sendActivitiesForPeriod(ctx, b, chatID, msgThreadID, month, endTime)
}

func (h *ActivityHandler) handleAddActivitySteps(ctx context.Context, b *bot.Bot, update *models.Update, userStateKey string, userState *UserState) {
	chatID := update.Message.Chat.ID
	msgThreadID := update.Message.MessageThreadID

	switch userState.Step {
	case 1:
		// Collect activity name
		userState.Activity.Name = update.Message.Text
		userState.Step = 2
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          chatID,
			MessageThreadID: msgThreadID,
			Text:            "Got it! Now please enter the start time (e.g., YYYY-MM-DD HH:MM).",
		})
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
		userState.Activity.StartedAt = startTime
		userState.Step = 3
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          chatID,
			MessageThreadID: msgThreadID,
			Text:            fmt.Sprintf("Next, please enter the name of the organizing committee. One of %v", AllOrgs),
		})
	case 3:
		// Collect organizing committee
		orgInput := Org(strings.ToUpper(update.Message.Text))

		// Check if the provided organization is valid
		isValidOrg := false
		for _, org := range AllOrgs {
			if orgInput == org {
				isValidOrg = true
				break
			}
		}

		if !isValidOrg {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:          chatID,
				MessageThreadID: msgThreadID,
				Text:            fmt.Sprintf("Invalid org. Please enter one of %v", AllOrgs),
			})
			return
		}
		userState.Activity.Org = orgInput
		userState.Step = 4
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          chatID,
			MessageThreadID: msgThreadID,
			Text:            "Now, please enter the name of the lead.",
		})
	case 4:
		// Collect lead
		userState.Activity.Lead = strings.TrimSpace(update.Message.Text)
		userState.Step = 5
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          chatID,
			MessageThreadID: msgThreadID,
			Text:            "Finally, please enter the name of the co-lead, separated by semicolon(e.g. Person A; Person B)",
		})
	case 5:
		// Collect co-lead
		coleads := strings.Split(update.Message.Text, ";")
		// Remove empty options
		var validColeads []string
		for _, colead := range coleads {
			if opt := strings.TrimSpace(colead); opt != "" {
				validColeads = append(validColeads, opt)
			}
		}
		userState.Activity.CoLeads = validColeads
		userState.Step = -1
	}

	if userState.Step > 0 {
		return
	}

	if _, err := h.activityDAO.Save(&userState.Activity); err != nil {
		log.Println("failed to save activity", userState.Activity, err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          chatID,
			MessageThreadID: msgThreadID,
			Text:            "Failed to save activity! Input anything to save again!",
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          chatID,
		MessageThreadID: msgThreadID,
		Text:            "Activity details collected successfully!\n" + userState.Activity.string(),
		ParseMode:       "HTML",
	})
	// Clean up user state
	delete(userStates, userStateKey)
}

func (h *ActivityHandler) handleDeleteActivitySteps(ctx context.Context, b *bot.Bot, update *models.Update, userStateKey string, userState *UserState) {
	chatID := update.Message.Chat.ID
	msgThreadID := update.Message.MessageThreadID
	activityIDStr := strings.TrimSpace(update.Message.Text)
	activityID, err := strconv.ParseInt(activityIDStr, 10, 64)
	if err != nil {
		log.Println("invalid activity ID", activityIDStr, err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          chatID,
			MessageThreadID: msgThreadID,
			Text:            "Invalid activity ID! Please enter a valid number.",
		})
		return
	}

	affectedRows, err := h.activityDAO.Delete(activityID)
	if err != nil {
		log.Println("failed to delete activity", activityID, err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          chatID,
			MessageThreadID: msgThreadID,
			Text:            "Failed to delete activity! Please try again.",
		})
		return
	}
	if affectedRows == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          chatID,
			MessageThreadID: msgThreadID,
			Text:            "No activity found with the given ID! Please try again.",
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          chatID,
		MessageThreadID: msgThreadID,
		Text:            "Activity deleted successfully!",
	})
	// Clean up user state
	delete(userStates, userStateKey)
}

// sendAllActivities from past 2 months, total 18 months
func (h *ActivityHandler) sendAllActivities(ctx context.Context, b *bot.Bot, chatID int64, msgThreadID int) {
	// Logic to view calendar of activities from past 2 months, total 18 months
	startTime := getCurrentMonthStartInUTC().AddDate(0, -2, 0)
	endTime := startTime.AddDate(0, 18, 0).Add(-time.Nanosecond)
	h.sendActivitiesForPeriod(ctx, b, chatID, msgThreadID, startTime, endTime)
}

func (h *ActivityHandler) sendActivitiesForPeriod(ctx context.Context, b *bot.Bot, chatID int64, msgThreadID int, start, end time.Time) {
	activities, err := h.activityDAO.GetByDuration(start, end)
	if err != nil {
		log.Println("error retrieving activities", start, end, err)
		return
	}
	startMonth := start.Format(monthFormat)
	endMonth := end.Format(monthFormat)
	periodStr := startMonth
	if startMonth != endMonth {
		periodStr += " - " + endMonth
	}

	messageText := fmt.Sprintf("Activities (%s):\n%s", periodStr, getActivitiesMessage(activities))
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          chatID,
		MessageThreadID: msgThreadID,
		Text:            messageText,
		ParseMode:       "HTML",
	})
}

func getActivitiesMessage(activities []Activity) string {
	if len(activities) == 0 {
		return "no activities found."
	}
	str := ""
	var year int
	var month time.Month
	for _, activity := range activities {
		y, m, _ := activity.StartedAt.Date()
		if y != year || m != month {
			year = y
			month = m
			str += fmt.Sprintf("<b><u>%v %d</u></b>\n\n", month, year)
		}
		str += activity.string() + "\n\n"
	}
	return str
}
