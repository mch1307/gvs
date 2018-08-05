# gvs

Simple container utility to read Vault secrets (kv v2) for a given application. Will put the secret(s) in a file on a tmpfs mount.

## Pre Requisites

* `sleep` and `rm` should be available in user PATH

## What It Does

When started, `gvs` will first read it's parameters from `GVS_` prefixed environment variables.

```
GVS_APPNAME                 string  Name of your application (can be stored in the container)
GVS_APPENV                  string  Environment where the app will run (ie dev, test,..)
GVS_VAULTURL                string  URL of the Vault server
GVS_SECRETPATH              string  Path where the Vault secret will be read. 
GVS_SECRETTARGETPATH        string  Path where the secret kv file will be written  (default /mnt/ramfs)
GVS_SECRETAVAILABLETIME     string  Number of seconds after which the secret file will be destroyed
GVS_VAULTROLEID             string  Path to file containing the Vault role id (default run/secret/role_id)
GVS_VAULTROLESECRETID       string  Path to file containing the Vault secret id (dedault /run/secret/secret_id)
```

It will then read the Vault role_id and secret_id from docker secrets. By convention, those should be called role_id and secret_id so that they are mounted in `/run/secret/role_id` and `/run/secret/secret_id`. This can be overriden by specifying the `GVS_VAULTROLEID` and `GVS_VAULTSECRETID` env variables.

Before reading the Vault secret kv(s), it will build the path from the `GVS_APPNAME` and `GVS_APPENV` variables, unless the `GVS_SECRETPATH` is specified.

After having read the secret kv(s) from Vault, it will write a file called `gvs` at `GVS_SECRETFILEPATH`. This file will contain the kv(s) from Vault in the form `key=value`, as in Vault.

This file will be deleted after `GVS_SECRETAVAILABLETIME` number of seconds.


## Example

Application called `demo` running in `dev` environment needs to read the secrets called `mysecret1` and `mysecret2`.

> As the app name is neither environment dependent nor sensitive, it could be stored in the docker image (`GVS_APPNAME`)

### Dockerfile

```Dockerfile
FROM alpine:3.7
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

> __tini__
> 
> In case you are using tini, you will need to add the `-s` tini flag in order to make sure tini will properly handle termination signals
>
> For example:
>
> `ENTRYPOINT ["gvs &&","/sbin/tini", "-s", "--", "/usr/local/bin/jenkins.sh"]`


### Run your container

Specify the environment dependent, non sensitive variables when running your container:

`docker run -d -e GVS_APPENV=dev -e GVS_VAULTADDR=https://vault.dev.domain:8200 ...`

Assuming the app and secrets list are stored in the docker image, gvs will get the `GVS_APPENV` and `GVS_VAULTURL` from the docker run command.

It will use them to connect to the right Vault cluster, read the application's secrets available at `$GVS_VAULTURL/v1/secret/data/demo-dev` and write them to `/mnt/ramfs/gvs` so that your app can read them. After `$GVS_SECRETAVAILABLETIME` this file will be removed.

```bash
MYSECRET1=mysecrevalue1
MYSECRET2=mysecrevalue2
```

## Build

`CGO_ENABLED="0" GOARCH="amd64" GOOS="linux" go build -a -installsuffix cgo -o gvs -ldflags="-s -w"`
