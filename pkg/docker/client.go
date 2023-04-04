package docker_client

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	container "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

type DockerClient interface {
    PullImage(ctx context.Context, imageName string, imageOpts types.ImagePullOptions, muted bool) error
	ListImagesTags(ctx context.Context, imageListOpts types.ImageListOptions) (imagetags []string,err error)
	ListContainers(ctx context.Context, containerListOpts types.ContainerListOptions) (containerNames map[string]string, err error)
	CreateContainer(ctx context.Context, imageName string, containerCfg *ContainerCfg, platformCfg *specs.Platform) (conttainerID string, err error)
	StartContainer(ctx context.Context, containerID string) error
	StopContainer(ctx context.Context, containerID string) error
}

type dockerClient struct {
	*client.Client
}

type ContainerCfg struct {
	Config  *container.Config
	Host    *container.HostConfig
	Network *network.NetworkingConfig
}

func NewDockerClient() (DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return &dockerClient{}, err
	}
	return &dockerClient{cli}, nil
}

func (client *dockerClient) PullImage(ctx context.Context, imageName string, imagePullOptions types.ImagePullOptions, muted bool) error {
    reader, err := client.Client.ImagePull(ctx, imageName, imagePullOptions)
    if err != nil {
        return err
    }
    defer reader.Close()

    if !muted {
        io.Copy(os.Stdout, reader)
    }

    return nil
}

func (client *dockerClient) ListImagesTags(ctx context.Context, opts types.ImageListOptions) ([]string, error) {
	images, err := client.Client.ImageList(ctx, opts)
	if err != nil {
		return []string{}, err
	}

	return extractTags(images), nil
}

func extractTags(images []types.ImageSummary) []string {
	var tags []string
	for _, image := range images {
		if len(image.RepoTags) != 0 {
			tags = append(tags, image.RepoTags[0])
		}
	}

    return tags
}

func (client *dockerClient) ListContainers(ctx context.Context, opts types.ContainerListOptions) (map[string]string, error) {
	containerList, err := client.Client.ContainerList(ctx, opts)
	if err != nil {
		return nil, err
	}

	return extractContainerNameToID(containerList), nil
}

func extractContainerNameToID(conatinersList []types.Container) map[string]string {
	containers := map[string]string{}

	for _, container := range conatinersList {
		containers[container.Names[0][1:]] = container.ID
	}

    return containers
}

func (client *dockerClient) CreateContainer(ctx context.Context, name string, containerCfg *ContainerCfg, platformCfg *specs.Platform) (string, error) {
	con, err := client.ContainerCreate(ctx, containerCfg.Config, containerCfg.Host, containerCfg.Network, platformCfg, name)
	if err != nil {
		return "", nil
	}
	return con.ID, nil
}

func (client *dockerClient) StartContainer(ctx context.Context, containerID string) error {
	return client.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
}

func (client *dockerClient) StopContainer(ctx context.Context, containerID string) error {
	return client.ContainerStop(ctx, containerID, nil)
}
