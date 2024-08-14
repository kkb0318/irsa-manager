## Setup for Self-Hosted

![](./IRSASetup-cr.png)

### Define and apply an IRSASetup custom resource according to your needs.

```yaml
apiVersion: irsa-manager.kkb0318.github.io/v1alpha1
kind: IRSASetup
metadata:
  name: irsa-init
  namespace: irsa-manager-system
spec:
  cleanup: false
  discovery:
    s3:
      region: <region>
      bucketName: <S3 bucket name>
```

Check the IRSASetup custom resource status to verify whether it is set to true.

> [!NOTE]
> Please ensure that only one IRSASetup resource is created.

### Modify kube-apiserver Settings

If the IRSASetup status is true, a key file (Name: `irsa-manager-key` , Namespace: `kube-system` ) will be created. This is used for signing tokens in the kubernetes API.
Execute the following commands on the control plane server to save the public and private keys locally for Kubernetes signatures:

```console
kubectl get secret -n kube-system irsa-manager-key -o jsonpath="{.data.ssh-privatekey}" | base64 --decode | sudo tee /path/to/file.key > /dev/null
kubectl get secret -n kube-system irsa-manager-key -o jsonpath="{.data.ssh-publickey}" | base64 --decode | sudo tee /path/to/file.pub > /dev/null
```

> [!NOTE]
> Path: `/path/to/file` can be any path you choose.
> If you use kubeadm, it is recommended to set `/etc/kubernetes/pki/irsa-manager.(key|pub)`

Then, modify the kube-apiserver settings to include the following parameters:

- API Audiences

```
--api-audiences=sts.amazonaws.com
```

- Service Account Issuer

```
--service-account-issuer=https://s3-<region>.amazonaws.com/<S3 bucket name>
```

> [!NOTE]
> Add this setting as the first element.
> When this flag is specified multiple times, the first is used to generate tokens and all are used to determine which issuers are accepted.

- Service Account Key File

The public key generated previously can be read by the API server. Add the path for this parameter flag:

```
--service-account-key-file=/path/to/file.pub
```

> [!NOTE]
> If you do not mount /path/to directory, you need to add the volumes field to this path.

- Service Account Signing Key File

The private key (oidc-issuer.key) generated previously can be read by the API server. Add the path for this parameter flag:

```
--service-account-signing-key-file=/path/to/file.key
```

> [!NOTE]
> Overwrite the existing settings.
> If you do not mount /path/to directory, you need to add the volumes field to this path.

For more details, refer to the [Kubernetes documentation](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#serviceaccount-token-volume-projection).
