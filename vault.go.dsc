package main

import (
	"github.com/hashicorp/vault/api"
)

var client *api.Client

// func (v *vaultConfig) Init() error {
// 	var err error

// 	config := api.DefaultConfig()
// 	client, err = api.NewClient(config)
// 	if err != nil {
// 		return err
// 	}

// 	err = client.SetAddress(v.Address)
// 	if err != nil {
// 		return err
// 	}
// 	v.secretRootPath = "secret" + "/" + path.Join(v.App, v.AppEnv)

// 	// Auth
// 	if len(v.RoleID) == 0 {
// 		log.Fatal("No Vault roleID provided")
// 	}
// 	if len(v.SecretID) == 0 {
// 		log.Fatal("No Vault secretID provided")
// 	}
// 	login := map[string]interface{}{
// 		"role_id":   v.RoleID,
// 		"secret_id": v.SecretID,
// 	}

// 	secret, err := client.Logical().Write("auth/approle/login", login)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	v.token = secret.Auth.ClientToken
// 	client.SetToken(v.token)
// 	client.SetToken("cb973b3b-2121-a26d-baeb-354cbb3233e5")

// 	return nil

// }

// func (v *vaultConfig) readSecret(name string) (val string, err error) {
// 	fmt.Println("get secret ", name)
// 	fmt.Println("secret path", v.secretRootPath)

// 	secret, err := client.Logical().Read("secret/data/demo/oldsecret")
// 	if err != nil {
// 		return "", err
// 	}
// 	tk, _ := secret.TokenID()
// 	fmt.Println("tokenid: " + tk)
// 	return secret.Data["data"].(string), nil
// }
