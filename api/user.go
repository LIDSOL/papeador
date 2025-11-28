package api

import (
	"fmt"
	"log"
	"net/http"

	"lidsol.org/papeador/security"
	"lidsol.org/papeador/store"
)

func (api *ApiContext) createUser(w http.ResponseWriter, r *http.Request) {
	var in store.User
	r.ParseForm()
	log.Println("CREATING USER")

	in.Email = r.FormValue("email")
	in.Username = r.FormValue("username")
	in.Password = r.FormValue("password")

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
		log.Println("Error", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name: "session",
		Value: in.JWT,
		Path: "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name: "username",
		Value: in.Username,
		Path: "/",
	})

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func (api *ApiContext) createUserView(w http.ResponseWriter, r *http.Request) {
	type prueba struct {
		Title string
	}
	a := prueba{Title: ""}
	templates.ExecuteTemplate(w, "createUser.html", &a)
}

func (api *ApiContext) loginUser(w http.ResponseWriter, r *http.Request) {
	var in store.User
	r.ParseForm()

	in.Username = r.FormValue("username")
	in.Password = r.FormValue("password")

	if err := api.Store.Login(r.Context(), &in); err != nil {
		if err == security.ErrInvalidCredentials {
			http.Error(w, "Datos inválidos", http.StatusForbidden)
			log.Println("Contraseña incorrecta")
			return
		} else if err == store.ErrNotFound {
			http.Error(w, "Datos inválidos", http.StatusNotFound)
			log.Println("Usuario invalido")
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name: "session",
		Value: in.JWT,
		Path: "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name: "username",
		Value: in.Username,
		Path: "/",
	})

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func (api *ApiContext) logoutUser(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name: "username",
		Value: "",
		MaxAge: 0,
	}
	http.SetCookie(w, &cookie)

	w.Header().Set("HX-Redirect", "/")
	// w.WriteHeader(http.StatusOK)
}

func (api *ApiContext) loginUserView(w http.ResponseWriter, r *http.Request) {
	type prueba struct {
		Title string
	}
	a := prueba{Title: ""}
	templates.ExecuteTemplate(w, "login.html", &a)
}

func (api *ApiContext) getUserByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("<h1>Not implemented: %v...</h1>", id)))
}
