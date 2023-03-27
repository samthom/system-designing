package main

import (
	"context"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	docker_client "github.com/samthom/system-designing/docker"
)

const TAG = "redis:latest"
const CONTAINER_NAME = "redis-latest-rate-limiter-demo"

func startRedis(ctx context.Context, client docker_client.DockerClient) (bool, string) {
	// create new docker client from env

    setupImage(ctx, client, TAG)
	containerID := setupContainer(ctx, client, CONTAINER_NAME)

	// start the container using containerID
    err := client.StartContainer(context.Background(), containerID)
	if err != nil {
		log.Panicf("StartContainer() failed: %v", err)
		return false, ""
	}
	return true, containerID
}

func stopRedis(ctx context.Context, client docker_client.DockerClient, containerID string) bool {
    err := client.StopContainer(ctx, containerID)
    if err != nil {
        log.Panicf("StopContainer(%q) failed: %v\nunable stop container", containerID, err)
        return false
    }
    return true
}

func setupImage(ctx context.Context, client docker_client.DockerClient, tag string) {
	imageExist := checkDockerImg(ctx, client, TAG)
	if !imageExist {
		downloadDockerImage(client, TAG)
	}
}

func setupContainer(ctx context.Context, client docker_client.DockerClient, containerName string) string {
	containerID, containerExist := checkContainer(ctx, client, CONTAINER_NAME)
	if !containerExist {
		// create new container
		log.Printf("couldn't find container %q \ncreating new container.\n", CONTAINER_NAME)
		containerID, err := createContainer(ctx, client)
		if err != nil {
			log.Panicf("container create failed: %v", err)
			return ""
		}
		log.Printf("conatiner %q created.\ncontainer id: %q", CONTAINER_NAME, containerID)
		return containerID
	}

	return containerID
}

func checkContainer(ctx context.Context, client docker_client.DockerClient, containerName string) (string, bool) {
	containers, err := client.ListContainers(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		log.Panicf("ListContainers() failed: %v", err)
	}
	id, ok := containers[containerName]
	if ok {
		return id, true
	} else {
		return "", false
	}
}

func checkDockerImg(ctx context.Context, client docker_client.DockerClient, tag string) bool {
	// Get all the images inside the host and see if the latest redis image is present or not
	images, err := client.ListImages(ctx, types.ImageListOptions{})
	if err != nil {
		log.Panicf("ListImages() failed: %v", err)
	}
	return arrayContains(images, tag)
}

func downloadDockerImage(client docker_client.DockerClient, tag string) bool {
	log.Printf("couldn't find image %q, trying to download.", tag)
	// @TODO Need to add lib method for downloading images with tag
	return true
}

func createContainer(ctx context.Context, client docker_client.DockerClient) (string, error) {
	// Create new container
	portStr := "6379"
	port, err := nat.NewPort("tcp", portStr)
	if err != nil {
		log.Panicf("unable to create port: %v", err)
		return "", err
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
		log.Panicf("Unable to create container config: %v", err)
		return "", err
	}

	// createing the actual container
	// return client.Create.Create(context.Background(), CONTAINER_NAME, &v1.Platform{})
	return client.CreateContainer(ctx, CONTAINER_NAME, containerCfg, &v1.Platform{})
}

func arrayContains(items []string, item string) bool {
	for _, it := range items {
		if it == item {
			return true
		}
	}

	return false
}
