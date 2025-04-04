package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-telegram/bot/models"
)

func getUserFullName(user *models.User) string {
	if user.LastName == "" {
		return user.FirstName
	}
	return user.FirstName + " " + user.LastName
}

func getCommandArgument(update *models.Update) string {
	if update.Message == nil {
		return ""
	}
	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

func getUserStateKey(chatID int64, msgThreadID int, user *models.User) string {
	return fmt.Sprintf("%d:%d:%s", chatID, msgThreadID, user.Username)
}

func getCurrentMonthInUTC() time.Time {
	// for the ease of user interaction, user input time is just stored as UTC while user has the expectation of in local time
	// hence, we would need to convert to UTC while still preserving the local time's month
	now := time.Now().In(AppConfig.Timezone)
	// Convert to UTC while preserving the local time's month
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// add app local timezone to the event but keep the clock unchanged
func addLocalTimezone(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), AppConfig.Timezone)
}
