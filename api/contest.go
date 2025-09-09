package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type Contest struct {
	ContestID   int    `json:"contest_id"`
	ContestName string `json:"contest_name"`
	Description string `json:"description"`
}

func createContest(w http.ResponseWriter, r *http.Request) {
	var newContest Contest
	if err := json.NewDecoder(r.Body).Decode(&newContest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var contestName string
	querry := "SELECT contest_name FROM contest WHERE contest_name=?"
	err = db.QueryRow(querry, newContest).Scan(&contestName)
	if err != sql.ErrNoRows {
		http.Error(w, "El nombre del contest ya esta registrado", http.StatusConflict)
		log.Println("El nombre del contest ya esta registrado")
		return
	}

	_, err := db.Exec(
		"INSERT INTO CONTEST (contest_id, contest_name, description) VALUES (?, ?, ?)",
		newContest.ContestID, newContest.ContestName, newContest.Description,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newContest)
}
