package main

import (
	"bufio"
	// "context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	// "github.com/containers/podman/v5/pkg/bindings"
	// "github.com/containers/podman/v5/pkg/bindings/containers"
	// "github.com/containers/podman/v5/pkg/bindings/images"
	// "github.com/containers/podman/v5/pkg/specgen"
	_ "modernc.org/sqlite"
)

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

func connectToPodman(programString string) string {

	// Crear el contenedor con la multistaged build
	// Ejecutar el contenedor
	// Comparar su salida con la esperada

	// Escribir esta cadena a un archivo para cargalo con el Dockerfile
	path := "./podman/program.go"
	f, err := os.Create(path);
	if err != nil {
		panic(err)
	}

	w := bufio.NewWriter(f)
	_, err = w.WriteString(programString)
	if err != nil {
		panic(err)
	}
	w.Flush()
	f.Close()

	command := exec.Command("podman", "build", "-t", "hello", "podman")
	output, err := command.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		log.Fatal(err)
	}


	command = exec.Command("podman", "run", "hello")
	output, err = command.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		log.Fatal(err)
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

	return string(output)
}

func submitProgram(w http.ResponseWriter, r *http.Request) {
	file, fileHeader, err := r.FormFile("program")
	if err != nil {
		log.Printf("Could not get form file: %v\n", err)
		http.Error(w, "Could not get form file", http.StatusBadRequest)
		return
	}

	var buf []byte = make([]byte, 5 * 1024)
	n, err := file.Read(buf)
	if err != nil {
		log.Printf("Could not read file: %v\n", err)
		http.Error(w, "Could not read file", http.StatusInternalServerError)
		return
	}

	fmt.Println(n)
	fmt.Println(string(buf))
	fmt.Println(fileHeader.Size, fileHeader.Filename)

	output := connectToPodman(string(buf)[:n])

	expected := "10\n";
	if output == expected {
		fmt.Println("Respuesta correcta")
	} else {
		fmt.Println("Respuesta incorrecta")
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Archivo recibido\n"))
}

func main() {
	mux := http.NewServeMux()
	port := 8080

	mux.HandleFunc("POST /program", submitProgram)
	log.Printf("Starting server at :%v\n", port)
	http.ListenAndServe(fmt.Sprintf(":%v", port), mux)
}
