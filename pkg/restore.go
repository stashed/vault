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
	"path/filepath"
	"strings"

	api_v1beta1 "stash.appscode.dev/apimachinery/apis/stash/v1beta1"
	"stash.appscode.dev/apimachinery/pkg/restic"

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
	vaultapi "kubevault.dev/apimachinery/apis/kubevault/v1alpha2"
)

func NewCmdRestore() *cobra.Command {
	var (
		masterURL      string
		kubeconfigPath string

		opt = VaultOptions{
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

			opt.KubeClient, err = kubernetes.NewForConfig(config)
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
				Name:       opt.AppBindingName,
				Namespace:  opt.AppBindingNamespace,
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
	cmd.Flags().StringVar(&opt.AppBindingName, "appbinding", opt.AppBindingName, "Name of the app binding")
	cmd.Flags().StringVar(&opt.AppBindingNamespace, "appbinding-namespace", opt.AppBindingNamespace, "Namespace of the app binding")
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

	cmd.Flags().BoolVar(&opt.force, "force", opt.force, "Specify whether to force restore or not")
	cmd.Flags().Int64Var(&opt.SecretShares, "secret-shares", opt.SecretShares, "number of secret shares")

	// for unseal mode google kms gcs
	cmd.Flags().StringVar(&opt.OldUnsealMode, "old-unseal-mode", opt.OldUnsealMode, "specifies the mode of storing old token & unseal keys")
	cmd.Flags().StringVar(&opt.OldKmsCryptoKey, "old-kms-crypto-key", opt.OldKmsCryptoKey, "crypto key")
	cmd.Flags().StringVar(&opt.OldKmsKeyRing, "old-kms-key-ring", opt.OldKmsKeyRing, "key ring")
	cmd.Flags().StringVar(&opt.OldKmsLocation, "old-kms-location", opt.OldKmsKeyRing, "key ring")
	cmd.Flags().StringVar(&opt.OldKmsProject, "old-kms-project", opt.OldKmsKeyRing, "kms project")
	cmd.Flags().StringVar(&opt.OldBucket, "old-bucket", opt.OldKmsKeyRing, "bucket")
	cmd.Flags().StringVar(&opt.OldCredentialSecretRef, "old-credential-secret-ref", opt.OldKmsKeyRing, "credential secret")

	cmd.Flags().StringVar(&opt.NewUnsealMode, "new-unseal-mode", opt.NewUnsealMode, "specifies the mode of storing new token & unseal keys")
	cmd.Flags().StringVar(&opt.NewKmsCryptoKey, "new-kms-crypto-key", opt.NewKmsCryptoKey, "crypto key")
	cmd.Flags().StringVar(&opt.NewKmsLocation, "new-kms-location", opt.NewKmsLocation, "key ring")
	cmd.Flags().StringVar(&opt.NewKmsKeyRing, "new-kms-key-ring", opt.NewKmsKeyRing, "kms location")
	cmd.Flags().StringVar(&opt.NewKmsProject, "new-kms-project", opt.NewKmsProject, "kms project")
	cmd.Flags().StringVar(&opt.NewBucket, "new-bucket", opt.NewBucket, "bucket")
	cmd.Flags().StringVar(&opt.NewCredentialSecretRef, "new-credential-secret-ref", opt.NewCredentialSecretRef, "credential secret")

	// for unseal mode kubernetes secret
	cmd.Flags().StringVar(&opt.OldSecretName, "old-secret-name", opt.OldSecretName, "old k8s secret name")

	cmd.Flags().StringVar(&opt.NewSecretName, "new-secret-name", opt.NewSecretName, "new k8s secret name")

	// for unseal mode aws kms
	cmd.Flags().StringVar(&opt.OldKmsKeyID, "old-kms-key-id", opt.OldKmsKeyID, "old kms key id")
	cmd.Flags().StringVar(&opt.OldSsmKeyPrefix, "old-ssm-key-prefix", opt.OldSsmKeyPrefix, "old ssm key prefix")
	cmd.Flags().StringVar(&opt.OldRegion, "old-region", opt.OldRegion, "old region")
	cmd.Flags().StringVar(&opt.OldEndpoint, "old-endpoint", opt.OldEndpoint, "old endpoint")

	cmd.Flags().StringVar(&opt.NewKmsKeyID, "new-kms-key-id", opt.NewKmsKeyID, "new kms key id")
	cmd.Flags().StringVar(&opt.NewSsmKeyPrefix, "new-ssm-key-prefix", opt.NewSsmKeyPrefix, "new ssm key prefix")
	cmd.Flags().StringVar(&opt.NewRegion, "new-region", opt.NewRegion, "new region")
	cmd.Flags().StringVar(&opt.NewEndpoint, "new-endpoint", opt.NewEndpoint, "new endpoint")
	return cmd
}

func (opt *VaultOptions) restoreVault(targetRef api_v1beta1.TargetRef) (*restic.RestoreOutput, error) {
	var err error

	err = license.CheckLicenseEndpoint(opt.config, licenseApiService, SupportedProducts)
	if err != nil {
		return nil, err
	}

	opt.setupOptions.StorageSecret, err = opt.KubeClient.CoreV1().Secrets(opt.storageSecret.Namespace).Get(context.TODO(), opt.storageSecret.Name, metav1.GetOptions{})
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

	appBinding, err := opt.catalogClient.AppcatalogV1alpha1().AppBindings(opt.AppBindingNamespace).Get(context.TODO(), opt.AppBindingName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if err = clearDir(opt.interimDataDir); err != nil {
		return nil, err
	}

	session := opt.newSessionWrapper(VaultCMD)

	vaultClient, err := newVaultClient(appBinding)
	if err != nil {
		return nil, err
	}

	err = session.setVaultToken(opt.KubeClient, appBinding)
	if err != nil {
		return nil, err
	}

	err = session.setVaultConnectionParameters(vaultClient, appBinding)
	if err != nil {
		return nil, err
	}

	err = session.setTLSParameters(appBinding, opt.setupOptions.ScratchDir)
	if err != nil {
		return nil, err
	}

	err = session.waitForVaultReady(vaultClient, opt.waitTimeout)
	if err != nil {
		return nil, err
	}

	if opt.force {
		klog.Infof("Try to migrate keys from %s to %s\n", opt.OldUnsealMode, opt.NewUnsealMode)
		err = opt.migrateTokenKeys()
		if err != nil {
			return nil, err
		}
		klog.Infoln("Successfully migrated keys")
	}

	opt.restoreOptions.RestorePaths = []string{opt.interimDataDir}

	resticWrapper, err := restic.NewResticWrapper(opt.setupOptions)
	if err != nil {
		return nil, err
	}

	restoreOutput, err := resticWrapper.RunRestore(opt.restoreOptions, targetRef)
	if err != nil {
		return nil, err
	}

	err = opt.restoreVaultSnapshot(session)
	if err != nil {
		return nil, err
	}

	return restoreOutput, nil
}

func (opt *VaultOptions) restoreVaultSnapshot(session *sessionWrapper) error {
	session.cmd.Args = append(session.cmd.Args, "operator", "raft", "snapshot", "restore")

	// -force is required for different vault cluster snapshot restoration
	if opt.force {
		session.cmd.Args = append(session.cmd.Args, "-force")
	}

	session.cmd.Args = append(session.cmd.Args, filepath.Join(opt.interimDataDir, VaultSnapshotFile))

	session.sh.ShowCMD = true
	session.setUserArgs(opt.vaultArgs)
	session.sh.Command(VaultCMD, session.cmd.Args...)

	if err := session.sh.Run(); err != nil {
		return err
	}

	return nil
}

func (opt *VaultOptions) migrateTokenKeys() error {
	sts, err := opt.KubeClient.AppsV1().StatefulSets(opt.AppBindingNamespace).Get(context.TODO(), opt.AppBindingName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	var keyPrefix string
	for _, cont := range sts.Spec.Template.Spec.Containers {
		if cont.Name != vaultapi.VaultUnsealerContainerName {
			continue
		}
		for _, arg := range cont.Args {
			if strings.HasPrefix(arg, "--key-prefix=") {
				keyPrefix = arg[1+strings.Index(arg, "="):]
			}
		}
	}

	opt.KeyPrefix = keyPrefix

	keys, err := opt.GetTokenKeys()
	if err != nil {
		klog.Infoln("failed to get token keys: ", err.Error())
		return err
	}

	return opt.SetTokenKeys(keys)
}
