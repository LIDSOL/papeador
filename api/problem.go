package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type Problem struct{
	problem_Id int `json:"problem_id"`
	contest_id int `json:"contest_id"`
	creator_id int `json:"creator_id"`
	problem_name string `json:problem_name`
	description string `json:description`
}

func createProblem (w http.ResponseWriter, r *http.Request){
	var newProblem Problem
	if err := json.NewDecoder(r.Body).Decode(&newProblem); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if newProblem.contest_id != nil {
		var contest_id int
		query := "SELECT contest_id FROM CONTEST WHERE contest_id=?"
		err := db.QueryRow(query, newProblem.contest_id).Scan(&contest_id)
		if err == sql.ErrNoRows {
			http.Error(w, "El concurso no existe", http.StatusConflict)
			log.Println("El concurso no existe")
			return
		}
	}

	_, err = db.Exec(
		"INSERT INTO PROBLEM (problem_id, contest_id, creator_id, problem_name, description) VALUES (?, ?, ?, ?, ?)",
		 newProblem.problem_Id, newProblem.contest_id, newProblem.creator_id, newProblem.problem_name,
		 newProblem.description
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newProblem)
}
