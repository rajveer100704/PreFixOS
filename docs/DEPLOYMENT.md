# Deployment Guide - PrefixOS

This guide covers local Docker Compose deployment, Kubernetes StatefulSet deployment, and configuration tuning for PrefixOS.

---

## 1. Local Docker Cluster Deployment

```bash
# Start a 3-node local PrefixOS cluster with Prometheus and Grafana
docker-compose up -d

# Check cluster status
docker-compose ps
```

---

## 2. Kubernetes StatefulSet Deployment

Deployment manifests are located under `deployments/kubernetes/`:

```bash
# Apply ConfigMap, Service, and StatefulSet
kubectl apply -f deployments/kubernetes/configmap.yaml
kubectl apply -f deployments/kubernetes/service.yaml
kubectl apply -f deployments/kubernetes/statefulset.yaml

# Verify pod rollout
kubectl rollout status statefulset/prefixos
```
