package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"lidsol.org/papeador/security"
	"lidsol.org/papeador/store"
)

type ContestRequestContent struct {
	store.Contest
	Username string `json:"username"`
	JWT      string `json:"jwt"`
}

func (api *ApiContext) createContest(w http.ResponseWriter, r *http.Request) {
	var in ContestRequestContent
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ok, err := security.ValidateJWT(in.JWT)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "Token inválida", http.StatusUnauthorized)
		return
	}

	contData := store.Contest{ContestID: in.ContestID, ContestName: in.ContestName}

	if err := api.Store.CreateContest(r.Context(), &contData); err != nil {
		if err == store.ErrAlreadyExists {
			http.Error(w, "El nombre del contest ya esta registrado", http.StatusConflict)
			log.Println("El nombre del contest ya esta registrado")
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

func (api *ApiContext) getContests(w http.ResponseWriter, r *http.Request) {
	log.Println("getcontests")
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("<h1>getContests: Not implemented:...</h1>"))
}

func (api *ApiContext) getContestByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("<h1>Not implemented: %v...</h1>", id)))
}
