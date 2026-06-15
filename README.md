# Forge

Forge is a Kubernetes-native mini deployment platform inspired by Render/Railway.

It allows users to create projects, deploy container images to Kubernetes using Helm, check pod status, view logs, delete deployments, view deployment history, and get launch instructions for Minikube services.

## Tech Stack

* Frontend: React + TypeScript + Vite
* Backend: Go
* Database: PostgreSQL
* Containerization: Docker / Docker Compose
* Orchestration: Kubernetes / Minikube
* Deployment Engine: Helm
* Local Registry Source: Docker Hub / GHCR images

## Features

* Create deployment projects
* Store projects in PostgreSQL
* Deploy apps to Kubernetes using Helm
* View Kubernetes pod status
* View pod logs
* Delete Helm deployments
* Track deployment history
* Launch deployed apps through Minikube service command
* Dockerized frontend, backend, and PostgreSQL setup

## Project Structure

```text
Forge/
├── Backend/
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
│
├── frontend/
│   ├── src/
│   ├── package.json
│   └── Dockerfile
│
├── charts/
│   └── app/
│       ├── Chart.yaml
│       ├── values.yaml
│       └── templates/
│
├── infra/
│   └── docker-compose.yml
│
└── README.md
```

## Local Development

Start PostgreSQL:

```bash
cd infra
docker compose up postgres
```

Start backend:

```bash
cd Backend
go run main.go
```

Start frontend:

```bash
cd frontend
npm run dev
```

Open frontend:

```text
http://localhost:5173
```

## Kubernetes Requirements

Make sure Minikube is running:

```bash
minikube start
kubectl get nodes
helm version
```

## Deployment Flow

```text
User creates project
↓
Backend stores project in PostgreSQL
↓
User clicks Deploy
↓
Backend runs Helm upgrade/install
↓
Helm deploys Kubernetes Deployment + Service
↓
User can check status, logs, history and launch command
```

## Example Project

```text
Project Name: test-nginx
Image Name: nginx:latest
```

## Important Note

The backend must be run locally for deployment features because it needs access to local `helm`, `kubectl`, `minikube`, and kubeconfig.

Docker Compose currently works for containerizing Forge services, but containerized backend deployment requires additional Kubernetes access setup.

## Future Roadmap

* Add backend retry logic for PostgreSQL startup
* Add image validation before deployment
* Add Ingress-based public URLs
* Deploy Forge itself to Kubernetes
* Add Kubernetes ServiceAccount and RBAC
* Add Terraform AWS infrastructure
* Deploy on AWS EKS
* Add GitHub Actions CI/CD
* Add GitHub repo-based deployment
* Add automatic image build and push to GHCR
