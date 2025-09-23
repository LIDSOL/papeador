package api

import (
	"encoding/json"
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
