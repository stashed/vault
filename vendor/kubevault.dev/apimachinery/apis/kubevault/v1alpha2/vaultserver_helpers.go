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

package v1alpha2

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"kubevault.dev/apimachinery/apis"
	"kubevault.dev/apimachinery/apis/kubevault"
	"kubevault.dev/apimachinery/crds"

	promapi "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gomodules.xyz/pointer"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	appslister "k8s.io/client-go/listers/apps/v1"
	"k8s.io/utils/ptr"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/apiextensions"
	apps_util "kmodules.xyz/client-go/apps/v1"
	clustermeta "kmodules.xyz/client-go/cluster"
	meta_util "kmodules.xyz/client-go/meta"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

func (*VaultServer) Hub() {}

func (_ VaultServer) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourceVaultServers))
}

func (_ VaultServer) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourceVaultServers, kubevault.GroupName)
}

func (v VaultServer) GetKey() string {
	return v.Namespace + "/" + v.Name
}

func (v VaultServer) OffshootName() string {
	return v.Name
}

func (v VaultServer) ServiceAccountName() string {
	return v.Name
}

func (v VaultServer) ServiceAccountForTokenReviewer() string {
	return meta_util.NameWithSuffix(v.Name, "k8s-token-reviewer")
}

func (v VaultServer) PolicyNameForPolicyController() string {
	return meta_util.NameWithSuffix(v.Name, "policy-controller")
}

func (v VaultServer) PolicyNameForAuthMethodController() string {
	return meta_util.NameWithSuffix(v.Name, "auth-method-controller")
}

func (v VaultServer) PolicyNameForAuthMethod(typ AuthMethodType, path string) string {
	return fmt.Sprintf("%s-%s-auth-policy", string(typ), path)
}

func (v VaultServer) AppBindingName() string {
	return v.Name
}

func (v VaultServer) OffshootSelectors() map[string]string {
	return map[string]string{
		meta_util.NameLabelKey:      v.ResourceFQN(),
		meta_util.InstanceLabelKey:  v.Name,
		meta_util.ManagedByLabelKey: kubevault.GroupName,
	}
}

func (v VaultServer) OffshootLabels() map[string]string {
	return meta_util.FilterKeys("kubevault.com", v.OffshootSelectors(), v.Labels)
}

func (v VaultServer) ConfigSecretName() string {
	return meta_util.NameWithSuffix(v.Name, "vault-config")
}

func (v VaultServer) TLSSecretName() string {
	return meta_util.NameWithSuffix(v.Name, "vault-tls")
}

func (v VaultServer) IsValid() error {
	return nil
}

func (v VaultServer) StatsServiceName() string {
	return meta_util.NameWithSuffix(v.Name, "stats")
}

func (v VaultServer) ServiceName(alias ServiceAlias) string {
	if alias == VaultServerServiceVault {
		return v.Name
	}
	return meta_util.NameWithSuffix(v.Name, string(alias))
}

func (v VaultServer) StatsLabels() map[string]string {
	labels := v.OffshootLabels()
	labels["feature"] = "stats"
	return labels
}

// Returns the default certificate secret name for given alias.
func (vs *VaultServer) DefaultCertSecretName(alias string) string {
	return meta_util.NameWithSuffix(fmt.Sprintf("%s-%s", vs.Name, alias), "certs")
}

// Returns certificate secret name for given alias if exists,
// otherwise returns the default certificate secret name.
func (vs *VaultServer) GetCertSecretName(alias string) string {
	if vs.Spec.TLS != nil {
		sName, valid := kmapi.GetCertificateSecretName(vs.Spec.TLS.Certificates, alias)
		if valid {
			return sName
		}
	}

	return vs.DefaultCertSecretName(alias)
}

func (v VaultServer) StatsService() mona.StatsAccessor {
	return &vaultServerStatsService{&v}
}

type vaultServerStatsService struct {
	*VaultServer
}

func (v vaultServerStatsService) ServiceMonitorAdditionalLabels() map[string]string {
	return v.VaultServer.OffshootLabels()
}

func (v vaultServerStatsService) GetNamespace() string {
	return v.VaultServer.GetNamespace()
}

func (v vaultServerStatsService) ServiceName() string {
	return v.StatsServiceName()
}

func (v vaultServerStatsService) ServiceMonitorName() string {
	return v.ServiceName()
}

func (v vaultServerStatsService) Path() string {
	return "/v1/sys/metrics"
}

func (v vaultServerStatsService) Scheme() string {
	if v.Spec.TLS != nil {
		return "https"
	}
	return ""
}

