package main

import (
	"flag"
	"strings"

	"database/sql"
	"fmt"
	"log"
	"os/exec"

	"lidsol.org/papeador/api"
	"lidsol.org/papeador/judge"

	"lidsol.org/papeador/store"
	_ "modernc.org/sqlite"
)

func testDB(port int) {
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
	s := store.NewSQLiteStore(db)
	api.API(s, port)
}

func main() {
	uriPtr := flag.String("u", "unix:///run/user/1000/podman/podman.sock", "uri for Podman API service connection")
	var port int
	flag.IntVar(&port, "p", 8080, "port to listen from")
	flag.Parse()

	uri := *uriPtr
	for _, u := range strings.Split(uri, " ") {
		_, err := judge.ConnectToPodman(u)
		if err != nil {
			// log.Fatalf("Could not connect to Podman: %v", err)
		}
	}

	testDB(port)
}
