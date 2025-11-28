package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"lidsol.org/papeador/security"
	"lidsol.org/papeador/store"
)

type ContestRequestContent struct {
	store.Contest
	Username string `json:"username"`
	JWT      string `json:"jwt"`
}

func (api *ApiContext) createContest(w http.ResponseWriter, r *http.Request) {
	var in store.Contest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	in.ContestName = r.FormValue("contest-name")
	in.StartDate = r.FormValue("start-date")
	in.EndDate = r.FormValue("end-date")

	cookieUsername, err := r.Cookie("username")
	username := cookieUsername.Value

	id, err := api.Store.GetUserID(r.Context(), username)

	// We shouldn't ever get here, handle it anyways
	if err != nil {
		http.Error(w, "This user does not exist", http.StatusNotFound)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(in)
}

func (api *ApiContext) createContestView(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "createContext.html", nil)
}

func (api *ApiContext) getContests(w http.ResponseWriter, r *http.Request) {
	type contestsInfo struct {
		contests []store.Contest
	}

	var info contestsInfo
	contests, err := api.Store.GetContests(r.Context())
	info.contests = contests

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		templates.ExecuteTemplate(w, "404.html", &info)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	templates.ExecuteTemplate(w, "contest.html", &info)
}

func (api *ApiContext) getContestByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, _ := strconv.Atoi(idStr)

	type contestInfo struct {
		store.Contest
		problems []store.Problem
	}

	c, err := api.Store.GetContestByID(r.Context(), id)
	info := contestInfo{Contest: c}

	w.Header().Set("Content-Type", "text/html")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		templates.ExecuteTemplate(w, "404.html", &info)
		return
	}

	problems, err := api.Store.GetContestProblems(r.Context(), int(info.ContestID))
	info.problems = problems

	w.WriteHeader(http.StatusOK)
	templates.ExecuteTemplate(w, "contest.html", &info)
}
