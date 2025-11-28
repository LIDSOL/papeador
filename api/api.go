package api

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"lidsol.org/papeador/auth"
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

	mux.HandleFunc("POST /users/login", apiCtx.loginUser)
	mux.HandleFunc("GET /login", apiCtx.loginUserView)
	mux.HandleFunc("GET /logout", apiCtx.logoutUser)

	mux.HandleFunc("POST /users/create", apiCtx.createUser)
	// mux.HandleFunc("GET /users", apiCtx.createUserView)
	mux.HandleFunc("GET /users/{id}", apiCtx.getUserByID)

	mux.HandleFunc("POST /contests/new", auth.RequireAuth(apiCtx.createContest))
	mux.HandleFunc("GET /new-contest", auth.RequireAuth(apiCtx.createContestView))
	mux.HandleFunc("GET /contests", apiCtx.getContests)
	mux.HandleFunc("GET /contests/{id}", apiCtx.getContestByID)

	mux.HandleFunc("POST /contests/{id}/problems/new", auth.RequireAuth(apiCtx.createProblem))
	mux.HandleFunc("GET /contests/{id}/new-problem", auth.RequireAuth(apiCtx.createProblemView))
	mux.HandleFunc("GET /contests/{contestID}/problems/{problemID}", apiCtx.getProblemByID)
	mux.HandleFunc("GET /contests/{contestID}/problems/{problemID}/pdf", apiCtx.getProblemStatementByID)

	mux.HandleFunc("POST /contests/{constestID}/problems/{problemID}/submit", auth.RequireAuth(apiCtx.submitProgram))
	// mux.HandleFunc("GET /contests/{constestID}/problems/{problemID}/submit/{submitID}", auth.RequireAuth(apiCtx.getSubmissionByID))
	mux.HandleFunc("GET /contests/{constestID}/problems/{problemID}/submit", auth.RequireAuth(apiCtx.getSubmissions))
	// mux.HandleFunc("GET /contests/{constestID}/problems/{problemID}/last-submit", auth.RequireAuth(apiCtx.getLastSubmission))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
