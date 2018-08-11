package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
)

var vaultAuthResponse VaultAuthResponse
var vaultSecret VaultSecretv2

func (a *gvsConfig) getKVVersion(name string) error {

	url := a.VaultURL + "/v1/sys/internal/ui/mounts"
	client := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", a.VaultToken)

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	var vaultRsp VaultMountListRespone

	jsonErr := json.Unmarshal(body, &vaultRsp)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	var mountInfo VaultSecretMounts = make(map[string]*VaultSecretMount)

	err = mountInfo.UnmarshalJSON([]byte(vaultRsp.Data.Secret))
	if err != nil {
		log.Fatal(err)
	}
	var version int
	for _, v := range mountInfo {
		if v.Name == name {
			if len(v.Options) > 0 {
				switch v.Options["version"].(type) {
				case int:
					version = v.Options["version"].(int)
				default:
					version = 1
				}
			} else {
				//kv v1
				version = 1
			}
		}
	}
	a.vaultKVVersion = version

	return nil
}

func (a *gvsConfig) getVaultAppRoleToken() error {
	client := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}
	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(a.VaultCredentials)
	req, err := http.NewRequest("POST", a.VaultURL+"/v1/auth/approle/login", payload)
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
	a.VaultToken = vaultAuthResponse.Auth.ClientToken

	return nil
}

func (a *gvsConfig) getVaultSecret(path string) (kv map[string]string, err error) {
	secretsList := make(map[string]string)
	url := a.VaultURL + filepath.Join("/v1", path)
	client := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", a.VaultToken)

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	// parse to Vx and get a simple kv map back
	if a.vaultKVVersion == 2 {
		secretsList, err = parseKVv2(body)
		if err != nil {
			log.Fatal(err)
		}
	} else if a.vaultKVVersion == 1 {
		secretsList, err = parseKVv1(body)
		if err != nil {
			log.Fatal(err)
		}
	}
	return secretsList, nil
}

func (a *gvsConfig) publishVaultSecret() error {
	secretsList := make(map[string]string)

	if len(a.SecretList[0]) > 0 {
		for _, v := range a.SecretList {
			kvMap, err := a.getVaultSecret(filepath.Join(a.VaultSecretPath, v))
			if err != nil {
				return err
			}
			for k, v := range kvMap {
				secretsList[k] = v
			}
		}
	} else {
		kvMap, err := a.getVaultSecret(a.VaultSecretPath)
		if err != nil {
			return err
		}
		for k, v := range kvMap {
			secretsList[k] = v
		}
	}

	// create the secret file
	f, err := os.Create(a.SecretFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if a.OutputFormat == "yaml" {
		output, _ := yaml.Marshal(&secretsList)
		_, _ = f.Write(output)
	} else {
		for k, v := range secretsList {
			_, _ = f.WriteString(k + "=" + v + "\n")
		}
	}
	f.Sync()
	return err

}

func parseKVv2(rawSecret []byte) (gs map[string]string, err error) {
	var secret VaultSecretv2
	gs = make(map[string]string)
	err = json.Unmarshal(rawSecret, &secret)
	if err != nil {
		log.Fatal(err)
	}
	for k, v := range secret.Data.Data {
		gs[k] = v
	}

	return gs, nil
}

func parseKVv1(rawSecret []byte) (gs map[string]string, err error) {
	var secret VaultSecret
	gs = make(map[string]string)
	err = json.Unmarshal(rawSecret, &secret)
	if err != nil {
		log.Fatal(err)
	}
	for k, v := range secret.Data {
		gs[k] = v
	}

	return gs, nil
}

// UnmarshalJSON unmarshal Vault JSON response to /v1/sys/internal/ui/mounts
func (p *VaultSecretMounts) UnmarshalJSON(data []byte) error {
	var transient = make(map[string]*VaultSecretMount)

	err := json.Unmarshal(data, &transient)
	if err != nil {
		return err
	}
	for k, v := range transient {
		v.Name = k
		(*p)[k] = v
	}
	return nil
}
