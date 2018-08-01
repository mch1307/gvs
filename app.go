package main

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
)

type vaultConfig struct {
	App            string   //get from env
	AppEnv         string   // get from env
	Address        string   // get from env
	Secrets        []string // get from env
	RoleID         string   // get from docker secret
	SecretID       string   // get from docker secret
	token          string
	secretRootPath string
	credentials    VaultAppRoleCredntials
}

var vaultCfg vaultConfig

func main() {
	err := envconfig.Process("gvs", &vaultCfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	//fmt.Println("printing config")
	//fmt.Println(vaultCfg.Address)
	vaultCfg.credentials.RoleID, err = getDockerSecret("role_id")
	if err != nil {
		log.Fatal(err)
	}
	vaultCfg.credentials.SecretID, err = getDockerSecret("secret_id")
	if err != nil {
		log.Fatal(err)
	}
	// err = vaultCfg.Init()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	vaultCfg.token, err = auth(vaultCfg.credentials)
	//fmt.Println(vaultCfg.token)
	// publish secret
	for _, v := range vaultCfg.Secrets {
		_ = publishVaultSecret(v)
	}

}

func getDockerSecret(name string) (secret string, err error) {
	// read from docker secret
	dat, err := ioutil.ReadFile(filepath.Join("/run/secret", name))
	if err != nil {
		return "", err
	}
	secret = string(dat)
	return secret, nil
}

// func publishSecret(name string) error {
// 	secret, err := readVaultSecret(name)
// 	if err != nil {
// 		return err
// 	}
// 	os.Setenv("GVS_"+strings.ToUpper(name), secret)

// 	return nil
// }
