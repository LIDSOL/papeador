package judge

import (
	"bufio"
	"context"
	"io"

	"log"
	"os"

	buildahDefine "github.com/containers/buildah/define"
	"github.com/containers/podman/v5/pkg/api/handlers"
	"github.com/containers/podman/v5/pkg/bindings"
	"github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/containers/podman/v5/pkg/bindings/images"
	"github.com/containers/podman/v5/pkg/domain/entities/types"
	"github.com/containers/podman/v5/pkg/specgen"
	dockerContainer "github.com/docker/docker/api/types/container"
	_ "modernc.org/sqlite"
)

var podmanConn context.Context

func GetConn() context.Context {
	return podmanConn
}

func ConnectToPodman(connURI string) (context.Context, error) {
	conn, err := bindings.NewConnection(context.Background(), connURI)
	if err != nil {
		return nil, err
	}

	podmanConn = conn

	return conn, nil
}

func CreateSandbox(conn context.Context, programStr, testInputStr, expectedOutputStr string) (types.ContainerCreateResponse, error) {
	err := os.Chdir("/vol/podman")
	if err != nil {
		return types.ContainerCreateResponse{}, err
	}

	programPath := "./papeador-submission.go"
	testInputPath := "./papeador-input.txt"
	expectedOutputPath := "./papeador-output.txt"

	err = writeStringToFile(programPath, programStr)
	if err != nil {
		return types.ContainerCreateResponse{}, err
	}

	err = writeStringToFile(testInputPath, testInputStr)
	if err != nil {
		return types.ContainerCreateResponse{}, err
	}

	err = writeStringToFile(expectedOutputPath, expectedOutputStr)
	if err != nil {
		return types.ContainerCreateResponse{}, err
	}

	options := types.BuildOptions{
		BuildOptions: buildahDefine.BuildOptions{
			Output:           "program:latest",
			ConfigureNetwork: buildahDefine.NetworkDisabled,
		},
	}

	log.Println("Building")
	_, err = images.Build(conn, []string{"./Dockerfile"}, options)
	if err != nil {
		return types.ContainerCreateResponse{}, err
	}

	log.Println("Creating with spec")
	s := specgen.NewSpecGenerator("program:latest", false)
	s.Command = []string{"/bin/sh", "-c", "sleep 5"}
	s.Name = "submission-sandbox" + genRandStr(8)
	createReponse, err := containers.CreateWithSpec(conn, s, nil)
	if err != nil {
		return types.ContainerCreateResponse{}, err
	}

	os.Remove(programPath)
	os.Remove(testInputPath)
	os.Remove(expectedOutputPath)
	return createReponse, nil
}

func StartSandbox(conn context.Context, createResponse types.ContainerCreateResponse, w io.Writer) error {
	log.Println(createResponse.ID)
	if err := containers.Start(conn, createResponse.ID, &containers.StartOptions{}); err != nil {
		return err
	}
	dockerExecOpts := dockerContainer.ExecOptions{
		Cmd:          []string{"/bin/papeador"},
		AttachStderr: true,
		AttachStdout: true,
		AttachStdin:  false,
		Tty:          true,
	}
	execConfig := &handlers.ExecCreateConfig{dockerExecOpts}
	execId, err := containers.ExecCreate(conn, createResponse.ID, execConfig)
	if err != nil {
		return err
	}
	_ = execId

	var stderr io.Writer = os.Stdout
	input := bufio.NewReader(os.Stdin)
	_ = input

	attachOptions := new(containers.ExecStartAndAttachOptions)
	attachOptions.WithAttachError(true).WithAttachInput(false).WithAttachOutput(true).WithErrorStream(stderr).WithOutputStream(w)
	err = containers.ExecStartAndAttach(conn, execId, attachOptions)

	if err != nil {
		return err
	}

	return nil
}
