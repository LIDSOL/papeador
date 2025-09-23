package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type Problem struct {
	ProblemID   int     `json:"problem_id"`
	ContestID   *int    `json:"contest_id"`
	CreatorID   int     `json:"creator_id"`
	ProblemName string  `json:"problem_name"`
	Description string  `json:"description"`
}

func (api *ApiContext) createProblem(w http.ResponseWriter, r *http.Request) {
	var newProblem Problem
	if err := json.NewDecoder(r.Body).Decode(&newProblem); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if newProblem.ContestID != nil {
		var contestID int
		query := "SELECT contest_id FROM CONTEST WHERE contest_id=?"
		err := api.DB.QueryRow(query, *newProblem.ContestID).Scan(&contestID)
		if err == sql.ErrNoRows {
			http.Error(w, "El concurso no existe", http.StatusConflict)
			log.Println("El concurso no existe")
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}

	res, err := api.DB.Exec(
		"INSERT INTO PROBLEM (contest_id, creator_id, problem_name, description) VALUES (?, ?, ?, ?)",
		newProblem.ContestID, newProblem.CreatorID, newProblem.ProblemName, newProblem.Description,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Obtener el ID del problema recién insertado
	if id, err := res.LastInsertId(); err == nil {
		newProblem.ProblemID = int(id)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newProblem)
}
