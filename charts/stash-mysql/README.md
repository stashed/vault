# stash-mysql

[stash-mysql](https://github.com/stashed/mysql) - MySQL database backup/restore plugin for [Stash by AppsCode](https://stash.run)

## TL;DR;

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm install stash-mysql-8.0.14 appscode/stash-mysql -n kube-system --version=8.0.14
```

## Introduction

This chart deploys necessary `Function` and `Task` definition to backup or restore MySQL database 8.0.14 using Stash on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

- Kubernetes 1.11+

## Installing the Chart

To install the chart with the release name `stash-mysql-8.0.14`:

```console
$ helm install stash-mysql-8.0.14 appscode/stash-mysql -n kube-system --version=8.0.14
```

The command deploys necessary `Function` and `Task` definition to backup or restore MySQL database 8.0.14 using Stash on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `stash-mysql-8.0.14`:

```console
$ helm delete stash-mysql-8.0.14 -n kube-system
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the `stash-mysql` chart and their default values.

|    Parameter     |                                                         Description                                                         |       Default       |
|------------------|-----------------------------------------------------------------------------------------------------------------------------|---------------------|
| nameOverride     | Overrides name template                                                                                                     | `""`                |
| fullnameOverride | Overrides fullname template                                                                                                 | `""`                |
| image.registry   | Docker registry used to pull MySQL addon image                                                                              | `stashed`           |
| image.repository | Docker image used to backup/restore MySQL database                                                                          | `stash-mysql`       |
| image.tag        | Tag of the image that is used to backup/restore MySQL database. This is usually same as the database version it can backup. | `"8.0.14"`          |
| backup.args      | Arguments to pass to `mysqldump` command  during bakcup process                                                             | `"--all-databases"` |
| restore.args     | Arguments to pass to `mysql` command during restore process                                                                 | `""`                |
| waitTimeout      | Number of seconds to wait for the database to be ready before backup/restore process.                                       | `300`               |


Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example:

```console
$ helm install stash-mysql-8.0.14 appscode/stash-mysql -n kube-system --version=8.0.14 --set image.registry=stashed
```

Alternatively, a YAML file that specifies the values for the parameters can be provided while
installing the chart. For example:

```console
$ helm install stash-mysql-8.0.14 appscode/stash-mysql -n kube-system --version=8.0.14 --values values.yaml
```
