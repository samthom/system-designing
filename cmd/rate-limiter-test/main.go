package main

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/docker/docker/api/types"
	docker_client "github.com/samthom/system-designing/docker"
)

const TAG = "redis:latest"
const CONTAINER_NAME = "redis-latest-rate-limiter-demo"

func main() {

	// create new docker client from env
	client, err := docker_client.NewDockerClient()
	if err != nil {
		panic(err)
	}

	// Get all the images inside the host and see if the latest redis image is present or not
	imageExist := checkImgExist(context.Background(), TAG, client)
	if !imageExist {
		// @TODO push the image and proceed
		panic("redis:latest images in not found inside docker. Pull image using 'docker pull redis:latest'")
	}

    if imageExist {
        fmt.Println("image ", TAG, " found, creating container ", CONTAINER_NAME)
    }

	// Create new container
    portStr := "6379"
    port, err := nat.NewPort("tcp", portStr)
    if err != nil {
        e := fmt.Errorf("unable to create port: %v", err)
        panic(e)
    }
    portSet := make(nat.PortSet)
    portSet[port] = struct{}{}
	containerCfg := &container.Config{
        Image: TAG,
        ExposedPorts: portSet,
    }

    portBinding := nat.PortBinding{
        HostIP: "0.0.0.0",
        HostPort: portStr,
    }
    portMap := make(nat.PortMap)
    portMap[port] = []nat.PortBinding{portBinding}
    hostCfg := &container.HostConfig{
        PortBindings: portMap,
        NetworkMode: "host",
    }

    // creating container config
	container, err := docker_client.NewContainer(client, docker_client.WithContainerConfig(containerCfg), docker_client.WithHostConfig(hostCfg))
    if err != nil {
        e := fmt.Errorf("unable to create container config: %v", err)
        panic(e)
    }

    // createing the actual container
    containerID, err := container.Create(context.Background(), CONTAINER_NAME, &v1.Platform{})
    if err != nil {
        e := fmt.Errorf("container create failed: %v", err)
        panic(e)
    }

    fmt.Println("container created, ID: ", containerID)
}

func checkImgExist(ctx context.Context, image string, cli *docker_client.DockerClient) bool {
	images, err := cli.ListImages(ctx, types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	for _, image := range images {
		if image == TAG {
			return true
		}
	}
    return false
}
