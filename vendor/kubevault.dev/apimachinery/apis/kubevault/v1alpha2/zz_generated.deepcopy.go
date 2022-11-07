//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha2

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	apiv1 "kmodules.xyz/client-go/api/v1"
	v1alpha1 "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	monitoringagentapiapiv1 "kmodules.xyz/monitoring-agent-api/api/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AllowedSecretEngines) DeepCopyInto(out *AllowedSecretEngines) {
	*out = *in
	if in.Namespaces != nil {
		in, out := &in.Namespaces, &out.Namespaces
		*out = new(SecretEngineNamespaces)
		(*in).DeepCopyInto(*out)
	}
	if in.SecretEngines != nil {
		in, out := &in.SecretEngines, &out.SecretEngines
		*out = make([]SecretEngineType, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AllowedSecretEngines.
func (in *AllowedSecretEngines) DeepCopy() *AllowedSecretEngines {
	if in == nil {
		return nil
	}
	out := new(AllowedSecretEngines)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuthMethod) DeepCopyInto(out *AuthMethod) {
	*out = *in
	if in.KubernetesConfig != nil {
		in, out := &in.KubernetesConfig, &out.KubernetesConfig
		*out = new(KubernetesConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.OIDCConfig != nil {
		in, out := &in.OIDCConfig, &out.OIDCConfig
		*out = new(JWTOIDCConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.JWTConfig != nil {
		in, out := &in.JWTConfig, &out.JWTConfig
		*out = new(JWTOIDCConfig)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AuthMethod.
func (in *AuthMethod) DeepCopy() *AuthMethod {
	if in == nil {
		return nil
	}
	out := new(AuthMethod)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuthMethodStatus) DeepCopyInto(out *AuthMethodStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AuthMethodStatus.
func (in *AuthMethodStatus) DeepCopy() *AuthMethodStatus {
	if in == nil {
		return nil
	}
	out := new(AuthMethodStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AwsKmsSsmSpec) DeepCopyInto(out *AwsKmsSsmSpec) {
	*out = *in
	if in.CredentialSecretRef != nil {
		in, out := &in.CredentialSecretRef, &out.CredentialSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AwsKmsSsmSpec.
func (in *AwsKmsSsmSpec) DeepCopy() *AwsKmsSsmSpec {
	if in == nil {
		return nil
	}
	out := new(AwsKmsSsmSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AzureKeyVault) DeepCopyInto(out *AzureKeyVault) {
	*out = *in
	if in.TLSSecretRef != nil {
		in, out := &in.TLSSecretRef, &out.TLSSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.CredentialSecretRef != nil {
		in, out := &in.CredentialSecretRef, &out.CredentialSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzureKeyVault.
func (in *AzureKeyVault) DeepCopy() *AzureKeyVault {
	if in == nil {
		return nil
	}
	out := new(AzureKeyVault)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AzureSpec) DeepCopyInto(out *AzureSpec) {
	*out = *in
	if in.CredentialSecretRef != nil {
		in, out := &in.CredentialSecretRef, &out.CredentialSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzureSpec.
func (in *AzureSpec) DeepCopy() *AzureSpec {
	if in == nil {
		return nil
	}
	out := new(AzureSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackendStorageSpec) DeepCopyInto(out *BackendStorageSpec) {
	*out = *in
	if in.Inmem != nil {
		in, out := &in.Inmem, &out.Inmem
		*out = new(InmemSpec)
		**out = **in
	}
	if in.Etcd != nil {
		in, out := &in.Etcd, &out.Etcd
		*out = new(EtcdSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Gcs != nil {
		in, out := &in.Gcs, &out.Gcs
		*out = new(GcsSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.S3 != nil {
		in, out := &in.S3, &out.S3
		*out = new(S3Spec)
		(*in).DeepCopyInto(*out)
	}
	if in.Azure != nil {
		in, out := &in.Azure, &out.Azure
		*out = new(AzureSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.PostgreSQL != nil {
		in, out := &in.PostgreSQL, &out.PostgreSQL
		*out = new(PostgreSQLSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.MySQL != nil {
		in, out := &in.MySQL, &out.MySQL
		*out = new(MySQLSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.File != nil {
		in, out := &in.File, &out.File
		*out = new(FileSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.DynamoDB != nil {
		in, out := &in.DynamoDB, &out.DynamoDB
		*out = new(DynamoDBSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Swift != nil {
		in, out := &in.Swift, &out.Swift
		*out = new(SwiftSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Consul != nil {
		in, out := &in.Consul, &out.Consul
		*out = new(ConsulSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Raft != nil {
		in, out := &in.Raft, &out.Raft
		*out = new(RaftSpec)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackendStorageSpec.
func (in *BackendStorageSpec) DeepCopy() *BackendStorageSpec {
	if in == nil {
		return nil
	}
	out := new(BackendStorageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConsulSpec) DeepCopyInto(out *ConsulSpec) {
	*out = *in
	if in.TLSSecretRef != nil {
		in, out := &in.TLSSecretRef, &out.TLSSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConsulSpec.
func (in *ConsulSpec) DeepCopy() *ConsulSpec {
	if in == nil {
		return nil
	}
	out := new(ConsulSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DynamoDBSpec) DeepCopyInto(out *DynamoDBSpec) {
	*out = *in
	if in.CredentialSecretRef != nil {
		in, out := &in.CredentialSecretRef, &out.CredentialSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DynamoDBSpec.
func (in *DynamoDBSpec) DeepCopy() *DynamoDBSpec {
	if in == nil {
		return nil
	}
	out := new(DynamoDBSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EtcdSpec) DeepCopyInto(out *EtcdSpec) {
	*out = *in
	if in.CredentialSecretRef != nil {
		in, out := &in.CredentialSecretRef, &out.CredentialSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.TLSSecretRef != nil {
		in, out := &in.TLSSecretRef, &out.TLSSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EtcdSpec.
func (in *EtcdSpec) DeepCopy() *EtcdSpec {
	if in == nil {
		return nil
	}
	out := new(EtcdSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FileSpec) DeepCopyInto(out *FileSpec) {
	*out = *in
	in.VolumeClaimTemplate.DeepCopyInto(&out.VolumeClaimTemplate)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FileSpec.
func (in *FileSpec) DeepCopy() *FileSpec {
	if in == nil {
		return nil
	}
	out := new(FileSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GcsSpec) DeepCopyInto(out *GcsSpec) {
	*out = *in
	if in.CredentialSecretRef != nil {
		in, out := &in.CredentialSecretRef, &out.CredentialSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GcsSpec.
func (in *GcsSpec) DeepCopy() *GcsSpec {
	if in == nil {
		return nil
	}
	out := new(GcsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GoogleKmsGcsSpec) DeepCopyInto(out *GoogleKmsGcsSpec) {
	*out = *in
	if in.CredentialSecretRef != nil {
		in, out := &in.CredentialSecretRef, &out.CredentialSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GoogleKmsGcsSpec.
func (in *GoogleKmsGcsSpec) DeepCopy() *GoogleKmsGcsSpec {
	if in == nil {
		return nil
	}
	out := new(GoogleKmsGcsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InmemSpec) DeepCopyInto(out *InmemSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InmemSpec.
func (in *InmemSpec) DeepCopy() *InmemSpec {
	if in == nil {
		return nil
	}
	out := new(InmemSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JWTOIDCConfig) DeepCopyInto(out *JWTOIDCConfig) {
	*out = *in
	if in.AuditNonHMACRequestKeys != nil {
		in, out := &in.AuditNonHMACRequestKeys, &out.AuditNonHMACRequestKeys
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AuditNonHMACResponseKeys != nil {
		in, out := &in.AuditNonHMACResponseKeys, &out.AuditNonHMACResponseKeys
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.PassthroughRequestHeaders != nil {
		in, out := &in.PassthroughRequestHeaders, &out.PassthroughRequestHeaders
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.CredentialSecretRef != nil {
		in, out := &in.CredentialSecretRef, &out.CredentialSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.TLSSecretRef != nil {
		in, out := &in.TLSSecretRef, &out.TLSSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.ProviderConfig != nil {
		in, out := &in.ProviderConfig, &out.ProviderConfig
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.JWTValidationPubkeys != nil {
		in, out := &in.JWTValidationPubkeys, &out.JWTValidationPubkeys
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.JWTSupportedAlgs != nil {
		in, out := &in.JWTSupportedAlgs, &out.JWTSupportedAlgs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JWTOIDCConfig.
func (in *JWTOIDCConfig) DeepCopy() *JWTOIDCConfig {
	if in == nil {
		return nil
	}
	out := new(JWTOIDCConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubernetesConfig) DeepCopyInto(out *KubernetesConfig) {
	*out = *in
	if in.AuditNonHMACRequestKeys != nil {
		in, out := &in.AuditNonHMACRequestKeys, &out.AuditNonHMACRequestKeys
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AuditNonHMACResponseKeys != nil {
		in, out := &in.AuditNonHMACResponseKeys, &out.AuditNonHMACResponseKeys
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.PassthroughRequestHeaders != nil {
		in, out := &in.PassthroughRequestHeaders, &out.PassthroughRequestHeaders
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubernetesConfig.
func (in *KubernetesConfig) DeepCopy() *KubernetesConfig {
	if in == nil {
		return nil
	}
	out := new(KubernetesConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubernetesSecretSpec) DeepCopyInto(out *KubernetesSecretSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubernetesSecretSpec.
func (in *KubernetesSecretSpec) DeepCopy() *KubernetesSecretSpec {
	if in == nil {
		return nil
	}
	out := new(KubernetesSecretSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModeSpec) DeepCopyInto(out *ModeSpec) {
	*out = *in
	if in.KubernetesSecret != nil {
		in, out := &in.KubernetesSecret, &out.KubernetesSecret
		*out = new(KubernetesSecretSpec)
		**out = **in
	}
	if in.GoogleKmsGcs != nil {
		in, out := &in.GoogleKmsGcs, &out.GoogleKmsGcs
		*out = new(GoogleKmsGcsSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.AwsKmsSsm != nil {
		in, out := &in.AwsKmsSsm, &out.AwsKmsSsm
		*out = new(AwsKmsSsmSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.AzureKeyVault != nil {
		in, out := &in.AzureKeyVault, &out.AzureKeyVault
		*out = new(AzureKeyVault)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModeSpec.
func (in *ModeSpec) DeepCopy() *ModeSpec {
	if in == nil {
		return nil
	}
	out := new(ModeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MySQLSpec) DeepCopyInto(out *MySQLSpec) {
	*out = *in
	if in.CredentialSecretRef != nil {
		in, out := &in.CredentialSecretRef, &out.CredentialSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.TLSSecretRef != nil {
		in, out := &in.TLSSecretRef, &out.TLSSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.DatabaseRef != nil {
		in, out := &in.DatabaseRef, &out.DatabaseRef
		*out = new(v1alpha1.AppReference)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MySQLSpec.
func (in *MySQLSpec) DeepCopy() *MySQLSpec {
	if in == nil {
		return nil
	}
	out := new(MySQLSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NamedServiceTemplateSpec) DeepCopyInto(out *NamedServiceTemplateSpec) {
	*out = *in
	in.ServiceTemplateSpec.DeepCopyInto(&out.ServiceTemplateSpec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NamedServiceTemplateSpec.
func (in *NamedServiceTemplateSpec) DeepCopy() *NamedServiceTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(NamedServiceTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgreSQLSpec) DeepCopyInto(out *PostgreSQLSpec) {
	*out = *in
	if in.CredentialSecretRef != nil {
		in, out := &in.CredentialSecretRef, &out.CredentialSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.DatabaseRef != nil {
		in, out := &in.DatabaseRef, &out.DatabaseRef
		*out = new(v1alpha1.AppReference)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgreSQLSpec.
func (in *PostgreSQLSpec) DeepCopy() *PostgreSQLSpec {
	if in == nil {
		return nil
	}
	out := new(PostgreSQLSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RaftSpec) DeepCopyInto(out *RaftSpec) {
	*out = *in
	if in.TrailingLogs != nil {
		in, out := &in.TrailingLogs, &out.TrailingLogs
		*out = new(int64)
		**out = **in
	}
	if in.SnapshotThreshold != nil {
		in, out := &in.SnapshotThreshold, &out.SnapshotThreshold
		*out = new(int64)
		**out = **in
	}
	if in.MaxEntrySize != nil {
		in, out := &in.MaxEntrySize, &out.MaxEntrySize
		*out = new(int64)
		**out = **in
	}
	if in.Storage != nil {
		in, out := &in.Storage, &out.Storage
		*out = new(v1.PersistentVolumeClaimSpec)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RaftSpec.
func (in *RaftSpec) DeepCopy() *RaftSpec {
	if in == nil {
		return nil
	}
	out := new(RaftSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *S3Spec) DeepCopyInto(out *S3Spec) {
	*out = *in
	if in.CredentialSecretRef != nil {
		in, out := &in.CredentialSecretRef, &out.CredentialSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new S3Spec.
func (in *S3Spec) DeepCopy() *S3Spec {
	if in == nil {
		return nil
	}
	out := new(S3Spec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretEngineNamespaces) DeepCopyInto(out *SecretEngineNamespaces) {
	*out = *in
	if in.From != nil {
		in, out := &in.From, &out.From
		*out = new(FromNamespaces)
		**out = **in
	}
	if in.Selector != nil {
		in, out := &in.Selector, &out.Selector
		*out = new(metav1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretEngineNamespaces.
func (in *SecretEngineNamespaces) DeepCopy() *SecretEngineNamespaces {
	if in == nil {
		return nil
	}
	out := new(SecretEngineNamespaces)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SwiftSpec) DeepCopyInto(out *SwiftSpec) {
	*out = *in
	if in.CredentialSecretRef != nil {
		in, out := &in.CredentialSecretRef, &out.CredentialSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SwiftSpec.
func (in *SwiftSpec) DeepCopy() *SwiftSpec {
	if in == nil {
		return nil
	}
	out := new(SwiftSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TLSPolicy) DeepCopyInto(out *TLSPolicy) {
	*out = *in
	if in.CABundle != nil {
		in, out := &in.CABundle, &out.CABundle
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TLSPolicy.
func (in *TLSPolicy) DeepCopy() *TLSPolicy {
	if in == nil {
		return nil
	}
	out := new(TLSPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UnsealerSpec) DeepCopyInto(out *UnsealerSpec) {
	*out = *in
	in.Mode.DeepCopyInto(&out.Mode)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UnsealerSpec.
func (in *UnsealerSpec) DeepCopy() *UnsealerSpec {
	if in == nil {
		return nil
	}
	out := new(UnsealerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultServer) DeepCopyInto(out *VaultServer) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultServer.
func (in *VaultServer) DeepCopy() *VaultServer {
	if in == nil {
		return nil
	}
	out := new(VaultServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VaultServer) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultServerList) DeepCopyInto(out *VaultServerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VaultServer, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultServerList.
func (in *VaultServerList) DeepCopy() *VaultServerList {
	if in == nil {
		return nil
	}
	out := new(VaultServerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VaultServerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultServerSpec) DeepCopyInto(out *VaultServerSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.ConfigSecret != nil {
		in, out := &in.ConfigSecret, &out.ConfigSecret
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.DataSources != nil {
		in, out := &in.DataSources, &out.DataSources
		*out = make([]v1.VolumeSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(apiv1.TLSConfig)
		(*in).DeepCopyInto(*out)
	}
	in.Backend.DeepCopyInto(&out.Backend)
	if in.Unsealer != nil {
		in, out := &in.Unsealer, &out.Unsealer
		*out = new(UnsealerSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.AuthMethods != nil {
		in, out := &in.AuthMethods, &out.AuthMethods
		*out = make([]AuthMethod, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Monitor != nil {
		in, out := &in.Monitor, &out.Monitor
		*out = new(monitoringagentapiapiv1.AgentSpec)
		(*in).DeepCopyInto(*out)
	}
	in.PodTemplate.DeepCopyInto(&out.PodTemplate)
	if in.ServiceTemplates != nil {
		in, out := &in.ServiceTemplates, &out.ServiceTemplates
		*out = make([]NamedServiceTemplateSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.AllowedSecretEngines != nil {
		in, out := &in.AllowedSecretEngines, &out.AllowedSecretEngines
		*out = new(AllowedSecretEngines)
		(*in).DeepCopyInto(*out)
	}
	in.HealthChecker.DeepCopyInto(&out.HealthChecker)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultServerSpec.
func (in *VaultServerSpec) DeepCopy() *VaultServerSpec {
	if in == nil {
		return nil
	}
	out := new(VaultServerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultServerStatus) DeepCopyInto(out *VaultServerStatus) {
	*out = *in
	in.VaultStatus.DeepCopyInto(&out.VaultStatus)
	if in.UpdatedNodes != nil {
		in, out := &in.UpdatedNodes, &out.UpdatedNodes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]apiv1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.AuthMethodStatus != nil {
		in, out := &in.AuthMethodStatus, &out.AuthMethodStatus
		*out = make([]AuthMethodStatus, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultServerStatus.
func (in *VaultServerStatus) DeepCopy() *VaultServerStatus {
	if in == nil {
		return nil
	}
	out := new(VaultServerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultStatus) DeepCopyInto(out *VaultStatus) {
	*out = *in
	if in.Standby != nil {
		in, out := &in.Standby, &out.Standby
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Sealed != nil {
		in, out := &in.Sealed, &out.Sealed
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Unsealed != nil {
		in, out := &in.Unsealed, &out.Unsealed
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultStatus.
func (in *VaultStatus) DeepCopy() *VaultStatus {
	if in == nil {
		return nil
	}
	out := new(VaultStatus)
	in.DeepCopyInto(out)
	return out
}
