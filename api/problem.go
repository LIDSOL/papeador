package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"lidsol.org/papeador/store"
)

func getRequestFileContents(r *http.Request, formField string) ([]byte, error) {
	// Retrieve the file from the form data
	file, _, err := r.FormFile(formField)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read file contents into []byte
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}

func getFileGroup(r *http.Request, formField string) ([][]byte, error) {
	files := r.MultipartForm.File[formField]

	output := make([][]byte, 0)
	for _, fileHeader := range files {
		file, _ := fileHeader.Open()
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}

		output = append(output, content)
	}

	return output, nil
}

func (api *ApiContext) createProblem(w http.ResponseWriter, r *http.Request) {
	var in store.Problem

	contestIDStr := r.PathValue("id")
	log.Println("ID", contestIDStr)
	contestID, err := strconv.Atoi(contestIDStr)
	if err != nil {
		http.Error(w, "Error en ruta", http.StatusBadRequest)
		log.Println("Error", err)
		return
	}
	contestID64 := int64(contestID)

	in.ContestID = &contestID64
	in.ProblemName = r.FormValue("problem-name")

	timelimitStr := r.FormValue("time-limit")
	timelimit, err := strconv.Atoi(timelimitStr)
	if err != nil {
		http.Error(w, "Error en ruta", http.StatusBadRequest)
		log.Println("Error", err)
		return
	}

	err = r.ParseMultipartForm(8 << 20)
	if err != nil {
		http.Error(w, "Los archivos deben ser, como máximo, de 8 MiB", http.StatusBadRequest)
		log.Println("Error", err)
		return
	}

	description, err := getRequestFileContents(r, "description")
	if err != nil {
		http.Error(w, "Error al leer el enunciado", http.StatusInternalServerError)
		log.Println("Error", err)
		return
	}
	in.Description = description

	inputs, err := getFileGroup(r, "inputs")
	if err != nil {
		http.Error(w, "Error al leer los archivos de entrada", http.StatusInternalServerError)
		log.Println("Error", err)
		return
	}

	outputs, err := getFileGroup(r, "outputs")
	if err != nil {
		http.Error(w, "Error al leer los archivos de salida", http.StatusInternalServerError)
		log.Println("Error", err)
		return
	}

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

	in.CreatorID = int64(id)

	log.Println("creandoproblema", contestIDStr)
	if err := api.Store.CreateProblem(r.Context(), &in); err != nil {
		if err == store.ErrNotFound {
			http.Error(w, "El concurso no existe", http.StatusConflict)
			log.Println("El concurso no existe")
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	for k, _ := range inputs {
		var t store.TestCase
		t.GivenInput = inputs[k]
		t.ExpectedOut = outputs[k]
		t.NumTestCase = int64(k)
		t.TimeLimit = int64(timelimit)
		t.ProblemID = in.ProblemID

		if err = api.Store.CreateTestCase(r.Context(), &t); err != nil {
			http.Error(w, "No se pudo crear el caso de prueba", http.StatusInternalServerError)
			log.Println("error", err)
			return
		}
	}

	path := fmt.Sprintf("/contests/%v/problems/%v", *in.ContestID, in.ProblemID)
	w.Header().Set("HX-Redirect", path)
	w.WriteHeader(http.StatusOK)
}

func (api *ApiContext) createProblemView(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	type probleminfo struct {
		ID string
	}

	info := probleminfo{ID: idStr}
	templates.ExecuteTemplate(w, "createProblem.html", &info)
}

func (api *ApiContext) getProblemByID(w http.ResponseWriter, r *http.Request) {
	contestIDStr := r.PathValue("contestID")
	contestID, err := strconv.Atoi(contestIDStr)
	if err != nil {
		http.Error(w, "Error en ruta", http.StatusBadRequest)
		log.Println("Error", err)
		return
	}

	problemIDStr := r.PathValue("problemID")
	problemID, err := strconv.Atoi(problemIDStr)
	if err != nil {
		http.Error(w, "Error en ruta", http.StatusBadRequest)
		log.Println("Error", err)
		return
	}

	problem, err := api.Store.GetProblemByIDs(r.Context(), contestID, problemID)
	if err != nil {
		http.Error(w, "Error al buscar problema", http.StatusBadRequest)
		log.Println("Error", err)
		return
	}
	problem.SampleInputStr = string(problem.SampleInput)
	problem.SampleOutStr = string(problem.SampleOut)

	type pageInfo struct {
		store.Problem
		ContestID int
		Username string
	}

	info := pageInfo{Problem: *problem, ContestID: contestID}


	cookieUsername, err := r.Cookie("username")
	if err == nil {
		username := cookieUsername.Value
		info.Username = username
	}

	templates.ExecuteTemplate(w, "problem.html", &info)
}

func (api *ApiContext) getProblemStatementByID(w http.ResponseWriter, r *http.Request) {
	contestIDStr := r.PathValue("contestID")
	contestID, err := strconv.Atoi(contestIDStr)
	if err != nil {
		http.Error(w, "Error en ruta", http.StatusBadRequest)
		log.Println("Error", err)
		return
	}

	problemIDStr := r.PathValue("problemID")
	problemID, err := strconv.Atoi(problemIDStr)
	if err != nil {
		http.Error(w, "Error en ruta", http.StatusBadRequest)
		log.Println("Error", err)
		return
	}

	problem, err := api.Store.GetProblemByIDs(r.Context(), contestID, problemID)
	if err != nil {
		http.Error(w, "Error al buscar problema", http.StatusBadRequest)
		log.Println("Error", err)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.WriteHeader(http.StatusOK)
	w.Write(problem.Description)
}
