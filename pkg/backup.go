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
	"os"
	"path/filepath"

	api_v1beta1 "stash.appscode.dev/apimachinery/apis/stash/v1beta1"
	stash "stash.appscode.dev/apimachinery/client/clientset/versioned"
	"stash.appscode.dev/apimachinery/pkg/restic"
	api_util "stash.appscode.dev/apimachinery/pkg/util"
	"stash.appscode.dev/vault/pkg/store"

	"github.com/pkg/errors"
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

func NewCmdBackup() *cobra.Command {
	var (
		masterURL      string
		kubeconfigPath string
		opt            = VaultOptions{
			setupOptions: restic.SetupOptions{
				ScratchDir:  restic.DefaultScratchDir,
				EnableCache: false,
			},
			backupOptions: restic.BackupOptions{
				Host: restic.DefaultHost,
			},
		}
	)

	cmd := &cobra.Command{
		Use:               "backup-vault",
		Short:             "Takes a backup of Vault",
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

			opt.stashClient, err = stash.NewForConfig(config)
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
			var backupOutput *restic.BackupOutput
			backupOutput, err = opt.backupVault(targetRef)
			if err != nil {
				backupOutput = &restic.BackupOutput{
					BackupTargetStatus: api_v1beta1.BackupTargetStatus{
						Ref: targetRef,
						Stats: []api_v1beta1.HostBackupStats{
							{
								Hostname: opt.backupOptions.Host,
								Phase:    api_v1beta1.HostBackupFailed,
								Error:    err.Error(),
							},
						},
					},
				}
			}
			// If output directory specified, then write the output in "output.json" file in the specified directory
			if opt.outputDir != "" {
				return backupOutput.WriteOutput(filepath.Join(opt.outputDir, restic.DefaultOutputFileName))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opt.vaultArgs, "vault-args", opt.vaultArgs, "Additional arguments")
	cmd.Flags().Int32Var(&opt.waitTimeout, "wait-timeout", opt.waitTimeout, "Time limit to wait for the database to be ready")

	cmd.Flags().StringVar(&masterURL, "master", masterURL, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	cmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", kubeconfigPath, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	cmd.Flags().StringVar(&opt.namespace, "namespace", "default", "Namespace of Backup/Restore Session")
	cmd.Flags().StringVar(&opt.backupSessionName, "backupsession", opt.backupSessionName, "Name of the Backup Session")
	cmd.Flags().StringVar(&opt.appBindingName, "appbinding", opt.appBindingName, "Name of the app binding")
	cmd.Flags().StringVar(&opt.appBindingNamespace, "appbinding-namespace", opt.appBindingNamespace, "Namespace of the app binding")
	cmd.Flags().StringVar(&opt.setupOptions.Provider, "provider", opt.setupOptions.Provider, "Backend provider (i.e. gcs, s3, azure etc)")
	cmd.Flags().StringVar(&opt.setupOptions.Bucket, "bucket", opt.setupOptions.Bucket, "Name of the cloud bucket/container (keep empty for local backend)")
	cmd.Flags().StringVar(&opt.setupOptions.Endpoint, "endpoint", opt.setupOptions.Endpoint, "Endpoint for s3/s3 compatible backend or REST backend URL")
	cmd.Flags().StringVar(&opt.setupOptions.Region, "region", opt.setupOptions.Region, "Region for s3/s3 compatible backend")
	cmd.Flags().StringVar(&opt.setupOptions.Path, "path", opt.setupOptions.Path, "Directory inside the bucket where backup will be stored")
	cmd.Flags().StringVar(&opt.setupOptions.ScratchDir, "scratch-dir", opt.setupOptions.ScratchDir, "Temporary directory")
	cmd.Flags().BoolVar(&opt.setupOptions.EnableCache, "enable-cache", opt.setupOptions.EnableCache, "Specify whether to enable caching for restic")
	cmd.Flags().Int64Var(&opt.setupOptions.MaxConnections, "max-connections", opt.setupOptions.MaxConnections, "Specify maximum concurrent connections for GCS, Azure and B2 backend")
	cmd.Flags().StringVar(&opt.storageSecret.Name, "storage-secret-name", opt.storageSecret.Name, "Name of the storage secret")
	cmd.Flags().StringVar(&opt.storageSecret.Namespace, "storage-secret-namespace", opt.storageSecret.Namespace, "Namespace of the storage secret")

	cmd.Flags().StringVar(&opt.backupOptions.Host, "hostname", opt.backupOptions.Host, "Name of the host machine")

	cmd.Flags().Int64Var(&opt.backupOptions.RetentionPolicy.KeepLast, "retention-keep-last", opt.backupOptions.RetentionPolicy.KeepLast, "Specify value for retention strategy")
	cmd.Flags().Int64Var(&opt.backupOptions.RetentionPolicy.KeepHourly, "retention-keep-hourly", opt.backupOptions.RetentionPolicy.KeepHourly, "Specify value for retention strategy")
	cmd.Flags().Int64Var(&opt.backupOptions.RetentionPolicy.KeepDaily, "retention-keep-daily", opt.backupOptions.RetentionPolicy.KeepDaily, "Specify value for retention strategy")
	cmd.Flags().Int64Var(&opt.backupOptions.RetentionPolicy.KeepWeekly, "retention-keep-weekly", opt.backupOptions.RetentionPolicy.KeepWeekly, "Specify value for retention strategy")
	cmd.Flags().Int64Var(&opt.backupOptions.RetentionPolicy.KeepMonthly, "retention-keep-monthly", opt.backupOptions.RetentionPolicy.KeepMonthly, "Specify value for retention strategy")
	cmd.Flags().Int64Var(&opt.backupOptions.RetentionPolicy.KeepYearly, "retention-keep-yearly", opt.backupOptions.RetentionPolicy.KeepYearly, "Specify value for retention strategy")
	cmd.Flags().StringSliceVar(&opt.backupOptions.RetentionPolicy.KeepTags, "retention-keep-tags", opt.backupOptions.RetentionPolicy.KeepTags, "Specify value for retention strategy")
	cmd.Flags().BoolVar(&opt.backupOptions.RetentionPolicy.Prune, "retention-prune", opt.backupOptions.RetentionPolicy.Prune, "Specify whether to prune old snapshot data")
	cmd.Flags().BoolVar(&opt.backupOptions.RetentionPolicy.DryRun, "retention-dry-run", opt.backupOptions.RetentionPolicy.DryRun, "Specify whether to test retention policy without deleting actual data")

	cmd.Flags().StringVar(&opt.outputDir, "output-dir", opt.outputDir, "Directory where output.json file will be written (keep empty if you don't need to write output in file)")
	cmd.Flags().StringVar(&opt.interimDataDir, "interim-data-dir", opt.interimDataDir, "Directory where the targeted data will be stored temporarily before uploading to the backend")
	cmd.Flags().StringVar(&opt.keyPrefix, "key-prefix", opt.keyPrefix, "prefix that will be append to root-token & unseal-keys")

	return cmd
}

func (opt *VaultOptions) backupVault(targetRef api_v1beta1.TargetRef) (*restic.BackupOutput, error) {
	var err error
	err = license.CheckLicenseEndpoint(opt.config, licenseApiService, SupportedProducts)
	if err != nil {
		return nil, err
	}

	opt.setupOptions.StorageSecret, err = opt.kubeClient.CoreV1().Secrets(opt.storageSecret.Namespace).Get(context.TODO(), opt.storageSecret.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	// if any pre-backup actions has been assigned to it, execute them
	actionOptions := api_util.ActionOptions{
		StashClient:       opt.stashClient,
		TargetRef:         targetRef,
		SetupOptions:      opt.setupOptions,
		BackupSessionName: opt.backupSessionName,
		Namespace:         opt.namespace,
	}

	if err := api_util.ExecutePreBackupActions(actionOptions); err != nil {
		return nil, err
	}

	// wait until the backend repository has been initialized.
	if err := api_util.WaitForBackendRepository(actionOptions); err != nil {
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

	// get the vault appbinding which has necessary information about vault & vault backup token
	appBinding, err := opt.catalogClient.AppcatalogV1alpha1().AppBindings(opt.appBindingNamespace).Get(context.TODO(), opt.appBindingName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	parameters := vaultconfig.VaultServerConfiguration{}
	if appBinding.Spec.Parameters != nil {
		if err = json.Unmarshal(appBinding.Spec.Parameters.Raw, &parameters); err != nil {
			klog.Errorf("unable to unmarshal appBinding.Spec.Parameters.Raw. Reason: %v", err)
		}
	}

	// update this while adding support for more backend options for backup (consul, s3, etc.)
	if parameters.Backend != VaultStorageBackendRaft {
		return nil, errors.New("Backend must be Raft for backup snapshots")
	}

	// clean the interim directory where the snapshot, unseal keys & root token will be stored
	if err = clearDir(opt.interimDataDir); err != nil {
		return nil, err
	}

	session := opt.newSessionWrapper(VaultCMD)

	// create a new vault client to interact with vault
	// this is needed for running commands like: vault operator raft snapshot save backup.snap or restore backup.snap
	vaultClient, err := newVaultClient(appBinding)
	if err != nil {
		return nil, err
	}

	// if the vault is TLS enabled then set the env variable for vault TLS
	if err := session.setTLSParameters(appBinding, opt.setupOptions.ScratchDir); err != nil {
		return nil, err
	}

	// wait until the vault is ready (vault must be unsealed when ready)
	if err := session.waitForVaultReady(vaultClient, opt.waitTimeout); err != nil {
		return nil, err
	}

	// set the vault token that has the necessary permission to save or restore snapshot
	if err := session.setVaultToken(opt.kubeClient, appBinding, parameters.BackupTokenSecretRef); err != nil {
		return nil, err
	}

	// set the vault connection parameters, essentially the vault leader node address
	if err := session.setVaultConnectionParameters(vaultClient, appBinding); err != nil {
		return nil, err
	}

	klog.Infof("Trying to backup for VaultServer %s/%s\n", appBinding.Namespace, appBinding.Name)

	// save the vault snapshot in the interim directory running command like: vault operator raft snapshot save backup.snap
	if err := opt.saveVaultSnapshot(session); err != nil {
		return nil, err
	}

	// save the unseal keys & root token too, for complete backup
	// write these in the interim directory along with the snapshot file
	if err := opt.writeVaultTokenKeys(appBinding, parameters); err != nil {
		return nil, err
	}

	// now take the backup of this directory using stash/restic
	opt.backupOptions.BackupPaths = []string{opt.interimDataDir}
	resticWrapper, err := restic.NewResticWrapper(opt.setupOptions)
	if err != nil {
		return nil, err
	}

	return resticWrapper.RunBackup(opt.backupOptions, targetRef)
}

func (opt *VaultOptions) saveVaultSnapshot(session *sessionWrapper) error {
	klog.Infoln("Trying to save snapshot")
	session.cmd.Args = append(session.cmd.Args, "operator", "raft", "snapshot", "save", filepath.Join(opt.interimDataDir, VaultSnapshotFile))

	session.sh.ShowCMD = false
	session.setUserArgs(opt.vaultArgs)
	session.sh.Command(VaultCMD, session.cmd.Args...)

	if err := session.sh.Run(); err != nil {
		return err
	}

	klog.Infoln("snapshot saved successfully")
	return nil
}

func (opt *VaultOptions) writeVaultTokenKeys(appBinding *appcatalog.AppBinding, params vaultconfig.VaultServerConfiguration) error {
	if params.Unsealer == nil {
		return errors.New("unsealer spec is nil")
	}

	klog.Infoln("Trying to get, write unseal keys & root token")
	// create a new store interface
	// for backup:
	// i. Get the unseal key & root token from store based on the unseal mode
	// ii. write them into the interim directory which will be backed up
	st, err := store.NewStore(opt.kubeClient, appBinding, params.Unsealer)
	if err != nil {
		return err
	}

	var keys []string
	keys = append(keys, opt.tokenName(opt.keyPrefix))
	for i := 0; i < int(params.Unsealer.SecretShares); i++ {
		keys = append(keys, opt.unsealKeyName(opt.keyPrefix, i))
	}

	for _, key := range keys {
		value, err := st.Get(key)
		if err != nil {
			klog.Errorf("failed to get key %s with %s\n", key, err.Error())
			return err
		}

		if err := opt.write(key, value); err != nil {
			klog.Errorf("failed to write key %s with %s\n", key, err.Error())
			return err
		}
	}

	klog.Infoln("Successfully stored unseal keys & root token")
	return nil
}

func (opt *VaultOptions) write(key, value string) error {
	byteStreams, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(opt.interimDataDir, key), byteStreams, 0o644)
}
