package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"lidsol.org/papeador/judge"

	"github.com/containers/podman/v5/pkg/bindings/containers"
)

func (api *ApiContext) submitProgram(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp))
}

func (api *ApiContext) getSubmissionByID(w http.ResponseWriter, r *http.Request) {
	contestId := r.PathValue("contestID")
	problemId := r.PathValue("problemID")
	submitId := r.PathValue("submitID")

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("<h1>Not implemented: %v, %v, %v...</h1>", contestId, problemId, submitId)))
}

func (api *ApiContext) getSubmissions(w http.ResponseWriter, r *http.Request) {
	contestId := r.PathValue("contestID")
	problemId := r.PathValue("problemID")

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("<h1>Not implemented: %v, %v...</h1>", contestId, problemId)))
}

func (api *ApiContext) getLastSubmission(w http.ResponseWriter, r *http.Request) {
	contestId := r.PathValue("contestID")
	problemId := r.PathValue("problemID")

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("<h1>Not implemented: %v, %v...</h1>", contestId, problemId)))
}
