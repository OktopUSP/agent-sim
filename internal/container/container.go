package container

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log/slog"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

func CreateDockerClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func BuildDockerImage(ctx context.Context, cli *client.Client, imageName string, dockerfilePath string) error {

	/* ------------------------- Open + Read Dokcerfile ------------------------- */
	dockerFileReader, err := os.Open(dockerfilePath)
	if err != nil {
		slog.Error("Error to open dockerfile", "error", err)
		os.Exit(1)
	}
	readDockerFile, err := ioutil.ReadAll(dockerFileReader)
	if err != nil {
		slog.Error("Error to read dockerfile", "error", err)
		os.Exit(1)
	}
	/* -------------------------------------------------------------------------- */

	/* ------------------ Transform Dockerfile into tar format ------------------ */
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	tarHeader := &tar.Header{
		Name: "Dockerfile",
		Size: int64(len(readDockerFile)),
	}
	err = tw.WriteHeader(tarHeader)
	if err != nil {
		slog.Error("Error to write tar header", "error", err)
		os.Exit(1)
	}
	_, err = tw.Write(readDockerFile)
	if err != nil {
		slog.Error("Error to write tar body", "error", err)
		os.Exit(1)
	}
	dockerFileTarReader := bytes.NewReader(buf.Bytes())
	/* -------------------------------------------------------------------------- */

	buildOptions := types.ImageBuildOptions{
		Tags:    []string{imageName},
		Context: dockerFileTarReader,
		Remove:  true,
	}

	buildResponse, err := cli.ImageBuild(ctx, dockerFileTarReader, buildOptions)
	if err != nil {
		return err
	}
	defer buildResponse.Body.Close()

	//Copy the build output to stdout
	_, err = io.Copy(os.Stdout, buildResponse.Body)
	if err != nil {
		return err
	}

	//log.Printf("Image --> '%s' built succesfuly!\n", imageName)
	return nil
}

func DownloadDockerImage(ctx context.Context, cli *client.Client) error {
	reader, err := cli.ImagePull(ctx, "oktopusp/obuspa:latest", types.ImagePullOptions{})
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, reader)
	return nil
}

func RunDockerContainer(
	ctx context.Context,
	cli *client.Client,
	imageName,
	bridgeName,
	containerName,
	confBind,
	tlsBind string,
) (string, error) {

	hostConfig := container.HostConfig{
		NetworkMode: "bridge",
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: confBind,
				Target: "/etc/factory_reset_example.txt",
			},
		},
		ExtraHosts: []string{
			"host.docker.internal:host-gateway",
		},
	}

	containerConfig := container.Config{
		Image: "oktopusp/obuspa:latest",
		// Cmd:   []string{"obuspa", "-p", "-v4", "-i", "lo", "-r", "/etc/factory_reset_example.txt"},
		Cmd: []string{"obuspa", "-p", "-v4", "-r", "/etc/factory_reset_example.txt"},
		Tty: true,
	}

	if tlsBind != "" {
		containerConfig.Cmd = append(containerConfig.Cmd, "-t", "/etc/chain.pem")
		hostConfig.Mounts = append(hostConfig.Mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: tlsBind,
			Target: "/etc/chain.pem",
		})
	}

	networking_config := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			bridgeName: {}},
	}

	resp, err := cli.ContainerCreate(
		ctx,
		&containerConfig,
		&hostConfig,
		networking_config,
		nil,
		containerName,
	)
	if err != nil {
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	slog.Debug("Container created and started", "container", containerName)
	return resp.ID, nil
}

func DeleteDockerContainer(ctx context.Context, cli *client.Client, containerName string) error {
	slog.Debug("Stopping docker container ...", "container", containerName)
	err := stopDockerContainer(ctx, cli, containerName)
	if err != nil {
		return err
	}

	slog.Debug("Removing docker container ...", "container", containerName)
	err = removeDockerContainer(ctx, cli, containerName)
	if err != nil {
		return err
	}

	return nil
}

func stopDockerContainer(ctx context.Context, cli *client.Client, containerName string) error {
	if err := cli.ContainerStop(ctx, containerName, container.StopOptions{}); err != nil {
		return err
	}

	//log.Printf("Container %s stopped\n", containerName)
	return nil
}

func removeDockerContainer(ctx context.Context, cli *client.Client, containerName string) error {
	options := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}
	if err := cli.ContainerRemove(ctx, containerName, options); err != nil {
		return err
	}

	//log.Printf("Container %s removed\n", containerName)
	return nil
}
