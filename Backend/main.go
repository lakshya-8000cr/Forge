package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/projects", projectsHandler)
	mux.HandleFunc("/projects/", projectDetailHandler)

	log.Println("Forge backend running on :8080")
	http.ListenAndServe(":8080", mux)
}

func projectDetailHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/projects/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 0 || parts[0] == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "project id required",
		})
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid project id",
		})
		return
	}

	if len(parts) == 1 && r.Method == http.MethodGet {
		getProjectByID(w, id)
		return
	}

	if len(parts) == 2 && parts[1] == "deploy" && r.Method == http.MethodPost {
		deployProject(w, id)
		return
	}

	writeJSON(w, http.StatusNotFound, map[string]string{
		"error": "route not found",
	})
}

func getProjectByID(w http.ResponseWriter, id int) {
	for _, project := range projects {
		if project.ID == id {
			writeJSON(w, http.StatusOK, project)
			return
		}
	}

	writeJSON(w, http.StatusNotFound, map[string]string{
		"error": "project not found",
	})
}

func deployProject(w http.ResponseWriter, id int) {
	for i := range projects {
		if projects[i].ID == id {
			projects[i].Status = "running"

			writeJSON(w, http.StatusOK, map[string]any{
				"message": "deployment simulated",
				"project": projects[i],
			})
			return
		}
	}

	writeJSON(w, http.StatusNotFound, map[string]string{
		"error": "project not found",
	})
}

func writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}


func projectsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("method:", r.Method)
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, projects)
		return
	}

	if r.Method == http.MethodPost {
		createProject(w, r)
		return
	}

	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
		"error": "method not allowed",
	})
}

func createProject(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name      string `json:"name"`
		ImageName string `json:"image_name"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json",
		})
		return
	}

	project := Project{
		ID:        nextID,
		Name:      req.Name,
		ImageName: req.ImageName,
		Status:    "pending",
	}

	nextID++
	projects = append(projects, project)

	writeJSON(w, http.StatusCreated, project)
}

type Project struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ImageName string `json:"image_name"`
	Status    string `json:"status"`
}


var projects = []Project{}
var nextID = 1


func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"app":    "forge-backend",
	})
}