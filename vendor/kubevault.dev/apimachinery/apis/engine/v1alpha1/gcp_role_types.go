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
)

const (
	ResourceKindGCPRole = "GCPRole"
	ResourceGCPRole     = "gcprole"
	ResourceGCPRoles    = "gcproles"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=gcproles,singular=gcprole,categories={vault,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type GCPRole struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              GCPRoleSpec `json:"spec,omitempty"`
	Status            RoleStatus  `json:"status,omitempty"`
}

// +kubebuilder:validation:Enum=access_token;service_account_key
type GCPSecretType string

const (
	GCPSecretAccessToken       GCPSecretType = "access_token"
	GCPSecretServiceAccountKey GCPSecretType = "service_account_key"
)

// GCPRoleSpec contains connection information, GCP role info, etc
// More info: https://www.vaultproject.io/api/secret/gcp/index.html#parameters
type GCPRoleSpec struct {
	// SecretEngineRef is the name of a Secret Engine
	SecretEngineRef core.LocalObjectReference `json:"secretEngineRef"`

	// Path defines the path of the Google Cloud secret engine
	// default: gcp
	// More info: https://www.vaultproject.io/docs/auth/gcp.html#via-the-cli-helper
	// +optional
	Path string `json:"path,omitempty"`

	// Specifies the type of secret generated for this role set
	SecretType GCPSecretType `json:"secretType"`

	// Name of the GCP project that this roleset's service account will belong to.
	// Cannot be updated.
	Project string `json:"project"`

	// Bindings configuration string (expects HCL or JSON format in raw
	// or base64-encoded string)
	Bindings string `json:"bindings"`

	// List of OAuth scopes to assign to access_token secrets generated
	// under this role set (access_token role sets only)
	// +optional
	TokenScopes []string `json:"tokenScopes,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type GCPRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is a list of GCPRole objects
	Items []GCPRole `json:"items,omitempty"`
}

const (
	GCPSACredentialJson = "sa.json"
)
