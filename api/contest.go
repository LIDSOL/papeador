package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"lidsol.org/papeador/store"
)

type ContestRequestContent struct {
	store.Contest
	Username string `json:"username"`
	JWT      string `json:"jwt"`
}

func (api *ApiContext) createContest(w http.ResponseWriter, r *http.Request) {
	var in store.Contest

	in.ContestName = r.FormValue("contest-name")
	in.StartDate = r.FormValue("start-date")
	in.EndDate = r.FormValue("end-date")

	cookieUsername, err := r.Cookie("username")
	username := cookieUsername.Value

	// We shouldn't ever get here, handle it anyways
	if err != nil {
		log.Println("Sin cookie")
		http.Error(w, "No hay sesión iniciada", http.StatusNotFound)
		return
	}

	id, err := api.Store.GetUserID(r.Context(), username)

	// We shouldn't ever get here, handle it anyways
	if err != nil {
		log.Println("Usuario no existe")
		http.Error(w, "Este usuario no existe", http.StatusNotFound)
		return
	}

	in.OrganizerID = int64(id)

	if err = api.Store.CreateContest(r.Context(), &in); err != nil {
		if err == store.ErrAlreadyExists {
			http.Error(w, "El nombre del contest ya esta registrado", http.StatusConflict)
			log.Println("El nombre del contest ya esta registrado")
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	path := fmt.Sprintf("/contests/%v", in.ContestID)
	w.Header().Set("HX-Redirect", path)
	w.WriteHeader(http.StatusOK)
}

func (api *ApiContext) createContestView(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "createContest.html", nil)
}

func (api *ApiContext) getContests(w http.ResponseWriter, r *http.Request) {
	type contestsInfo struct {
		Contests []store.Contest
	}

	var info contestsInfo
	contests, err := api.Store.GetContests(r.Context())
	info.Contests = contests

	if err != nil {
		log.Println("ERROR", err)
		w.WriteHeader(http.StatusNotFound)
		templates.ExecuteTemplate(w, "404.html", &info)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	templates.ExecuteTemplate(w, "contests.html", &info)
}

func (api *ApiContext) getContestByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, _ := strconv.Atoi(idStr)

	type contestInfo struct {
		store.Contest
		Problems []store.Problem
	}

	c, err := api.Store.GetContestByID(r.Context(), id)
	info := contestInfo{Contest: c}

	w.Header().Set("Content-Type", "text/html")
	if err != nil {
		log.Println("ERRROR", err)
		w.WriteHeader(http.StatusNotFound)
		templates.ExecuteTemplate(w, "404.html", &info)
		return
	}

	problems, err := api.Store.GetContestProblems(r.Context(), int(info.ContestID))
	info.Problems = problems
	log.Println("info", info)

	w.WriteHeader(http.StatusOK)
	templates.ExecuteTemplate(w, "contest.html", &info)
}
