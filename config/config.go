package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	log "github.com/sirupsen/logrus"
)

var Configuration Config

func init() {
	err := cleanenv.ReadConfig("config.yml", &Configuration)
	if err != nil {
		log.Error("Can't load configuration: %v", err)
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
		Debug bool   `yaml:"debug" env:"TG_BOT_DEBUG" env-upd`
	} `yaml:"bot"`
	DockerOrchestrator struct {
		ConfigLocation string `yaml:"config-location env: "DOCKER_ORCHESTRATOR_CONFIG_LOCATION" env-required:"true"`
	} `yaml:"docker-orchestrator"`
}

//https://dev.to/ilyakaznacheev/a-clean-way-to-pass-configs-in-a-go-application-1g64
