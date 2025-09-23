package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type User struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Passhash string `json:"passhash"`
	Email    string `json:"email"`
}

func (api *ApiContext) createUser(w http.ResponseWriter, r *http.Request) {
	var new_user User

	err := json.NewDecoder(r.Body).Decode(&new_user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Verify if the user already exists (unique username and email)
	var username string
	query := "SELECT username FROM user WHERE username=? OR email=?"
	err = api.DB.QueryRow(query, new_user.Username, new_user.Email).Scan(&username)
	if err == nil {
		http.Error(w, "El usuario ya está registrado", http.StatusConflict)
		log.Println("El usuario ya está registrado")
		return
	} else if err != sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	res, err := api.DB.Exec("INSERT INTO user (username,passhash,email) VALUES (?, ?, ?)", new_user.Username, new_user.Passhash, new_user.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Obtener el ID del usuario recién insertado
	if id, err := res.LastInsertId(); err == nil {
		new_user.User_ID = int(id)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(new_user)
}