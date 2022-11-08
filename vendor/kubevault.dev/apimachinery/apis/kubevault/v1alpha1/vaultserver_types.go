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
	"time"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

const (
	ResourceKindVaultServer = "VaultServer"
	ResourceVaultServer     = "vaultserver"
	ResourceVaultServers    = "vaultservers"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=vaultservers,singular=vaultserver,shortName=vs,categories={vault,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Replicas",type="string",JSONPath=".spec.replicas"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type VaultServer struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VaultServerSpec   `json:"spec,omitempty"`
	Status            VaultServerStatus `json:"status,omitempty"`
}

type VaultServerSpec struct {
	// Version of VaultServer to be deployed.
	Version string `json:"version"`

	// Number of instances to deploy for a VaultServer.
	Replicas *int32 `json:"replicas,omitempty"`

	// ConfigSecret is an optional field to provide extra configuration for vault.
	// This secret contain extra config for vault
	// File name should be 'vault.hcl'.
	// If specified, this file will be appended to the controller configuration file.
	// +optional
	ConfigSecret *core.LocalObjectReference `json:"configSecret,omitempty"`

	// DataSources is a list of Configmaps/Secrets in the same namespace as the VaultServer
	// object, which shall be mounted into the VaultServer Pods.
	// The data are mounted into /etc/vault/data/<name>.
	// The first data will be named as "data-0", second one will be named as "data-1" and so on.
	// +optional
	DataSources []core.VolumeSource `json:"dataSources,omitempty"`

	// TLS policy of vault nodes
	// +optional
	TLS *kmapi.TLSConfig `json:"tls,omitempty"`

	// backend storage configuration for vault
	Backend BackendStorageSpec `json:"backend"`

	// Unsealer configuration for vault
	// +optional
	Unsealer *UnsealerSpec `json:"unsealer,omitempty"`

	// Specifies the list of auth methods to enable
	// +optional
	AuthMethods []AuthMethod `json:"authMethods,omitempty"`

	// Monitor is used monitor database instance
	// +optional
	Monitor *mona.AgentSpec `json:"monitor,omitempty"`

	// PodTemplate is an optional configuration for pods used to run vault
	// +optional
	PodTemplate ofst.PodTemplateSpec `json:"podTemplate,omitempty"`

	// ServiceTemplates is an optional configuration for services used to expose database
	// +optional
	ServiceTemplates []NamedServiceTemplateSpec `json:"serviceTemplates,omitempty"`

	// Indicates that the vault server is halted and all offshoot Kubernetes resources except PVCs are deleted.
	// +optional
	Halted bool `json:"halted,omitempty"`

	// TerminationPolicy controls the delete operation for vault server
	// +optional
	TerminationPolicy TerminationPolicy `json:"terminationPolicy,omitempty"`

	// AllowedSecretEngines defines the types of Secret Engines that MAY be attached to a
	// Listener and the trusted namespaces where those Route resources MAY be
	// present.
	//
	// Although a client request may match multiple route rules, only one rule
	// may ultimately receive the request. Matching precedence MUST be
	// determined in order of the following criteria:
	//
	// * The most specific match as defined by the Route type.
	// * The oldest Route based on creation timestamp. For example, a Route with
	//   a creation timestamp of "2020-09-08 01:02:03" is given precedence over
	//   a Route with a creation timestamp of "2020-09-08 01:02:04".
	// * If everything else is equivalent, the Route appearing first in
	//   alphabetical order (namespace/name) should be given precedence. For
	//   example, foo/bar is given precedence over foo/baz.
	//
	// All valid rules within a Route attached to this Listener should be
	// implemented. Invalid Route rules can be ignored (sometimes that will mean
	// the full Route). If a Route rule transitions from valid to invalid,
	// support for that Route rule should be dropped to ensure consistency. For
	// example, even if a filter specified by a Route rule is invalid, the rest
	// of the rules within that Route should still be supported.
	//
	// Support: Core
	// +kubebuilder:default={namespaces:{from: Same}}
	// +optional
	AllowedSecretEngines *AllowedSecretEngines `json:"allowedSecretEngines,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type VaultServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultServer `json:"items,omitempty"`
}

type VaultServerStatus struct {
	// ObservedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Phase indicates the state this Vault server jumps in.
	// +optional
	Phase VaultServerPhase `json:"phase,omitempty"`

	// Initialized indicates if the Vault service is initialized.
	// +optional
	Initialized bool `json:"initialized,omitempty"`

	// ServiceName is the LB service for accessing vault nodes.
	// +optional
	ServiceName string `json:"serviceName,omitempty"`

	// ClientPort is the port for vault client to access.
	// It's the same on client LB service and vault nodes.
	// +optional
	ClientPort int64 `json:"clientPort,omitempty"`

	// VaultStatus is the set of Vault node specific statuses: Active, Standby, and Sealed
	// +optional
	VaultStatus VaultStatus `json:"vaultStatus,omitempty"`

	// PodNames of updated Vault nodes. Updated means the Vault container image version
	// matches the spec's version.
	// +optional
	UpdatedNodes []string `json:"updatedNodes,omitempty"`

	// Represents the latest available observations of a VaultServer current state.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty"`

	// Status of the vault auth methods
	// +optional
	AuthMethodStatus []AuthMethodStatus `json:"authMethodStatus,omitempty"`
}

// AllowedSecretEngines defines which Secret Engines may be attached to this Listener.
type AllowedSecretEngines struct {
	// Namespaces indicates namespaces from which Secret Engines may be attached to this
	// Listener. This is restricted to the namespace of this VaultServer by default.
	//
	// +optional
	// +kubebuilder:default={from: Same}
	Namespaces *SecretEngineNamespaces `json:"namespaces,omitempty"`

	// SecretEngines specifies the types of Secret Engines that are allowed to bind
	// to this VaultServer. When unspecified or empty, all types of Secret Engines
	// are allowed.
	//
	// +optional
	SecretEngines []SecretEngineType `json:"secretEngines,omitempty"`
}

// +kubebuilder:validation:Enum=kv;pki;aws;azure;gcp;postgres;mongodb;mysql;elasticsearch
type SecretEngineType string

const (
	SecretEngineTypeKV            SecretEngineType = "kv"
	SecretEngineTypePKI           SecretEngineType = "pki"
	SecretEngineTypeAWS           SecretEngineType = "aws"
	SecretEngineTypeAzure         SecretEngineType = "azure"
	SecretEngineTypeGCP           SecretEngineType = "gcp"
	SecretEngineTypePostgres      SecretEngineType = "postgres"
	SecretEngineTypeMongoDB       SecretEngineType = "mongodb"
	SecretEngineTypeMySQL         SecretEngineType = "mysql"
	SecretEngineTypeElasticsearch SecretEngineType = "elasticsearch"
)

// FromNamespaces specifies namespace from which Secret Engines may be attached to a
// VaultServer.
//
// +kubebuilder:validation:Enum=All;Selector;Same
type FromNamespaces string

const (
	// Secret Engines in all namespaces may be attached to this VaultServer.
	NamespacesFromAll FromNamespaces = "All"
	// Only Secret Engines in namespaces selected by the selector may be attached to
	// this VaultServer.
	NamespacesFromSelector FromNamespaces = "Selector"
	// Only Secret Engines in the same namespace as the VaultServer may be attached to this
	// VaultServer.
	NamespacesFromSame FromNamespaces = "Same"
)

// SecretEngineNamespaces indicate which namespaces Secret Engines should be selected from.
type SecretEngineNamespaces struct {
	// From indicates where Secret Engines will be selected for this VaultServer. Possible
	// values are:
	// * All: Secret Engines in all namespaces may be used by this VaultServer.
	// * Selector: Secret Engines in namespaces selected by the selector may be used by
	//   this VaultServer.
	// * Same: Only Secret Engines in the same namespace may be used by this VaultServer.
	//
	// +optional
	// +kubebuilder:default=Same
	From *FromNamespaces `json:"from,omitempty"`

	// Selector must be specified when From is set to "Selector". In that case,
	// only Secret Engines in Namespaces matching this Selector will be selected by this
	// VaultServer. This field is ignored for other values of "From".
	//
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
}

type VaultStatus struct {
	// PodName of the active Vault node. Active node is unsealed.
	// Only active node can serve requests.
	// Vault service only points to the active node.
	// +optional
	Active string `json:"active,omitempty"`

	// PodNames of the standby Vault nodes. Standby nodes are unsealed.
	// Standby nodes do not process requests, and instead redirect to the active Vault.
	// +optional
	Standby []string `json:"standby,omitempty"`

	// PodNames of Sealed Vault nodes. Sealed nodes MUST be unsealed to
	// become standby or leader.
	// +optional
	Sealed []string `json:"sealed,omitempty"`

	// PodNames of Unsealed Vault nodes.
	// +optional
	Unsealed []string `json:"unsealed,omitempty"`
}

// TLSPolicy defines the TLS policy of the vault nodes
// If this is not set, operator will auto-gen TLS assets and secrets.
type TLSPolicy struct {
	// TLSSecret is the secret containing TLS certs used by each vault node
	// for the communication between the vault server and its clients.
	// The secret should contain three files:
	// 	- tls.crt
	// 	- tls.key
	//
	// The server certificate must allow the following wildcard domains:
	// 	- localhost
	// 	- *.<namespace>.pod
	// 	- <vaultServer-name>.<namespace>.svc
	TLSSecret string `json:"tlsSecret"`

	// CABundle is a PEM encoded CA bundle which will be used to validate the serving certificate.
	// +optional
	CABundle []byte `json:"caBundle,omitempty"`
}

// TODO : set defaults and validation
// BackendStorageSpec defines storage backend configuration of vault
type BackendStorageSpec struct {
	// ref: https://www.vaultproject.io/docs/configuration/storage/in-memory.html
	// +optional
	Inmem *InmemSpec `json:"inmem,omitempty"`

	// +optional
	Etcd *EtcdSpec `json:"etcd,omitempty"`

	// +optional
	Gcs *GcsSpec `json:"gcs,omitempty"`

	// +optional
	S3 *S3Spec `json:"s3,omitempty"`

	// +optional
	Azure *AzureSpec `json:"azure,omitempty"`

	// +optional
	PostgreSQL *PostgreSQLSpec `json:"postgresql,omitempty"`

	// +optional
	MySQL *MySQLSpec `json:"mysql,omitempty"`

	// +optional
	File *FileSpec `json:"file,omitempty"`

	// +optional
	DynamoDB *DynamoDBSpec `json:"dynamodb,omitempty"`

	// +optional
	Swift *SwiftSpec `json:"swift,omitempty"`

	// +optional
	Consul *ConsulSpec `json:"consul,omitempty"`

	// +optional
	Raft *RaftSpec `json:"raft,omitempty"`
}

// ref: https://www.vaultproject.io/docs/configuration/storage/consul.html
//
// ConsulSpec defines the configuration to set up consul as backend storage in vault
type ConsulSpec struct {
	// Specifies the address of the Consul agent to communicate with.
	// This can be an IP address, DNS record, or unix socket.
	// +optional
	Address string `json:"address,omitempty"`

	// Specifies the check interval used to send health check information
	// back to Consul.
	// This is specified using a label suffix like "30s" or "1h".
	// +optional
	CheckTimeout string `json:"checkTimeout,omitempty"`

	// Specifies the Consul consistency mode.
	// Possible values are "default" or "strong".
	// +optional
	ConsistencyMode string `json:"consistencyMode,omitempty"`

	// Specifies whether Vault should register itself with Consul.
	// Possible values are "true" or "false"
	// +optional
	DisableRegistration string `json:"disableRegistration,omitempty"`

	// Specifies the maximum number of concurrent requests to Consul.
	// +optional
	MaxParallel string `json:"maxParallel,omitempty"`

	// Specifies the path in Consul's key-value store
	// where Vault data will be stored.
	// +optional
	Path string `json:"path,omitempty"`

	// Specifies the scheme to use when communicating with Consul.
	// This can be set to "http" or "https".
	// +optional
	Scheme string `json:"scheme,omitempty"`

	// Specifies the name of the service to register in Consul.
	// +optional
	Service string `json:"service,omitempty"`

	// Specifies a comma-separated list of tags
	// to attach to the service registration in Consul.
	// +optional
	ServiceTags string `json:"serviceTags,omitempty"`

	// Specifies a service-specific address to set on the service registration
	// in Consul.
	// If unset, Vault will use what it knows to be the HA redirect address
	// - which is usually desirable.
	// Setting this parameter to "" will tell Consul to leverage the configuration
	// of the node the service is registered on dynamically.
	// +optional
	ServiceAddress string `json:"serviceAddress,omitempty"`

	// Specifies the secret name that contains ACL token with permission
	// to read and write from the path in Consul's key-value store.
	// secret data:
	//  - aclToken:<value>
	// +optional
	ACLTokenSecretName string `json:"aclTokenSecretName,omitempty"`

	// Specifies the minimum allowed session TTL.
	// Consul server has a lower limit of 10s on the session TTL by default.
	// +optional
	SessionTTL string `json:"sessionTTL,omitempty"`

	// Specifies the wait time before a lock lock acquisition is made.
	// This affects the minimum time it takes to cancel a lock acquisition.
	// +optional
	LockWaitTime string `json:"lockWaitTime,omitempty"`

	// Specifies the secret name that contains tls_ca_file, tls_cert_file and tls_key_file
	// for consul communication
	// Secret data:
	//  - ca.crt
	//  - client.crt
	//  - client.key
	// +optional
	TLSSecretName string `json:"tlsSecretName,omitempty"`

	// Specifies the minimum TLS version to use.
	// Accepted values are "tls10", "tls11" or "tls12".
	// +optional
	TLSMinVersion string `json:"tlsMinVersion,omitempty"`

	// Specifies if the TLS host verification should be disabled.
	// It is highly discouraged that you disable this option.
	// +optional
	TLSSkipVerify bool `json:"tlsSkipVerify,omitempty"`
}

// ref: https://www.vaultproject.io/docs/configuration/storage/in-memory.html
type InmemSpec struct{}

// TODO : set defaults and validation
// vault doc: https://www.vaultproject.io/docs/configuration/storage/etcd.html
//
// EtcdSpec defines configuration to set up etcd as backend storage in vault
type EtcdSpec struct {
	// Specifies the addresses of the etcd instances
	Address string `json:"address"`

	// Specifies the version of the API to communicate with etcd
	// +optional
	EtcdApi string `json:"etcdApi,omitempty"`

	// Specifies if high availability should be enabled
	// +optional
	HAEnable bool `json:"haEnable,omitempty"`

	// Specifies the path in etcd where vault data will be stored
	// +optional
	Path string `json:"path,omitempty"`

	// Specifies whether to sync list of available etcd services on startup
	// +optional
	Sync bool `json:"sync,omitempty"`

	// Specifies the domain name to query for SRV records describing cluster endpoints
	// +optional
	DiscoverySrv string `json:"discoverySrv,omitempty"`

	// Specifies the secret name that contain username and password to use when authenticating with the etcd server
	// secret data:
	//  - username:<value>
	//  - password:<value>
	// +optional
	CredentialSecretName string `json:"credentialSecretName,omitempty"`

	// Specifies the secret name that contains tls_ca_file, tls_cert_file and tls_key_file for etcd communication
	// secret data:
	//  - ca.crt
	//  - client.crt
	//  - client.key
	// +optional
	TLSSecretName string `json:"tlsSecretName,omitempty"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/google-cloud-storage.html
//
// GcsSpec defines configuration to set up Google Cloud Storage as backend storage in vault
type GcsSpec struct {
	// Specifies the name of the bucket to use for storage.
	Bucket string `json:"bucket"`

	// Specifies the maximum size (in kilobytes) to send in a single request. If set to 0,
	// it will attempt to send the whole object at once, but will not retry any failures.
	// +optional
	ChunkSize string `json:"chunkSize,omitempty"`

	//  Specifies the maximum number of parallel operations to take place.
	// +optional
	MaxParallel int64 `json:"maxParallel,omitempty"`

	// Specifies if high availability mode is enabled.
	// +optional
	HAEnabled bool `json:"haEnabled,omitempty"`

	// Secret containing Google application credential
	// secret data:
	//  - sa.json:<value>
	// +optional
	CredentialSecret string `json:"credentialSecret,omitempty"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/s3.html
//
// S3Spec defines configuration to set up Amazon S3 Storage as backend storage in vault
type S3Spec struct {
	// Specifies the name of the bucket to use for storage.
	Bucket string `json:"bucket"`

	// Specifies an alternative, AWS compatible, S3 endpoint.
	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// Specifies the AWS region
	// +optional
	Region string `json:"region,omitempty"`

	// Specifies the secret name containing AWS access key and AWS secret key
	// secret data:
	//  - access_key=<value>
	//  - secret_key=<value>
	// +optional
	CredentialSecret string `json:"credentialSecret,omitempty"`

	// Specifies the secret name containing AWS session token
	// secret data:
	//  - session_token:<value>
	// +optional
	SessionTokenSecret string `json:"sessionTokenSecret,omitempty"`

	// Specifies the maximum number of parallel operations to take place.
	// +optional
	MaxParallel int64 `json:"maxParallel,omitempty"`

	// Specifies whether to use host bucket style domains with the configured endpoint.
	// +optional
	ForcePathStyle bool `json:"forcePathStyle,omitempty"`

	// Specifies if SSL should be used for the endpoint connection
	// +optional
	DisableSSL bool `json:"disableSSL,omitempty"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/azure.html
//
// AzureSpec defines configuration to set up Google Cloud Storage as backend storage in vault
type AzureSpec struct {
	// Specifies the Azure Storage account name.
	AccountName string `json:"accountName"`

	// Specifies the secret containing Azure Storage account key.
	// secret data:
	//  - account_key:<value>
	AccountKeySecret string `json:"accountKeySecret"`

	// Specifies the Azure Storage Blob container name.
	Container string `json:"container"`

	//  Specifies the maximum number of concurrent operations to take place.
	// +optional
	MaxParallel int64 `json:"maxParallel,omitempty"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/postgresql.html
//
// PostgreSQLSpec defines configuration to set up PostgreSQL storage as backend storage in vault
type PostgreSQLSpec struct {
	// Specifies the name of the secret containing the connection string to use to authenticate and connect to PostgreSQL.
	// A full list of supported parameters can be found in the pq library documentation(https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters).
	// secret data:
	//  - connection_url:<data>
	ConnectionURLSecret string `json:"connectionURLSecret"`

	// Specifies the name of the table in which to write Vault data.
	// This table must already exist (Vault will not attempt to create it).
	// +optional
	Table string `json:"table,omitempty"`

	//  Specifies the maximum number of concurrent requests to take place.
	// +optional
	MaxParallel int64 `json:"maxParallel,omitempty"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/mysql.html
//
// MySQLSpec defines configuration to set up MySQL Storage as backend storage in vault
type MySQLSpec struct {
	// Specifies the address of the MySQL host.
	// +optional
	Address string `json:"address"`

	// Specifies the name of the database. If the database does not exist, Vault will attempt to create it.
	// +optional
	Database string `json:"database,omitempty"`

	// Specifies the name of the table. If the table does not exist, Vault will attempt to create it.
	// +optional
	Table string `json:"table,omitempty"`

	// Specifies the MySQL username and password to connect to the database
	// secret data:
	//  - username=<value>
	//  - password=<value>
	UserCredentialSecret string `json:"userCredentialSecret"`

	// Specifies the name of the secret containing the CA certificate to connect using TLS.
	// secret data:
	//  - tls_ca_file=<ca_cert>
	// +optional
	TLSCASecret string `json:"tlsCASecret,omitempty"`

	//  Specifies the maximum number of concurrent requests to take place.
	// +optional
	MaxParallel int64 `json:"maxParallel,omitempty"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/filesystem.html
//
// FileSpec defines configuration to set up File system Storage as backend storage in vault
type FileSpec struct {
	// The absolute path on disk to the directory where the data will be stored.
	// If the directory does not exist, Vault will create it.
	Path string `json:"path"`

	// volumeClaimTemplate is a claim that pods are allowed to reference.
	// The VaultServer controller is responsible for deploying the claim
	// and update the volumeMounts in the Vault server container in the template.
	VolumeClaimTemplate ofst.PersistentVolumeClaim `json:"volumeClaimTemplate"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/dynamodb.html
//
// DynamoDBSpec defines configuration to set up DynamoDB Storage as backend storage in vault
type DynamoDBSpec struct {
	// Specifies an alternative, AWS compatible, DynamoDB endpoint.
	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// Specifies the AWS region
	// +optional
	Region string `json:"region,omitempty"`

	// Specifies whether this backend should be used to run Vault in high availability mode.
	// +optional
	HaEnabled bool `json:"haEnabled,omitempty"`

	// Specifies the maximum number of reads consumed per second on the table
	// +optional
	ReadCapacity int64 `json:"readCapacity,omitempty"`

	// Specifies the maximum number of writes performed per second on the table.
	// +optional
	WriteCapacity int64 `json:"writeCapacity,omitempty"`

	// Specifies the name of the DynamoDB table in which to store Vault data.
	// If the specified table does not yet exist, it will be created during initialization.
	// default: vault-dynamodb-backend
	// +optional
	Table string `json:"table,omitempty"`

	// Specifies the secret name containing AWS access key and AWS secret key
	// secret data:
	//  - access_key=<value>
	//  - secret_key=<value>
	// +optional
	CredentialSecret string `json:"credentialSecret,omitempty"`

	// Specifies the secret name containing AWS session token
	// secret data:
	//  - session_token:<value>
	// +optional
	SessionTokenSecret string `json:"sessionTokenSecret,omitempty"`

	// Specifies the maximum number of parallel operations to take place.
	// +optional
	MaxParallel int64 `json:"maxParallel,omitempty"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/swift.html
//
// SwiftSpec defines configuration to set up Swift Storage as backend storage in vault
type SwiftSpec struct {
	// Specifies the OpenStack authentication endpoint.
	AuthURL string `json:"authURL"`

	// Specifies the name of the Swift container.
	Container string `json:"container"`

	// Specifies the name of the secret containing the OpenStack account/username and password
	// secret data:
	//  - username=<value>
	//  - password=<value>
	CredentialSecret string `json:"credentialSecret"`

	// Specifies the name of the tenant. If left blank, this will default to the default tenant of the username.
	// +optional
	Tenant string `json:"tenant,omitempty"`

	// Specifies the name of the region.
	// +optional
	Region string `json:"region,omitempty"`

	// Specifies the id of the tenant.
	// +optional
	TenantID string `json:"tenantID,omitempty"`

	// Specifies the name of the user domain.
	// +optional
	Domain string `json:"domain,omitempty"`

	// Specifies the name of the project's domain.
	// +optional
	ProjectDomain string `json:"projectDomain,omitempty"`

	// Specifies the id of the trust.
	// +optional
	TrustID string `json:"trustID,omitempty"`

	// Specifies storage URL from alternate authentication.
	// +optional
	StorageURL string `json:"storageURL,omitempty"`

	// Specifies secret containing auth token from alternate authentication.
	// secret data:
	//  - auth_token=<value>
	// +optional
	AuthTokenSecret string `json:"authTokenSecret,omitempty"`

	//  Specifies the maximum number of concurrent requests to take place.
	// +optional
	MaxParallel int64 `json:"maxParallel,omitempty"`
}

// UnsealerSpec contain the configuration for auto vault initialize/unseal
type UnsealerSpec struct {
	// Total count of secret shares that exist
	// +optional
	SecretShares int64 `json:"secretShares,omitempty"`

	// Minimum required secret shares to unseal
	// +optional
	SecretThreshold int64 `json:"secretThreshold,omitempty"`

	// How often to attempt to unseal the vault instance
	// +optional
	RetryPeriodSeconds time.Duration `json:"retryPeriodSeconds,omitempty"`

	// overwrite existing unseal keys and root tokens, possibly dangerous!
	// +optional
	OverwriteExisting bool `json:"overwriteExisting,omitempty"`

	// should the root token be stored in the key store (default true)
	// +optional
	StoreRootToken bool `json:"storeRootToken,omitempty"`

	// mode contains unseal mechanism
	// +optional
	Mode ModeSpec `json:"mode,omitempty"`
}

// ModeSpec contain unseal mechanism
type ModeSpec struct {
	// +optional
	KubernetesSecret *KubernetesSecretSpec `json:"kubernetesSecret,omitempty"`

	// +optional
	GoogleKmsGcs *GoogleKmsGcsSpec `json:"googleKmsGcs,omitempty"`

	// +optional
	AwsKmsSsm *AwsKmsSsmSpec `json:"awsKmsSsm,omitempty"`

	// +optional
	AzureKeyVault *AzureKeyVault `json:"azureKeyVault,omitempty"`
}

// KubernetesSecretSpec contain the fields that required to unseal using kubernetes secret
type KubernetesSecretSpec struct {
	SecretName string `json:"secretName"`
}

// GoogleKmsGcsSpec contain the fields that required to unseal vault using google kms
type GoogleKmsGcsSpec struct {
	// The name of the Google Cloud KMS crypto key to use
	KmsCryptoKey string `json:"kmsCryptoKey"`

	// The name of the Google Cloud KMS key ring to use
	KmsKeyRing string `json:"kmsKeyRing"`

	// The Google Cloud KMS location to use (eg. 'global', 'europe-west1')
	KmsLocation string `json:"kmsLocation"`

	// The Google Cloud KMS project to use
	KmsProject string `json:"kmsProject"`

	// The name of the Google Cloud Storage bucket to store values in
	Bucket string `json:"bucket"`

	// Secret containing Google application credential
	// secret data:
	//  - sa.json:<value>
	// +optional
	CredentialSecret string `json:"credentialSecret,omitempty"`
}

// AwsKmsSsmSpec contain the fields that required to unseal vault using aws kms ssm
type AwsKmsSsmSpec struct {
	// The ID or ARN of the AWS KMS key to encrypt values
	KmsKeyID string `json:"kmsKeyID"`

	// +optional
	// An optional Key prefix for SSM Parameter store
	SsmKeyPrefix string `json:"ssmKeyPrefix,omitempty"`

	Region string `json:"region,omitempty"`

	// Specifies the secret name containing AWS access key and AWS secret key
	// secret data:
	//  - access_key:<value>
	//  - secret_key:<value>
	// +optional
	CredentialSecret string `json:"credentialSecret,omitempty"`

	// Used to make AWS KMS requests. This is useful,
	// for example, when connecting to KMS over a VPC Endpoint.
	// If not set, Vault will use the default API endpoint for your region.
	Endpoint string `json:"endpoint,omitempty"`
}

// RaftSpec defines the configuration for the Raft integrated storage.
// https://www.vaultproject.io/docs/configuration/storage/raft
type RaftSpec struct {
	// Path (string: "") specifies the filesystem path where the vault data gets stored.
	// This value can be overridden by setting the VAULT_RAFT_PATH environment variable.
	// default: ""
	// +optional
	Path string `json:"path,omitempty"`

	// An integer multiplier used by servers to scale key Raft timing parameters.
	// Tuning this affects the time it takes Vault to detect leader failures and to perform leader elections,
	// at the expense of requiring more network and CPU resources for better performance.
	// default: 0
	// +optional
	PerformanceMultiplier int64 `json:"performanceMultiplier,omitempty"`

	// This controls how many log entries are left in the log store on disk after a snapshot is made.
	// default: 10000
	// +optional
	TrailingLogs *int64 `json:"trailingLogs,omitempty"`

	// This controls the minimum number of raft commit entries between snapshots that are saved to disk.
	// default: 8192
	// +optional
	SnapshotThreshold *int64 `json:"snapshotThreshold,omitempty"`

	// This configures the maximum number of bytes for a raft entry. It applies to both Put operations and transactions.
	// default: 1048576
	// +optional
	MaxEntrySize *int64 `json:"maxEntrySize,omitempty"`

	// This is the interval after which autopilot will pick up any state changes.
	// default: ""
	// +optional
	AutopilotReconcileInterval string `json:"autopilotReconcileInterval,omitempty"`

	// Storage to specify how storage shall be used.
	Storage *core.PersistentVolumeClaimSpec `json:"storage,omitempty"`
}

// AzureKeyVault contain the fields that required to unseal vault using azure key vault
type AzureKeyVault struct {
	// Azure key vault url, for example https://myvault.vault.azure.net
	VaultBaseURL string `json:"vaultBaseURL"`

	// The cloud environment identifier
	// default: "AZUREPUBLICCLOUD"
	// +optional
	Cloud string `json:"cloud,omitempty"`

	// The AAD Tenant ID
	TenantID string `json:"tenantID"`

	// Specifies the name of secret containing client cert and client cert password
	// secret data:
	//  - client-cert:<value>
	// 	- client-cert-password: <value>
	// +optional
	ClientCertSecret string `json:"clientCertSecret,omitempty"`

	// Specifies the name of secret containing client id and client secret of AAD application
	// secret data:
	//  - client-id:<value>
	//  - client-secret:<value>
	// +optional
	AADClientSecret string `json:"aadClientSecret,omitempty"`

	// Use managed service identity for the virtual machine
	// +optional
	UseManagedIdentity bool `json:"useManagedIdentity,omitempty"`
}

// +kubebuilder:validation:Enum=kubernetes;aws;gcp;userpass;cert;azure
type AuthMethodType string

const (
	AuthTypeKubernetes AuthMethodType = "kubernetes"
	AuthTypeAws        AuthMethodType = "aws"
	AuthTypeGcp        AuthMethodType = "gcp"
	AuthTypeUserPass   AuthMethodType = "userpass"
	AuthTypeCert       AuthMethodType = "cert"
	AuthTypeAzure      AuthMethodType = "azure"
)

// AuthMethod contains the information to enable vault auth method
// links: https://www.vaultproject.io/api/system/auth.html
type AuthMethod struct {
	//  Specifies the name of the authentication method type, such as "github" or "token".
	Type string `json:"type"`

	// Specifies the path in which to enable the auth method.
	// Default value is the same as the 'type'
	Path string `json:"path"`

	// Specifies a human-friendly description of the auth method.
	// +optional
	Description string `json:"description,omitempty"`

	// Specifies configuration options for this auth method.
	// +optional
	Config *AuthConfig `json:"config,omitempty"`

	// Specifies the name of the auth plugin to use based from the name in the plugin catalog.
	// Applies only to plugin methods.
	// +optional
	PluginName string `json:"pluginName,omitempty"`

	// Specifies if the auth method is a local only. Local auth methods are not replicated nor (if a secondary) removed by replication.
	// +optional
	Local bool `json:"local,omitempty"`
}

// +kubebuilder:validation:Enum=EnableSucceeded;EnableFailed;DisableSucceeded;DisableFailed
type AuthMethodEnableDisableStatus string

const (
	AuthMethodEnableSucceeded  AuthMethodEnableDisableStatus = "EnableSucceeded"
	AuthMethodEnableFailed     AuthMethodEnableDisableStatus = "EnableFailed"
	AuthMethodDisableSucceeded AuthMethodEnableDisableStatus = "DisableSucceeded"
	AuthMethodDisableFailed    AuthMethodEnableDisableStatus = "DisableFailed"
)

// AuthMethodStatus specifies the status of the auth method maintained by the auth method controller
type AuthMethodStatus struct {
	//  Specifies the name of the authentication method type, such as "github" or "token".
	Type string `json:"type"`

	// Specifies the path in which to enable the auth method.
	Path string `json:"path"`

	// Specifies whether auth method is enabled or not
	Status AuthMethodEnableDisableStatus `json:"status"`

	// Specifies the reason why failed to enable auth method
	// +optional
	Reason string `json:"reason,omitempty"`
}

type AuthConfig struct {
	// The default lease duration, specified as a string duration like "5s" or "30m".
	// +optional
	DefaultLeaseTTL string `json:"defaultLeaseTTL,omitempty"`

	// The maximum lease duration, specified as a string duration like "5s" or "30m".
	// +optional
	MaxLeaseTTL string `json:"maxLeaseTTL,omitempty"`

	// The name of the plugin in the plugin catalog to use.
	// +optional
	PluginName string `json:"pluginName,omitempty"`

	// List of keys that will not be HMAC'd by audit devices in the request data object.
	// +optional
	AuditNonHMACRequestKeys []string `json:"auditNonHMACRequestKeys,omitempty"`

	// List of keys that will not be HMAC'd by audit devices in the response data object.
	// +optional
	AuditNonHMACResponseKeys []string `json:"auditNonHMACResponseKeys,omitempty"`

	// Speficies whether to show this mount in the UI-specific listing endpoint.
	// +optional
	ListingVisibility string `json:"listingVisibility,omitempty"`

	// List of headers to whitelist and pass from the request to the backend.
	// +optional
	PassthroughRequestHeaders []string `json:"passthroughRequestHeaders,omitempty"`
}
