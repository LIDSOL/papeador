package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"lidsol.org/papeador/store"
)

func (api *ApiContext) createProblem(w http.ResponseWriter, r *http.Request) {
	var in store.Problem
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := api.Store.CreateProblem(r.Context(), &in); err != nil {
		if err == store.ErrNotFound {
			http.Error(w, "El concurso no existe", http.StatusConflict)
			log.Println("El concurso no existe")
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(in)
}

func (api *ApiContext) getProblems(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("<h1>Not implemented: %v...</h1>", id)))
}

func (api *ApiContext) getProblemByID(w http.ResponseWriter, r *http.Request) {
	contestId := r.PathValue("contestID")
	problemId := r.PathValue("problemID")

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("<h1>Not implemented: %v, %v...</h1>", contestId, problemId)))
}
