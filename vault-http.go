package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var vaultAuthResponse VaultAuthResponse
var vaultSecret VaultSecretv2

func (a *gvsConfig) getVaultAppRoleToken() error {
	client := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}
	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(a.VaultCredentials)
	req, err := http.NewRequest("POST", a.VaultURL+"/v1/auth/approle/login", payload)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), errInfo())
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), errInfo())
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return errors.Wrap(errors.WithStack(err), errInfo())
	}

	jsonErr := json.Unmarshal(body, &vaultAuthResponse)
	if jsonErr != nil {
		return errors.Wrap(errors.WithStack(err), errInfo())
	}

	a.VaultToken = vaultAuthResponse.Auth.ClientToken

	return nil
}

func (a *gvsConfig) getKVVersion(name string) error {

	url := a.VaultURL + "/v1/sys/internal/ui/mounts"
	log.Debug("vault url: ", a.VaultURL+"/v1/sys/internal/ui/mounts")
	client := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), errInfo())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", a.VaultToken)

	res, err := client.Do(req)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), errInfo())
	}
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return errors.Wrap(errors.WithStack(readErr), errInfo())
	}

	var vaultRsp VaultMountListRespone

	jsonErr := json.Unmarshal(body, &vaultRsp)
	if jsonErr != nil {
		return errors.Wrap(errors.WithStack(jsonErr), errInfo())
	}
	log.Debug("Vault mounts: ", string(body))
	var mountInfo VaultSecretMounts = make(map[string]*VaultSecretMount)

	err = mountInfo.UnmarshalJSON([]byte(vaultRsp.Data.Secret))
	if err != nil {
		return errors.Wrap(errors.WithStack(err), errInfo())
	}
	var version string
	for _, v := range mountInfo {
		if v.Name == name {
			log.Debugf("Selected Vault mount: %+v", v)
			if len(v.Options) > 0 {
				switch v.Options["version"].(type) {
				case string:
					version = v.Options["version"].(string)
				default:
					version = "1"
				}
			} else {
				//kv v1
				version = "1"
			}
		}
	}
	a.VaultKvVersion = version
	log.Debugf("%v is vault kv v %v", name, version)
	return nil
}

func (a *gvsConfig) publishVaultSecret() error {
	secretsList := make(map[string]string)
	if len(a.SecretList) > 0 {
		for _, v := range a.SecretList {
			kvMap, err := a.getVaultSecret(a.VaultSecretPath + "/" + v)
			if err != nil {
				return errors.Wrap(errors.WithStack(err), errInfo())
			}
			for k, v := range kvMap {
				secretsList[k] = v
			}
		}
	} else {
		kvMap, err := a.getVaultSecret(a.VaultSecretPath)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), errInfo())
		}
		for k, v := range kvMap {
			secretsList[k] = v
		}
	}

	// add GVS_APPNAME & GVS_APPENV to secretfile
	secretsList["GVS_APPNAME"] = gvs.AppName
	secretsList["GVS_APPENV"] = gvs.AppEnv

	log.Debugf("our secretsList: %v", secretsList)
	// create the secret file
	f, err := os.Create(a.SecretFilePath)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), errInfo())
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

func (a *gvsConfig) getVaultSecret(path string) (kv map[string]string, err error) {
	secretsList := make(map[string]string)
	url := a.VaultURL + "/v1/" + path
	log.Debug("Vault url: ", url)
	client := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return secretsList, errors.Wrap(errors.WithStack(err), errInfo())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", a.VaultToken)

	res, err := client.Do(req)
	if err != nil {
		return secretsList, errors.Wrap(errors.WithStack(err), errInfo())
	}
	body, readErr := ioutil.ReadAll(res.Body)
	log.Debugf("Vault response: %v", string(body))
	if readErr != nil {
		return secretsList, errors.Wrap(errors.WithStack(err), errInfo())
	}
	log.Debugf("Vault kv at %v is v%v, sending to parse function", path, a.VaultKvVersion)
	// parse to Vx and get a simple kv map back
	if a.VaultKvVersion == "2" {
		secretsList, err = parseKVv2(body)
		if err != nil {
			return secretsList, errors.Wrap(errors.WithStack(err), errInfo())
		}
	} else if a.VaultKvVersion == "1" {
		secretsList, err = parseKVv1(body)
		if err != nil {
			return secretsList, errors.Wrap(errors.WithStack(err), errInfo())
		}
	}
	log.Debugf("%+v", secretsList)
	return secretsList, nil
}

func parseKVv2(rawSecret []byte) (gs map[string]string, err error) {
	var secret VaultSecretv2
	gs = make(map[string]string)
	err = json.Unmarshal(rawSecret, &secret)
	if err != nil {
		return gs, errors.Wrap(errors.WithStack(err), errInfo())
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
		return gs, errors.Wrap(errors.WithStack(err), errInfo())
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
		return errors.Wrap(errors.WithStack(err), errInfo())
	}
	for k, v := range transient {
		v.Name = k
		(*p)[k] = v
	}
	return nil
}
