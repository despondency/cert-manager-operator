# 🛠️ cert-manager-operator

> A Kubernetes Operator built with [Kubebuilder](https://book.kubebuilder.io) in Go for managing custom resources.

---

## 📦 Overview

**cert-manager-operator** is a Kubernetes operator that manages `Certificate` custom resources. 
It automates creation, updates, and lifecycle management within your cluster.

---

## 📋 Features

- ✅ Reconciliation logic for Certificate resource, creating and updating it if it expires

---

## 🚀 Getting Started

### Prerequisites

- [Go](https://golang.org/) >= 1.23
- [Docker](https://www.docker.com/)
- [Kubebuilder](https://book.kubebuilder.io/quick-start.html)
- [Kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/)
- kind for e2e tests

## 📂 Project Structure

- /api contains all the apis this operator manages (Certificate)
- /config contains all the kustomize needed to deploy + samples
- /test are all the e2e tests

### 🧪 Running Locally

```bash
# will boot up kind cluster, make manifests, generate, build docker image and run all in the kind cluster
make run-in-kind 
# will boot up kind cluster, make manifests, generate, build docker image and run the operator locally 
make run
```

```bash
# apply the sample
kubectl apply -f ./config/samples/certs.k8c.io_v1_certificate.yaml
```