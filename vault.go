package main

import (
	"log"

	vault "github.com/hashicorp/vault/api"
)

type vaultConfig struct {
	app      string   //get from env
	appEnv   string   // get from env
	address  string   // get from env
	secrets  []string // get from env
	roleID   string   // get from docker secret
	secretID string
	token    string // get from docker secret
}

var client *vault.Client

func (v *vaultConfig) Init() error {
	var err error
	var token string
	var credential string

	config := vault.DefaultConfig()
	client, err = vault.NewClient(config)
	if err != nil {
		return err
	}

	err = client.SetAddress(v.address)
	if err != nil {
		return err
	}

	// Auth
	if len(v.roleID) == 0 {
		log.Fatal("No Vault credential provided")
	}
	if len(v.secretID) == 0 {
		log.Fatal("No Vault secretID provided")
	}
	login := map[string]interface{}{
		"role_id":   v.roleID,
		"secret_id": v.secretID,
	}

	secret, err := client.Logical().Write("auth/approle/login", login)
	v.token = secret.Auth.ClientToken

	return nil

}
