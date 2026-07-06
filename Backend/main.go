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
	"os"
	_ "net/http/pprof"
	"time"
	"runtime"
	"context"

_ "github.com/lib/pq"
)

var db *sql.DB

type Deployment struct {
	ID        int    `json:"id"`
	ProjectID int    `json:"project_id"`
	ImageName string `json:"image_name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
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
		"[DB Pool] Open=%d InUse=%d Idle=%d WaitCount=%d WaitDuration=%v MaxIdleClosed=%d MaxLifetimeClosed=%d",
		stats.OpenConnections,
		stats.InUse,
		stats.Idle,
		stats.WaitCount,
		stats.WaitDuration,
		stats.MaxIdleClosed,
		stats.MaxLifetimeClosed,
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
	);
	`

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
	);
	`

	_, err = db.Exec(deploymentsQuery)
	if err != nil {
		log.Fatal("deployments table creation error:", err)
	}

	log.Println("Projects table ready")
	log.Println("Deployments table ready")
}

func main() {
	connectDB()
    createTables()
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/projects", projectsHandler)
	mux.HandleFunc("/projects/", projectDetailHandler)
	go func() {  // this is where we have setup the profilling in go 
		//profilling ?=> can show how much resourcres are alloacted for the code parts , everything from varibale to func , so that we can optimize the algos
		//problem was-> we thought that we are not creating the connection pool , but in actually we were creating but error was coming bcz , every second new connection pool or connections werea getting creatted , so we limit thata thing 
		//next thing was thread , 1023 thread were gettig created before 1 thtread == 1 mb , that was taking a lot of space 
    log.Println("pprof server running on :6060")
    log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

	log.Println("Forge backend running on :8080")
	srv := &http.Server{
	Addr:              ":8080",
	Handler:           mux,
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

	if len(parts) == 2 && parts[1] == "logs" && r.Method == http.MethodGet {
	getProjectLogs(w, r,id)
	return
    }
	

    if len(parts) == 2 && parts[1] == "deployments" && r.Method == http.MethodGet {
	getDeployments(w, id)
	return
    }


if len(parts) == 2 && parts[1] == "url" && r.Method == http.MethodGet {
	getProjectURL(w,r, id)
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
		return "", fmt.Errorf("%s",string(ipOutput))
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
		return "", fmt.Errorf("%s",string(portOutput))
	}

	ip := strings.TrimSpace(string(ipOutput))
	port := strings.TrimSpace(string(portOutput))

	if ip == "" || port == "" {
		return "", fmt.Errorf("minikube ip or nodePort not found")
	}

	return fmt.Sprintf("http://%s:%s", ip, port), nil
}

func getDeployments(w http.ResponseWriter, id int) {

	rows, err := db.Query(
		`SELECT id, project_id, image_name, status, created_at
		 FROM deployments
		 WHERE project_id = $1
		 ORDER BY created_at DESC`,
		id,
	)

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


func getProjectStatus(w http.ResponseWriter, id int) {

	var project Project

	err := db.QueryRow(
		`SELECT id, name, image_name, status
		 FROM projects
		 WHERE id = $1`,
		id,
	).Scan(
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
}


func getProjectLogs(w http.ResponseWriter, r *http.Request, id int) {  // updated get projcts
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
}

func getKubernetesLogs(appName string) (string, error) {  // updated
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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
	"./charts/app",
	"--namespace",
	"forge-apps",
	"--create-namespace",
	"--set", fmt.Sprintf("appName=%s", project.Name),
	"--set", fmt.Sprintf("image.repository=%s", repository),
	"--set", fmt.Sprintf("image.tag=%s", tag),
	"--set", "service.type=ClusterIP",
	"--set", "ingress.enabled=true",
	"--set", "ingress.basePath=/apps",
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
	var project Project

	err := db.QueryRow(
		`SELECT id, name, image_name, status FROM projects WHERE id = $1`,
		id,
	).Scan(&project.ID, &project.Name, &project.ImageName, &project.Status)

	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "project not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, project)
}

func deployProject(w http.ResponseWriter, id int) {
	var project Project

	err := db.QueryRow(
		`SELECT id, name, image_name, status FROM projects WHERE id = $1`,
		id,
	).Scan(&project.ID, &project.Name, &project.ImageName, &project.Status)

	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "project not found",
		})
		return
	}

	_, err = db.Exec(`UPDATE projects SET status = 'deploying' WHERE id = $1`, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// err = validateImage(project.ImageName)
     
	if err != nil {
	db.Exec(`UPDATE projects SET status = 'failed' WHERE id = $1`, id)

	db.Exec(
		`INSERT INTO deployments (project_id, image_name, status)
		 VALUES ($1, $2, $3)`,
		project.ID,
		project.ImageName,
		"failed",
	)

	//writeJSON(w, http.StatusBadRequest, map[string]any{
	//	"error":   "image validation failed",
	//	"details": err.Error(),
	//})
	//return
    }

	err = ensureForgeNamespace()

   if err != nil {

	writeJSON(
		w,
		http.StatusInternalServerError,

		map[string]string{

			"error": "failed to create forge namespace",
		},
	)

	return
   }

	err = runHelmDeploy(project)

	if err != nil {
		db.Exec(`UPDATE projects SET status = 'failed' WHERE id = $1`, id)

		db.Exec(
			`INSERT INTO deployments (project_id, image_name, status)
			 VALUES ($1, $2, $3)`,
			project.ID,
			project.ImageName,
			"failed",
		)

		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error":   "helm deploy failed",
			"details": err.Error(),
		})
		return
	}

	_, err = db.Exec(`UPDATE projects SET status = 'running' WHERE id = $1`, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	_, err = db.Exec(
		`INSERT INTO deployments (project_id, image_name, status)
		 VALUES ($1, $2, $3)`,
		project.ID,
		project.ImageName,
		"success",
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	project.Status = "running"

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "deployment successful",
		"project": project,
	})
}


func ensureForgeNamespace() error {

	cmd := exec.Command(
		"kubectl",
		"create",
		"namespace",
		"forge-apps",
	)

	err := cmd.Run()

	if err == nil {
		return nil
	}

	cmd = exec.Command(
		"kubectl",
		"get",
		"namespace",
		"forge-apps",
	)

	return cmd.Run()
}


/*func validateImage(imageName string) error {
	cmd := exec.Command("docker", "manifest", "inspect", imageName)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("invalid or inaccessible image: %s", strings.TrimSpace(string(output)))
	}

	return nil
}*/


func writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}


func projectsHandler(w http.ResponseWriter, r *http.Request) {  // updated version

	switch r.Method {

	case http.MethodGet:

		// Create a timeout context for the database query
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

		// Check if iteration ended because of an error
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