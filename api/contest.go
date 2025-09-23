package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Contest struct {
	ContestID   int    `json:"contest_id"`
	ContestName string `json:"contest_name"`
}

func (api *ApiContext) createContest(w http.ResponseWriter, r *http.Request) {
	var newContest Contest
	if err := json.NewDecoder(r.Body).Decode(&newContest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Verificar si el nombre del concurso ya existe
	var contestName string
	query := "SELECT contest_name FROM contest WHERE contest_name=?"
	err := api.DB.QueryRow(query, newContest.ContestName).Scan(&contestName)
	if err == nil {
		// existe
		http.Error(w, "El nombre del contest ya esta registrado", http.StatusConflict)
		log.Println("El nombre del contest ya esta registrado")
		return
	} else if err != sql.ErrNoRows {
		// error distinto a no rows
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	res, err := api.DB.Exec(
		"INSERT INTO contest (contest_name) VALUES (?)",
		newContest.ContestName,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Intentar obtener el id generado
	if id, ierr := res.LastInsertId(); ierr == nil {
		newContest.ContestID = int(id)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newContest)
}
