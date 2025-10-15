package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"io"
	"strings"

	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

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

var uri string

func randomString(length int) string {
	b := make([]byte, length+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2:length+2]
}

func testDB() {
	db, err := sql.Open("sqlite", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	command := exec.Command("bash", "-c", "sqlite3 test.db <schema.sql")
	output, err := command.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		log.Fatal(err)
	}
}

func writeStringToFile(path, content string) (error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	if err != nil {
		return err
	}

	f.Sync()
	fmt.Println("Cerrando: ", path)

	return nil
}

func connectToPodman(connURI string) (context.Context, error) {
	conn, err := bindings.NewConnection(context.Background(), connURI)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func createSandbox(conn context.Context, programStr, testInputStr, expectedOutputStr string) (types.ContainerCreateResponse, error) {
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
			Output: "program:latest",
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
	s.Name = "submission-sandbox" + randomString(8)
	createReponse, err := containers.CreateWithSpec(conn, s, nil)
	if err != nil {
		return types.ContainerCreateResponse{}, err
	}

	os.Remove(programPath)
	os.Remove(testInputPath)
	os.Remove(expectedOutputPath)
	return createReponse, nil
}

func startSandbox(conn context.Context, createResponse types.ContainerCreateResponse, w io.Writer) (error) {
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

func submitProgram(w http.ResponseWriter, r *http.Request) {
	file, fileHeader, err := r.FormFile("program")
	if err != nil {
		log.Printf("Could not get form file: %v\n", err)
		http.Error(w, "Could not get form file", http.StatusBadRequest)
		return
	}

	var buf []byte = make([]byte, 5*1024)
	n, err := file.Read(buf)
	if err != nil {
		log.Printf("Could not read file: %v\n", err)
		http.Error(w, "Could not read file", http.StatusInternalServerError)
		return
	}

	fmt.Println(n)
	fmt.Println(string(buf))
	fmt.Println(fileHeader.Size, fileHeader.Filename)

	conn, err := connectToPodman(uri)
	if err != nil {
		log.Fatalln(err)
	}

	createResponse, err := createSandbox(conn, string(buf)[:n], "10\n", "10\n")
	if err != nil {
		s := fmt.Sprintf("Could not create sandbox: %v", err)
		http.Error(w, s, http.StatusInternalServerError)
		return
	}

	stdoutBuf := new(strings.Builder)
	err = startSandbox(conn, createResponse, stdoutBuf)
	if err != nil {
		log.Println(err)
		http.Error(w, "Could not start sandbox", http.StatusInternalServerError)
		return
	}

	log.Println("Container created successfully")
	_, err = containers.Wait(conn, createResponse.ID, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "Could not stop sandbox", http.StatusInternalServerError)
		return
	}

	resp := stdoutBuf.String() + "\n"
	log.Printf("Resultado: %v", resp)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp))
}

func main() {
	uriPtr := flag.String("u", "", "uri for podman connection")
	var port int
	flag.IntVar(&port, "p", 8080, "port to listen from")
	flag.Parse()

	uri = *uriPtr
	fmt.Println(uri, port)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /program", submitProgram)
	log.Printf("Starting server at :%v\n", port)
	http.ListenAndServe(fmt.Sprintf(":%v", port), mux)
}
