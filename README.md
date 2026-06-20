# Forge 

**Forge** is a Kubernetes-native mini Platform-as-a-Service (PaaS) inspired by Render, Railway and modern internal developer platforms.

Forge allows users to deploy **real Docker images** to AWS EKS clusters using Helm, manage deployments from a dashboard, and expose applications through a shared Nginx Ingress with public URLs.

The goal of Forge is to provide a lightweight self-hosted platform where developers can deploy, monitor and manage applications without directly interacting with Kubernetes.

---

## Architecture

```text
Internet
↓
AWS ELB
↓
Nginx Ingress Controller
↓
Forge Frontend
↓
Forge Backend
↓
PostgreSQL
↓
Helm
↓
EKS Kubernetes API
↓
forge-apps Namespace
↓
User Applications
```

---

## Features

### Platform Features

* Create deployment projects
* Store project metadata in PostgreSQL
* Deploy real Docker images to AWS EKS
* Automatic Helm-based deployments
* Dedicated `forge-apps` namespace
* Generate public application URLs
* Shared AWS ELB + Nginx Ingress routing
* View deployment history
* View application logs
* Delete deployments
* RBAC-secured Kubernetes access

---

## Tech Stack

### Frontend

* React
* TypeScript
* Vite
* Nginx

### Backend

* Go

### Database

* PostgreSQL

### Cloud Infrastructure

* AWS EKS
* Amazon ECR
* AWS ELB
* Amazon VPC

### DevOps

* Docker
* Kubernetes
* Helm
* Terraform
* GitHub Actions

### Networking

* Nginx Ingress Controller

---

## Deployment Flow

```text
User
↓
Create Project
↓
Deploy
↓
Forge Backend
↓
Helm
↓
EKS Kubernetes API
↓
Deployment + Service + Ingress
↓
Public Application URL
```

---

## Verified Public Image Deployments

### Example 1

```text
Project Name: nginx-prod

Docker Image:
nginx:latest
```

### Example 2

```text
Project Name: whoami-demo

Docker Image:
traefik/whoami:latest
```

Example Public URL:

```text
http://<INGRESS_ELB>/apps/whoami-demo
```

---

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
│   ├── forge/
│   └── app/
│
├── k8s/
│
├── infra/
│   └── aws/
│
├── .github/
│   └── workflows/
│
└── README.md
```

---

## Local Development

### Start PostgreSQL

```bash
cd infra

docker compose up postgres
```

### Start Backend

```bash
cd Backend

go run main.go
```

### Start Frontend

```bash
cd frontend

npm run dev
```

Frontend URL:

```text
http://localhost:5173
```

---

## AWS Deployment

Configure Kubernetes:

```bash
aws eks update-kubeconfig \
--region eu-north-1 \
--name forge-eks
```

Verify:

```bash
kubectl get nodes

helm version
```

---

## CI/CD

Forge uses GitHub Actions for:

* Backend build verification
* Frontend build verification
* Docker image builds
* Amazon ECR image pushes
* Automatic EKS deployment updates

---

## Kubernetes Resources Used

* Deployments
* Services
* Namespaces
* Ingress
* Service Accounts
* RBAC
* Cluster Roles
* Cluster Role Bindings
* Persistent Volume Claims

---

## Lessons Learned

During development, several production-level issues were debugged and resolved:

* PVC binding failures
* EBS CSI provisioning issues
* ImagePullBackOff errors
* CrashLoopBackOff errors
* AWS ELB configuration issues
* Ingress routing problems
* ServiceAccount permission errors
* RBAC misconfigurations
* Kubernetes scheduling limitations
* Node capacity constraints

---

## Future Roadmap

* Monitoring with Prometheus
* Grafana dashboards
* Loki log aggregation
* HTTPS with cert-manager
* Custom domains
* GitHub repository deployments
* Automatic image builds
* Registry validation
* Authentication & multi-user support

---

## License

MIT License

---


