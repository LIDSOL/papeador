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
	"github.com/containers/podman/v5/pkg/bindings"
	"github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/containers/podman/v5/pkg/bindings/images"
	"github.com/containers/podman/v5/pkg/domain/entities/types"
	"github.com/containers/podman/v5/pkg/specgen"
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

	w := bufio.NewWriter(f)
	_, err = w.WriteString(content)
	if err != nil {
		return err
	}

	w.Flush()

	return nil
}

func connectToPodman(connURI string) (context.Context, error) {
	conn, err := bindings.NewConnection(context.Background(), connURI)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func createSandbox(conn context.Context, programStr, testInputStr, expectedOutputStr string) (types.ContainerCreateResponse, *io.PipeReader, error) {
	programPath := "/tmp/papeador-submission.go"
	testInputPath := "/tmp/papeador-input.txt"
	expectedOutputPath := "/tmp/papeador-output.txt"

	err := writeStringToFile(programPath, programStr)
	if err != nil {
		return types.ContainerCreateResponse{}, nil, err
	}

	err = writeStringToFile(testInputPath, testInputStr)
	if err != nil {
		return types.ContainerCreateResponse{}, nil, err
	}

	err = writeStringToFile(expectedOutputPath, expectedOutputStr)
	if err != nil {
		return types.ContainerCreateResponse{}, nil, err
	}

	defer func() {
		os.Remove(programPath)
		os.Remove(testInputPath)
		os.Remove(expectedOutputPath)
	}()

	r, w := io.Pipe()

	options := types.BuildOptions{
		BuildOptions: buildahDefine.BuildOptions{
			Output: "program:latest",
			ConfigureNetwork: buildahDefine.NetworkDisabled,
			Out: w, // Read stdout later
		},
	}

	_, err = images.Build(conn, []string{"./podman/Dockerfile"}, options)
	if err != nil {
		return types.ContainerCreateResponse{}, nil, err
	}

	s := specgen.NewSpecGenerator("program:latest", false)
	s.Name = "submission-sandbox" + randomString(8)
	createReponse, err := containers.CreateWithSpec(conn, s, nil)
	if err != nil {
		return types.ContainerCreateResponse{}, nil, err
	}

	return createReponse, r, nil
}

func startSandbox(conn context.Context, createResponse types.ContainerCreateResponse) (error) {
	if err := containers.Start(conn, createResponse.ID, nil); err != nil {
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

	createResponse, stdoutPipe, err := createSandbox(conn, string(buf)[:n], "10\n", "10\n")
	if err != nil {
		http.Error(w, "Could not create sandbox environment", http.StatusInternalServerError)
		return
	}

	err = startSandbox(conn, createResponse)
	if err != nil {
		http.Error(w, "Could not start sandbox", http.StatusInternalServerError)
		return
	}

	log.Println("Container created successfully")

	stdoutBuf := new(strings.Builder)
	_, err = io.Copy(stdoutBuf, stdoutPipe)

	stdoutStr := strings.TrimSpace(stdoutBuf.String())

	resp := fmt.Sprintf("Result: %v\n", stdoutStr)


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
