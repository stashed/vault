/*
Copyright AppsCode Inc. and Contributors

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
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kmapi "kmodules.xyz/client-go/api/v1"
)

const (
	ResourceKindVaultPolicy = "VaultPolicy"
	ResourceVaultPolicy     = "vaultpolicy"
	ResourceVaultPolicies   = "vaultpolicies"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=vaultpolicies,singular=vaultpolicy,shortName=vp,categories={vault,policy,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type VaultPolicy struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VaultPolicySpec   `json:"spec,omitempty"`
	Status            VaultPolicyStatus `json:"status,omitempty"`
}

// More info: https://www.vaultproject.io/docs/concepts/policies.html
type VaultPolicySpec struct {
	// VaultRef is the name of a AppBinding referencing to a Vault Server
	VaultRef core.LocalObjectReference `json:"vaultRef"`

	// VaultPolicyName is the policy name set inside Vault.
	// This defaults to following format: k8s.${cluster}.${metadata.namespace}.${metadata.name}
	// +optional
	VaultPolicyName string `json:"vaultPolicyName,omitempty"`

	// PolicyDocument specifies a vault policy in hcl format.
	// For example:
	// path "secret/*" {
	//   capabilities = ["create", "read", "update", "delete", "list"]
	// }
	// +optional
	PolicyDocument string `json:"policyDocument,omitempty"`

	// Policy specifies a vault policy in json format.
	// +optional
	// +kubebuilder:validation:EmbeddedResource
	// +kubebuilder:pruning:PreserveUnknownFields
	Policy *runtime.RawExtension `json:"policy,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type VaultPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultPolicy `json:"items,omitempty"`
}

// +kubebuilder:validation:Enum=Success;Failed
type PolicyPhase string

const (
	PolicySuccess PolicyPhase = "Success"
	PolicyFailed  PolicyPhase = "Failed"
)

type VaultPolicyStatus struct {
	// ObservedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Phase indicates whether the policy successfully applied in vault or not or in progress
	// +optional
	Phase PolicyPhase `json:"phase,omitempty"`

	// Represents the latest available observations of a VaultPolicy.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty"`
}
