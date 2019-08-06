[![Go Report Card](https://goreportcard.com/badge/stash.appscode.dev/mysql)](https://goreportcard.com/report/stash.appscode.dev/mysql)
[![Build Status](https://travis-ci.org/stashed/mysql.svg?branch=master)](https://travis-ci.org/stashed/mysql)
[![Docker Pulls](https://img.shields.io/docker/pulls/stashed/stash-mysql.svg)](https://hub.docker.com/r/stashed/stash-mysql/)
[![Slack](https://slack.appscode.com/badge.svg)](https://slack.appscode.com)
[![Twitter](https://img.shields.io/twitter/follow/appscodehq.svg?style=social&logo=twitter&label=Follow)](https://twitter.com/intent/follow?screen_name=AppsCodeHQ)

# MySQL

MySQL backup and restore plugin for [Stash by AppsCode](https://appscode.com/products/stash).

## Install

Install MySQL 8.0.14 backup or restore plugin for Stash as below.

```console
helm repo add appscode https://charts.appscode.com/stable/
helm repo update
helm install appscode/stash-mysql --name=stash-mysql-8.0.14 --version=8.0.14
```

To install catalog for all supported MySQL versions, please visit [here](https://github.com/stashed/catalog).

## Uninstall

Uninstall MySQL 8.0.14 backup or restore plugin for Stash as below.

```console
helm delete stash-mysql-8.0.14
```

## Support

We use Slack for public discussions. To chit chat with us or the rest of the community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

If you have found a bug with Stash or want to request for new features, please [file an issue](https://github.com/stashed/stash/issues/new).
