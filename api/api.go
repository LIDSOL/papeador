package api

import (
	"fmt"
	"log"
	"net/http"

	"lidsol.org/papeador/store"
)

type ApiContext struct{
	Store store.Store
}

func API(s store.Store, port int) {
	apiCtx := ApiContext{
		Store: s,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/program", methodHandler("POST", apiCtx.submitProgram))
	mux.HandleFunc("/users", methodHandler("POST", apiCtx.createUser))
	mux.HandleFunc("/contests", methodHandler("POST", apiCtx.createContest))
	mux.HandleFunc("/problems", methodHandler("POST", apiCtx.createProblem))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), mux))
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
