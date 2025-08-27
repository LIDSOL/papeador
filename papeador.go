package main

import (
	"database/sql"
	"fmt"
	"log"
	"os/exec"

	"lidsol.org/papeador/api"
	_ "modernc.org/sqlite"
)

func main() {
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
	api.API(db)
}
