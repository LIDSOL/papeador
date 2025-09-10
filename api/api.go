package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)
var db *sql.DB

func API(sqlitedb *sql.DB) {
	db = sqlitedb

	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", createUser)
	mux.HandleFunc("POST /contests", createContest)
	mux.HandleFunc("POST /problems", createProblem)

	log.Fatal(http.ListenAndServe(":8000", mux))
}