package container

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
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
		log.Fatal(err, " :unable to open Dockerfile")
	}
	readDockerFile, err := ioutil.ReadAll(dockerFileReader)
	if err != nil {
		log.Fatal(err, " :unable to read dockerfile")
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
		log.Fatal(err, " :unable to write tar header")
	}
	_, err = tw.Write(readDockerFile)
	if err != nil {
		log.Fatal(err, " :unable to write tar body")
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

func RunDockerContainer(
	ctx context.Context,
	cli *client.Client,
	imageName string,
	containerName string,
	confBind string,
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
	}

	containerConfig := container.Config{
		Image: imageName,
		Cmd:   []string{"obuspa", "-p", "-v", "4", "-r", "/etc/factory_reset_example.txt"},
		Tty:   true,
	}

	resp, err := cli.ContainerCreate(
		ctx,
		&containerConfig,
		&hostConfig,
		nil,
		nil,
		containerName,
	)
	if err != nil {
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	log.Printf("Container %s started\n", containerName)
	return resp.ID, nil
}

func DeleteDockerContainer(ctx context.Context, cli *client.Client, containerName string) error {
	//log.Printf("Stopping docker container: %s ...", containerName)
	err := stopDockerContainer(ctx, cli, containerName)
	if err != nil {
		return err
	}

	//log.Printf("Removing docker container: %s ...", containerName)
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
