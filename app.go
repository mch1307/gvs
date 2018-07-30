package main

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

var vaultCfg vaultConfig

func main() {
	err := envconfig.Process("gvs", &vaultCfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	vaultCfg.credential, err = getDockerSecret("credential")
	if err != nil {
		log.Fatal(err)
	}
	vaultCfg.secretID, err = getDockerSecret("secret")
	if err != nil {
		log.Fatal(err)
	}
	err = vaultCfg.Init()
	if err != nil {
		log.Fatal(err)
	}

}

func getDockerSecret(name string) (secret string, err error) {
	// read from docker secret
	secret = "toto"
	return secret, nil
}

// 	GVS_APP        string   //get from env
//	GVS_APPENV     string   // get from env
//	GVS_ADDRESS    string   // get from env
//	GVS_SECRETS    []string // get from env
