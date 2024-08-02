/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// IRSASetupKind represents the kind attribute of an IRSASetup resource.
	IRSASetupKind = "IRSASetup"
)

// IRSASetupSpec defines the desired state of IRSASetup
type IRSASetupSpec struct {
	// Cleanup, when enabled, allows the IRSASetup to perform garbage collection
	// of resources that are no longer needed or managed.
	// +required
	Cleanup bool `json:"cleanup"`

	// Mode specifies the operation mode of the controller.
	// Possible values:
	//   - "selfhosted": For self-managed Kubernetes clusters.
	//   - "eks": For Amazon EKS environments.
	// Default: "selfhosted"
	Mode SetupMode `json:"mode,omitempty"`

	// Discovery configures the IdP Discovery process, essential for setting up IRSA by locating
	// the OIDC provider information.
	// Only applicable when Mode is "selfhosted".
	Discovery Discovery `json:"discovery"`

	// IamOIDCProvider configures IAM OIDC IamOIDCProvider Name
	// Only applicable when Mode is "eks".
	IamOIDCProvider string `json:"provider,omitempty"`
}

// +kubebuilder:default=selfhosted
// +kubebuilder:validation:Enum=selfhosted;eks
// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
type SetupMode string

const (
	ModeSelfhosted = SetupMode("selfhosted")
	ModeEks        = SetupMode("eks")
)

// Discovery holds the configuration for IdP Discovery, which is crucial for locating
// the OIDC provider in a self-hosted environment.
type Discovery struct {
	// S3 specifies the AWS S3 bucket details where the OIDC provider's discovery information is hosted.
	S3 S3Discovery `json:"s3,omitempty"`
}

// S3Discovery contains the specifics of the S3 bucket used for hosting OIDC provider discovery information.
type S3Discovery struct {
	// Region denotes the AWS region where the S3 bucket is located.
	Region string `json:"region"`

	// BucketName is the name of the S3 bucket that hosts the OIDC discovery information.
	BucketName string `json:"bucketName"`
}

// IRSASetupStatus defines the observed state of IRSASetup
type IRSASetupStatus struct {
	SelfHostedSetup []metav1.Condition `json:"selfHostedSetup,omitempty"`
}

// GetSelfhostedStatusConditions returns a pointer to the Status.Conditions slice
func (in *IRSASetup) GetSelfhostedStatusConditions() *[]metav1.Condition {
	return &in.Status.SelfHostedSetup
}

func SetupSelfHostedStatusReady(irsa IRSASetup, reason, message string) IRSASetup {
	newCondition := metav1.Condition{
		Type:    ReadyCondition,
		Status:  metav1.ConditionTrue,
		Reason:  reason,
		Message: message,
	}
	apimeta.SetStatusCondition(irsa.GetSelfhostedStatusConditions(), newCondition)
	return irsa
}

func SelfHostedStatusNotReady(irsa IRSASetup, reason, message string) IRSASetup {
	newCondition := metav1.Condition{
		Type:    ReadyCondition,
		Status:  metav1.ConditionFalse,
		Reason:  reason,
		Message: message,
	}
	apimeta.SetStatusCondition(irsa.GetSelfhostedStatusConditions(), newCondition)
	return irsa
}

// SelfHostedReadyStatus
func SelfHostedReadyStatus(irsa IRSASetup) *metav1.Condition {
	if c := apimeta.FindStatusCondition(irsa.Status.SelfHostedSetup, ReadyCondition); c != nil {
		return c
	}
	return nil
}

// HasConditionReason
func HasConditionReason(cond *metav1.Condition, reasons ...string) bool {
	if cond == nil {
		return false
	}
	for _, reason := range reasons {
		if cond.Reason == reason {
			return true
		}
	}
	return false
}

func IsSelfHostedReadyConditionTrue(irsa IRSASetup) bool {
	return apimeta.IsStatusConditionTrue(irsa.Status.SelfHostedSetup, ReadyCondition)
}

type SelfHostedReason string

const (
	SelfHostedReasonFailedWebhook SelfHostedReason = "SelfHostedSetupFailedWebhookCreation"
	SelfHostedReasonFailedOidc    SelfHostedReason = "SelfHostedSetupFailedOidcCreation"
	SelfHostedReasonFailedKeys    SelfHostedReason = "SelfHostedSetupFailedKeysCreation"
	SelfHostedReasonReady         SelfHostedReason = "SelfHostedSetupReady"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="SelfHostedReady",type="string",JSONPath=".status.selfHostedSetup[?(@.type==\"Ready\")].status",description=""

// IRSASetup represents a configuration for setting up IAM Roles for Service Accounts (IRSA) in a Kubernetes cluster.
type IRSASetup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IRSASetupSpec   `json:"spec,omitempty"`
	Status IRSASetupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IRSASetupList contains a list of IRSASetup
type IRSASetupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IRSASetup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IRSASetup{}, &IRSASetupList{})
}
