# kubehelper
This golang project patches the replicas field of a kubernetes deployment for scaling purposes, this project is mainly just for fun and learning, but perhaps could have a usecase for automation for easy programmatic access.

# Testing the project
1. `go mody tidy`
2. `go build`
3. run executable
4. send request:
```bash
curl --location --request PATCH 'http://localhost:8081/api/v1/deployment/scale' \
--header 'Content-Type: application/json' \
--data-raw '{
    "Deployment": "nginx-deployment",
    "Namespace": "default",
    "Replicas": 1
}'
```
# Create local kind cluster
1. `kind create cluster --config kind-example-config.yaml`
2. switch context to kind
3. `kubectl apply -f test-deployment.yaml` (for testing)
