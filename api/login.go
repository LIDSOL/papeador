package api

import (
	"log"
	"net/http"

	"lidsol.org/papeador/security"
	"lidsol.org/papeador/store"
)

func (api *ApiContext) login(w http.ResponseWriter, r *http.Request) {
	var in store.User
	r.ParseForm()

	in.Email = r.FormValue("email")
	in.Username = r.FormValue("username")
	in.Password = r.FormValue("password")

	if (in.Username == "" && in.Email == "") || (in.Password == "") {
		http.Error(w, "Usuario y contraseña son requeridos", http.StatusBadRequest)
		log.Println("Campos requeridos vacíos")
		return
	}

	if err := api.Store.Login(r.Context(), &in); err != nil {
		if err == store.ErrNotFound {
			http.Error(w, "El usuario ingresado no existe", http.StatusUnauthorized)
			log.Println("El usuario ingresado no existe")
			return
		} else if err == security.ErrInvalidPassword {
			http.Error(w, "La contraseña ingresada es incorrecta", http.StatusUnauthorized)
			log.Println("La contraseña ingresada es incorrecta")
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	log.Println("SE PUDO INICIAR SESIÓN")
	http.SetCookie(w, &http.Cookie{
		Name:  "session",
		Value: in.JWT,
		Path:  "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:  "username",
		Value: in.Username,
		Path:  "/",
	})
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)

}
func (api *ApiContext) createLoginView(w http.ResponseWriter, r *http.Request) {
	type prueba struct {
		Title string
	}
	a := prueba{Title: "Iniciar sesión"}
	if err := templates.ExecuteTemplate(w, "login.html", &a); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error ejecutando template login.html:", err)
	}
}

