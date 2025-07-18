# Open WebUI Ollama Scaler API

This is a simple API to count the number of active users from the Open Web-UI API. The API is intended to be used as a KEDA metric-api scaler to adjust the number of ollama instnces dynamically based on the active users.

## Features
Fetches and counts the number of active users using the Open Web-UI API.
Caches the result for a configurable amount of time to reduce the number of API calls.
Exposes an HTTP endpoint to return the active users count in JSON format.

## Installation

In kubernetes style:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: openwebui-ollama-scaler
  labels:
    app: openwebui-ollama-scaler
spec:
  replicas: 1
  selector:
    matchLabels:
      app: openwebui-ollama-scaler
  template:
    metadata:
      labels:
        app: openwebui-ollama-scaler
    spec:
      containers:
      - name: scaler
        image: ghcr.io/krisek/openwebui-ollama-scaler:latest
        env:
        - name: API_URL
          valueFrom:
            secretKeyRef:
              name: api-secrets
              key: API_URL
        - name: TOKEN
          valueFrom:
            secretKeyRef:
              name: api-secrets
              key: TOKEN
        - name: CACHE_TIMEOUT
          value: "60"  # Cache timeout in seconds
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: openwebui-ollama-scaler
  labels:
    app: openwebui-ollama-scaler
spec:
  selector:
    app: openwebui-ollama-scaler
  ports:
    - protocol: TCP
      port: 8080
      targetPort: 8080
  type: ClusterIP
---
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: ollama
spec:
  scaleTargetRef:
    name: ollama
  pollingInterval: 120
  cooldownPeriod: 600
  minReplicaCount: 0
  maxReplicaCount: 1
  triggers:
  - type: metrics-api
    metadata:
      targetValue: "1"
      format: "json"
      url: "http://openwebui-ollama-scaler.ai-lab:8080/active_users"
      valueLocation: "active_users"
```

You might need to adopt the metrics-api url depending on your namespace.

## Example Request
```bash
curl http://localhost:8080/active_users
Example Response:
json
{
  "active_users": 5
}
```

## Environment Variables

- `API_URL`: The base URL for the Open Web-UI API.
  Default: None (must be set)
- `TOKEN`: A token with necessary permissions to access the Open Web-UI API.
  Permissions: Not specified in this documentation, refer to the Open Web-UI API documentation
- `CACHE_TIMEOUT`: Cache duration in seconds.
  Default: 60 seconds
- `PORT`: The port the server will listen on.
  Default: 8080
