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

> [!NOTE]
> You may encounter an error during the deployment. Proceed with the following steps and create the "aws-secret" secret to eliminate the error.

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

```
apiVersion: irsa.kkb0318.github.io/v1alpha1
kind: IRSASetup
metadata:
  name: irsa-init
  namespace: irsa-manager-system
spec:
  cleanup: false
  mode: selfhosted
  discovery:
    s3:
      region: <region>
      bucketName: <S3 bucket name>
```

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
--service-account-issuer=https://s3-<region>.amazonaws.com/<S3 bucket name>
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
> Overwrite the existing settings.

For more details, refer to the [Kubernetes documentation](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#serviceaccount-token-volume-projection).

5. Check the status

Check the IRSASetup custom resource status. If the status is true, you are ready to use IRSA.

## How To Use

You can set IRSA for the Kubernetes ServiceAccount.

The following example shows that irsa-manager sets the `irsa1-sa` ServiceAccount in the kube-system and default namespaces with the AmazonS3FullAccess policy:

```
apiVersion: irsa.kkb0318.github.io/v1alpha1
kind: IRSA
metadata:
  name: irsa-sample
  namespace: irsa-manager-system
spec:
  cleanup: true
  serviceAccount:
    name: irsa1-sa
    namespaces:
      - kube-system
      - default
  iamRole:
    name: irsa1-role
  iamPolicies:
    - AmazonS3FullAccess
```

For more details, please see the API Reference.

## Verification

To verify the above example and ensure the IRSA works correctly, you can check the following job.
There is a Kubernetes job that will put one file into the S3 bucket, confirming that the Pod can assume the role to get S3 write permission:

```
cd validation
sh s3-echoer.sh
```

## API Reference

You can find the reference in the [Reference](./docs/api.md) file.

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.

## Acknowledgments

In creating this OSS project, I referred to several sources and would like to express my gratitude for their valuable information and insights.

The necessity of this project was realized through discussions in the following issue:

- https://github.com/kubernetes-sigs/cluster-api-provider-aws/issues/3560

Additionally, the implementation was guided by the following repositories:

- [smalltown/aws-irsa-example](https://github.com/smalltown/aws-irsa-example)
- [aws/amazon-eks-pod-identity-webhook](https://github.com/aws/amazon-eks-pod-identity-webhook)
