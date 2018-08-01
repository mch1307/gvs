# vault-secret

Utility to read Vault secrets for a given application. Will put the secret(s) as env var when the container starts.

Read role and secret id from docker secret (role_id and secret_id)

Read Vault address, app name/id, and app environment variable

```
GVS_APP        string
GVS_APPENV     string
GVS_ADDRESS    string
	
```

Get the list of secret(s) to get from environment variable

```
GVS_SECRETS    []string
```

The requested Vault secrets will be available to the container as env var, prefixed with GVS_.


## Example

Application called MYAPP running in DEV environment needs to read the secrets called mysecret1 and mysecret2:

```
export GVS_APP=myapp
export GVS_APPENV=dev
export GVS_ADDRESS="http://myvault.mydomain:8200"
export GVS_SECRETS="mysecret1,mysecret2"
```

Copy the gvs executable in the Dockerfile and embed it in the entrypoint.

The secret will be available to the running container as env variables

```
$GVS_MYSECRET1
$GVS_MYSECRET2
```