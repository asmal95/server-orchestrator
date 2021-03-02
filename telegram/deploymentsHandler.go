package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"server-orchestrator/docker"
	"server-orchestrator/docker/deployment"
	"strings"
	"time"
)

type DeploymentsHandler struct{}

func (d DeploymentsHandler) HandleUpdate(update tgbotapi.Update) bool {
	if update.Message != nil {
		return d.handleMessage(*update.Message)
	} else if update.CallbackQuery != nil {
		return d.handleCallbackQuery(*update.CallbackQuery)
	}
	return false
}

func (d DeploymentsHandler) handleMessage(message tgbotapi.Message) bool {
	switch command(message.Text) {
	case "/deployments":
		return showDeployments(message, false)
	}
	return false
}

func (d DeploymentsHandler) handleCallbackQuery(callbackQuery tgbotapi.CallbackQuery) bool {

	s := strings.Split(callbackQuery.Data, ":")
	action, target := s[0], s[1]
	switch action {
	case "deployment":
		return showDeployment(target, *callbackQuery.Message)
	case "create":
		return createDeployment(target, *callbackQuery.Message)
	case "start":
		return startDeployment(target, *callbackQuery.Message)
	case "stop":
		return stopDeployment(target, *callbackQuery.Message)
	case "delete":
		return deleteDeployment(target, *callbackQuery.Message)
	case "logs":
		getLogs(target, *callbackQuery.Message)
	case "pull":
		pull(target, *callbackQuery.Message)
	case "return":
		showDeployments(*callbackQuery.Message, true)
	case "return_new":
		showDeployments(*callbackQuery.Message, false)
	}
	return false
}

func createDeployment(deploymentName string, message tgbotapi.Message) bool {
	dep, err := deployment.GetDeployment(deploymentName)
	if err != nil {
		log.Errorf("Can't load deployment: %v", err)
		return false
	}
	dep, err = docker.CreateContainer(dep)
	deployment.SaveDeployment(dep)
	if err != nil {
		log.Errorf("Can't create deployemnt: %v", err)
		return false
	}
	return showDeployment(deploymentName, message)
}

func startDeployment(deploymentName string, message tgbotapi.Message) bool {
	dep, err := deployment.GetDeployment(deploymentName)
	if err != nil {
		log.Errorf("Can't load deployment: %v", err)
		return false
	}
	dep, err = docker.StartContainer(dep)
	deployment.SaveDeployment(dep)
	if err != nil {
		log.Errorf("Can't start deployemnt: %v", err)
		return false
	}
	return showDeployment(deploymentName, message)
}

func stopDeployment(deploymentName string, message tgbotapi.Message) bool {
	dep, err := deployment.GetDeployment(deploymentName)
	if err != nil {
		log.Errorf("Can't load deployment: %v", err)
		return false
	}
	dep, err = docker.StopContainer(dep)
	deployment.SaveDeployment(dep)
	if err != nil {
		log.Errorf("Can't stop deployemnt: %v", err)
		return false
	}
	return showDeployment(deploymentName, message)
}

func deleteDeployment(deploymentName string, message tgbotapi.Message) bool {
	dep, err := deployment.GetDeployment(deploymentName)
	if err != nil {
		log.Errorf("Can't load deployment: %v", err)
		return false
	}
	dep, err = docker.RemoveContainer(dep)
	deployment.SaveDeployment(dep)
	if err != nil {
		log.Errorf("Can't delete deployemnt: %v", err)
		return false
	}
	return showDeployment(deploymentName, message)
}

func showDeployment(target string, message tgbotapi.Message) bool {
	deployments := deployment.GetDeployments()
	for _, dep := range deployments {
		if target == dep.Name {

			text := fmt.Sprintf("Deployment: %v\n"+
				" image: %v\n"+
				" status: %v\n", dep.Name, dep.GetImage(), dep.Container.Status)

			edited := tgbotapi.NewEditMessageText(message.Chat.ID, message.MessageID, text)
			edited.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{
				InlineKeyboard: buildDeploymentControls(dep),
			}
			_, err := Bot.Send(edited)
			if err != nil {
				log.Errorf("Can't send message: %v", err)
			}
			return true
		}
	}
	return false
}

