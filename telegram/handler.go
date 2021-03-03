package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"server-orchestrator/config"
	"strings"
)

type Handler interface {
	HandleUpdate(update tgbotapi.Update) bool
}

var (
	status = "status"
	logs   = "logs"
)

type GreetingHandler struct{}

func (g GreetingHandler) HandleUpdate(update tgbotapi.Update) bool {

	if update.Message != nil {
		return g.handleMessage(*update.Message)
	}
	return false
}

func (g GreetingHandler) handleMessage(message tgbotapi.Message) bool {
	switch command(message.Text) {
	case "/start":
		msg := tgbotapi.NewMessage(message.Chat.ID, "Hi! I can help you to deploy your service in the docker right here, in the telegram chat!\n"+
			"To create a new deployment please use /new_deployment command. \n"+
			"To list existing deployments please use /deployments command.")
		_, err := Bot.Send(msg)
		if err != nil {
			log.Errorf("Can't send message to the chat: %v", err)
			return false
		}
		return true
	}
	return false
}

// Removes bot suffix from end of command. Used for more simple handling of the commands.
func command(input string) string {
	suffix := fmt.Sprintf("@%v", config.Configuration.Bot.Name)
	return strings.TrimSuffix(input, suffix)
}
