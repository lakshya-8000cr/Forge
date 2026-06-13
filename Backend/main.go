package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/projects", projectsHandler)

	log.Println("Forge backend running on :8080")
	http.ListenAndServe(":8080", mux)
}


func writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}


func projectsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	writeJSON(w, http.StatusOK, projects)
}


type Project struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ImageName string `json:"image_name"`
	Status    string `json:"status"`
}


var projects = []Project{}


func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"app":    "forge-backend",
	})
}