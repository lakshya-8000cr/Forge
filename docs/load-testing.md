Uncontrolled Stress Test 01:
57.1k requests, 42.3% failures, p95 4s, p99 6s.
Finding: system saturated under extreme load; controlled baseline testing required

# Load Testing

## Objective

The goal of this document is to understand how Forge and deployed applications behave under increasing load, identify performance bottlenecks, validate architectural decisions, and measure the impact of future optimizations.

Rather than focusing only on successful deployments, these tests intentionally push the system toward failure to observe how different components behave under stress.

---

# Test Environment

Infrastructure

- Kubernetes: Amazon EKS
- Ingress: NGINX Ingress Controller
- Load Balancer: AWS ELB
- Container Runtime: containerd
- Deployment Engine: Helm

Testing Tool

- k6

Application

- traefik/whoami

---

# Test 001

## Name

Uncontrolled Stress Test

## Purpose

Establish an initial performance baseline without any application tuning or infrastructure optimization.

The objective was not to achieve maximum throughput, but to intentionally overload the platform and observe how it fails.

---

## Configuration

Virtual Users

100000

Duration

60 Seconds

Target

/apps/whoami-demo

---

## Results

Total Requests

57,100

Successful Requests

57.7%

Failed Requests

42.3%

Average Response Time

1s

P95

4s

P99

6s

Maximum Response Time

30s

---

## Observations

The application remained operational but began rejecting a significant percentage of requests as concurrency increased.

Latency increased rapidly once request throughput exceeded the available processing capacity.

The Kubernetes API also became temporarily difficult to reach from kubectl, producing TLS handshake timeouts during the test.

No tuning or autoscaling had been configured before this experiment.

---

## Initial Findings

The current deployment is functional but not designed for extreme concurrent traffic.

The primary objective now is to determine which component becomes the bottleneck first:

- Application
- NGINX Ingress
- Kubernetes Networking
- Node Resources
- Database
- Forge Backend

Future experiments will isolate each component individually.

---

## Next Steps

- Install Metrics Server
- Record CPU and Memory utilization
- Configure Resource Requests and Limits
- Increase replica count
- Introduce Horizontal Pod Autoscaler
- Repeat benchmark
- Compare before and after results