func (v vaultServerStatsService) TLSConfig() *promapi.TLSConfig {
	if v.Spec.TLS != nil {
		return &promapi.TLSConfig{
			SafeTLSConfig: promapi.SafeTLSConfig{
				CA: promapi.SecretOrConfigMap{
					Secret: &core.SecretKeySelector{
						LocalObjectReference: core.LocalObjectReference{
							Name: v.VaultServer.GetCertSecretName(string(VaultServerCert)),
						},
						Key: core.TLSCertKey,
					},
				},
				ServerName: ptr.To(fmt.Sprintf("%s.%s.svc", v.VaultServer.ServiceName(VaultServerServiceVault), v.VaultServer.Namespace)),
			},
		}
	}
	return nil
}

func (vs *VaultServer) GetCertificateCN(alias VaultCertificateAlias) string {
	return fmt.Sprintf("%s-%s", vs.Name, string(alias))
}

func (vs *VaultServer) Scheme() string {
	if vs.Spec.TLS != nil {
		return "https"
	}
	return "http"
}

// UnsealKeyID is the ID that used as key name when storing unseal key
func (vs *VaultServer) UnsealKeyID(id int) string {
	return strings.Join([]string{vs.KeyPrefix(), fmt.Sprintf("unseal-key-%d", id)}, "-")
}

// RootTokenID is the ID that used as key name when storing root token
func (vs *VaultServer) RootTokenID() string {
	return strings.Join([]string{vs.KeyPrefix(), "root-token"}, "-")
}

func (vs *VaultServer) KeyPrefix() string {
	cluster := "-"
	if clustermeta.ClusterName() != "" {
		cluster = clustermeta.ClusterName()
	}
	return fmt.Sprintf("k8s.%s.%s.%s", cluster, vs.Namespace, vs.Name)
}

func (vsb *BackendStorageSpec) GetBackendType() (VaultServerBackend, error) {
	switch {
	case vsb.Inmem != nil:
		return VaultServerInmem, nil
	case vsb.Etcd != nil:
		return VaultServerEtcd, nil
	case vsb.Gcs != nil:
		return VaultServerGcs, nil
	case vsb.S3 != nil:
		return VaultServerS3, nil
	case vsb.Azure != nil:
		return VaultServerAzure, nil
	case vsb.PostgreSQL != nil:
		return VaultServerPostgreSQL, nil
	case vsb.MySQL != nil:
		return VaultServerMySQL, nil
	case vsb.File != nil:
		return VaultServerFile, nil
	case vsb.DynamoDB != nil:
		return VaultServerDynamoDB, nil
	case vsb.Swift != nil:
		return VaultServerSwift, nil
	case vsb.Consul != nil:
		return VaultServerConsul, nil
	case vsb.Raft != nil:
		return VaultServerRaft, nil
	default:
		return "", errors.New("unknown backened type")
	}
}

func (v *VaultServer) CertificateMountPath(alias VaultCertificateAlias) string {
	return filepath.Join(apis.CertificatePath, string(alias))
}

func (v *VaultServer) ReplicasAreReady(lister appslister.StatefulSetLister) (bool, string, error) {
	// Desired number of statefulSets
	expectedItems := 1
	return checkReplicas(lister.StatefulSets(v.Namespace), labels.SelectorFromSet(v.OffshootLabels()), expectedItems)
}

func checkReplicas(lister appslister.StatefulSetNamespaceLister, selector labels.Selector, expectedItems int) (bool, string, error) {
	items, err := lister.List(selector)
	if err != nil {
		return false, "", err
	}

	if len(items) < expectedItems {
		return false, fmt.Sprintf("All StatefulSets are not available. Desire number of StatefulSet: %d, Available: %d", expectedItems, len(items)), nil
	}

	// return isReplicasReady, message, error
	ready, msg := apps_util.StatefulSetsAreReady(items)
	return ready, msg, nil
}

// GetServiceTemplate returns a pointer to the desired serviceTemplate referred by "alias". Otherwise, it returns nil.
func (vs *VaultServer) GetServiceTemplate(alias ServiceAlias) ofst.ServiceTemplateSpec {
	templates := vs.Spec.ServiceTemplates
	for i := range templates {
		c := templates[i]
		if c.Alias == alias {
			return c.ServiceTemplateSpec
		}
	}
	return ofst.ServiceTemplateSpec{}
}

const (
	VaultServerAnnotationName      = "vaultservers.kubevault.com/name"
	VaultServerAnnotationNamespace = "vaultservers.kubevault.com/namespace"
)

func (vs *VaultServer) SetHealthCheckerDefaults() {
	if vs.Spec.HealthChecker.PeriodSeconds == nil {
		vs.Spec.HealthChecker.PeriodSeconds = pointer.Int32P(10)
	}
	if vs.Spec.HealthChecker.TimeoutSeconds == nil {
		vs.Spec.HealthChecker.TimeoutSeconds = pointer.Int32P(10)
	}
	if vs.Spec.HealthChecker.FailureThreshold == nil {
		vs.Spec.HealthChecker.FailureThreshold = pointer.Int32P(1)
	}
}

func (vs *VaultServer) BackupSecretName() string {
	return meta_util.NameWithSuffix(vs.Name, "backup-token")
}
