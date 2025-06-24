# Deploying Agentry

This directory contains Kubernetes manifests and a Helm chart for deploying Agentry workers.

The manifests under `k8s/` are raw YAML files that can be applied directly:

```bash
kubectl apply -f k8s/
```

The `helm/agentry` chart provides the same resources with configurable values for
queue address, autoscaler settings and storage:

```bash
helm install agentry helm/agentry
```

Adjust values by creating a YAML file and passing it with `-f`. See
`helm/agentry/values.yaml` for all available options.

When enabled, the autoscaler monitors NATS queue lag and scales the
`agentry-worker` deployment accordingly.
