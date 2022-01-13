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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	stash "stash.appscode.dev/apimachinery/client/clientset/versioned"
	"stash.appscode.dev/apimachinery/pkg/restic"

	shell "gomodules.xyz/go-sh"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	appcatalog "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
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
	storageSecret     kmapi.ObjectReference

	setupOptions  restic.SetupOptions
	backupOptions restic.BackupOptions
	dumpOptions   restic.DumpOptions
}

type sessionWrapper struct {
	sh  *shell.Session
	cmd *restic.Command
}

func (opt *mysqlOptions) newSessionWrapper(cmd string) *sessionWrapper {
	return &sessionWrapper{
		sh: shell.NewSession(),
		cmd: &restic.Command{
			Name: cmd,
		},
	}
}

func (session *sessionWrapper) setDatabaseCredentials(kubeClient kubernetes.Interface, appBinding *appcatalog.AppBinding) error {
	appBindingSecret, err := kubeClient.CoreV1().Secrets(appBinding.Namespace).Get(context.TODO(), appBinding.Spec.Secret.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	err = appBinding.TransformSecret(kubeClient, appBindingSecret.Data)
	if err != nil {
		return err
	}

	session.cmd.Args = append(session.cmd.Args, "-u", string(appBindingSecret.Data[MySqlUser]))
	session.sh.SetEnv(EnvMySqlPassword, string(appBindingSecret.Data[MySqlPassword]))
	return nil
}

func (session *sessionWrapper) setDatabaseConnectionParameters(appBinding *appcatalog.AppBinding) error {
	hostname, err := appBinding.Hostname()
	if err != nil {
		return err
	}
	session.cmd.Args = append(session.cmd.Args, "-h", hostname)

	port, err := appBinding.Port()
	if err != nil {
		return err
	}
	// if port is specified, append port in the arguments
	if port != 0 {
		session.cmd.Args = append(session.cmd.Args, fmt.Sprintf("--port=%d", port))
	}
	return nil
}

func (session *sessionWrapper) setUserArgs(args string) {
	for _, arg := range strings.Fields(args) {
		session.cmd.Args = append(session.cmd.Args, arg)
	}
}

func (session *sessionWrapper) setTLSParameters(appBinding *appcatalog.AppBinding, scratchDir string) error {
	// if ssl enabled, add ca.crt in the arguments
	if appBinding.Spec.ClientConfig.CABundle != nil {
		if err := ioutil.WriteFile(filepath.Join(scratchDir, MySQLTLSRootCA), appBinding.Spec.ClientConfig.CABundle, os.ModePerm); err != nil {
			return err
		}
		tlsCreds := []interface{}{
			fmt.Sprintf("--ssl-ca=%v", filepath.Join(scratchDir, MySQLTLSRootCA)),
		}
		session.cmd.Args = append(session.cmd.Args, tlsCreds)
	}
	return nil
}

func (session sessionWrapper) waitForDBReady(waitTimeout int32) error {
	klog.Infoln("Waiting for the database to be ready....")

	sh := shell.NewSession()
	for k, v := range session.sh.Env {
		sh.SetEnv(k, v)
	}

	// Execute "SELECT 1" query to the database. It should return an error when mysqld is not ready.
	args := append(session.cmd.Args, "-e", "SELECT 1;")

	// don't show the output of the query
	sh.Stdout = nil

	return wait.PollImmediate(5*time.Second, time.Duration(waitTimeout)*time.Second, func() (done bool, err error) {
		if err := sh.Command("mysql", args...).Run(); err == nil {
			klog.Infoln("Database is accepting connection....")
			return true, nil
		}
		klog.Infof("Unable to connect with the database. Reason: %v.\nRetrying after 5 seconds....", err)
		return false, nil
	})
}
