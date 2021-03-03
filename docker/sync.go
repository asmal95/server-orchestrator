package docker

import (
	"fmt"
	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
	"server-orchestrator/config"
	"server-orchestrator/docker/deployment"
	"time"
)

var ticker *time.Ticker

func init() {
	interval, err := time.ParseDuration(config.Configuration.DockerOrchestrator.SynchronizationInterval)
	if err != nil {
		panic(fmt.Sprintf("Can't parse sync interval: %v", err))
	}
	ticker = time.NewTicker(interval)
	go func() {
		for t := range ticker.C {
			log.Debugf("Start synchronization job at %v", t)
			synchronize()
		}
	}()
}

func synchronize() {
	containers, err := GetManagedContainers()
	if err != nil {
		log.Errorf("Can't get docker containers to syncronize it: %v", err)
	}

	dockerContainers := make(map[string]types.Container)
	for _, container := range containers {
		if deploymentName, ok := container.Labels["deployment_name"]; ok {
			dockerContainers[deploymentName] = container
			_, err := deployment.GetDeployment(deploymentName) //check that deployment exist in the service
			if err != nil {
				log.Errorf("Can't find deployment by %v name for %v container", deploymentName, container.ID)
				continue
			}
		}
	}

	deployments := deployment.GetDeployments()
	for _, dep := range deployments {
		var touched = false
		if container, ok := dockerContainers[dep.Name]; ok {
			var status = deployment.NotCreated
			switch container.State {
			case "created":
				status = deployment.Created
			case "restarting", "running":
				status = deployment.Running
			case "paused", "exited":
				status = deployment.Stopped
			case "dead":
				status = deployment.Failed
			}

			if dep.Container.Status != status {
				dep.Container.Status = status
				touched = true
			}
			if dep.Container.ContainerId != container.ID {
				dep.Container.ContainerId = container.ID
				touched = true
			}
		} else {
			if dep.Container.Status != deployment.NotCreated {
				dep.Container.Status = deployment.NotCreated
				touched = true
			}
		}
		if touched {
			log.Infof("%v deploument container has been changed in the docker and force syncronized", dep.Name)
			_, err := deployment.SaveDeployment(dep)
			if err != nil {
				log.Errorf("Can't save deployment during sychronization: %v", err)
			}
		}
	}
}

//states description: https://stackoverflow.com/questions/32427684/what-are-the-possible-states-for-a-docker-container
