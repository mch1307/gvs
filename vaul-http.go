package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// VaultRawSecret holds both v1 and v2 Vault secret
type VaultRawSecret struct {
	RequestID     string          `json:"request_id"`
	LeaseID       string          `json:"lease_id"`
	Renewable     bool            `json:"renewable"`
	LeaseDuration int             `json:"lease_duration"`
	Data          json.RawMessage `json:"data"`
	WrapInfo      interface{}     `json:"wrap_info"`
	Warnings      interface{}     `json:"warnings"`
	Auth          interface{}     `json:"auth"`
}

// VaultV2SecretData holds the Vault kv v2 secret data
type VaultV2SecretData struct {
	Data struct {
		Data     map[string]string `json:"data"`
		Metadata struct {
			CreatedTime  time.Time `json:"created_time"`
			DeletionTime string    `json:"deletion_time"`
			Destroyed    bool      `json:"destroyed"`
			Version      int       `json:"version"`
		} `json:"metadata"`
	} `json:"data"`
}

// VaultV1SecretData holds the Vault kv v1 secret data
type VaultV1SecretData struct {
	Data map[string]string `json:"data"`
}

// VaultSecretv2 holds the Vault secret (kv v2)
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

// VaultSecret holds the Vault secret (kv v1)
type VaultSecret struct {
	RequestID     string            `json:"request_id"`
	LeaseID       string            `json:"lease_id"`
	Renewable     bool              `json:"renewable"`
	LeaseDuration int               `json:"lease_duration"`
	Data          map[string]string `json:"data"`
	WrapInfo      interface{}       `json:"wrap_info"`
	Warnings      interface{}       `json:"warnings"`
	Auth          interface{}       `json:"auth"`
}

// VaultAppRoleCredentials holds the role and secret id for Vault approle auth
type VaultAppRoleCredentials struct {
	RoleID   string `json:"role_id"`
	SecretID string `json:"secret_id"`
}

// VaultAuthResponse holds the Vault auth response, used to get the Clienttoken
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

func auth(a VaultAppRoleCredentials) (token string, err error) {
	client := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}
	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(a)
	req, err := http.NewRequest("POST", appCfg.VaultAddr+"/v1/auth/approle/login", payload)
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

func publishVaultSecret(path string) error {
	url := appCfg.VaultAddr + filepath.Join("/v1/secret/data/", appCfg.secretRootPath)
	//url := appCfg.VaultAddr + "/v1/kv/demo/" + name
	client := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", appCfg.Token)

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

	// create a shell script that will export the secret to env variables
	f, err := os.Create("/tmp/gvs.sh")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString("#!/bin/bash\n")

	for k, v := range vaultSecret.Data.Data {
		if k != "value" {
			_, _ = f.WriteString("export " + strings.ToUpper(appCfg.App) + "_" + strings.ToUpper(k) + "=" + v + "\n")
			//os.Setenv("GVS_"+strings.ToUpper(k), v)
		} else {
			_, _ = f.WriteString("export " + strings.ToUpper(appCfg.App) + "_" + strings.ToUpper(path) + "=" + v + "\n")
			//os.Setenv("GVS_"+strings.ToUpper(name), v)
		}
	}
	_, err = f.WriteString("rm -rf /tmp/gvs.sh\n")
	f.Sync()
	return nil

}
