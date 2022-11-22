package docker_client

import (
	"context"

	"github.com/docker/docker/api/types"
	containr "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

type DockerClient struct {
	*client.Client
}

func NewDockerClient() (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return &DockerClient{}, err
	}
	return &DockerClient{cli}, nil
}

func (client *DockerClient)ListImages(ctx context.Context, opts types.ImageListOptions) ([]string, error) {
    images, err := client.Client.ImageList(ctx, opts)
    if err != nil {
        return []string{}, err
    }

    var tags []string
    for _, image := range images {
        tags = append(tags, image.RepoTags[0])
    }

    return tags, nil
}

type Container interface {
	Create(context.Context, string, *specs.Platform) (string, error)
}

type container struct {
	Client  *DockerClient
	Config  *containr.Config
	Host    *containr.HostConfig
	Network *network.NetworkingConfig
}

type ContainerOption func(*container)

func WithContainerConfig(cfg *containr.Config) ContainerOption {
    return func(c *container) {
        c.Config = cfg
    }
}

func WithHostConfig(hostCfg *containr.HostConfig) ContainerOption {
    return func(c *container) {
        c.Host = hostCfg
    }
}

func WithNetworkConfig(networkCfg *network.NetworkingConfig) ContainerOption {
    return func(c *container) {
        c.Network = networkCfg
    }
}

func NewContainer(client *DockerClient, opts... ContainerOption) (Container, error) {
    c := &container{
        Client: client,
    }
    for _, opt := range opts {
        opt(c)
    }

	return c, nil
}

func (c *container) Create(ctx context.Context, name string, platformCfg *specs.Platform) (string, error) {
	con, err := c.Client.ContainerCreate(ctx, c.Config, c.Host, c.Network, platformCfg, name)
	if err != nil {
		return "", nil
	}
	return con.ID, nil
}
