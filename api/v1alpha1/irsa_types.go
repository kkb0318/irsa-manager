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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IRSASpec defines the desired state of IRSA
type IRSASpec struct {
	// ServiceAccount represents the Kubernetes service account associated with the IRSA
	ServiceAccount IRSAServiceAccount `json:"serviceAccount,omitempty"`
	// IamRole represents the IAM role details associated with the IRSA
	IamRole IamRole `json:"iamRole,omitempty"`
	// IamPolicies represents the list of IAM policies to be attached to the IAM role
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
	// Name represents the name of the IAM role
	Name string `json:"name,omitempty"`
}

// IRSAStatus defines the observed state of IRSA
type IRSAStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

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
