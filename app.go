package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// holds our config
type appConfig struct {
	AppName             string // get from $GVS_APP
	AppEnv              string // get from $GVS_APPENV
	VaultURL            string // get from $GVS_VAULTADDR
	SecretPath          string // get from $GVS_SECRETPATH, computed if not provided
	VaultRoleID         string // get from $GVS_VAULTROLEID, default
	VaultSecretID       string // get from $GVS_VAULTSECRETID, default
	SecretFilePath      string // default: /mnt/ramfs
	SecretAvailabletime string // default: 60 sec
	Token               string
	VaultCredentials    VaultAppRoleCredentials
	vaultKVVersion      int
}

var appCfg appConfig

func init() {
	var err error
	fmt.Println("starting init")
	// get part of config from env
	appCfg.vaultKVVersion = 2
	appCfg.AppName = os.Getenv("GVS_APPNAME")
	appCfg.AppEnv = os.Getenv("GVS_APPENV")
	appCfg.VaultURL = os.Getenv("GVS_VAULTURL")
	appCfg.SecretPath = os.Getenv("GVS_SECRETPATH")
	appCfg.SecretFilePath = os.Getenv("GVS_SECRETFILEPATH")
	appCfg.SecretAvailabletime = os.Getenv("GVS_SECRETAVAILABLETIME")
	appCfg.VaultRoleID = os.Getenv("GVS_VAULTROLEID")
	appCfg.VaultSecretID = os.Getenv("GVS_VAULTSECRETID")

	if len(appCfg.SecretFilePath) == 0 {
		appCfg.SecretFilePath = "/dev/shm/gvs"
	} else {
		appCfg.SecretFilePath = filepath.Join(appCfg.SecretFilePath, "gvs")
	}

	if len(appCfg.SecretAvailabletime) == 0 {
		appCfg.SecretAvailabletime = "60"
	}

	if len(appCfg.SecretPath) == 0 {
		// default to Vault kv v2, default secret/data path
		appCfg.SecretPath = filepath.Join("secret/data", appCfg.AppName, appCfg.AppEnv)
	} else {
		secretParam := strings.Split(appCfg.SecretPath, "/")
		if len(secretParam) == 1 {
			// we get a name, not a full secret path -> default to Vault kv v2, default secret/data path
			appCfg.SecretPath = filepath.Join("secret/data", secretParam[0])
		} else {
			// assume we got a full secret path
			// check kv version
			appCfg.vaultKVVersion, err = getKVVersion(secretParam[0] + "/")
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// read secret to get Vault App Role credentials
	if len(appCfg.VaultRoleID) == 0 {
		appCfg.VaultRoleID = "/run/secret/role_id"
	}

	appCfg.VaultCredentials.RoleID, err = getSecret(appCfg.VaultRoleID)
	if err != nil {
		log.Fatal("Error reading role_id docker secret: ", err)
	}

	if len(appCfg.VaultSecretID) == 0 {
		appCfg.VaultSecretID = "/run/secret/secret_id"
	}

	appCfg.VaultCredentials.SecretID, err = getSecret(appCfg.VaultSecretID)
	if err != nil {
		log.Fatal("Error reading secret_id docker secret: ", err)
	}
	appCfg.Token, err = auth(appCfg.VaultCredentials)
	if err != nil {
		log.Fatal("Vault auth error: ", err)
	}
}

func main() {
	var err error
	fmt.Println("Checking secretfileok")
	secretFileOK, step, errSecretFile := isSecretFilePathOK(appCfg.SecretFilePath)
	if errSecretFile != nil {
		log.Fatal(step, errSecretFile)
	}
	if secretFileOK {
		fmt.Println("secretfileok")
		// read Vault Secrets write them in kv file
		err = publishVaultSecret(appCfg.SecretPath)
		if err != nil {
			log.Fatal("Error processing Vault Secret:", err)
		}
	}
	_ = destroySecretFile(appCfg.SecretFilePath, appCfg.SecretAvailabletime)
}

func getSecret(path string) (secret string, err error) {
	// read from docker secret
	dat, err := ioutil.ReadFile(filepath.Join(path))
	if err != nil {
		return "", err
	}
	return string(dat), nil
}

func mountFS(path, timeout string) error {
	mntPath, err := exec.LookPath("mount")
	if err != nil {
		return (err)
	}
	cmd := exec.Command(mntPath, path)

	var out bytes.Buffer
	cmd.Stdout = &out

	err = cmd.Run()
	if err != nil {
		return (err)
	}
	umntPath, err := exec.LookPath("umount")
	if err != nil {
		return (err)
	}
	sleepPath, err := exec.LookPath("sleep")
	if err != nil {
		return (err)
	}

	cmdUmount := exec.Command("/bin/sh", "-c", sleepPath+" "+timeout+" && "+umntPath+" "+path)
	// capture STDOUT
	cmdUmount.Stdout = &out

	// run cmd
	err = cmdUmount.Start()
	if err != nil {
		return (err)
	}

	return nil
}

func isSecretFilePathOK(path string) (isOK bool, step string, err error) {
	testFile := appCfg.SecretFilePath + ".tmp"
	// create tmp test file
	f, err := os.Create(testFile)
	if err != nil {
		return false, "Create File", err
	}
	defer f.Close()

	_, err = f.WriteString("test\n")
	if err != nil {
		return false, "Write File", err
	}

	f.Sync()

	// remove tmp test file
	err = destroySecretFile(testFile, "0")
	if err != nil {
		return false, "Destroy File", err
	}
	return true, "", nil

}

func destroySecretFile(path, delay string) error {
	rmPath, err := exec.LookPath("rm")
	if err != nil {
		return (err)
	}

	sleepPath, err := exec.LookPath("sleep")
	if err != nil {
		return (err)
	}

	cmdUmount := exec.Command("/bin/sh", "-c", sleepPath+" "+delay+" && "+rmPath+" "+path)
	err = cmdUmount.Start()
	if err != nil {
		return (err)
	}

	return nil
}
