package api

import (
	"log"
	"net/http"

	"lidsol.org/papeador/store"
)

type ApiContext struct {
	Store store.Store
}

func API(s store.Store) {
	apiCtx := ApiContext{
		Store: s,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /users", apiCtx.createUser)
	mux.HandleFunc("GET /users/{id}", apiCtx.getUserByID)

	mux.HandleFunc("POST /contests", apiCtx.createContest)
	mux.HandleFunc("GET /contests", apiCtx.getContests)
	mux.HandleFunc("GET /contests/{id}", apiCtx.getContestByID)

	mux.HandleFunc("POST /contests/{id}/problems", apiCtx.createProblem)
	mux.HandleFunc("GET /contests/{id}/problems", apiCtx.getProblems)
	mux.HandleFunc("GET /contests/{contestID}/problems/{problemID}", apiCtx.getProblemByID)

	mux.HandleFunc("POST /contests/{constestID}/problems/{problemID}/submit", apiCtx.submitProgram)
	mux.HandleFunc("GET /contests/{constestID}/problems/{problemID}/submit/{submitID}", apiCtx.getSubmissionByID)
	mux.HandleFunc("GET /contests/{constestID}/problems/{problemID}/submit", apiCtx.getSubmissions)
	mux.HandleFunc("GET /contests/{constestID}/problems/{problemID}/last-submit", apiCtx.getLastSubmission)

	log.Fatal(http.ListenAndServe(":8000", mux))
}
