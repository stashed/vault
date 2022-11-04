/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pkg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	stash "stash.appscode.dev/apimachinery/client/clientset/versioned"
	"stash.appscode.dev/apimachinery/pkg/restic"

	"github.com/hashicorp/vault/api"
	shell "gomodules.xyz/go-sh"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	appcatalog "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcatalog_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
)

// commands:
// vault operator raft snapshot save backup.snap
// vault operator raft snapshot restore backup.snap

// vault operator raft snapshot restore -force backup.snap

// Command Options:
//
// -force
// Allow the operation to continue with no key=value pairs. This allows
// writing to keys that do not need or expect data. This is aliased as
// "-f". The default is false.

// flags:
// -address=<string>
// Address of the Vault server. The default is https://127.0.0.1:8200. This
// can also be specified via the VAULT_ADDR environment variable.
// env: VAULT_ADDR

// -tls-skip-verify
// Disable verification of TLS certificates. Using this option is highly
// discouraged as it decreases the security of data transmissions to and
// from the Vault server. The default is false. This can also be specified
// via the VAULT_SKIP_VERIFY environment variable.
// env: VAULT_SKIP_VERIFY=true

// VAULT_TOKEN=<token>

const (
	VaultToken        = "token"
	VaultSnapshotFile = "backup.snap"
	VaultCMD          = "vault"
	VaultTLSRootCA    = "ca.crt"
	// VaultSnapshotMD       = ""
	// VaultRestoreCMD       = ""
	EnvVaultAddress       = "VAULT_ADDR"
	EnvVaultToken         = "VAULT_TOKEN"
	EnvVaultSkipVerifyTLS = "VAULT_SKIP_VERIFY"
	EnvVaultCACert        = "VAULT_CACERT"
)

type vaultOptions struct {
	kubeClient    kubernetes.Interface
	stashClient   stash.Interface
	catalogClient appcatalog_cs.Interface

	namespace           string
	backupSessionName   string
	appBindingName      string
	appBindingNamespace string
	vaultArgs           string
	waitTimeout         int32
	outputDir           string
	storageSecret       kmapi.ObjectReference

	setupOptions   restic.SetupOptions
	backupOptions  restic.BackupOptions
	restoreOptions restic.RestoreOptions
	config         *restclient.Config

	interimDataDir string

	force bool

	OldUnsealMode          string
	OldKmsCryptoKey        string
	OldKmsKeyRing          string
	OldKmsLocation         string
	OldKmsProject          string
	OldBucket              string
	OldCredentialSecretRef string

	NewUnsealMode          string
	NewKmsCryptoKey        string
	NewKmsKeyRing          string
	NewKmsLocation         string
	NewKmsProject          string
	NewBucket              string
	NewCredentialSecretRef string
}

type sessionWrapper struct {
	sh  *shell.Session
	cmd *restic.Command
}

func (opt *vaultOptions) newSessionWrapper(cmd string) *sessionWrapper {
	return &sessionWrapper{
		sh: shell.NewSession(),
		cmd: &restic.Command{
			Name: cmd,
		},
	}
}

func (session *sessionWrapper) setVaultToken(kubeClient kubernetes.Interface, appBinding *appcatalog.AppBinding) error {
	tokenSecret, err := kubeClient.CoreV1().Secrets(appBinding.Namespace).Get(context.TODO(), backupSecretName(appBinding), metav1.GetOptions{})
	if err != nil {
		return err
	}

	err = appBinding.TransformSecret(kubeClient, tokenSecret.Data)
	if err != nil {
		return err
	}

	session.sh.SetEnv(EnvVaultToken, string(tokenSecret.Data[VaultToken]))

	return nil
}

func (session *sessionWrapper) setVaultConnectionParameters(vc *api.Client, appBinding *appcatalog.AppBinding) error {
	// use leader pod addr to take the snapshot
	// known issue: https://github.com/hashicorp/vault/issues/15258

	leaderAddr, err := getLeaderAddress(vc, appBinding)
	if err != nil {
		return err
	}

	session.sh.SetEnv(EnvVaultAddress, leaderAddr)

	return nil
}

func (session *sessionWrapper) setUserArgs(args string) {
	for _, arg := range strings.Fields(args) {
		session.cmd.Args = append(session.cmd.Args, arg)
	}
}

func (session *sessionWrapper) setTLSParameters(appBinding *appcatalog.AppBinding, scratchDir string) error {
	if appBinding.Spec.ClientConfig.CABundle != nil {
		if err := os.WriteFile(filepath.Join(scratchDir, VaultTLSRootCA), appBinding.Spec.ClientConfig.CABundle, os.ModePerm); err != nil {
			return err
		}

		session.sh.SetEnv(EnvVaultCACert, filepath.Join(scratchDir, VaultTLSRootCA))
	}
	return nil
}

func (session sessionWrapper) waitForVaultReady(vc *api.Client, waitTimeout int32) error {
	klog.Infoln("Waiting for the vault to be ready....")

	sh := shell.NewSession()
	for k, v := range session.sh.Env {
		sh.SetEnv(k, v)
	}

	return wait.PollImmediate(5*time.Second, time.Duration(waitTimeout)*time.Second, func() (done bool, err error) {
		resp, err := vc.Sys().Health()
		if err != nil {
			klog.Infof("Unable to connect with the VaultServer. Reason: %v.\nRetrying after 5 seconds....", err)
			return false, nil
		}

		if resp == nil {
			klog.Infof("Unable to connect with the VaultServer. Reason: Empty Health response")
			return false, nil
		}

		if resp.Sealed {
			klog.Infof("Unable to connect with the VaultServer. Reason: VaultServer is Sealed")
			return false, nil
		}

		klog.Infoln("VaultServer is Unsealed & Accepting Connection")
		return true, nil
	})
}

func clearDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("unable to clean datadir: %v. Reason: %v", dir, err)
	}
	return os.MkdirAll(dir, os.ModePerm)
}

func newVaultClient(appBinding *appcatalog.AppBinding) (*api.Client, error) {
	url, err := appBinding.URL()
	if err != nil {
		return nil, err
	}

	cfg := api.DefaultConfig()
	cfg.Address = url

	tlsConfig := &api.TLSConfig{
		Insecure: true,
	}

	if err = cfg.ConfigureTLS(tlsConfig); err != nil {
		return nil, err
	}

	return api.NewClient(cfg)
}

func getLeaderAddress(vc *api.Client, appBinding *appcatalog.AppBinding) (string, error) {
	port, err := appBinding.Port()
	if err != nil {
		return "", err
	}

	resp, err := vc.Sys().Leader()
	if err != nil {
		return "", err
	}

	addr := resp.LeaderClusterAddress
	if len(addr) == 0 {
		return "", errors.New("addr is empty")
	}

	addr = addr[strings.LastIndex(addr, "/")+1 : strings.LastIndex(addr, ":")]

	leaderAddr := fmt.Sprintf("%s://%s.%s.svc:%d", appBinding.Spec.ClientConfig.Service.Scheme, addr, appBinding.Namespace, port)

	return leaderAddr, nil
}

func backupSecretName(appBinding *appcatalog.AppBinding) string {
	return fmt.Sprintf("%s-%s-backup-token", appBinding.Name, appBinding.Namespace)
}
