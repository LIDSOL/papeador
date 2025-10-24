package main

import (
	"flag"

	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"lidsol.org/papeador/api"
	"lidsol.org/papeador/judge"

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

func main() {
	uriPtr := flag.String("u", "unix:///run/user/1000/podman/podman.sock", "uri for Podman API service connection")
	var port int
	flag.IntVar(&port, "p", 8080, "port to listen from")
	flag.Parse()

	uri := *uriPtr
	_, err := judge.ConnectToPodman(uri)
	if err != nil {
		log.Fatalf("Could not connect to Podman: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /program", api.SubmitProgram)
	log.Printf("Starting server at :%v\n", port)
	http.ListenAndServe(fmt.Sprintf(":%v", port), mux)
}
