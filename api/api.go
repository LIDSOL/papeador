package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

type User struct {
	User_ID  int    `json:"user_id"`
	Username string `json:"username"`
	Passhash string `json:"passhash"`
	Email    string `json:"email"`
}

var db *sql.DB

func API(sqlitedb *sql.DB) {
	db = sqlitedb

	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", createUser)
	//mux.HandleFunc("POST /users", createContest)
	//mux.HandleFunc("POST /users", createProblem)

	log.Fatal(http.ListenAndServe(":8000", mux))
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var new_user User

	err := json.NewDecoder(r.Body).Decode(&new_user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Verify if the user already exists (unique username and email)
	var username string
	query := "SELECT username FROM user WHERE username=? OR email=?"
	err = db.QueryRow(query, new_user.Username, new_user.Email).Scan(&username)
	if err != sql.ErrNoRows {
		http.Error(w, "El usuario ya está registrado", http.StatusConflict)
		log.Println("El usuario ya está registrado")
		return
	}

	_, err = db.Query("INSERT INTO user (username,passhash,email) VALUES (?, ?, ?)", new_user.Username, new_user.Passhash, new_user.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(new_user)
}
