package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"server-orchestrator/docker/deployment"
	"server-orchestrator/telegram/chat"
	"strings"
)

type ConfigsHandler struct{}

func (c ConfigsHandler) HandleUpdate(update tgbotapi.Update) bool {
	if update.Message != nil {
		return c.handleMessage(*update.Message) || c.handleAction(*update.Message)
	} else if update.CallbackQuery != nil {
		return c.handleCallbackQuery(*update.CallbackQuery)
	}
	return false
}

func (c ConfigsHandler) handleAction(message tgbotapi.Message) bool {
	state := chat.GetState(message.Chat.ID)
	switch state.Action {
	case "waiting_for_config":
		return applyConfig(message)
	case "waiting_for_new_config":
		return applyNewConfig(message)
	}
	return false
}

func (c ConfigsHandler) handleMessage(message tgbotapi.Message) bool {

	switch command(message.Text) {
	case "/new_deployment":
		return newDeployment(message)
	}
	return false
}

func (c ConfigsHandler) handleCallbackQuery(callbackQuery tgbotapi.CallbackQuery) bool {

	s := strings.Split(callbackQuery.Data, ":")
	if len(s) < 2 {
		log.Warnf("Can't parse %v query data")
		return false
	}
	action, target := s[0], s[1]
	switch action {
	case "config":
		return showConfig(target, *callbackQuery.Message, true)
	case "edit_config":
		return editConfig(target, *callbackQuery.Message)
	case "cancel_edit_config":
		//return cancelEditConfig(target, *callbackQuery.Message)
	case "save_edit_config":
		//return saveEditConfig(target, *callbackQuery.Message)
	case "return_deployment":
		return showDeployment(target, *callbackQuery.Message)
	}
	return false
}

func showConfig(deploymentName string, message tgbotapi.Message, editMessage bool) bool {
	dep, err := deployment.GetDeployment(deploymentName)
	if err != nil {
		log.Errorf("Can't get deployment: %v", err)
	}

	depMessConf := ConvertToDeploymentMessageConfig(dep)
	jsonString, _ := yaml.Marshal(depMessConf)

	editAct := fmt.Sprintf("edit_config:%v", deploymentName)
	returnAct := fmt.Sprintf("return_deployment:%v", deploymentName)

	markup := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				{
					Text:         "Edit",
					CallbackData: &editAct,
				},
				{
					Text:         "Return to deployment",
					CallbackData: &returnAct,
				},
			},
		},
	}

	var resultMessage tgbotapi.Chattable
	if editMessage {
		editedMessage := tgbotapi.NewEditMessageText(message.Chat.ID, message.MessageID, string(jsonString))
		editedMessage.ReplyMarkup = &markup
		resultMessage = editedMessage
	} else {
		newMessage := tgbotapi.NewMessage(message.Chat.ID, string(jsonString))
		newMessage.ReplyMarkup = markup
		resultMessage = newMessage
	}

	_, err = Bot.Send(resultMessage)
	if err != nil {
		log.Errorf("Can't send message: %v", err)
		return false
	}

	return true
}

type DeploymentMessageConfig struct {
	Name             string            `yaml:"name"`
	DockerRepository string            `yaml:"docker_repository"`
	DockerTag        string            `yaml:"docker_tag"`
	Environment      map[string]string `yaml:"environment"`
	Entrypoint       []string          `yaml:"entrypoint"`
	PortBinding      map[string]string `yaml:"port_binding"` //host : container
}

func ConvertToDeploymentMessageConfig(dep deployment.Deployment) DeploymentMessageConfig {
	res := DeploymentMessageConfig{
		Name:             dep.Name,
		DockerRepository: dep.DockerRepository,
		DockerTag:        dep.DockerTag,
		Environment:      maskSensitiveParameters(dep.Environment),
		Entrypoint:       dep.Entrypoint,
		PortBinding:      dep.PortBinding,
	}
	return res
}

func editConfig(target string, message tgbotapi.Message) bool {
	edited := tgbotapi.NewEditMessageText(message.Chat.ID, message.MessageID, message.Text)
	_, err := Bot.Send(edited)
	if err != nil {
		log.Errorf("Can't edit message: %v", err)
	}

	returnAct := fmt.Sprintf("return_deployment:%v", target)

	newMessage := tgbotapi.NewMessage(message.Chat.ID, "Send `yaml` config to apply")
	newMessage.ParseMode = tgbotapi.ModeMarkdown
	markup := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				{
					Text:         "Cancel",
					CallbackData: &returnAct,
				},
			},
		},
	}
	newMessage.ReplyMarkup = &markup
	_, err = Bot.Send(newMessage)
	if err != nil {
		log.Errorf("Can't send message: %v", err)
		return false
	}

	chat.SetMeta(message.Chat.ID, "deployment", target)
	chat.SetAction(message.Chat.ID, "waiting_for_config")
	return true
}

