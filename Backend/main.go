package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
	"sync"

	_ "github.com/lib/pq"
	_ "net/http/pprof"
)

//global variable 
var (
	inFlightMu          sync.Mutex
	inFlightDeployments = make(map[int]bool)
)

func markDeployment(id int) bool {
	inFlightMu.Lock()
	defer inFlightMu.Unlock()

	if inFlightDeployments[id] {
		return false
	}

	inFlightDeployments[id] = true
	return true
}

func clearDeployment(id int) {
	inFlightMu.Lock()
	defer inFlightMu.Unlock()

	delete(inFlightDeployments, id)
}

//$

var db *sql.DB

// 🚀 ASYNC JOB QUEUE CONFIGURATION
type DeployJob struct {
	ProjectID int
	Project   Project
}

var deployQueue chan DeployJob

type Deployment struct {
	ID        int    `json:"id"`
	ProjectID int    `json:"project_id"`
	ImageName string `json:"image_name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type Project struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ImageName string `json:"image_name"`
	Status    string `json:"status"`
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func connectDB() {
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "forge")
	password := getEnv("POSTGRES_PASSWORD", "forge123")
	dbname := getEnv("POSTGRES_DB", "forge")

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user,
		password,
		host,
		port,
		dbname,
	)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("DB open error:", err)
	}

	// after using the go profiling i realised that my connection pools were unstable , not rightly configured
	// so i just configured the database pooling 

	db.SetMaxIdleConns(25)                  // Warm authenticated connections
	db.SetMaxOpenConns(50)                  // Prevent unlimited connections
	db.SetConnMaxLifetime(5 * time.Minute)  // Recycle periodically
	db.SetConnMaxIdleTime(2 * time.Minute)  // Close stale idle connections

	err = db.Ping()
	if err != nil {
		log.Fatal("DB ping error:", err)
	}

	log.Println("Connected to PostgreSQL")
}

func logDBStats() {
	stats := db.Stats()
	log.Printf(
		"[DB Pool Status] Open=%d InUse=%d Idle=%d WaitCount=%d WaitDuration=%v",
		stats.OpenConnections,
		stats.InUse,
		stats.Idle,
		stats.WaitCount,
		stats.WaitDuration,
	)
}

func createTables() {
	projectsQuery := `
	CREATE TABLE IF NOT EXISTS projects (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		image_name TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT NOW()
	);`

	_, err := db.Exec(projectsQuery)
	if err != nil {
		log.Fatal("projects table creation error:", err)
	}

	deploymentsQuery := `
	CREATE TABLE IF NOT EXISTS deployments (
		id SERIAL PRIMARY KEY,
		project_id INT NOT NULL,
		image_name TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	);`

	_, err = db.Exec(deploymentsQuery)
	if err != nil {
		log.Fatal("deployments table creation error:", err)
	}

	log.Println("Projects table ready")
	log.Println("Deployments table ready")
}

// 🚀 GO CHANNEL WORKER POOL ENGINE
func startDeployWorkers(workerCount int, queueSize int) {
	deployQueue = make(chan DeployJob, queueSize)

	for i := 1; i <= workerCount; i++ {
		go func(workerID int) {
			log.Printf("[Worker Pool] Background Worker #%d Warm & Initialized", workerID)
			for job := range deployQueue {
				log.Printf("[Worker #%d] Processing deployment for project: %s", workerID, job.Project.Name)
				processAsyncDeployment(job)
			}
		}(i)
	}
}

func processAsyncDeployment(job DeployJob) {
	defer clearDeployment(job.ProjectID)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	err := runHelmDeployCtx(ctx, job.Project)
	if err != nil {
		failDeployment(job.ProjectID, job.Project, "helm deploy failed")
		return
	}

	_, _ = db.Exec(`UPDATE projects SET status = 'running' WHERE id = $1`, job.ProjectID)
	_, _ = db.Exec(
		`INSERT INTO deployments (project_id, image_name, status) VALUES ($1, $2, $3)`,
		job.Project.ID, job.Project.ImageName, "success",
	)

	log.Printf("[Worker Pool] Deployment SUCCESS for project: %s", job.Project.Name)
}


func failDeployment(id int, project Project, reason string) {
	log.Printf("[Worker Pool] Deployment FAILED for %s: %s", project.Name, reason)
	_, _ = db.Exec(`UPDATE projects SET status = 'failed' WHERE id = $1`, id)
	_, _ = db.Exec(
		`INSERT INTO deployments (project_id, image_name, status) VALUES ($1, $2, $3)`,
		project.ID, project.ImageName, "failed",
	)
}

func main() {
	connectDB()
	createTables()

	startDeployWorkers(3, 200)
	startRuntimeMonitor()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/projects", projectsHandler)
	mux.HandleFunc("/projects/", projectDetailHandler)

	go func() {  // this is where we have setup the profilling in go 
		//profilling ?=> can show how much resourcres are alloacted for the code parts , everything from varibale to func , so that we can optimize the algos
		log.Println("pprof server running on :6060")
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	log.Println("Forge backend running on :8080")
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           cors(mux), 
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func startRuntimeMonitor() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			log.Printf(
				"[Runtime] Goroutines=%d",
				runtime.NumGoroutine(),
			)
		}
	}()
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
		getProjectByID(w, r, id)
		return
	}

	if len(parts) == 2 && parts[1] == "deploy" && r.Method == http.MethodPost {
		deployProject(w, r, id)
		return
	}

	if len(parts) == 2 && parts[1] == "status" && r.Method == http.MethodGet {
		getProjectStatus(w, r, id)
		return
	}

	if len(parts) == 2 && parts[1] == "logs" && r.Method == http.MethodGet {
		getProjectLogs(w, r, id)
		return
	}

	if len(parts) == 2 && parts[1] == "deployments" && r.Method == http.MethodGet {
		getDeployments(w, r, id)
		return
	}

	if len(parts) == 2 && parts[1] == "url" && r.Method == http.MethodGet {
		getProjectURL(w, r, id)
		return
	}

	writeJSON(w, http.StatusNotFound, map[string]string{
		"error": "route not found",
	})
}

func getProjectURL(w http.ResponseWriter, r *http.Request, id int) {  // updated url
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var project Project

	err := db.QueryRowContext(ctx, `
		SELECT id, name, image_name, status
		FROM projects
		WHERE id = $1
	`, id).Scan(&project.ID, &project.Name, &project.ImageName, &project.Status)

	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "project not found",
		})
		return
	}

	baseURL := getEnv(
		"FORGE_PUBLIC_BASE_URL",
		"http://a6228a9f5a97041acb1f84aa6ada2478-774918050.eu-north-1.elb.amazonaws.com",
	)

	writeJSON(w, http.StatusOK, map[string]string{
		"url": baseURL + "/apps/" + project.Name,
	})
}

func getMinikubeServiceURL(serviceName string) (string, error) {
	ipCmd := exec.Command("minikube", "ip")
	ipOutput, err := ipCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", string(ipOutput))
	}

	nodePortCmd := exec.Command(
		"kubectl",
		"get",
		"svc",
		serviceName,
		"-o",
		"jsonpath={.spec.ports[0].nodePort}",
	)

	portOutput, err := nodePortCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", string(portOutput))
	}

	ip := strings.TrimSpace(string(ipOutput))
	port := strings.TrimSpace(string(portOutput))

	if ip == "" || port == "" {
		return "", fmt.Errorf("minikube ip or nodePort not found")
	}

	return fmt.Sprintf("http://%s:%s", ip, port), nil
}

func getDeployments(w http.ResponseWriter, r *http.Request, id int) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, `
		SELECT id, project_id, image_name, status, created_at
		FROM deployments
		WHERE project_id = $1
		ORDER BY created_at DESC
	`, id)

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}
	defer rows.Close()

	deployments := []Deployment{}
	for rows.Next() {
		var d Deployment
		err := rows.Scan(
			&d.ID,
			&d.ProjectID,
			&d.ImageName,
			&d.Status,
			&d.CreatedAt,
		)
		if err != nil {
			continue
		}
		deployments = append(deployments, d)
	}

	writeJSON(w, http.StatusOK, deployments)
}

func getProjectStatus(w http.ResponseWriter, r *http.Request, id int) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var project Project
	err := db.QueryRowContext(ctx, `
		SELECT id, name, image_name, status
		FROM projects
		WHERE id = $1
	`, id).Scan(
		&project.ID,
		&project.Name,
		&project.ImageName,
		&project.Status,
	)

	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "project not found",
		})
		return
	}

	status, err := getKubernetesStatusCtx(ctx, project.Name)
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
}

func getProjectLogs(w http.ResponseWriter, r *http.Request, id int) {  // updated get projcts
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var project Project
	err := db.QueryRowContext(ctx, `
		SELECT id, name, image_name, status
		FROM projects
		WHERE id = $1
	`, id).Scan(
		&project.ID,
		&project.Name,
		&project.ImageName,
		&project.Status,
	)

	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "project not found",
		})
		return
	}

	logs, err := getKubernetesLogsCtx(ctx, project.Name)
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
}

func getKubernetesLogsCtx(ctx context.Context, appName string) (string, error) {  // updated
	cmd := exec.CommandContext(
		ctx,
		"kubectl",
		"logs",
		"-n",
		"forge-apps",
		"-l",
		fmt.Sprintf("app=%s", appName),
		"--tail=100",
	)

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("kubectl logs timed out")
	}

	if err != nil {
		return "", fmt.Errorf("%s", string(output))
	}

	return string(output), nil
}

func getKubernetesStatusCtx(ctx context.Context, appName string) (string, error) {
	cmd := exec.CommandContext(
		ctx,
		"kubectl",
		"get",
		"pods",
		"-n",
		"forge-apps",
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

func getAppChartPath() string {  // this will resolve the p.th issue
	if path := os.Getenv("FORGE_APP_CHART_PATH"); path != "" {
		return path
	}

	candidates := []string{
		"./charts/app",
		"../charts/app",
	}

	for _, path := range candidates {
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			return path
		}
	}

	return "./charts/app"
}

func runHelmDeployCtx(ctx context.Context, project Project) error {
	imageParts := strings.Split(project.ImageName, ":")
	chartPath := getAppChartPath()
	repository := imageParts[0]
	tag := "latest"
	if len(imageParts) > 1 {
		tag = imageParts[1]
	}

	cmd := exec.CommandContext(ctx,
		"helm", "upgrade", "--install", project.Name, chartPath, // added the path function
		"--namespace", "forge-apps", "--create-namespace",
		"--set", fmt.Sprintf("appName=%s", project.Name),
		"--set", fmt.Sprintf("image.repository=%s", repository),
		"--set", fmt.Sprintf("image.tag=%s", tag),
		"--set", "service.type=ClusterIP",
		"--set", "ingress.enabled=true",
		"--set", "ingress.basePath=/apps",
	)
	
	output, err := cmd.CombinedOutput() // 1. Catch the 'err' here!
	log.Println(string(output))
	
	return err // 2. Return the raw error! If err == nil, it means exit code was 0.
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

func getProjectByID(w http.ResponseWriter, r *http.Request, id int) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var project Project
	err := db.QueryRowContext(ctx, `
		SELECT id, name, image_name, status FROM projects WHERE id = $1
	`, id).Scan(&project.ID, &project.Name, &project.ImageName, &project.Status)

	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "project not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, project)
}

func deployProject(w http.ResponseWriter, r *http.Request, id int) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var project Project
	err := db.QueryRowContext(ctx, `
		SELECT id, name, image_name, status FROM projects WHERE id = $1
	`, id).Scan(&project.ID, &project.Name, &project.ImageName, &project.Status)

	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "project not found",
		})
		return
	}

	// 1. Guard: Check if deployment is already in progress
	if !markDeployment(id) {
		writeJSON(w, http.StatusConflict, map[string]string{
			"error": "deployment already in progress for this project",
		})
		return
	}

	_, err = db.ExecContext(ctx, `UPDATE projects SET status = 'deploying' WHERE id = $1`, id)
	if err != nil {
		// 2. Cleanup: Clear tracking flag if state update fails
		clearDeployment(id)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	select {
	case deployQueue <- DeployJob{ProjectID: id, Project: project}:
		// 3. Updated Success Response Message
		writeJSON(w, http.StatusAccepted, map[string]any{
			"message": "deployment queued successfully",
			"status":  "deploying",
		})
	default:
		// 4. Cleanup: Clear tracking flag if channel buffer is full
		clearDeployment(id)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "deployment engine processing capacity saturated, retry later",
		})
	}
}


func ensureForgeNamespaceCtx(ctx context.Context) error {
	cmd := exec.CommandContext(ctx,
		"kubectl",
		"create",
		"namespace",
		"forge-apps",
	)

	err := cmd.Run()
	if err == nil {
		return nil
	}

	cmd = exec.CommandContext(ctx,
		"kubectl",
		"get",
		"namespace",
		"forge-apps",
	)

	return cmd.Run()
}

func writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func projectsHandler(w http.ResponseWriter, r *http.Request) {  // updated version
	switch r.Method {
	case http.MethodGet:
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		rows, err := db.QueryContext(
			ctx,
			`SELECT id, name, image_name, status
			 FROM projects
			 ORDER BY id DESC`,
		)
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
			if err := rows.Scan(
				&p.ID,
				&p.Name,
				&p.ImageName,
				&p.Status,
			); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{
					"error": err.Error(),
				})
				return
			}
			list = append(list, p)
		}

		if err := rows.Err(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, list)

	case http.MethodPost:
		createProject(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
	}
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

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"app":    "forge-backend",
	})
}