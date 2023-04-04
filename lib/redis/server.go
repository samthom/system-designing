package redis_container

import (
	"context"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	docker_client "github.com/samthom/system-designing/pkg/docker"
)


type RedisContainer interface {
	StartRedis(ctx context.Context) (containerID string, err error)
	StopRedis(ctx context.Context) (ok bool)
}

type redisContainer struct {
	ctx          context.Context
	dockerClient docker_client.DockerClient
	containerID  string
}

func GetRedis(ctx context.Context, dockerClient docker_client.DockerClient) RedisContainer {
	return &redisContainer{ctx, dockerClient, ""}
}

func (redisClient *redisContainer) StartRedis(ctx context.Context) (string, error) {
    _, err := setupImage(ctx, redisClient.dockerClient, TAG)
    if err != nil {
        return "", err
    }
	containerID := setupContainer(ctx, redisClient.dockerClient, CONTAINER_NAME)
    redisClient.containerID = containerID

	err = redisClient.dockerClient.StartContainer(context.Background(), containerID)
	if err != nil {
		log.Panicf("StartContainer() failed: %v", err)
		return "", err
	}
	return containerID, nil

}

func (redisClient *redisContainer) StopRedis(ctx context.Context) (ok bool) {
	err := redisClient.dockerClient.StopContainer(ctx, redisClient.containerID)
	if err != nil {
		log.Panicf("StopContainer(%q) failed: %v\nunable stop container", redisClient.containerID, err)
		return false
	}
	return true
}

func setupImage(ctx context.Context, client docker_client.DockerClient, tag string) (bool, error) {
	imageExist, _ := checkRedisImg(ctx, client)
	if !imageExist {
        err := downloadDockerImage(ctx, client, TAG)
        if err != nil {
            return false, err
        }
        return true, nil
	} else {
		return imageExist, nil
	}
}

func setupContainer(ctx context.Context, client docker_client.DockerClient, containerName string) string {
	containerID, containerExist := checkContainer(ctx, client, CONTAINER_NAME)
	if !containerExist {
		// create new container
		log.Printf("couldn't find container %q. creating new container.\n", CONTAINER_NAME)
		containerID, err := createContainer(ctx, client)
		if err != nil {
			log.Panicf("container create failed: %v", err)
			return ""
		}
		log.Printf("conatiner %q created. container id: %q", CONTAINER_NAME, containerID)
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

func checkRedisImg(ctx context.Context, client docker_client.DockerClient) (bool, error) {
	images, err := client.ListImagesTags(ctx, types.ImageListOptions{})
	if err != nil {
		log.Panicf("ListImages() failed: %v", err)
		return false, err
	}
	return arrayContains(images, TAG), nil
}

func downloadDockerImage(ctx context.Context, client docker_client.DockerClient, tag string) error {
	log.Printf("couldn't find image %q, trying to download.", tag)
	return client.PullImage(ctx, tag, types.ImagePullOptions{}, false)
}

func createContainer(ctx context.Context, client docker_client.DockerClient) (string, error) {
    containerCfg, err := createRedisContainerConfig()
    if err != nil {
        return "", err
    }
	return client.CreateContainer(ctx, CONTAINER_NAME, containerCfg, &v1.Platform{})
}

func createRedisContainerConfig() (*docker_client.ContainerCfg, error){
    containerCfg := &docker_client.ContainerCfg{
        Network: &network.NetworkingConfig{},
    }

	port, err := nat.NewPort("tcp", PORT)
	if err != nil {
		log.Panicf("unable to create port: %v", err)
		return nil, err
	}
	portSet := make(nat.PortSet)
	portSet[port] = struct{}{}
	containerCfg.Config = &container.Config{
		Image:        TAG,
		ExposedPorts: portSet,
	}

	portBinding := nat.PortBinding{
		HostIP:   "",
		HostPort: PORT,
	}
	portMap := make(nat.PortMap)
	portMap[port] = []nat.PortBinding{portBinding}
	containerCfg.Host = &container.HostConfig{
		PortBindings: portMap,
		NetworkMode:  "default",
	}
    return containerCfg, nil
}

func arrayContains(items []string, item string) bool {
	for _, it := range items {
		if it == item {
			return true
		}
	}

	return false
}
