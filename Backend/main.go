package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"os/exec"
    "fmt"
    "database/sql"

_ "github.com/lib/pq"
)

var db *sql.DB

func connectDB() {
	connStr := "postgres://forge:forge123@localhost:5432/forge?sslmode=disable"

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("DB open error:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("DB ping error:", err)
	}

	log.Println("Connected to PostgreSQL")
}

func createTables() {
	query := `
	CREATE TABLE IF NOT EXISTS projects (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		image_name TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT NOW()
	);
	`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("table creation error:", err)
	}

	log.Println("Projects table ready")
}

func main() {
	connectDB()
    createTables()
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/projects", projectsHandler)
	mux.HandleFunc("/projects/", projectDetailHandler)

	log.Println("Forge backend running on :8080")
	http.ListenAndServe(":8080", cors(mux))
	
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}


func deleteProjectDeployment(w http.ResponseWriter, id int) {
	for i := range projects {
		if projects[i].ID == id {
			err := deleteHelmRelease(projects[i].Name)
			if err != nil {
				projects[i].Status = "delete_failed"
				writeJSON(w, http.StatusInternalServerError, map[string]any{
					"error":   "delete failed",
					"details": err.Error(),
				})
				return
			}

			projects[i].Status = "deleted"

			writeJSON(w, http.StatusOK, map[string]any{
				"message": "deployment deleted",
				"project": projects[i],
			})
			return
		}
	}

	writeJSON(w, http.StatusNotFound, map[string]string{
		"error": "project not found",
	})
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

	if len(parts) == 2 && parts[1] == "status" && r.Method == http.MethodGet {
	getProjectStatus(w, id)
	return
    }

	if len(parts) == 2 && parts[1] == "delete" && r.Method == http.MethodDelete {
	deleteProjectDeployment(w, id)
	return
    }

	if len(parts) == 2 && parts[1] == "logs" && r.Method == http.MethodGet {
	getProjectLogs(w, id)
	return
    }
	
	writeJSON(w, http.StatusNotFound, map[string]string{
		"error": "route not found",
	})
}

func getProjectStatus(w http.ResponseWriter, id int) {
	for _, project := range projects {
		if project.ID == id {
			status, err := getKubernetesStatus(project.Name)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{
					"error":   "failed to get kubernetes status",
					"details": err.Error(),
				})
				return
			}

			writeJSON(w, http.StatusOK, map[string]string{
				"project": project.Name,
				"status":  status,
			})
			return
		}
	}

	writeJSON(w, http.StatusNotFound, map[string]string{
		"error": "project not found",
	})
}


func getProjectLogs(w http.ResponseWriter, id int) {
	for _, project := range projects {
		if project.ID == id {
			logs, err := getKubernetesLogs(project.Name)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{
					"error":   "failed to get logs",
					"details": err.Error(),
				})
				return
			}

			writeJSON(w, http.StatusOK, map[string]string{
				"project": project.Name,
				"logs":    logs,
			})
			return
		}
	}

	writeJSON(w, http.StatusNotFound, map[string]string{
		"error": "project not found",
	})
}


func getKubernetesLogs(appName string) (string, error) {
	cmd := exec.Command(
		"kubectl",
		"logs",
		"-l",
		fmt.Sprintf("app=%s", appName),
		"--tail=100",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", string(output))
	}

	return string(output), nil
}


func getKubernetesStatus(appName string) (string, error) {
	cmd := exec.Command(
		"kubectl",
		"get",
		"pods",
		"-l",
		fmt.Sprintf("app=%s", appName),
		"-o",
		"jsonpath={.items[0].status.phase}",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", string(output))
	}

	status := strings.TrimSpace(string(output))
	if status == "" {
		return "NotFound", nil
	}

	return status, nil
}


func runHelmDeploy(project Project) error {
	imageParts := strings.Split(project.ImageName, ":")

	repository := imageParts[0]
	tag := "latest"

	if len(imageParts) > 1 {
		tag = imageParts[1]
	}

	cmd := exec.Command(
		"helm",
		"upgrade",
		"--install",
		project.Name,
		"../charts/app",
		"--set", fmt.Sprintf("appName=%s", project.Name),
		"--set", fmt.Sprintf("image.repository=%s", repository),
		"--set", fmt.Sprintf("image.tag=%s", tag),
	)

	output, err := cmd.CombinedOutput()
	log.Println(string(output))

	return err
}


func deleteHelmRelease(appName string) error {
	cmd := exec.Command(
		"helm",
		"uninstall",
		appName,
	)

	output, err := cmd.CombinedOutput()
	log.Println(string(output))

	return err
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
			projects[i].Status = "deploying"

			err := runHelmDeploy(projects[i])
			if err != nil {
				projects[i].Status = "failed"
				writeJSON(w, http.StatusInternalServerError, map[string]any{
					"error":   "helm deploy failed",
					"details": err.Error(),
				})
				return
			}

			projects[i].Status = "running"

			writeJSON(w, http.StatusOK, map[string]any{
				"message": "deployment successful",
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
	if r.Method == http.MethodGet {
		rows, err := db.Query(`SELECT id, name, image_name, status FROM projects ORDER BY id DESC`)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
			return
		}
		defer rows.Close()

		list := []Project{}

		for rows.Next() {
			var p Project
			err := rows.Scan(&p.ID, &p.Name, &p.ImageName, &p.Status)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{
					"error": err.Error(),
				})
				return
			}
			list = append(list, p)
		}

		writeJSON(w, http.StatusOK, list)
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

	var project Project

	err = db.QueryRow(
		`INSERT INTO projects (name, image_name, status)
		 VALUES ($1, $2, 'pending')
		 RETURNING id, name, image_name, status`,
		req.Name,
		req.ImageName,
	).Scan(&project.ID, &project.Name, &project.ImageName, &project.Status)

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

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