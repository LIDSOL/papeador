package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"lidsol.org/papeador/store"
)

func (api *ApiContext) createUser(w http.ResponseWriter, r *http.Request) {
	var in store.User

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("E: %v\n", in)
	if err := api.Store.CreateUser(r.Context(), &in); err != nil {
		if err == store.ErrAlreadyExists {
			http.Error(w, "El usuario ya está registrado", http.StatusConflict)
			log.Println("El usuario ya está registrado")
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
