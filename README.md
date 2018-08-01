# vault-secret

Utility to read Vault secrets for a given application. Will put the secret(s) as env var when the container starts.

Read role and secret id from docker secrets. By convention, those should be called role_id and secret_id.

Read Vault address, app name/id, and app environment variable

```
GVS_APP         string
GVS_APPENV      string
GVS_VAULTADDR    string	
```

Get the list of secret(s) to get from environment variable

```
GVS_SECRETS    []string
```

The requested Vault secrets will be available to the container as env var, prefixed with GVS_.


## Example

Application called `demo` running in `dev` environment needs to read the secrets called `mysecret1` and `mysecret2`.

> `docker run -e` only supports simple env variables (no array), so the list of secret (`GVS_SECRETS`) will need to be stored in the docker image.

> As the app name is not environment dependent, it should also be stored in the docker image (`GVS_APP`)

### Dockerfile

```Dockerfile
FROM alpine:3.7
ENV GVS_APP=demo
    GVS_SECRETS="myscret1,mysecret2"
COPY gvs /tmp/gvs
COPY demo.app /demo.app
ENTRYPOINT ["/docker-entrypoint.sh"]
```

### Entrypoint shell

Include gvs in your startup script so that the secrets will be available to your application when starting:

```bash
#!/bin/bash
gvs
source /tmp/gvs.sh
.
.
code to start your app
```

### Run your container

Specify the environment dependent variables when running your container:

`docker run -d -e GVS_APPENV=dev -e GVS_VAULTADDR=http://vault.dev.domain:8200 ...`

Assuming the app and secrets list are stored in the docker image, gvs will get the `GVS_APPENV` and `GVS_VAULTADDR` from the docker run command.

It will use them to connect to the right Vault cluster, read the application's secrets (`GVS_MYSECRET1` & `GVS_MYSECRET2`) and make them accessible to your application
