# Forge

> A lightweight, self-hosted Platform-as-a-Service (PaaS) inspired by Render and Railway that allows developers to deploy real Docker images to Kubernetes without directly interacting with cluster infrastructure.

> ⚠️ The project is currently under active development.

---

# The Problem

Deploying applications to Kubernetes usually requires developers to manually configure deployments, services, ingresses, namespaces, RBAC permissions and cloud infrastructure.

For small teams and personal projects, this introduces significant operational overhead.

Most developers don't want to SSH into servers, write Kubernetes manifests or manage networking just to deploy an application.

They simply want:

```text
Docker Image
↓

Deploy

↓

Get a Public URL
```

without worrying about the underlying infrastructure.

---

# How Forge Solves It

Forge provides a lightweight self-hosted alternative to modern deployment platforms.

Users create a project from a dashboard, provide a Docker image and Forge automatically provisions the necessary Kubernetes resources.

Forge handles:

* Deployment creation
* Service provisioning
* Ingress routing
* Public URL generation
* Namespace management
* Deployment history
* Log viewing
* RBAC-secured cluster access

All deployments are orchestrated through Helm and run inside AWS EKS.

---

# Features

### Platform Features

* Create deployment projects
* Deploy real Docker images
* Automatic Helm-based deployments
* Dedicated `forge-apps` namespace
* Public application URLs
* Shared Nginx Ingress routing
* View deployment history
* View application logs
* Delete deployments
* RBAC-secured Kubernetes access

---

# Architecture

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

AWS EKS Kubernetes API

↓

forge-apps Namespace

↓

User Applications
```

---

# How It Works

```text
User
↓

Creates Project

↓

Clicks Deploy

↓

Forge Backend

↓

Helm

↓

Kubernetes API

↓

Deployment

↓

Service

↓

Ingress

↓

Public URL Generated
```

The Forge backend acts as a mini platform engineer by communicating with the Kubernetes API and provisioning infrastructure automatically.

---

# Example Deployment

Input:

```text
Project Name:
whoami-demo

Docker Image:
traefik/whoami:latest
```

Output:

```text
http://<INGRESS_ELB>/apps/whoami-demo
```

The application becomes publicly accessible without writing any Kubernetes manifests.

---

# Tech Stack

### Frontend

* React
* TypeScript
* Vite

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

# Project Structure

```text
Forge/

├── Backend/
│
├── frontend/
│
├── charts/
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

# Local Development

Start PostgreSQL

```bash
cd infra

docker compose up postgres
```

Start Backend

```bash
cd Backend

go run main.go
```

Start Frontend

```bash
cd frontend

npm run dev
```

Open:

```text
http://localhost:5173
```

---

# AWS Setup

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

# CI/CD

Forge uses GitHub Actions to:

* Verify backend builds
* Verify frontend builds
* Build Docker images
* Push images to Amazon ECR
* Update EKS deployments

---

# Kubernetes Resources Used

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

# Production Debugging Challenges Solved

During development, several production-level issues were debugged and resolved:

* PVC binding failures
* EBS CSI provisioning issues
* ImagePullBackOff errors
* CrashLoopBackOff errors
* AWS ELB configuration issues
* Ingress routing issues
* ServiceAccount permission errors
* RBAC misconfigurations
* Kubernetes scheduling limitations
* Node capacity constraints

---

# Future Roadmap

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

# License

MIT License
---


