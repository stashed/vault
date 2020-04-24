# stash-mysql

[stash-mysql](https://github.com/stashed/mysql) - MySQL database backup/restore plugin for [Stash by AppsCode](https://appscode.com/products/stash/).

## TL;DR;

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm install stash-mysql-8.0.14 appscode/stash-mysql -n kube-system --version=8.0.14
```

## Introduction

This chart installs necessary `Functions` and `Tasks` definitions to take backup of MySQL-8.0.14 databases and restore them using Stash.

## Prerequisites

- Kubernetes 1.11+

## Installing the Chart

- Add AppsCode chart repository to your helm repository list,

```console
$ helm repo add appscode https://charts.appscode.com/stable/
```

- Update helm repositories to fetch latest charts from the remove repository,

```console
$ helm repo update
```

- Install the chart with the release name `stash-mysql-8.0.14` run the following command,

```console
$ helm install stash-mysql-8.0.14 appscode/stash-mysql -n kube-system --version=8.0.14
```

The above commands installs `Functions` and `Task` CRDs that are necessary to take backup of MySQL-8.0.14 databases and restore them using Stash.

## Uninstalling the Chart

To uninstall/delete the `stash-mysql-8.0.14` run the following command,

```console
helm uninstall stash-mysql-8.0.14 -n kube-system --purge
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the `stash-mysql` chart and their default values.

|     Parameter      |                                                                    Description                                                                     |      Default      |
| :----------------: | -------------------------------------------------------------------------------------------------------------------------------------------------- | :---------------: |
| `image.registry`   | Docker registry used to pull respective images                                                                                                     |     `stashed`     |
| `image.repository` | Docker image used to take backup of MySQL databases and restore them                                                                               |   `stash-mysql`   |
| `image.tag`        | Tag of the image that is used to take backup of MySQL databases and restore them. This is usually same as the database version it can take backup. |       `8.0.14`    |
| `backup.args`      | Optional arguments to pass to `mysqldump` command  during bakcup process                                                                           | `--all-databases` |
| `restore.args`     | Optional arguments to pass to `mysql` command during restore process                                                                               |        ""         |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

For example:

```console
helm install stash-mysql-8.0.14 appscode/stash-mysql -n kube-system ---set docker.registry=my-registry
```
