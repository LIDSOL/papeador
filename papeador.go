package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"

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

func connectToPodman(connURI string) (context.Context, error) {

	// Crear el contenedor con la multistaged build
	// Ejecutar el contenedor
	// Comparar su salida con la esperada

	conn, err := bindings.NewConnection(context.Background(), connURI)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func createSandbox(conn context.Context, programString string) (types.ContainerCreateResponse, error) {
	sourcePath := "/tmp/papeador-submission.go"
	f, err := os.Create(sourcePath)
	if err != nil {
		return types.ContainerCreateResponse{}, err
	}

	w := bufio.NewWriter(f)
	_, err = w.WriteString(programString)
	if err != nil {
		return types.ContainerCreateResponse{}, err
	}
	w.Flush()
	f.Close()

	options := types.BuildOptions{
		BuildOptions: buildahDefine.BuildOptions{
			Output: "program:latest",
			ConfigureNetwork: buildahDefine.NetworkDisabled,

		},
	}

	_, err = images.Build(conn, []string{"./podman/Dockerfile"}, options)
	if err != nil {
		return types.ContainerCreateResponse{}, err
	}

	s := specgen.NewSpecGenerator("program:latest", false)
	s.Name = "submission-sandbox" + randomString(8)
	createReponse, err := containers.CreateWithSpec(conn, s, nil)
	if err != nil {
		return types.ContainerCreateResponse{}, err
	}

	return createReponse, nil
}

func runOnSandbox(conn context.Context, createResponse types.ContainerCreateResponse) (string, error) {

	if err := containers.Start(conn, createResponse.ID, nil); err != nil {
		return "", err
	}

	// Ejecuta el programa midiendo el tiempo y con timeout
	// Si hubo timeout, reportar
	// Comparar las salidas con diff y reportar errores
	// Reportar el tiempo de ejecución

	// #!/bin/sh

	// timeout 3s time hello 2>time.output >output
	// # si $? == 143, timeout
	// # si $? != 0, runtime error

	// diff output /tmp/expected-output
	// # si $? == 1, incorrecto

	// awk 'NR==1{print $3}' time.output
	// report time

	return "submission-sandbox", nil
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

	createResponse, err := createSandbox(conn, string(buf)[:n])
	output, err := runOnSandbox(conn, createResponse)
	log.Println("Container created successfully")

	expected := "10\n"
	if output == expected {
		fmt.Println("Respuesta correcta")
	} else {
		fmt.Println("Respuesta incorrecta")
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Archivo recibido\n"))
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
