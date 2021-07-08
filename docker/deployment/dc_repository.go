package deployment

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"server-orchestrator/config"
	"sync"
)

var deployments []Deployment
var deploymentConfigLocation string
var mutex sync.Mutex

func init() {
	deploymentConfigLocation = config.Configuration.DockerOrchestrator.ConfigLocation
	var err error
	deployments, err = loadData()
	if err != nil {
		log.Fatalf("Can't load data from '%v' file: %v", deploymentConfigLocation, err)
		panic(err)
	}
}

func SaveDeployment(dc Deployment) (Deployment, error) {
	mutex.Lock()
	defer mutex.Unlock()

	for i, dep := range deployments {
		if dc.Name == dep.Name {
			deployments[i] = dc
			return dc, flushData(deployments)
		}
	}
	deployments = append(deployments, dc)
	return dc, flushData(deployments)
}

func FindByQuery(query func(d Deployment) bool) []Deployment {
	res := make([]Deployment, 0)
	for _, dep := range deployments {
		if query(dep) {
			res = append(res, dep)
		}
	}
	return res
}

func FindFirstByQuery(query func(d Deployment) bool) (Deployment, error) {
	res := FindByQuery(query)
	if len(res) == 0 {
		return Deployment{}, fmt.Errorf("can't find dc by query")
	}
	return res[0], nil
}

func GetDeployment(name string) (Deployment, error) {
	for _, dep := range deployments {
		if name == dep.Name {
			return dep, nil
		}
	}
	return Deployment{}, fmt.Errorf("can't find dc with %v name", name)
}

func DeleteDeployment(name string) (Deployment, error) {
	mutex.Lock()
	defer mutex.Unlock()

	index := -1
	for i, dep := range deployments {
		if name == dep.Name {
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

func ClearDeployments() ([]Deployment, error) {
	mutex.Lock()
	defer mutex.Unlock()

	copyData := deployments
	deployments = make([]Deployment, 0)
	return copyData, flushData(deployments)
}

func loadData() ([]Deployment, error) {
	if _, err := os.Stat(deploymentConfigLocation); err != nil && os.IsNotExist(err) {
		return make([]Deployment, 0), nil // file doesn't exist - return empty data
	} else if err != nil {
		return nil, err //other trouble with the file
	}
	jsonFile, err := os.Open(deploymentConfigLocation)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close() //todo need to find better way to handle it
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	var configs = make([]Deployment, 0)
	err = json.Unmarshal(byteValue, &configs)
	if err != nil {
		log.Errorf("Can't open deployment configs file: %v", err)
		return nil, err
	}
	return configs, nil
}

func flushData(configs []Deployment) error {
	jsonString, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(deploymentConfigLocation, jsonString, os.ModePerm)
}