func applyConfig(message tgbotapi.Message) bool {
	deploymentName := chat.GetState(message.Chat.ID).Meta["deployment"]

	conf := DeploymentMessageConfig{}

	err := yaml.Unmarshal([]byte(message.Text), &conf)
	if err != nil {
		log.Errorf("Can't parse configuration: %v", err)
		return false
	}
	dep, err := deployment.GetDeployment(deploymentName)
	if err != nil {
		log.Errorf("Can't find deployment: %v", err)
		return false
	}
	if dep.Name != conf.Name && conf.Name != "" {
		_, _ = Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "You can't change deployment name"))
		return false
	}
	dep.DockerRepository = conf.DockerRepository
	dep.DockerTag = conf.DockerTag
	dep.PortBinding = conf.PortBinding
	dep.Environment = unmaskSensitiveParameters(conf.Environment, dep.Environment)
	dep.Entrypoint = conf.Entrypoint

	_, err = deployment.SaveDeployment(dep)
	if err != nil {
		log.Errorf("Can't save deploymnt: %v", err)
		return false
	}
	chat.SetAction(message.Chat.ID, "none")
	return showConfig(deploymentName, message, false)
}

func newDeployment(message tgbotapi.Message) bool {
	edited := tgbotapi.NewEditMessageText(message.Chat.ID, message.MessageID, message.Text)
	_, err := Bot.Send(edited)
	if err != nil {
		log.Errorf("Can't edit message: %v", err)
	}

	returnAct := fmt.Sprintf("return:deployments")

	newMessage := tgbotapi.NewMessage(message.Chat.ID, "Send `yaml` config to create. Example:\n"+
		"```\n"+
		"name: deployment-name\n"+
		"docker_repository: <username>/<image_name>\n"+
		"docker_tag: <tag>\n"+
		"environment:\n"+
		"  ENV_VAR_1: value_1\n"+
		"  ENV_VAR_2: value_2\n"+
		"port_binding:\n"+
		"  \"<host_port>\": \"<container_port>\"\n"+
		"```")
	newMessage.ParseMode = tgbotapi.ModeMarkdown
	markup := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				{
					Text:         "Cancel",
					CallbackData: &returnAct,
				},
			},
		},
	}
	newMessage.ReplyMarkup = &markup
	_, err = Bot.Send(newMessage)
	if err != nil {
		log.Errorf("Can't send message: %v", err)
		return false
	}

	chat.SetAction(message.Chat.ID, "waiting_for_new_config")
	return true
}

func applyNewConfig(message tgbotapi.Message) bool {

	conf := DeploymentMessageConfig{}

	err := yaml.Unmarshal([]byte(message.Text), &conf)
	if err != nil {
		log.Errorf("Can't parse configuration: %v", err)
		return false
	}
	_, err = deployment.GetDeployment(conf.Name)
	if err == nil {
		_, _ = Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Deployment with the name already exist"))
		return false
	}

	dep := deployment.Deployment{
		Name:             conf.Name,
		DockerRepository: conf.DockerRepository,
		DockerTag:        conf.DockerTag,
		Environment:      conf.Environment,
		Entrypoint:       conf.Entrypoint,
		PortBinding:      conf.PortBinding,
	}

	dep, err = deployment.SaveDeployment(dep)
	if err != nil {
		log.Errorf("Can't save deploymnt: %v", err)
		return false
	}
	chat.SetAction(message.Chat.ID, "none")
	return showConfig(dep.Name, message, false)
}

func maskSensitiveParameters(env map[string]string) map[string]string {
	result := make(map[string]string)
	for key, value := range env {
		if strings.Contains(strings.ToLower(key), "password") ||
			strings.Contains(strings.ToLower(key), "token") ||
			strings.Contains(strings.ToLower(key), "secret") {

			result[key] = "[HIDDEN]"
		} else {
			result[key] = value
		}
	}
	return result
}

func unmaskSensitiveParameters(masked map[string]string, original map[string]string) map[string]string {
	result := make(map[string]string)
	for key, value := range masked {
		if value == "[HIDDEN]" {
			if originalValue, ok := original[key]; ok {
				result[key] = originalValue
			} else {
				log.Warnf("Can't find original value for the masked env variable: %v", key)
			}
		} else {
			result[key] = value
		}
	}
	return result
}
