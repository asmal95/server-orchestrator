package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	log "github.com/sirupsen/logrus"
)

var Configuration Config

func init() {
	err := cleanenv.ReadConfig("config.yaml", &Configuration)
	if err != nil {
		log.Errorf("Can't load configuration: %v", err)
		panic(err)
	}
}

func reloadConfig() {
	err := cleanenv.UpdateEnv(&Configuration)
	if err != nil {
		log.Errorf("Can't reload configuration: %v", err)
	}
}

type Config struct {
	Bot struct {
		Name  string `yaml:"name" env:"TG_BOT_NAME" env-upd`
		Token string `yaml:"token" env:"TG_BOT_TOKEN" env-required:"true" env-upd`
		Debug bool   `yaml:"debug" env:"TG_BOT_DEBUG" env-default:"false" env-upd`
	} `yaml:"bot"`
	DockerOrchestrator struct {
		ConfigLocation          string `yaml:"config-location" env:"DOCKER_ORCHESTRATOR_CONFIG_LOCATION" env-required:"true"`
		SynchronizationInterval string `yaml:"synchronization-interval" env:"DOCKER_ORCHESTRATOR_SYNCHRONIZATION_INTERVAL" env-default:"30s"`
	} `yaml:"docker-orchestrator"`
	AdminChatId int64 `yaml:"admin-chat-id"`
}

//https://dev.to/ilyakaznacheev/a-clean-way-to-pass-configs-in-a-go-application-1g64
