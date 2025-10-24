package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"lidsol.org/papeador/judge"

	"github.com/containers/podman/v5/pkg/bindings/containers"
)

func SubmitProgram(w http.ResponseWriter, r *http.Request) {
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

	conn := judge.GetConn()

	createResponse, err := judge.CreateSandbox(conn, string(buf)[:n], "10\n", "10\n")
	if err != nil {
		s := fmt.Sprintf("Could not create sandbox: %v", err)
		http.Error(w, s, http.StatusInternalServerError)
		return
	}

	stdoutBuf := new(strings.Builder)
	err = judge.StartSandbox(conn, createResponse, stdoutBuf)
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
