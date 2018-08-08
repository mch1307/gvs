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

	"gopkg.in/yaml.v2"
)

var vaultAuthResponse VaultAuthResponse
var vaultSecret VaultSecretv2

func getKVVersion(name string) (version int, err error) {

	url := appCfg.VaultURL + "/v1/sys/internal/ui/mounts"
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

	return version, nil
}

func auth(a VaultAppRoleCredentials) (token string, err error) {
	client := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}
	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(a)
	req, err := http.NewRequest("POST", appCfg.VaultURL+"/v1/auth/approle/login", payload)
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
	url := appCfg.VaultURL + filepath.Join("/v1", appCfg.SecretPath)
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
	var secretsList map[string]string
	// parse to Vx and get a simple kv map back
	if appCfg.vaultKVVersion == 2 {
		secretsList, err = parseKVv2(body)
		if err != nil {
			log.Fatal(err)
		}
	} else if appCfg.vaultKVVersion == 1 {
		secretsList, err = parseKVv1(body)
		if err != nil {
			log.Fatal(err)
		}
	}

	// create the secret file
	f, err := os.Create(appCfg.SecretFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if appCfg.OutputFormat == "yaml" {
		output, _ := yaml.Marshal(&secretsList)
		_, _ = f.Write(output)
	} else {
		for k, v := range secretsList {
			_, _ = f.WriteString(strings.ToUpper(k) + "=" + v + "\n")
		}
		f.Sync()
	}
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
