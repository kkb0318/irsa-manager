# irsa-manager

[irsa-manager](https://github.com/kkb0318/irsa-manager) IRSA manager allows you to easily set up IAM Roles for Service Accounts (IRSA) on non-EKS Kubernetes clusters.

## Setup

* Get Repo Info
```console
helm repo add kkb0318 https://kkb0318.github.io/irsa-manager
helm repo update
```

* Install Chart

```console
helm install irsa-manager kkb0318/irsa-manager -n irsa-manager-system --create-namespace
```

* Set AWS Secret for IRSA Manager

```console
kubectl create secret generic aws-secret -n irsa-manager-system \
  --from-literal=aws-access-key-id=<your-access-key-id> \
  --from-literal=aws-secret-access-key=<your-secret-access-key> \
  --from-literal=aws-session-token=<your-aws-session-token> # Optional \
  --from-literal=aws-region=<your-region> \
  --from-literal=aws-role-arn=<your-role-arn>  # Optional: Set this if you want to switch roles

```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| controllerManager.manager.args[0] | string | `"--leader-elect"` |  |
| controllerManager.manager.containerSecurityContext.allowPrivilegeEscalation | bool | `false` |  |
| controllerManager.manager.containerSecurityContext.capabilities.drop[0] | string | `"ALL"` |  |
| controllerManager.manager.image.repository | string | `"ghcr.io/kkb0318/irsa-manager"` |  |
| controllerManager.manager.image.tag | string | `"APP_VERSION"` |  |
| controllerManager.manager.resources.limits.cpu | string | `"500m"` |  |
| controllerManager.manager.resources.limits.memory | string | `"128Mi"` |  |
| controllerManager.manager.resources.requests.cpu | string | `"10m"` |  |
| controllerManager.manager.resources.requests.memory | string | `"64Mi"` |  |
| controllerManager.replicas | int | `1` |  |
| controllerManager.serviceAccount.annotations | object | `{}` |  |
| kubernetesClusterDomain | string | `"cluster.local"` |  |
| metricsService.ports[0].name | string | `"https"` |  |
| metricsService.ports[0].port | int | `8443` |  |
| metricsService.ports[0].protocol | string | `"TCP"` |  |
| metricsService.ports[0].targetPort | string | `"https"` |  |
| metricsService.type | string | `"ClusterIP"` |  |

