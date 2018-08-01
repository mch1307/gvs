package main

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
)

// holds our config
type appConfig struct {
	App              string   //get from env $GVS_APP
	AppEnv           string   // get from env $GVS_APPENV
	VaultAddr        string   // get from env $GVS_VAULTADDR
	Secrets          []string // get from env $GVS_SECRETS
	Token            string
	secretRootPath   string
	VaultCredentials VaultAppRoleCredentials
}

var appCfg appConfig

func main() {
	// get part of config from env
	err := envconfig.Process("gvs", &appCfg)
	if err != nil {
		log.Fatal("Error processing env variables: ", err)
	}
	appCfg.secretRootPath = filepath.Join(appCfg.App, appCfg.AppEnv)
	// read docker secret to get Vault App Role credentials
	appCfg.VaultCredentials.RoleID, err = getDockerSecret("role_id")
	if err != nil {
		log.Fatal("Error reading role_id docker secret: ", err)
	}
	appCfg.VaultCredentials.SecretID, err = getDockerSecret("secret_id")
	if err != nil {
		log.Fatal("Error reading secret_id docker secret: ", err)
	}
	appCfg.Token, err = auth(appCfg.VaultCredentials)
	if err != nil {
		log.Fatal("Vault auth error: ", err)
	}

	// read Vault Secrets listed in $GVS_SECRETS and set them as GVS_ prefixed env variables
	for _, v := range appCfg.Secrets {
		err = publishVaultSecret(v)
		if err != nil {
			log.Fatal("Error processing Vault Secret:", err)
		}
	}

}

func getDockerSecret(name string) (secret string, err error) {
	// read from docker secret
	dat, err := ioutil.ReadFile(filepath.Join("/run/secret", name))
	if err != nil {
		return "", err
	}
	return string(dat), nil
}
