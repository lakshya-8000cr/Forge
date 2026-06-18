# Forge

Forge is a Kubernetes-native mini deployment platform inspired by Render/Railway.

It allows users to create projects, deploy real public Docker images to Kubernetes using Helm, expose deployed apps through Nginx Ingress on AWS EKS, check pod status, view logs, delete deployments, and track deployment history.

## Tech Stack

- Frontend: React + TypeScript + Vite, Nginx
- Backend: Go
- Database: PostgreSQL
- Containerization: Docker / Docker Compose
- Orchestration: Kubernetes, AWS EKS
- Deployment Engine: Helm
- Cloud Infrastructure: AWS, Terraform
- Registry: Amazon ECR, Docker Hub / GHCR public images
- Networking: Nginx Ingress Controller, AWS ELB
- Storage: EBS CSI Driver

## Features

- Create deployment projects
- Store projects and deployment history in PostgreSQL
- Deploy real Docker images to Kubernetes using Helm
- Deploy apps into a dedicated `forge-apps` namespace
- Expose deployed apps through shared Ingress routes
- Generate public app URLs like `/apps/<project-name>`
- View Kubernetes pod status
- View pod logs
- Delete Helm deployments
- Track deployment history
- Dockerized frontend, backend, and PostgreSQL setup
- Production-style AWS EKS deployment using Terraform

```Project Structure
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
## Example Project

```text
Project Name: whoami-demo
Image Name: traefik/whoami:latest
```
if you get the Hostname ip etc then test is passed.

## Important Note

Forge uses Kubernetes ServiceAccount + RBAC to allow the backend pod to create deployments, services, pods, logs, namespaces, and ingresses.
User apps are deployed into the forge-apps namespace.
Public app routes are exposed through a shared AWS ELB + Nginx Ingress Controller.
Current app URLs use path-based routing: /apps/<project-name>.
Image validation is currently skipped for EKS compatibility and can be re-added later using registry APIs.

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
