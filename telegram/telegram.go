package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"server-orchestrator/config"
)

var Bot *tgbotapi.BotAPI
var handlers = []Handler{GreetingHandler{}, DeploymentsHandler{}, ConfigsHandler{}}

func Start() {
	bot, err := tgbotapi.NewBotAPI(config.Configuration.Bot.Token)
	if err != nil {
		panic(err)
	}
	bot.Debug = config.Configuration.Bot.Debug
	Bot = bot

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates, err := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		for _, handler := range handlers {
			if handler.HandleUpdate(update) {
				break
			}
		}
	}
}

//if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {
