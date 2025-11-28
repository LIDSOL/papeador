package api

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"lidsol.org/papeador/store"
)

type ApiContext struct {
	Store store.Store
}

var templates = template.Must(template.ParseGlob("templates/*.html"))

func API(s store.Store, port int) {
	apiCtx := ApiContext{
		Store: s,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", apiCtx.getContests)

	mux.HandleFunc("POST /login", apiCtx.createUser)
	mux.HandleFunc("GET /login", apiCtx.createUser)

	mux.HandleFunc("POST /users", apiCtx.createUser)
	mux.HandleFunc("GET /users", apiCtx.createUserView)
	mux.HandleFunc("GET /users/{id}", apiCtx.getUserByID)

	mux.HandleFunc("POST /contests/new", apiCtx.createContest)
	mux.HandleFunc("GET /contests/new", apiCtx.createContestView)
	mux.HandleFunc("GET /contests", apiCtx.getContests)
	mux.HandleFunc("GET /contests/{id}", apiCtx.getContestByID)

	mux.HandleFunc("POST /contests/{id}/problems/new", apiCtx.createProblem)
	mux.HandleFunc("GET /contests/{id}/problems/new", apiCtx.createProblemView)
	mux.HandleFunc("GET /contests/{contestID}/problems/{problemID}", apiCtx.getProblemByID)
	mux.HandleFunc("GET /contests/{contestID}/problems/{problemID}/pdf", apiCtx.getProblemStatementByID)

	mux.HandleFunc("POST /contests/{constestID}/problems/{problemID}/submit", apiCtx.submitProgram)
	mux.HandleFunc("GET /contests/{constestID}/problems/{problemID}/submit/{submitID}", apiCtx.getSubmissionByID)
	mux.HandleFunc("GET /contests/{constestID}/problems/{problemID}/submit", apiCtx.getSubmissions)
	mux.HandleFunc("GET /contests/{constestID}/problems/{problemID}/last-submit", apiCtx.getLastSubmission)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
