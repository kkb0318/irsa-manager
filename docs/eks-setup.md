## Setup for EKS

Define and apply an IRSASetup custom resource.

```yaml
apiVersion: irsa-manager.kkb0318.github.io/v1alpha1
kind: IRSASetup
metadata:
  name: irsa-init
  namespace: irsa-manager-system
spec:
  mode: eks
  cleanup: false
  iamOIDCProvider: "oidc.eks.<region>.amazonaws.com/id/<id>"
```

Check the IRSASetup custom resource status to verify whether it is set to true.
