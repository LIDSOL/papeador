package api

import (
	"log"
	"net/http"

	"lidsol.org/papeador/store"
)

type ApiContext struct{
	Store store.Store
}

func API(s store.Store) {
	apiCtx := ApiContext{
		Store: s,
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