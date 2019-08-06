---


title: Backup MySQL | Stash
description: Backup MySQL database using Stash
menu:
  product_stash_0.8.3:
    identifier: database-mysql
    name: MySQL
    parent: database
    weight: 20
product_name: stash
menu_name: product_stash_0.8.3
section_menu_id: guides
---

# Backup and Restore MySQL database using Stash

Stash 0.9.0+ supports backup and restoration of MySQL databases. This guide will show you how you can backup and restore your MySQL database with Stash.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using Minikube.

- Install Stash in your cluster following the steps [here](https://appscode.com/products/stash/0.8.3/setup/install/).

- Install [KubeDB](https://kubedb.com) in your cluster following the steps [here](https://kubedb.com/docs/0.12.0/setup/install/). This step is optional. You can deploy your database using any method you want. We are using KubeDB because it automates some tasks that you have to do manually otherwise.

- If you are not familiar with how Stash takes backup of databases and restores them, please check the following guide:
  - [How Stash takes backup of databases and restores them](https://appscode.com/products/stash/0.8.3/guides/databases/overview/).

You have to be familiar with following custom resources:

- [AppBinding](https://appscode.com/products/stash/0.8.3/concepts/crds/appbinding/)
- [Function](https://appscode.com/products/stash/0.8.3/concepts/crds/function/)
- [Task](https://appscode.com/products/stash/0.8.3/concepts/crds/task/)
- [BackupConfiguration](https://appscode.com/products/stash/0.8.3/concepts/crds/backupconfiguration/)
- [RestoreSession](https://appscode.com/products/stash/0.8.3/concepts/crds/restoresession/)

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial. Create `demo` namespace if you haven't created yet.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored [here](https://github.com/stashed/mysql/tree/master/docs/examples).

## Install MySQL Catalog for Stash

Stash uses a `Function-Task` model to backup databases. We have to install MySQL catalogs (`stash-mysql`) for Stash. This catalog creates necessary `Function` and `Task` definitions to backup/restore MySQL databases.

You can install the catalog either as a helm chart or you can create only the YAMLs of the respective resources.

<ul class="nav nav-tabs" id="installerTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link" id="helm-tab" data-toggle="tab" href="#helm" role="tab" aria-controls="helm" aria-selected="false">Helm</a>
  </li>
  <li class="nav-item">
    <a class="nav-link active" id="script-tab" data-toggle="tab" href="#script" role="tab" aria-controls="script" aria-selected="true">Script</a>
  </li>
</ul>
<div class="tab-content" id="installerTabContent">
 <!-- ------------ Helm Tab Begins----------- -->
  <div class="tab-pane fade" id="helm" role="tabpanel" aria-labelledby="helm-tab">

### Install as chart release

Run the following script to install `stash-mysql` catalog as a Helm chart.

```console
curl -fsSL https://github.com/stashed/catalog/raw/master/deploy/chart.sh | bash -s -- --catalog=stash-mysql
```

</div>
<!-- ------------ Helm Tab Ends----------- -->

<!-- ------------ Script Tab Begins----------- -->
<div class="tab-pane fade show active" id="script" role="tabpanel" aria-labelledby="script-tab">

### Install only YAMLs

Run the following script to install `stash-mysql` catalog as Kubernetes YAMLs.

```console
curl -fsSL https://github.com/stashed/catalog/raw/master/deploy/script.sh | bash -s -- --catalog=stash-mysql
```

</div>
<!-- ------------ Script Tab Ends----------- -->
</div>

Once installed, this will create `mysql-backup-*` and `mysql-restore-*` Functions for all supported MySQL versions. To verify, run the following command:

```console
$ kubectl get functions.stash.appscode.com
NAME                    AGE
mysql-backup-8.0.14     20s
mysql-backup-5.7        20s
pvc-backup              7h6m
pvc-restore             7h6m
update-status           7h6m
```

Also, verify that the necessary `Task` have been created.

```console
$ kubectl get tasks.stash.appscode.com
NAME                    AGE
mysql-backup-8.0.14     2m7s
mysql-backup-5.7        2m7s
pvc-backup              7h7m
pvc-restore             7h7m
```

Now, Stash is ready to backup MySQL database.

## Backup MySQL

This section will demonstrate how to backup MySQL database. Here, we are going to deploy a MySQL database using KubeDB. Then, we are going to backup this database into a GCS bucket. Finally, we are going to restore the backed up data into another MySQL database.

### Deploy Sample MySQL Database

Let's deploy a sample MySQL database and insert some data into it.

**Create MySQL CRD:**

Below is the YAML of a sample MySQL CRD that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: sample-mysql
  namespace: demo
spec:
  version: "8.0.14"
  replicas: 1
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
  terminationPolicy: WipeOut
```

Create the above `MySQL` CRD,

```bash
$ kubectl apply -f ./docs/examples/backup/mysql.yaml
mysql.kubedb.com/sample-mysql created
```

KubeDB will deploy a MySQL database according to the above specification. It will also create the necessary Secrets and Services to access the database.

Let's check if the database is ready to use,

```bash
$ kubectl get my -n demo sample-mysql
NAME           VERSION   STATUS    AGE
sample-mysql   8.0.14    Running   4m22s
```

The database is `Running`. Verify that KubeDB has created a Secret and a Service for this database using the following commands,

```bash
$ kubectl get secret -n demo -l=kubedb.com/name=sample-mysql
NAME                TYPE     DATA   AGE
sample-mysql-auth   Opaque   2      4m58s

$ kubectl get service -n demo -l=kubedb.com/name=sample-mysql
NAME               TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
sample-mysql       ClusterIP   10.101.2.138   <none>        3306/TCP   5m33s
sample-mysql-gvr   ClusterIP   None           <none>        3306/TCP   5m33s
```

Here, we have to use service `sample-mysql` and secret `sample-mysql-auth` to connect with the database. KubeDB creates an [AppBinding](https://appscode.com/products/stash/0.8.3/concepts/crds/appbinding/) CRD that holds the necessary information to connect with the database.

**Verify AppBinding:**

Verify that the AppBinding has been created successfully using the following command,

```bash
$ kubectl get appbindings -n demo
NAME           AGE
sample-mysql   9m24s
```

Let's check the YAML of the above AppBinding,

```bash
$ kubectl get appbindings -n demo sample-mysql -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  creationTimestamp: "2019-08-02T05:13:37Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-mysql
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: mysql
    app.kubernetes.io/version: 8.0.14
    kubedb.com/kind: MySQL
    kubedb.com/name: sample-mysql
  name: sample-mysql
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1alpha1
    blockOwnerDeletion: false
    kind: MySQL
    name: sample-mysql
    uid: dab30216-485f-405a-af4f-09fe5f0ad88e
  resourceVersion: "7970"
  selfLink: /apis/appcatalog.appscode.com/v1alpha1/namespaces/demo/appbindings/sample-mysql
  uid: d2a932e5-924f-4321-b206-7b9da534cf12
spec:
  clientConfig:
    service:
      name: sample-mysql
      path: /
      port: 3306
      scheme: mysql
    url: tcp(sample-mysql:3306)/
  secret:
    name: sample-mysql-auth
  type: kubedb.com/mysql
  version: 8.0.14
```

Stash uses the AppBinding CRD to connect with the target database. It requires the following two fields to set in AppBinding's `.spec` section.

- `.spec.clientConfig.service.name` specifies the name of the Service that connects to the database.
- `.spec.secret` specifies the name of the Secret that holds necessary credentials to access the database.
- `spec.type` specifies the types of the app that this AppBinding is pointing to. KubeDB generated AppBinding follows the following format: `<app group>/<app resource type>`.

**Creating AppBinding Manually:**

If you deploy MySQL database without KubeDB, you have to create the AppBinding CRD manually in the same namespace as the service and secret of the database.

The following YAML shows a minimal AppBinding specification that you have to create if you deploy MySQL database without KubeDB.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: <my_custom_appbinding_name>
  namespace: <my_database_namespace>
spec:
  clientConfig:
    service:
      name: <my_database_service_name>
      port: <my_database_port_number>
  secret:
    name: <my_database_credentials_secret_name>
  # type field is optional. you can keep it empty.
  # if you keep it emtpty then the value of TARGET_APP_RESOURCE variable
  # will be set to "appbinding" during auto-backup.
  type: mysql
```

You have to replace the `<...>` quoted part with proper values in the above YAML.

**Insert Sample Data:**

Now, we are going to exec into the database pod and create some sample data. At first, find out the database Pod using the following command,

```bash
$ kubectl get pods -n demo --selector="kubedb.com/name=sample-mysql"
NAME             READY   STATUS    RESTARTS   AGE
sample-mysql-0   1/1     Running   0          33m
```

And copy the user name and password of the `root` user to access into `mysql` shell.

```bash
$ kubectl get secret -n demo  sample-mysql-auth -o jsonpath='{.data.username}'| base64 -d
root⏎

$ kubectl get secret -n demo  sample-mysql-auth -o jsonpath='{.data.password}'| base64 -d
5HEqoozyjgaMO97N⏎
```

Now, let's exec into the Pod to enter into `mysql` shell and create a database and a table,

```bash
$ kubectl exec -it -n demo sample-mysql-0 -- mysql --user=root --password=5HEqoozyjgaMO97N
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 10
Server version: 8.0.14 MySQL Community Server - GPL

Copyright (c) 2000, 2019, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> CREATE DATABASE playground;
Query OK, 1 row affected (0.01 sec)

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| mysql              |
| performance_schema |
| playground         |
| sys                |
+--------------------+
5 rows in set (0.00 sec)

mysql> CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));
Query OK, 0 rows affected (0.01 sec)

mysql> SHOW TABLES IN playground;
+----------------------+
| Tables_in_playground |
+----------------------+
| equipment            |
+----------------------+
1 row in set (0.01 sec)

mysql> INSERT INTO playground.equipment (type, quant, color) VALUES ("slide", 2, "blue");
Query OK, 1 row affected (0.01 sec)

mysql> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
1 row in set (0.00 sec)

mysql> exit
Bye
```

Now, we are ready to backup the database.

### Prepare Backend

We are going to store our backed up data into a GCS bucket. At first, we need to create a secret with GCS credentials then we need to create a `Repository` CRD. If you want to use a different backend, please read the respective backend configuration doc from [here](https://appscode.com/products/stash/0.8.3/guides/backends/overview/).

**Create Storage Secret:**

Let's create a secret called `gcs-secret` with access credentials to our desired GCS bucket,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ echo -n '<your-project-id>' > GOOGLE_PROJECT_ID
$ cat downloaded-sa-json.key > GOOGLE_SERVICE_ACCOUNT_JSON_KEY
$ kubectl create secret generic -n demo gcs-secret \
    --from-file=./RESTIC_PASSWORD \
    --from-file=./GOOGLE_PROJECT_ID \
    --from-file=./GOOGLE_SERVICE_ACCOUNT_JSON_KEY
secret/gcs-secret created
```

**Create Repository:**

Now, crete a `Respository` using this secret. Below is the YAML of Repository CRD we are going to create,

```yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  name: gcs-repo
  namespace: demo
spec:
  backend:
    gcs:
      bucket: appscode-qa
      prefix: /demo/mysql/sample-mysql
    storageSecretName: gcs-secret
```

Let's create the `Repository` we have shown above,

```bash
$ kubectl create -f ./docs/examples/backup/repository.yaml
repository.stash.appscode.com/gcs-repo created
```

Now, we are ready to backup our database to our desired backend.

### Backup

We have to create a `BackupConfiguration` targeting respective AppBinding CRD of our desired database. Then Stash will create a CronJob to periodically backup the database.

**Create BackupConfiguration:**

Below is the YAML for `BackupConfiguration` CRD to backup the `sample-mysql` database we have deployed earlier,

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: sample-mysql-backup
  namespace: demo
spec:
  schedule: "*/5 * * * *"
  task:
    name: mysql-backup-8.0.14
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-mysql
  retentionPolicy:
    keepLast: 5
    prune: true
```

Here,

- `.spec.schedule` specifies that we want to backup the database at 5 minutes interval.
- `.spec.task.name` specifies the name of the Task CRD that specifies the necessary Functions and their execution order to backup a MySQL database.
- `.spec.target.ref` refers to the AppBinding CRD that was created for `sample-mysql` database.

Let's create the `BackupConfiguration` CRD we have shown above,

```bash
$ kubectl create -f ./docs/examples/backup/backupconfiguration.yaml
backupconfiguration.stash.appscode.com/sample-mysql-backup created
```

**Verify CronJob:**

If everything goes well, Stash will create a CronJob with the schedule specified in `spec.schedule` field of `BackupConfiguration` CRD.

Verify that the CronJob has been created using the following command,

```bash
$ kubectl get cronjob -n demo
NAME                  SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
sample-mysql-backup   */5 * * * *   False     0        <none>          27s
```

**Wait for BackupSession:**

The `sample-mysql-backup` CronJob will trigger a backup on each scheduled slot by creating a `BackupSession` CRD.

Wait for a schedule to appear. Run the following command to watch `BackupSession` CRD,

```bash
$ kubectl get backupsession -n demo -w
NAME                             BACKUPCONFIGURATION   PHASE     AGE
sample-mysql-backup-1564729507   sample-mysql-backup   Running   51s
sample-mysql-backup-1564729507   sample-mysql-backup   Succeeded   51s
```

Here, the phase **`Succeeded`** means that the backupsession has been succeeded.

**Verify Backup:**

Now, we are going to verify whether the backed up data is in the backend. Once a backup is completed, Stash will update the respective `Repository` CRD to reflect the backup completion. Check that the repository `gcs-repo` has been updated by the following command,

```bash
$ kubectl get repository -n demo gcs-repo
NAME       INTEGRITY   SIZE        SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
gcs-repo   true        6.815 MiB   2                3m39s                    30m
```

Now, if we navigate to the GCS bucket, we will see the backed up data has been stored in `demo/mysql/sample-mysql` directory as specified by `.spec.backend.gcs.prefix` field of Repository CRD.

<figure align="center">
  <img alt="Backup data in GCS Bucket" src="/docs/images/sample-mysql-backup.png">
  <figcaption align="center">Fig: Backup data in GCS Bucket</figcaption>
</figure>

> Note: Stash keeps all the backed up data encrypted. So, data in the backend will not make any sense until they are decrypted.

## Restore MySQL

In this section, we are going to restore the database from the backup we have taken in the previous section. We are going to deploy a new database and initialize it from the backup.

**Stop Taking Backup of the Old Database:**

At first, let's stop taking any further backup of the old database so that no backup is taken during restore process. We are going to pause the `BackupConfiguration` crd that we had created to backup the `sample-mysql` database. Then, Stash will stop taking any further backup for this database.

Let's pause the `sample-mysql-backup` BackupConfiguration,

```console
$ kubectl patch backupconfiguration -n demo sample-mysql-backup --type="merge" --patch='{"spec": {"paused": true}}'
backupconfiguration.stash.appscode.com/sample-mysql-backup patched
```

Now, wait for a moment. Stash will pause the BackupConfiguration. Verify that the BackupConfiguration  has been paused,

```console
$ kubectl get backupconfiguration -n demo sample-mysql-backup
NAME                 TASK                  SCHEDULE      PAUSED   AGE
sample-mysql-backup  mysql-backup-8.0.14   */5 * * * *   true     26m
```

Notice the `PAUSED` column. Value `true` for this field means that the BackupConfiguration has been paused.

**Deploy Restored Database:**

Now, we have to deploy the restored database similarly as we have deployed the original `sample-mysql` database. However, this time there will be the following differences:

- We have to use the same secret that was used in the original database. We are going to specify it using `.spec.databaseSecret` field.
- We have to specify `.spec.init` section to tell KubeDB that we are going to use Stash to initialize this database from backup. KubeDB will keep the database phase to **`Initializing`** until Stash finishes its initialization.

Below is the YAML for `MySQL` CRD we are going deploy to initialize from backup,

```yaml
apiVersion: kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: restored-mysql
  namespace: demo
spec:
  version: "8.0.14"
  databaseSecret:
    secretName: sample-mysql-auth
  replicas: 1
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
  init:
    stashRestoreSession:
      name: sample-mysql-restore
  terminationPolicy: WipeOut
```

Here,

- `spec.init.stashRestoreSession.name` specifies the `RestoreSession` CRD name that we will use later to restore the database.

Let's create the above database,

```bash
$ kubectl apply -f ./docs/examples/restore/restored-mysql.yaml
mysql.kubedb.com/restored-mysql created
```

If you check the database status, you will see it is stuck in **`Initializing`** state.

```bash
$ kubectl get my -n demo restored-mysql
NAME             VERSION   STATUS         AGE
restored-mysql   8.0.14    Initializing   61s
```

**Create RestoreSession:**

Now, we need to create a RestoreSession CRD pointing to the AppBinding for this restored database.

Using the following command, check that another AppBinding object has been created for the `restored-mysql` object,

```bash
$ kubectl get appbindings -n demo restored-mysql
NAME             AGE
restored-mysql   6m6s
```

> If you are not using KubeDB to deploy database, create the AppBinding manually.

Below is the contents of YAML file of the RestoreSession CRD that we are going to create to restore backed up data into the newly created database provisioned by MySQL CRD named `restored-mysql`.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: sample-mysql-restore
  namespace: demo
  labels:
    kubedb.com/kind: MySQL # this label is mandatory if you are using KubeDB to deploy the database.
spec:
  task:
    name: mysql-restore-8.0.14
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: restored-mysql
  rules:
    - snapshots: [latest]
```

Here,

- `.metadata.labels` specifies a `kubedb.com/kind: MySQL` label that is used by KubeDB to watch this RestoreSession object.
- `.spec.task.name` specifies the name of the Task CRD that specifies the necessary Functions and their execution order to restore a MySQL database.
- `.spec.repository.name` specifies the Repository CRD that holds the backend information where our backed up data has been stored.
- `.spec.target.ref` refers to the newly created AppBinding object for the `restored-mysql` MySQL object.
- `.spec.rules` specifies that we are restoring data from the latest backup snapshot of the database.

> **Warning:** Label `kubedb.com/kind: MySQL` is mandatory if you are using KubeDB to deploy the database. Otherwise, the database will be stuck in **`Initializing`** state.

Let's create the RestoreSession CRD object we have shown above,

```bash
$ kubectl apply -f ./docs/examples/restore/restoresession.yaml
restoresession.stash.appscode.com/sample-mysql-restore created
```

Once, you have created the RestoreSession object, Stash will create a restore Job. We can watch the phase of the RestoreSession object to check whether the restore process has succeeded or not.

Run the following command to watch the phase of the RestoreSession object,

```bash
$ kubectl get restoresession -n demo sample-mysql-restore -w
NAME                   REPOSITORY-NAME   PHASE       AGE
sample-mysql-restore   gcs-repo          Running     3m15s
sample-mysql-restore   gcs-repo          Succeeded   3m28s
```

Here, we can see from the output of the above command that the restore process succeeded.

**Verify Restored Data:**

In this section, we are going to verify whether the desired data has been restored successfully. We are going to connect to the database server and check whether the database and the table we created earlier in the original database are restored.

At first, check if the database has gone into **`Running`** state by the following command,

```bash
$ kubectl get my -n demo restored-mysql
NAME             VERSION   STATUS    AGE
restored-mysql   8.0.14    Running   34m
```

Now, find out the database Pod by the following command,

```bash
$ kubectl get pods -n demo --selector="kubedb.com/name=restored-mysql"
NAME               READY   STATUS    RESTARTS   AGE
restored-mysql-0   1/1     Running   0          39m
```

And then copy the user name and password of the `root` user to access into `mysql` shell.

> Notice: We used the same Secret for the `restored-mysql` object. So, we will use the same commands as before.

```bash
$ kubectl get secret -n demo  sample-mysql-auth -o jsonpath='{.data.username}'| base64 -d
root⏎

$ kubectl get secret -n demo  sample-mysql-auth -o jsonpath='{.data.password}'| base64 -d
5HEqoozyjgaMO97N⏎
```

Now, let's exec into the Pod to enter into `mysql` shell and create a database and a table,

```bash
$ kubectl exec -it -n demo restored-mysql-0 -- mysql --user=root --password=5HEqoozyjgaMO97N
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 9
Server version: 8.0.14 MySQL Community Server - GPL

Copyright (c) 2000, 2019, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| mysql              |
| performance_schema |
| playground         |
| sys                |
+--------------------+
5 rows in set (0.00 sec)

mysql> SHOW TABLES IN playground;
+----------------------+
| Tables_in_playground |
+----------------------+
| equipment            |
+----------------------+
1 row in set (0.00 sec)

mysql> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
1 row in set (0.00 sec)

mysql> exit
Bye
```

So, from the above output, we can see that the `playground` database and the `equipment` table we created earlier in the original database and now, they are restored successfully.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete restoresession -n demo sample-mysql-restore
restoresession.stash.appscode.com "restore-sample-mysql" deleted

$ kubectl delete backupconfiguration -n demo sample-mysql-backup
backupconfiguration.stash.appscode.com "sample-mysql-backup" deleted

$ kubectl delete my -n demo restored-mysql
mysql.kubedb.com "restored-mysql" deleted

$ kubectl delete my -n demo sample-mysql
mysql.kubedb.com "sample-mysql" deleted
```

To cleanup the MySQL catalogs that we had created earlier, run the following:

<ul class="nav nav-tabs" id="uninstallerTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link" id="helm-uninstaller-tab" data-toggle="tab" href="#helm-uninstaller" role="tab" aria-controls="helm-uninstaller" aria-selected="false">Helm</a>
  </li>
  <li class="nav-item">
    <a class="nav-link active" id="script-uninstaller-tab" data-toggle="tab" href="#script-uninstaller" role="tab" aria-controls="script-uninstaller" aria-selected="true">Script</a>
  </li>
</ul>
<div class="tab-content" id="uninstallerTabContent">
 <!-- ------------ Helm Tab Begins----------- -->
  <div class="tab-pane fade" id="helm-uninstaller" role="tabpanel" aria-labelledby="helm-uninstaller-tab">

### Uninstall  `stash-mysql-*` charts

Run the following script to uninstall `stash-mysql` catalogs that was installed as a Helm chart.

```console
curl -fsSL https://github.com/stashed/catalog/raw/master/deploy/chart.sh | bash -s -- --uninstall --catalog=stash-mysql
```

</div>
<!-- ------------ Helm Tab Ends----------- -->

<!-- ------------ Script Tab Begins----------- -->
<div class="tab-pane fade show active" id="script-uninstaller" role="tabpanel" aria-labelledby="script-uninstaller-tab">

### Uninstall `stash-mysql` catalog YAMLs

Run the following script to uninstall `stash-mysql` catalog that was installed as Kubernetes YAMLs.

```console
curl -fsSL https://github.com/stashed/catalog/raw/master/deploy/script.sh | bash -s -- --uninstall --catalog=stash-mysql
```

</div>
<!-- ------------ Script Tab Ends----------- -->
</div>
