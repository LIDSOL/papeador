package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"lidsol.org/papeador/judge"

	"github.com/containers/podman/v5/pkg/bindings/containers"
)

var m sync.Mutex

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

	m.Lock()
	worker := <- *judge.WorkerQueueP
	m.Unlock()

	conn := worker.Ctx

	filenameSep := strings.Split(fileHeader.Filename, ".")
	filetype := filenameSep[len(filenameSep)-1]

	testcases := []judge.SubmissionTestCase{
		judge.SubmissionTestCase{Input: "10\n", Output: "20\n"},
		judge.SubmissionTestCase{Input: "5\n", Output: "10\n"},
		judge.SubmissionTestCase{Input: "20\n", Output: "40\n"},
	}
	timelimit := 1
	createResponse, err := judge.CreateSandbox(conn, filetype, string(buf)[:n], testcases, timelimit)
	if err != nil {
		*judge.WorkerQueueP <- worker
		s := fmt.Sprintf("Could not create sandbox: %v", err)
		http.Error(w, s, http.StatusInternalServerError)
		return
	}

	stdoutBuf := new(strings.Builder)
	err = judge.StartSandbox(conn, createResponse, stdoutBuf)
	if err != nil {
		*judge.WorkerQueueP <- worker
		log.Println(err)
		http.Error(w, "Could not start sandbox", http.StatusInternalServerError)
		return
	}

	log.Println("Container created successfully")
	_, err = containers.Wait(conn, createResponse.ID, nil)
	if err != nil {
		*judge.WorkerQueueP <- worker
		log.Println(err)
		http.Error(w, "Could not stop sandbox", http.StatusInternalServerError)
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(stdoutBuf.String()))
	resMap := make(map[int]string)

	i := 0
	for scanner.Scan() {
		res := scanner.Text()
		log.Printf("Testcase %v: %v", i, res)
		resMap[i] = res
		i++
	}

	// Metelo de nuevo
	*judge.WorkerQueueP <- worker

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resMap)
}
