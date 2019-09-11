package pkg

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/appscode/go/log"
	"k8s.io/client-go/kubernetes"
	appcatalog_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
	"stash.appscode.dev/stash/pkg/restic"
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
	catalogClient appcatalog_cs.Interface

	namespace      string
	appBindingName string
	myArgs         string
	outputDir      string

	setupOptions  restic.SetupOptions
	backupOptions restic.BackupOptions
	dumpOptions   restic.DumpOptions
}

func waitForDBReady(host string, port int32) {
	log.Infoln("Checking database connection")
	cmd := fmt.Sprintf(`nc "%s" "%d" -w 30`, host, port)
	for {
		if err := exec.Command(cmd).Run(); err != nil {
			break
		}
		log.Infoln("Waiting... database is not ready yet")
		time.Sleep(5 * time.Second)
	}
}
