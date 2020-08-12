/*
Copyright AppsCode Inc. and Contributors

Licensed under the PolyForm Noncommercial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/PolyForm-Noncommercial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pkg

import (
	"fmt"

	stash "stash.appscode.dev/apimachinery/client/clientset/versioned"
	"stash.appscode.dev/apimachinery/pkg/restic"

	"github.com/appscode/go/log"
	"github.com/codeskyblue/go-sh"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcatalog_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
)

const (
	MySqlUser        = "username"
	MySqlPassword    = "password"
	MySqlDumpFile    = "dumpfile.sql"
	MySqlDumpCMD     = "mysqldump"
	MySqlRestoreCMD  = "mysql"
	EnvMySqlPassword = "MYSQL_PWD"
)

type mysqlOptions struct {
	kubeClient    kubernetes.Interface
	stashClient   stash.Interface
	catalogClient appcatalog_cs.Interface

	namespace         string
	backupSessionName string
	appBindingName    string
	myArgs            string
	waitTimeout       int32
	outputDir         string

	setupOptions  restic.SetupOptions
	backupOptions restic.BackupOptions
	dumpOptions   restic.DumpOptions
}

func waitForDBReady(appBinding *v1alpha1.AppBinding, secret *core.Secret, waitTimeout int32) error {
	log.Infoln("Waiting for the database to be ready.....")
	shell := sh.NewSession()
	shell.SetEnv(EnvMySqlPassword, string(secret.Data[MySqlPassword]))
	args := []interface{}{
		"ping",
		"--host", appBinding.Spec.ClientConfig.Service.Name,
		"--user=root",
		fmt.Sprintf("--wait=%d", waitTimeout),
	}
	return shell.Command("mysqladmin", args...).Run()
}
