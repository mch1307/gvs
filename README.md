# gvs

[![Coverage Status](https://coveralls.io/repos/github/mch1307/gvs/badge.svg?branch=dev)](https://coveralls.io/github/mch1307/gvs?branch=dev) [![Go Report Card](https://goreportcard.com/badge/github.com/mch1307/gvs)](https://goreportcard.com/report/github.com/mch1307/gvs) [![Build Status](https://travis-ci.org/mch1307/gvs.svg?branch=master)](https://travis-ci.org/mch1307/gvs)

Simple (linux) container utility to read Vault secrets for a given application. Will put the secret(s) in a file and remove this file after x seconds.

Support secrets stored as key/value for both kv v1 and v2.

## Use Case

A containerized application needs to get secrets from Vault at runtime. The application cannot directly access Vault before boostraping.

The secrets cannot be stored in the image, nor in the source code repo.

Create a Vault app role with relevant policy, create docker or kubernetes secrets to store the role_id and secret_id.

`gvs` will authenticate to Vault with the provided role_id / secret_id and read the application secret(s).

The secret will be stored in a file so that the application can easily access them. This file will be erased after X seconds (60 by default)

## Prerequisites

* `sleep` and `rm` should be available in user PATH

## How It Works

When started, `gvs` will first read it's parameters from `GVS_` prefixed environment variables.

```bash
GVS_APPNAME                 Name of your application
GVS_APPENV                  Environment where the app will run (ie dev, test,..)
GVS_VAULTURL                URL of the Vault server
GVS_SECRETPATH              Path to the Vault secret.
GVS_SECRETTARGETPATH        Path where the secret kv file will be written  (default /dev/shm/gvs)
GVS_SECRETAVAILABLETIME     Number of seconds after which the secret file will be destroyed
GVS_VAULTROLEID             Path to file containing the Vault role id (default run/secrets/role_id)
GVS_VAULTROLESECRETID       Path to file containing the Vault secret id (default /run/secrets/secret_id)
GVS_OUTPUTFORMAT            yaml by default, key=value text if any other value supplied
```

`gvs` will read the Vault role_id and secret_id from files secrets. By convention, those should be called role_id and secret_id and mounted in `/run/secrets/role_id` and `/run/secrets/secret_id` (docker secret). This can be overriden by specifying the full file path in `GVS_VAULTROLEID` and `GVS_VAULTSECRETID` env variables.

Before reading the Vault secret kv(s), it will build the path from the `GVS_APPNAME` and `GVS_APPENV` variables, unless the `GVS_SECRETPATH` is specified.

After having read the secret kv(s) from Vault, it will write a file called `gvs` at `GVS_SECRETFILEPATH`. This file will contain the kv(s) from Vault in the form `key=value`, as in Vault.

This file will be deleted after `GVS_SECRETAVAILABLETIME` number of seconds.

## Example

Application called `demo` running in `dev` environment needs to read the secrets called `mysecret1` and `mysecret2`.

> As the app name is neither environment dependent nor sensitive, it could be stored in the docker image (`GVS_APPNAME`)

### Dockerfile

```Dockerfile
FROM alpine:3.8
ENV GVS_APP=demo
COPY gvs /usr/local/bin
COPY demo.app /demo.app
ENTRYPOINT ["/docker-entrypoint.sh"]
```

### Entrypoint shell

Include gvs in your startup script so that the secrets will be available to your application when starting:

```bash
#!/bin/bash
gvs
.
code to start your app
```

### Run your container

Specify the environment dependent, non sensitive variables when running your container:

`docker run -d -e GVS_APPENV=dev -e GVS_VAULTADDR=https://vault.dev.domain:8200 ...`

Assuming the app and secrets list are stored in the docker image, gvs will get the `GVS_APPENV` and `GVS_VAULTURL` from the docker run command.

It will use them to connect to the right Vault cluster, read the application's secrets available at `$GVS_VAULTURL/v1/secret/data/demo-dev` and write them to `/dev/shm/gvs` so that your app can read them. After `$GVS_SECRETAVAILABLETIME` this file will be removed.

```bash
MYSECRET1=mysecrevalue1
MYSECRET2=mysecrevalue2
```

## Build

```shell
VERSION=$(git log -n1 --pretty="format:%d" | sed "s/, /\n/g" | grep tag: | sed "s/tag: \|)//g") && \
VERSION=$VERSION-$(git log -1 --pretty=format:%h) && \
CGO_ENABLED="0" GOARCH="amd64" GOOS="linux" go build -a -installsuffix cgo -o gvs -ldflags="-s -w -X main.version=$VERSION"
```
