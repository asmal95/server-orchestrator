package deployment

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"server-orchestrator/config"
)

var deployments []Deployment
var deploymentConfigLocation string

func init() {
	deploymentConfigLocation = config.Configuration.DockerOrchestrator.ConfigLocation
	deployments = loadAllDeployments()
}

func SaveDeployment(dc Deployment) (Deployment, error) {
	for i, config := range deployments {
		if dc.Name == config.Name {
			deployments[i] = dc
			saveAllDeployments(deployments)
			return dc, nil
		}
	}
	deployments = append(deployments, dc)
	saveAllDeployments(deployments)
	return dc, nil
}

func GetDeployment(name string) (Deployment, error) {
	for _, config := range deployments {
		if name == config.Name {
			return config, nil
		}
	}
	return Deployment{}, fmt.Errorf("can't find dc with %v name", name)
}

func DeleteDeployment(name string) (Deployment, error) {
	index := -1
	for i, config := range deployments {
		if name == config.Name {
			index = i
			break
		}
	}
	if index != -1 {
		dc := deployments[index]
		deployments = append(deployments[:index], deployments[index+1:]...)
		return dc, nil
	}
	return Deployment{}, fmt.Errorf("can't find dc with %v name", name)
}

func GetDeployments() []Deployment {
	return deployments
}

func ClearDeployments() []Deployment {
	copy := deployments
	deployments = make([]Deployment, 0)
	saveAllDeployments(deployments)
	return copy
}

func loadAllDeployments() []Deployment {
	jsonFile, err := os.Open(deploymentConfigLocation)
	if err != nil {
		return make([]Deployment, 0)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var configs = make([]Deployment, 0)

	err = json.Unmarshal(byteValue, &configs)
	if err != nil {
		log.Errorf("Can't open deployment configs file: %v", err)
		return configs //todo ret error
	}

	return configs
}

func saveAllDeployments(configs []Deployment) {
	jsonString, _ := json.MarshalIndent(configs, "", "  ")
	err := ioutil.WriteFile(deploymentConfigLocation, jsonString, os.ModePerm)
	if err != nil {
		log.Errorf("Can't save deployment config: %v", err)
	}
}
