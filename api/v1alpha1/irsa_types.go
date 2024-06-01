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
	// IRSAKind represents the kind attribute of an IRSA resource.
	IRSAKind = "IRSA"
)

// IRSASpec defines the desired state of IRSA
type IRSASpec struct {
	// Cleanup, when enabled, allows the IRSA to perform garbage collection
	// of resources that are no longer needed or managed.
	// +required
	Cleanup bool `json:"cleanup"`

	// ServiceAccount represents the Kubernetes service account associated with the IRSA.
	// +required
	ServiceAccount IRSAServiceAccount `json:"serviceAccount,omitempty"`

	// IamRole represents the IAM role details associated with the IRSA.
	// +required
	IamRole IamRole `json:"iamRole,omitempty"`

	// IamPolicies represents the list of IAM policies to be attached to the IAM role.
	// You can set both the policy name (only AWS default policies) or the full ARN.
	// +required
	IamPolicies []string `json:"iamPolicies,omitempty"`
}

// IRSAServiceAccount represents the details of the Kubernetes service account
type IRSAServiceAccount struct {
	// Name represents the name of the Kubernetes service account
	Name string `json:"name,omitempty"`
	// Namespaces represents the list of namespaces where the service account is used
	Namespaces []string `json:"namespaces,omitempty"`
}

// IamRole represents the IAM role configuration
type IamRole struct {
	// Name represents the name of the IAM role.
	Name string `json:"name,omitempty"`
}

// IRSAStatus defines the observed state of IRSA.
type IRSAStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// GetIRSAStatusConditions returns a pointer to the Conditions slice
func (in *IRSA) GetIRSAStatusConditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func IRSAStatusReady(irsa IRSA, reason, message string) IRSA {
	newCondition := metav1.Condition{
		Type:    ReadyCondition,
		Status:  metav1.ConditionTrue,
		Reason:  reason,
		Message: message,
	}
	apimeta.SetStatusCondition(irsa.GetIRSAStatusConditions(), newCondition)
	return irsa
}

func IRSAStatusNotReady(irsa IRSA, reason, message string) IRSA {
	newCondition := metav1.Condition{
		Type:    ReadyCondition,
		Status:  metav1.ConditionFalse,
		Reason:  reason,
		Message: message,
	}
	apimeta.SetStatusCondition(irsa.GetIRSAStatusConditions(), newCondition)
	return irsa
}

type IRSAReason string

const (
	IRSAReasonFailedRoleUpdate IRSAReason = "IRSAFailedRoleUpdate"
	IRSAReasonFailedK8sApply   IRSAReason = "IRSAFailedApplyingResources"
	IRSAReasonReady            IRSAReason = "IRSAReady"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""

// IRSA is the Schema for the irsas API
type IRSA struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IRSASpec   `json:"spec,omitempty"`
	Status IRSAStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IRSAList contains a list of IRSA
type IRSAList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IRSA `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IRSA{}, &IRSAList{})
}