func buildDeploymentControls(dep deployment.Deployment) [][]tgbotapi.InlineKeyboardButton {
	createAct := fmt.Sprintf("create:%v", dep.Name)
	startAct := fmt.Sprintf("start:%v", dep.Name)
	stopAct := fmt.Sprintf("stop:%v", dep.Name)
	deleteAct := fmt.Sprintf("delete:%v", dep.Name)
	configAct := fmt.Sprintf("config:%v", dep.Name)
	logsAct := fmt.Sprintf("logs:%v", dep.Name)
	returnAct := fmt.Sprintf("return:deployments")
	pullAct := fmt.Sprintf("pull:%v", dep.Name)

	actions := make([][]tgbotapi.InlineKeyboardButton, 0)
	line := make([]tgbotapi.InlineKeyboardButton, 0)
	switch dep.Container.Status {
	case deployment.Created:
		line = append(line, tgbotapi.InlineKeyboardButton{Text: "Start Container", CallbackData: &startAct})
		line = append(line, tgbotapi.InlineKeyboardButton{Text: "Delete Container", CallbackData: &deleteAct})
		line = append(line, tgbotapi.InlineKeyboardButton{Text: "Config", CallbackData: &configAct})
	case deployment.Running:
		line = append(line, tgbotapi.InlineKeyboardButton{Text: "Stop Container", CallbackData: &stopAct})
		line = append(line, tgbotapi.InlineKeyboardButton{Text: "Config", CallbackData: &configAct})
		line = append(line, tgbotapi.InlineKeyboardButton{Text: "Logs", CallbackData: &logsAct})
	case deployment.Failed:
		line = append(line, tgbotapi.InlineKeyboardButton{Text: "Start Container", CallbackData: &startAct})
		line = append(line, tgbotapi.InlineKeyboardButton{Text: "Delete Container", CallbackData: &deleteAct})
		line = append(line, tgbotapi.InlineKeyboardButton{Text: "Config", CallbackData: &configAct})
		line = append(line, tgbotapi.InlineKeyboardButton{Text: "Logs", CallbackData: &logsAct})
	case deployment.Stopped:
		line = append(line, tgbotapi.InlineKeyboardButton{Text: "Start Container", CallbackData: &startAct})
		line = append(line, tgbotapi.InlineKeyboardButton{Text: "Delete Container", CallbackData: &deleteAct})
		line = append(line, tgbotapi.InlineKeyboardButton{Text: "Logs", CallbackData: &logsAct})
		line = append(line, tgbotapi.InlineKeyboardButton{Text: "Config", CallbackData: &configAct})
	default:
		line = append(line, tgbotapi.InlineKeyboardButton{Text: "Create Container", CallbackData: &createAct})
		line = append(line, tgbotapi.InlineKeyboardButton{Text: "Config", CallbackData: &configAct})
	}
	actions = append(actions, line)
	actions = append(actions, []tgbotapi.InlineKeyboardButton{
		{
			Text:         "Return to deployments list",
			CallbackData: &returnAct,
		},
		{
			Text:         "Pull",
			CallbackData: &pullAct,
		},
	})
	return actions
}

// Controls
// NOT CREATED: Create Config Return
// CREATED: Start Delete Config Return
// RUNNING: Stop Config Logs Return
// FAILED: Start Delete Config Logs Return
// STOPPED: Start Delete Config Logs Return

func showDeployments(message tgbotapi.Message, editMessage bool) bool {
	inlineResponse := make([][]tgbotapi.InlineKeyboardButton, 0)
	line := make([]tgbotapi.InlineKeyboardButton, 0)

	deployments := deployment.GetDeployments()
	for i, dep := range deployments {
		callback := fmt.Sprintf("deployment:%v", dep.Name)
		button := tgbotapi.InlineKeyboardButton{
			Text:         dep.Name,
			CallbackData: &callback,
		}
		line = append(line, button)

		if (i+1)/2 != 0 {
			inlineResponse = append(inlineResponse, line)
			line = make([]tgbotapi.InlineKeyboardButton, 0)
		}
	}
	if len(line) != 0 {
		inlineResponse = append(inlineResponse, line)
	}
	markup := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: inlineResponse,
	}

	var resultMessage tgbotapi.Chattable
	if editMessage {
		editedMessage := tgbotapi.NewEditMessageText(message.Chat.ID, message.MessageID, "Current deployments:")
		editedMessage.ReplyMarkup = &markup
		resultMessage = editedMessage
	} else {
		newMessage := tgbotapi.NewMessage(message.Chat.ID, "Current deployments:")
		newMessage.ReplyMarkup = markup
		resultMessage = newMessage
	}

	_, err := Bot.Send(resultMessage)
	if err != nil {
		log.Errorf("Can't send message to the bot: %v", err)
		return false
	}
	return true
}

func getLogs(target string, message tgbotapi.Message) bool {
	dep, err := deployment.GetDeployment(target)
	if err != nil {
		log.Errorf("Can't load deployment: %v", err)
		return false
	}
	fileName, err := docker.GetLogs(dep)
	if err != nil {
		return false
	}
	file, _ := os.Open(fileName)
	data, _ := ioutil.ReadFile(fileName)
	defer file.Close()

	fileBytes := tgbotapi.FileBytes{
		Name:  dep.Name + time.Now().Format("020106_150405") + ".txt",
		Bytes: data,
	}

	doc := tgbotapi.NewDocumentUpload(message.Chat.ID, fileBytes)
	doc.MimeType = "text/plain"
	_, err = Bot.Send(doc)

	if err != nil {
		log.Errorf("Can't send logs: %v", err)
	}

	return true
}

func pull(target string, message tgbotapi.Message) bool {
	dep, err := deployment.GetDeployment(target)
	if err != nil {
		log.Errorf("Can't load deployment: %v", err)
		return false
	}
	err = docker.PullImage(dep)
	if err != nil {
		_, _ = Bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Can't pull docker image: %v", err)))
		return false
	}
	_, _ = Bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("%v image pull completed!", dep.GetImage())))
	return true
}
