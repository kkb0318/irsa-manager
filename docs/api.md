# API Reference

## Packages
- [irsa.kkb0318.github.io/v1alpha1](#irsakkb0318githubiov1alpha1)


## irsa.kkb0318.github.io/v1alpha1

Package v1alpha1 contains API Schema definitions for the irsa v1alpha1 API group

### Resource Types
- [IRSASetup](#irsasetup)



#### Auth



Auth holds the authentication configuration details.



_Appears in:_
- [IRSASetupSpec](#irsasetupspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `secretRef` _[SecretRef](#secretref)_ | SecretRef specifies the reference to the Kubernetes secret containing authentication details. |  |  |


#### Discovery



Discovery holds the configuration for IdP Discovery, which is crucial for locating
the OIDC provider in a self-hosted environment.



_Appears in:_
- [IRSASetupSpec](#irsasetupspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `s3` _[S3Discovery](#s3discovery)_ | S3 specifies the AWS S3 bucket details where the OIDC provider's discovery information is hosted. |  |  |


#### IRSASetup



IRSASetup represents a configuration for setting up IAM Roles for Service Accounts (IRSA) in a Kubernetes cluster.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `irsa.kkb0318.github.io/v1alpha1` | | |
| `kind` _string_ | `IRSASetup` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[IRSASetupSpec](#irsasetupspec)_ |  |  |  |


#### IRSASetupSpec



IRSASetupSpec defines the desired state of IRSASetup



_Appears in:_
- [IRSASetup](#irsasetup)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `mode` _string_ | Mode specifies the mode of operation. Can be either "selfhosted" or "eks". |  |  |
| `discovery` _[Discovery](#discovery)_ | Discovery configures the IdP Discovery process, essential for setting up IRSA by locating<br />the OIDC provider information. |  |  |
| `auth` _[Auth](#auth)_ | Auth contains authentication configuration details. |  |  |




#### S3Discovery



S3Discovery contains the specifics of the S3 bucket used for hosting OIDC provider discovery information.



_Appears in:_
- [Discovery](#discovery)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `region` _string_ | Region denotes the AWS region where the S3 bucket is located. |  |  |
| `bucketName` _string_ | BucketName is the name of the S3 bucket that hosts the OIDC discovery information. |  |  |


#### SecretRef



SecretRef contains the reference to a Kubernetes secret.



_Appears in:_
- [Auth](#auth)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name specifies the name of the secret. |  |  |
| `namespace` _string_ | Namespace specifies the namespace of the secret. |  |  |


