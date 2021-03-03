package deployment

import (
	"fmt"
	"time"
)

type ContainerStatus string

const (
	NotCreated ContainerStatus = "NOT_CREATED"
	Created    ContainerStatus = "CREATED"
	Running    ContainerStatus = "RUNNING"
	Failed     ContainerStatus = "FAILED"
	Stopped    ContainerStatus = "STOPPED"
)

type Deployment struct {
	Name             string            `json:"name"`
	DockerRepository string            `json:"docker_repository"`
	DockerTag        string            `json:"docker_tag"`
	Environment      map[string]string `json:"environment"`
	PortBinding      map[string]string `json:"port_binding"` //host : container
	Entrypoint       []string          `json:"entrypoint"`   //host : container
	Container        struct {
		ContainerId string          `json:"container_id"`
		Status      ContainerStatus `json:"status"`
		Warnings    []string        `json:"warnings"`
	} `json:"container"`
}

func (d Deployment) GetImage() string {
	return fmt.Sprintf("%v:%v", d.DockerRepository, d.DockerTag)
}

func (dc Deployment) BuildEnvVariables() []string {
	res := make([]string, 0)
	for key, value := range dc.Environment {
		res = append(res, key+"="+value)
	}
	return res
}

type History struct {
	DeploymentName string
	Date           time.Time
	Data           map[string]string
}
