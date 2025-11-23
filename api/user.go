package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"lidsol.org/papeador/security"
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
		} else if err == security.ErrInvalidUsername {
			http.Error(w, "El nombre de usuario solo puede contener caracteres, número y guiones", http.StatusUnprocessableEntity)
			log.Println("La nombre de usuario que se intentó registrar es inválido")
			return
		} else if err == security.ErrInvalidPassword {
			http.Error(w, "La contraseña debe contener al menos de 12 a 64 caracteres, una mayúscula, una minúscula, un número y un caracter especial, sin espacios", http.StatusUnprocessableEntity)
			log.Println("La contraseña que se intentó registrar es insegura")
			return
		} else if err == security.ErrInvalidEmail {
			http.Error(w, "El correo que se intenta registrar es inválido", http.StatusUnprocessableEntity)
			log.Println("El correo que se intentó registrar es inválido")
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
