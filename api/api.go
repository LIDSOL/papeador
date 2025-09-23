package api

import (
	"database/sql"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

type ApiContext struct{
	DB *sql.DB
}

func API(sqlitedb *sql.DB) {
	apiCtx := ApiContext{
		DB: sqlitedb,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/users", methodHandler("POST", apiCtx.createUser))
	mux.HandleFunc("/contests", methodHandler("POST", apiCtx.createContest))
	mux.HandleFunc("/problems", methodHandler("POST", apiCtx.createProblem))

	log.Fatal(http.ListenAndServe(":8000", mux))
}

func methodHandler(method string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h(w, r)
	}
}