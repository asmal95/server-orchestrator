package docker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/docker/distribution/context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"server-orchestrator/docker/deployment"
)

var Client client.APIClient

func init() {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("Unable to create docker client: %v", err)
		panic(err)
	}
	Client = cli
}

func CreateContainer(dc deployment.Deployment) (deployment.Deployment, error) {

	portBinding := make(nat.PortMap)
	for hostPortValue, containerPortValue := range dc.PortBinding {
		hostBinding := nat.PortBinding{
			HostIP:   "0.0.0.0",
			HostPort: hostPortValue,
		}
		containerPort, err := nat.NewPort("tcp", containerPortValue)
		if err != nil {
			panic("Unable to get the port")
		}
		portBinding[containerPort] = []nat.PortBinding{hostBinding}
	}

	cont, err := Client.ContainerCreate(
		context.Background(),
		&container.Config{
			Image: dc.GetImage(),
			Env:   dc.BuildEnvVariables(),
			Tty:   true,
		},
		&container.HostConfig{
			PortBindings: portBinding,
		}, nil, dc.Name)
	if err != nil {
		dc.Container.Status = deployment.Failed
		log.Errorf("Can't create container: %v", err)
		return dc, err
	}

	dc.Container.ContainerId = cont.ID
	dc.Container.Warnings = cont.Warnings
	dc.Container.Status = deployment.Created

	return dc, nil
}

func StartContainer(dc deployment.Deployment) (deployment.Deployment, error) {
	op := types.ContainerStartOptions{}

	err := Client.ContainerStart(context.Background(), dc.Container.ContainerId, op)
	if err != nil {
		dc.Container.Status = deployment.Failed
		log.Errorf("Can't start container: %v", err)
		return dc, err
	}
	dc.Container.Status = deployment.Running
	log.Infof("Container %s is started", dc.Container.ContainerId)

	return dc, nil
}

func ListContainer() error {

	containers, err := Client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	if len(containers) > 0 {
		for _, container := range containers {
			fmt.Printf("Container ID: %s", container.ID)
		}
	} else {
		fmt.Println("There are no containers running")
	}
	return nil
}

func RemoveContainer(dep deployment.Deployment) (deployment.Deployment, error) {
	err := Client.ContainerRemove(context.Background(), dep.Container.ContainerId, types.ContainerRemoveOptions{
		RemoveVolumes: false,
		RemoveLinks:   false,
		Force:         true,
	})
	if err != nil {
		return dep, err
	}
	dep.Container.Status = deployment.NotCreated
	return dep, err
}

func StopContainer(dep deployment.Deployment) (deployment.Deployment, error) {

	err := Client.ContainerStop(context.Background(), dep.Container.ContainerId, nil)
	if err != nil {
		return dep, err
	}
	dep.Container.Status = deployment.Stopped
	return dep, err
}

func GetLogs(dc deployment.Deployment) (string, error) {
	ctx := context.Background()

	options := types.ContainerLogsOptions{ShowStdout: true}
	out, err := Client.ContainerLogs(ctx, dc.Container.ContainerId, options)
	if err != nil {
		log.Errorf("Can't obtain logs: %v", err)
		return "", err
	}
	file, err := ioutil.TempFile("", dc.Name)
	if err != nil {
		log.Errorf("Can't create temp file for logs: %v", err)
		return "", err
	}

	defer file.Close()
	defer out.Close()
	_, err = io.Copy(file, out)

	return file.Name(), nil
}

func PullImage(dep deployment.Deployment) error {
	ctx := context.Background()

	out, err := Client.ImagePull(ctx, "docker.io/"+dep.GetImage(), types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()
	io.Copy(os.Stdout, out)
	return nil
}

func list() {

	containers, err := Client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Println(container.ID)
	}
}

func pullAuth() {
	ctx := context.Background()

	authConfig := types.AuthConfig{
		Username: "username",
		Password: "password",
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		panic(err)
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	out, err := Client.ImagePull(ctx, "docker.io/library/alpine", types.ImagePullOptions{RegistryAuth: authStr})
	if err != nil {
		panic(err)
	}

	defer out.Close()
	io.Copy(os.Stdout, out)
}

func stopAll() {
	ctx := context.Background()

	containers, err := Client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, c := range containers {
		fmt.Print("Stopping container ", c.ID[:10], "... ")
		if err := Client.ContainerStop(ctx, c.ID, nil); err != nil {
			panic(err)
		}
		fmt.Println("Success")
	}
}

// https://medium.com/tarkalabs/controlling-the-docker-engine-in-go-826012f9671c
// https://habr.com/ru/post/449038/
