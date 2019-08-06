# stash-mysql

[stash-mysql](https://github.com/stashed/mysql) - MySQL database backup/restore plugin for [Stash by AppsCode](https://appscode.com/products/stash/).

## TL;DR;

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm install appscode/stash-mysql --name=stash-mysql-8.0.14 --version=8.0.14
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
$ helm install appscode/stash-mysql --name=stash-mysql-8.0.14 --version=8.0.14
```

The above commands installs `Functions` and `Task` CRDs that are necessary to take backup of MySQL-8.0.14 databases and restore them using Stash.

## Uninstalling the Chart

To uninstall/delete the `stash-mysql-8.0.14` run the following command,

```console
helm delete stash-mysql-8.0.14 --purge
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the `stash-mysql` chart and their default values.

|     Parameter     |                                                                    Description                                                                     |      Default      |
| :---------------: | -------------------------------------------------------------------------------------------------------------------------------------------------- | :---------------: |
| `docker.registry` | Docker registry used to pull respective images                                                                                                     |     `stashed`     |
| `docker.image`    | Docker image used to take backup of MySQL databases and restore them                                                                               |   `stash-mysql`   |
| `docker.tag`      | Tag of the image that is used to take backup of MySQL databases and restore them. This is usually same as the database version it can take backup. |       `8.0.14`       |
| `backup.myArgs`   | Optional arguments to pass to `mysqldump` command  during bakcup process                                                                           | `--all-databases` |
| `restore.myArgs`  | Optional arguments to pass to `mysql` command during restore process                                                                               |        ""         |
| `metrics.enabled` | Specifies whether to send Prometheus metrics                                                                                                       |      `true`       |
| `metrics.labels`  | Optional comma separated labels to add to the Prometheus metrics                                                                                   |        []         |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

For example:

```console
helm install --name stash-mysql-8.0.14 --set metrics.enabled=false appscode/stash-mysql
```

**Tips:** Use escape character (`\`) while providing multiple comma-separated labels for `metrics.labels`.

```console
 helm install chart/stash-mysql --set metrics.labels="k1=v1\,k2=v2"
```
