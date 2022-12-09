package main

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	docker_client "github.com/samthom/system-designing/docker"
)

const TAG = "redis:latest"
const CONTAINER_NAME = "redis-latest-rate-limiter-demo"

func startRedis() {
	// create new docker client from env
	client, err := docker_client.NewDockerClient()
	if err != nil {
		panic(err)
	}
	imageExist := checkDockerImg(client, TAG)
	if !imageExist {
		downloadDockerImage(client, TAG)
	}
    containerExist := checkContainer(client, CONTAINER_NAME)
	if !containerExist {
		// create new container
		fmt.Printf("couldn't find container '%v', creating new container.\n", CONTAINER_NAME)
		containerID, err := createContainer(client)
		if err != nil {
			panic(fmt.Sprintf("container create failed: %v", err))
		}
		fmt.Printf("container '%v' created. container id: %v", CONTAINER_NAME, containerID)
	}
}

func checkContainer(client docker_client.DockerClient, containerName string) bool {
	containers, err := client.ListContainers(context.Background(), types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		panic(err)
	}
    return arrayContains(containers, containerName)
}

func checkDockerImg(client docker_client.DockerClient, tag string) bool {
	// Get all the images inside the host and see if the latest redis image is present or not
	images, err := client.ListImages(context.Background(), types.ImageListOptions{})
	if err != nil {
		panic(err)
	}
	return arrayContains(images, tag)
}

func downloadDockerImage(client docker_client.DockerClient, tag string) bool {
	fmt.Printf("couldn't find image '%v', trying to download.", tag)
	// @TODO Need to add lib method for downloading images with tag
	return true
}

func createContainer(client docker_client.DockerClient) (string, error) {
	// Create new container
	portStr := "6379"
	port, err := nat.NewPort("tcp", portStr)
	if err != nil {
		e := fmt.Errorf("unable to create port: %v", err)
		panic(e)
	}
	portSet := make(nat.PortSet)
	portSet[port] = struct{}{}
	cfg := &container.Config{
		Image:        TAG,
		ExposedPorts: portSet,
	}

	portBinding := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: portStr,
	}
	portMap := make(nat.PortMap)
	portMap[port] = []nat.PortBinding{portBinding}
	hostCfg := &container.HostConfig{
		PortBindings: portMap,
		NetworkMode:  "host",
	}

	containerCfg := &docker_client.ContainerCfg{
		Config: cfg,
		Host:   hostCfg,
	}

	// creating container config
	if err != nil {
		e := fmt.Errorf("unable to create container config: %v", err)
		panic(e)
	}

	// createing the actual container
	// return client.Create.Create(context.Background(), CONTAINER_NAME, &v1.Platform{})
	return client.CreateContainer(context.Background(), CONTAINER_NAME, containerCfg, &v1.Platform{})
}

func arrayContains(items []string, item string) bool {
	for _, it := range items {
		if it == item {
			return true
		}
	}

	return false
}
