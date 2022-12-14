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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	api_v1beta1 "stash.appscode.dev/apimachinery/apis/stash/v1beta1"
	"stash.appscode.dev/apimachinery/pkg/restic"
	"stash.appscode.dev/vault/pkg/store"

	"github.com/spf13/cobra"
	license "go.bytebuilders.dev/license-verifier/kubernetes"
	"gomodules.xyz/flags"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	appcatalog "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcatalog_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
	v1 "kmodules.xyz/offshoot-api/api/v1"
	vaultconfig "kubevault.dev/apimachinery/apis/config/v1alpha1"
)

func NewCmdRestore() *cobra.Command {
	var (
		masterURL      string
		kubeconfigPath string

		opt = vaultOptions{
			setupOptions: restic.SetupOptions{
				ScratchDir:  restic.DefaultScratchDir,
				EnableCache: false,
			},
			waitTimeout: 300,
			restoreOptions: restic.RestoreOptions{
				Host: restic.DefaultHost,
			},
		}
	)

	cmd := &cobra.Command{
		Use:               "restore-vault",
		Short:             "Restores Vault Backup",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags.EnsureRequiredFlags(cmd, "appbinding", "provider", "storage-secret-name", "storage-secret-namespace")

			// prepare client
			config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
			if err != nil {
				return err
			}
			opt.config = config

			opt.kubeClient, err = kubernetes.NewForConfig(config)
			if err != nil {
				return err
			}

			opt.catalogClient, err = appcatalog_cs.NewForConfig(config)
			if err != nil {
				return err
			}

			targetRef := api_v1beta1.TargetRef{
				APIVersion: appcatalog.SchemeGroupVersion.String(),
				Kind:       appcatalog.ResourceKindApp,
				Name:       opt.appBindingName,
				Namespace:  opt.appBindingNamespace,
			}

			var restoreOutput *restic.RestoreOutput
			restoreOutput, err = opt.restoreVault(targetRef)
			if err != nil {
				restoreOutput = &restic.RestoreOutput{
					RestoreTargetStatus: api_v1beta1.RestoreMemberStatus{
						Ref: targetRef,
						Stats: []api_v1beta1.HostRestoreStats{
							{
								Hostname: opt.restoreOptions.Host,
								Phase:    api_v1beta1.HostRestoreFailed,
								Error:    err.Error(),
							},
						},
					},
				}
			}
			// If output directory specified, then write the output in "output.json" file in the specified directory
			if opt.outputDir != "" {
				return restoreOutput.WriteOutput(filepath.Join(opt.outputDir, restic.DefaultOutputFileName))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opt.vaultArgs, "vault-args", opt.vaultArgs, "Additional arguments")
	cmd.Flags().Int32Var(&opt.waitTimeout, "wait-timeout", opt.waitTimeout, "Time limit to wait for the database to be ready")

	cmd.Flags().StringVar(&masterURL, "master", masterURL, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	cmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", kubeconfigPath, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	cmd.Flags().StringVar(&opt.namespace, "namespace", "default", "Namespace of Backup/Restore Session")
	cmd.Flags().StringVar(&opt.appBindingName, "appbinding", opt.appBindingName, "Name of the app binding")
	cmd.Flags().StringVar(&opt.appBindingNamespace, "appbinding-namespace", opt.appBindingNamespace, "Namespace of the app binding")
	cmd.Flags().StringVar(&opt.setupOptions.Provider, "provider", opt.setupOptions.Provider, "Backend provider (i.e. gcs, s3, azure etc)")
	cmd.Flags().StringVar(&opt.setupOptions.Bucket, "bucket", opt.setupOptions.Bucket, "Name of the cloud bucket/container (keep empty for local backend)")
	cmd.Flags().StringVar(&opt.setupOptions.Endpoint, "endpoint", opt.setupOptions.Endpoint, "Endpoint for s3/s3 compatible backend or REST backend URL")
	cmd.Flags().StringVar(&opt.setupOptions.Region, "region", opt.setupOptions.Region, "Region for s3/s3 compatible backend")
	cmd.Flags().StringVar(&opt.setupOptions.Path, "path", opt.setupOptions.Path, "Directory inside the bucket where backup will be stored")
	cmd.Flags().StringVar(&opt.storageSecret.Name, "storage-secret-name", opt.storageSecret.Name, "Name of the storage secret")
	cmd.Flags().StringVar(&opt.storageSecret.Namespace, "storage-secret-namespace", opt.storageSecret.Namespace, "Namespace of the storage secret")

	cmd.Flags().StringVar(&opt.setupOptions.ScratchDir, "scratch-dir", opt.setupOptions.ScratchDir, "Temporary directory")
	cmd.Flags().BoolVar(&opt.setupOptions.EnableCache, "enable-cache", opt.setupOptions.EnableCache, "Specify whether to enable caching for restic")
	cmd.Flags().Int64Var(&opt.setupOptions.MaxConnections, "max-connections", opt.setupOptions.MaxConnections, "Specify maximum concurrent connections for GCS, Azure and B2 backend")

	cmd.Flags().StringVar(&opt.restoreOptions.Host, "hostname", opt.restoreOptions.Host, "Name of the host machine")
	cmd.Flags().StringVar(&opt.restoreOptions.SourceHost, "source-hostname", opt.restoreOptions.SourceHost, "Name of the host from where data will be restored")
	cmd.Flags().StringSliceVar(&opt.restoreOptions.Snapshots, "snapshot", opt.restoreOptions.Snapshots, "Snapshot to restore")

	cmd.Flags().StringVar(&opt.interimDataDir, "interim-data-dir", opt.interimDataDir, "Directory where the restored data will be stored temporarily before injecting into the desired NATS Server")
	cmd.Flags().StringVar(&opt.outputDir, "output-dir", opt.outputDir, "Directory where output.json file will be written (keep empty if you don't need to write output in file)")

	// vault related flags
	// -force implies that snapshot will be restore forcefully, required when restoring on a different vault server
	cmd.Flags().BoolVar(&opt.force, "force", opt.force, "Specify whether to force restore or not")

	cmd.Flags().StringVar(&opt.keyPrefix, "key-prefix", opt.keyPrefix, "prefix that will be append to root-token & unseal-keys")
	cmd.Flags().StringVar(&opt.oldKeyPrefix, "old-key-prefix", opt.oldKeyPrefix, "old prefix that was appended to root-token & unseal-keys")

	return cmd
}

func (opt *vaultOptions) restoreVault(targetRef api_v1beta1.TargetRef) (*restic.RestoreOutput, error) {
	var err error

	err = license.CheckLicenseEndpoint(opt.config, licenseApiService, SupportedProducts)
	if err != nil {
		return nil, err
	}

	opt.setupOptions.StorageSecret, err = opt.kubeClient.CoreV1().Secrets(opt.storageSecret.Namespace).Get(context.TODO(), opt.storageSecret.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// apply nice, ionice settings from env
	opt.setupOptions.Nice, err = v1.NiceSettingsFromEnv()
	if err != nil {
		return nil, err
	}
	opt.setupOptions.IONice, err = v1.IONiceSettingsFromEnv()
	if err != nil {
		return nil, err
	}

	appBinding, err := opt.catalogClient.AppcatalogV1alpha1().AppBindings(opt.appBindingNamespace).Get(context.TODO(), opt.appBindingName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	parameters := vaultconfig.VaultServerConfiguration{}
	if appBinding.Spec.Parameters != nil {
		if err = json.Unmarshal(appBinding.Spec.Parameters.Raw, &parameters); err != nil {
			return nil, fmt.Errorf("unable to unmarshal appBinding.Spec.Parameters.Raw: %w", err)
		}
	}

	if parameters.Backend != VaultStorageBackendRaft {
		return nil, fmt.Errorf("backend must be Raft for backup snapshots")
	}

	if err := clearDir(opt.interimDataDir); err != nil {
		return nil, err
	}

	session := opt.newSessionWrapper(VaultCMD)

	vaultClient, err := newVaultClient(appBinding)
	if err != nil {
		return nil, err
	}

	if err := session.setTLSParameters(appBinding, opt.setupOptions.ScratchDir); err != nil {
		return nil, err
	}

	if err := session.waitForVaultReady(vaultClient, opt.waitTimeout); err != nil {
		return nil, err
	}

	if err := session.setVaultToken(opt.kubeClient, appBinding, parameters.BackupTokenSecretRef); err != nil {
		return nil, err
	}

	if err := session.setVaultConnectionParameters(vaultClient, appBinding); err != nil {
		return nil, err
	}

	klog.Infof("Trying to restore snapshot for VaultServer %s/%s\n", appBinding.Namespace, appBinding.Name)

	opt.restoreOptions.RestorePaths = []string{opt.interimDataDir}

	resticWrapper, err := restic.NewResticWrapper(opt.setupOptions)
	if err != nil {
		return nil, err
	}

	restoreOutput, err := resticWrapper.RunRestore(opt.restoreOptions, targetRef)
	if err != nil {
		return nil, err
	}

	if err := opt.restoreVaultSnapshot(session); err != nil {
		return nil, err
	}

	if opt.force {
		if err := opt.migrateVaultTokenKeys(appBinding, parameters); err != nil {
			return nil, err
		}
	}

	return restoreOutput, nil
}

func (opt *vaultOptions) restoreVaultSnapshot(session *sessionWrapper) error {
	klog.Infoln("Trying to restore snapshot")
	session.cmd.Args = append(session.cmd.Args, "operator", "raft", "snapshot", "restore")

	// -force is required for different vault cluster snapshot restoration
	if opt.force {
		session.cmd.Args = append(session.cmd.Args, "-force")
	}

	session.cmd.Args = append(session.cmd.Args, filepath.Join(opt.interimDataDir, VaultSnapshotFile))

	session.sh.ShowCMD = false
	session.setUserArgs(opt.vaultArgs)
	session.sh.Command(VaultCMD, session.cmd.Args...)

	if err := session.sh.Run(); err != nil {
		return err
	}

	klog.Infoln("snapshot restored successfully")
	return nil
}

func (opt *vaultOptions) migrateVaultTokenKeys(appBinding *appcatalog.AppBinding, params vaultconfig.VaultServerConfiguration) error {
	if params.Unsealer == nil {
		return fmt.Errorf("unsealer spec is nil")
	}

	klog.Infoln("Trying to read, set unseal keys & root token")
	// for restore:
	// i. read the unseal keys & root token from the interim directory
	// ii. Set the unseal keys & root token to store based on the unseal mode
	st, err := store.NewStore(opt.kubeClient, appBinding, params.Unsealer)
	if err != nil {
		return err
	}

	var oldKeys []string
	oldKeys = append(oldKeys, opt.tokenName(opt.oldKeyPrefix))
	for i := 0; i < int(params.Unsealer.SecretShares); i++ {
		oldKeys = append(oldKeys, opt.unsealKeyName(opt.oldKeyPrefix, i))
	}

	var newKeys []string
	newKeys = append(newKeys, opt.tokenName(opt.keyPrefix))
	for i := 0; i < int(params.Unsealer.SecretShares); i++ {
		newKeys = append(newKeys, opt.unsealKeyName(opt.keyPrefix, i))
	}

	for idx, oldKey := range oldKeys {
		value, err := opt.read(oldKey)
		if err != nil {
			return fmt.Errorf("failed to read key %s. Reason: %w", oldKey, err)
		}

		if err := st.Set(newKeys[idx], value); err != nil {
			return fmt.Errorf("failed to set key %s. Reason: %w", newKeys[idx], err)
		}
	}

	return nil
}

func (opt *vaultOptions) read(key string) (string, error) {
	byteStreams, err := os.ReadFile(filepath.Join(opt.interimDataDir, key))
	if err != nil {
		return "", err
	}

	var data string
	if err := json.Unmarshal(byteStreams, &data); err != nil {
		return "", err
	}

	return data, nil
}
