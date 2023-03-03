package docker_client

import (
	"context"

	"github.com/docker/docker/api/types"
	containr "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

type DockerClient interface {
	ListImages(context.Context, types.ImageListOptions) ([]string, error)
    ListContainers(context.Context, types.ContainerListOptions) (map[string] string, error)
	CreateContainer(context.Context, string, *ContainerCfg, *specs.Platform) (string, error)
    StartContainer(context.Context, string) error
}

type dockerClient struct {
	*client.Client
}

type ContainerCfg struct {
	Config  *containr.Config
	Host    *containr.HostConfig
	Network *network.NetworkingConfig
}

func NewDockerClient() (DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return &dockerClient{}, err
	}
	return &dockerClient{cli}, nil
}

func (client *dockerClient) ListImages(ctx context.Context, opts types.ImageListOptions) ([]string, error) {
	var tags []string
	images, err := client.Client.ImageList(ctx, opts)
	if err != nil {
		return tags, err
	}

	for _, image := range images {
		tags = append(tags, image.RepoTags[0])
	}

	return tags, nil
}

func (client *dockerClient) ListContainers(ctx context.Context, opts types.ContainerListOptions) (map[string] string, error) {
    containers := map[string] string{}
	c, err := client.Client.ContainerList(ctx, opts)
	if err != nil {
		return containers, err
	}

    for _, container := range c {
        containers[container.Names[0][1:]] = container.ID
    }
	return containers, nil
}

func (client *dockerClient) CreateContainer(ctx context.Context, name string, config *ContainerCfg, platformCfg *specs.Platform) (string, error) {
	con, err := client.ContainerCreate(ctx, config.Config, config.Host, config.Network, platformCfg, name)
	if err != nil {
		return "", nil
	}
	return con.ID, nil
}


func (client *dockerClient) StartContainer(ctx context.Context, containerID string) error {
    return client.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
}
