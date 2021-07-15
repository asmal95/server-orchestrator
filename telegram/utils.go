package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"server-orchestrator/config"
)

func isAdminChat(update tgbotapi.Update) bool {
	if update.Message != nil {
		return config.Configuration.AdminChatId == update.Message.Chat.ID ||
			config.Configuration.SuperUserId == update.Message.Chat.ID
	} else if update.CallbackQuery != nil {
		return config.Configuration.AdminChatId == update.CallbackQuery.Message.Chat.ID ||
			config.Configuration.SuperUserId == update.CallbackQuery.Message.Chat.ID
	}
	return false
}

func isSuperUserChat(update tgbotapi.Update) bool {
	if update.Message != nil {
		return config.Configuration.SuperUserId == update.Message.Chat.ID
	} else if update.CallbackQuery != nil {
		return config.Configuration.SuperUserId == update.CallbackQuery.Message.Chat.ID
	}
	return false
}
