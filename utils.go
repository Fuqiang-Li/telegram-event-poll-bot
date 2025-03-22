package main

import (
	"fmt"
	"strings"

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
