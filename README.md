# irsa-manager

irsa-manager allows you to easily set up IAM Roles for Service Accounts (IRSA) on non-EKS Kubernetes clusters.

## Introduction

IRSA (IAM Roles for Service Accounts) allows Kubernetes service accounts to assume AWS IAM roles.
This is particularly useful for providing Kubernetes workloads with the necessary AWS permissions in a secure manner.

## Prerequisites

Before you begin, ensure you have the following:

- A running Kubernetes cluster (non-EKS).
- Helm installed on your local machine.
- AWS user credentials with appropriate permissions.

## Setup

Follow these steps to set up IRSA on your non-EKS cluster:

1. install helm

Add the irsa-manager Helm repository and install irsa-manager:

```
helm repo add kkb0318 https://kkb0318.github.io/irsa-manager
helm repo update
helm install irsa-manager kkb0318/irsa-manager -n irsa-manager-system --create-namespace
```

2. Set AWS Secret for IRSA Manager

Create a secret for irsa-manager to access AWS:

```
kubectl create secret generic aws-secret -n irsa-manager-system \
  --from-literal=aws-access-key-id=<your-access-key-id> \
  --from-literal=aws-secret-access-key=<your-secret-access-key> \
  --from-literal=aws-region=<your-region> \
  --from-literal=aws-role-arn=<your-role-arn>  # Optional: Set this if you want to switch roles

```

3. Create an IRSASetup Custom Resource

Define and apply an IRSASetup custom resource according to your needs.

4. Modify kube-apiserver Settings

Execute the following commands on the control plane server to save the public and private keys for Kubernetes signatures:

```
kubectl get secret -n kube-system irsa-manager-key -o jsonpath="{.data.ssh-privatekey}" | base64 --decode | sudo tee /etc/kubernetes/pki/irsa-manager.key > /dev/null
kubectl get secret -n kube-system irsa-manager-key -o jsonpath="{.data.ssh-publickey}" | base64 --decode | sudo tee /etc/kubernetes/pki/irsa-manager.pub > /dev/null
```

Then, modify the kube-apiserver.yaml file to include the following parameters:

- API Audiences

```
--api-audiences=sts.amazonaws.com
```

- Service Account Issuer

```
--service-account-issuer=https://s3-<region>.amazonaws.com/<bucketName>
```

- Service Account Key File

The public key (oidc-issuer.pub) generated previously can be read by the API server. Add the path for this parameter flag:

```
--service-account-key-file=/etc/kubernetes/pki/irsa-manager.pub
```

> [!NOTE]
> Add this setting as the first element. If specified multiple times, tokens signed by any of the specified keys are considered valid by the Kubernetes API server.

- Service Account Signing Key File

The private key (oidc-issuer.key) generated previously can be read by the API server. Add the path for this parameter flag:

```
--service-account-signing-key-file=/etc/kubernetes/pki/irsa-manager.key
```

> [!NOTE]
> Add these settings before the existing ones. If specified multiple times, tokens signed by any of the specified keys are considered valid by the Kubernetes API server.

For more details, refer to the [Kubernetes documentation](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#serviceaccount-token-volume-projection).

TODO

- [x] delete secret
- [x] delete s3
- [x] only once secret creation (status check)
- [x] no update secret
- [x] no update bucket object
- [x] delete idp
- [x] issue: when irsasetup was deleted, resource remained with some error occured
- [x] certificate
- [x] check keys.json keyid has to be empty or not
- [x] IRSA api
- [ ] use with cert-manager
- [ ] aws context
- [ ] temporary aws account
- [ ] cannot delete IRSASetup before existing IRSA

- [ ] validation webhook (invalid to change)
