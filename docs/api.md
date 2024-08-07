# API Reference

## Packages
- [irsa-manager.kkb0318.github.io/v1alpha1](#irsa-managerkkb0318githubiov1alpha1)


## irsa-manager.kkb0318.github.io/v1alpha1

Package v1alpha1 contains API Schema definitions for the irsa v1alpha1 API group

### Resource Types
- [IRSA](#irsa)
- [IRSASetup](#irsasetup)



#### Discovery



Discovery holds the configuration for IdP Discovery, which is crucial for locating
the OIDC provider in a self-hosted environment.



_Appears in:_
- [IRSASetupSpec](#irsasetupspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `s3` _[S3Discovery](#s3discovery)_ | S3 specifies the AWS S3 bucket details where the OIDC provider's discovery information is hosted. |  |  |




#### IRSA



IRSA is the Schema for the irsas API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `irsa-manager.kkb0318.github.io/v1alpha1` | | |
| `kind` _string_ | `IRSA` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[IRSASpec](#irsaspec)_ |  |  |  |






#### IRSAServiceAccount



IRSAServiceAccount represents the details of the Kubernetes service account



_Appears in:_
- [IRSASpec](#irsaspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name represents the name of the Kubernetes service account |  |  |
| `namespaces` _string array_ | Namespaces represents the list of namespaces where the service account is used |  |  |


#### IRSASetup



IRSASetup represents a configuration for setting up IAM Roles for Service Accounts (IRSA) in a Kubernetes cluster.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `irsa-manager.kkb0318.github.io/v1alpha1` | | |
| `kind` _string_ | `IRSASetup` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[IRSASetupSpec](#irsasetupspec)_ |  |  |  |


#### IRSASetupSpec



IRSASetupSpec defines the desired state of IRSASetup



_Appears in:_
- [IRSASetup](#irsasetup)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `cleanup` _boolean_ | Cleanup, when enabled, allows the IRSASetup to perform garbage collection<br />of resources that are no longer needed or managed. |  |  |
| `mode` _[SetupMode](#setupmode)_ | Mode specifies the operation mode of the controller.<br />Possible values:<br />  - "selfhosted": For self-managed Kubernetes clusters.<br />  - "eks": For Amazon EKS environments.<br />Default: "selfhosted" |  | Enum: [selfhosted eks] <br /> |
| `discovery` _[Discovery](#discovery)_ | Discovery configures the IdP Discovery process, essential for setting up IRSA by locating<br />the OIDC provider information.<br />Only applicable when Mode is "selfhosted". |  |  |
| `iamOIDCProvider` _string_ | IamOIDCProvider configures IAM OIDC IamOIDCProvider Name<br />Only applicable when Mode is "eks". |  |  |




#### IRSASpec



IRSASpec defines the desired state of IRSA



_Appears in:_
- [IRSA](#irsa)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `cleanup` _boolean_ | Cleanup, when enabled, allows the IRSA to perform garbage collection<br />of resources that are no longer needed or managed. |  |  |
| `serviceAccount` _[IRSAServiceAccount](#irsaserviceaccount)_ | ServiceAccount represents the Kubernetes service account associated with the IRSA. |  |  |
| `iamRole` _[IamRole](#iamrole)_ | IamRole represents the IAM role details associated with the IRSA. |  |  |
| `iamPolicies` _string array_ | IamPolicies represents the list of IAM policies to be attached to the IAM role.<br />You can set both the policy name (only AWS default policies) or the full ARN. |  |  |




#### IamRole



IamRole represents the IAM role configuration



_Appears in:_
- [IRSASpec](#irsaspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name represents the name of the IAM role. |  |  |


#### S3Discovery



S3Discovery contains the specifics of the S3 bucket used for hosting OIDC provider discovery information.



_Appears in:_
- [Discovery](#discovery)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `region` _string_ | Region denotes the AWS region where the S3 bucket is located. |  |  |
| `bucketName` _string_ | BucketName is the name of the S3 bucket that hosts the OIDC discovery information. |  |  |




#### SetupMode

_Underlying type:_ _string_



_Validation:_
- Enum: [selfhosted eks]

_Appears in:_
- [IRSASetupSpec](#irsasetupspec)



