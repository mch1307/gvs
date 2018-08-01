package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// type VaultSecretv2 struct {
// 	RequestID     string `json:"request_id"`
// 	LeaseID       string `json:"lease_id"`
// 	Renewable     bool   `json:"renewable"`
// 	LeaseDuration int    `json:"lease_duration"`
// 	Data          struct {
// 		Data struct {
// 			Value string `json:"value"`
// 		} `json:"data"`
// 		Metadata struct {
// 			CreatedTime  time.Time `json:"created_time"`
// 			DeletionTime string    `json:"deletion_time"`
// 			Destroyed    bool      `json:"destroyed"`
// 			Version      int       `json:"version"`
// 		} `json:"metadata"`
// 	} `json:"data"`
// 	WrapInfo interface{} `json:"wrap_info"`
// 	Warnings interface{} `json:"warnings"`
// 	Auth     interface{} `json:"auth"`
// }

type VaultSecretv2 struct {
	RequestID     string `json:"request_id"`
	LeaseID       string `json:"lease_id"`
	Renewable     bool   `json:"renewable"`
	LeaseDuration int    `json:"lease_duration"`
	Data          struct {
		Data     map[string]string `json:"data"`
		Metadata struct {
			CreatedTime  time.Time `json:"created_time"`
			DeletionTime string    `json:"deletion_time"`
			Destroyed    bool      `json:"destroyed"`
			Version      int       `json:"version"`
		} `json:"metadata"`
	} `json:"data"`
	WrapInfo interface{} `json:"wrap_info"`
	Warnings interface{} `json:"warnings"`
	Auth     interface{} `json:"auth"`
}

type VaultAppRoleCredntials struct {
	RoleID   string `json:"role_id"`
	SecretID string `json:"secret_id"`
}

type VaultAuthResponse struct {
	RequestID     string      `json:"request_id"`
	LeaseID       string      `json:"lease_id"`
	Renewable     bool        `json:"renewable"`
	LeaseDuration int         `json:"lease_duration"`
	Data          interface{} `json:"data"`
	WrapInfo      interface{} `json:"wrap_info"`
	Warnings      interface{} `json:"warnings"`
	Auth          struct {
		ClientToken string   `json:"client_token"`
		Accessor    string   `json:"accessor"`
		Policies    []string `json:"policies"`
		Metadata    struct {
			RoleName string `json:"role_name"`
		} `json:"metadata"`
		LeaseDuration int    `json:"lease_duration"`
		Renewable     bool   `json:"renewable"`
		EntityID      string `json:"entity_id"`
	} `json:"auth"`
}

var vaultAuthResponse VaultAuthResponse
var vaultSecret VaultSecretv2

func auth(a VaultAppRoleCredntials) (token string, err error) {
	client := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}
	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(a)
	req, err := http.NewRequest("POST", vaultCfg.Address+"/v1/auth/approle/login", payload)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	jsonErr := json.Unmarshal(body, &vaultAuthResponse)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return vaultAuthResponse.Auth.ClientToken, nil
}

func publishVaultSecret(name string) error {
	url := vaultCfg.Address + "/v1/secret/data/demo/" + name
	client := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", vaultCfg.token)

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	jsonErr := json.Unmarshal(body, &vaultSecret)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	fmt.Println(vaultSecret.Data.Data)
	for k, v := range vaultSecret.Data.Data {
		if k != "value" {
			os.Setenv("GVS_"+strings.ToUpper(k), v)
		} else {
			os.Setenv("GVS_"+strings.ToUpper(name), v)
		}
	}

	return nil

}
