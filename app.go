package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// holds our config
type gvsConfig struct {
	AppName             string   // get from $GVS_APP
	AppEnv              string   // get from $GVS_APPENV
	VaultURL            string   // get from $GVS_VAULTADDR
	VaultSecretPath     string   // get from $GVS_SECRETPATH, computed if not provided
	VaultRoleID         string   // get from $GVS_VAULTROLEID, default
	VaultSecretID       string   // get from $GVS_VAULTSECRETID, default
	SecretFilePath      string   // default: /mnt/ramfs
	SecretAvailabletime string   // default: 60 sec
	SecretList          []string // list of secrets the application want to get
	VaultToken          string
	VaultCredentials    VaultAppRoleCredentials
	vaultKVVersion      int
	OutputFormat        string
}

func errInfo() (info string) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.Function + ":" + strconv.Itoa(frame.Line)
}

var gvs gvsConfig

func init() {
	var err error
	// get config from env
	gvs.vaultKVVersion = 2
	gvs.AppName = os.Getenv("GVS_APPNAME")
	gvs.AppEnv = os.Getenv("GVS_APPENV")
	gvs.VaultURL = os.Getenv("GVS_VAULTURL")
	gvs.VaultSecretPath = strings.TrimSuffix(strings.TrimPrefix(os.Getenv("GVS_SECRETPATH"), "/"), "/")
	gvs.SecretFilePath = os.Getenv("GVS_SECRETFILEPATH")
	gvs.SecretAvailabletime = os.Getenv("GVS_SECRETAVAILABLETIME")
	gvs.VaultRoleID = os.Getenv("GVS_VAULTROLEID")
	gvs.VaultSecretID = os.Getenv("GVS_VAULTSECRETID")
	gvs.OutputFormat = os.Getenv("GVS_OUTPUTFORMAT")

	// replace nil values with defaults when applicable
	if len(gvs.SecretFilePath) == 0 {
		gvs.SecretFilePath = "/dev/shm/gvs"
	} else {
		gvs.SecretFilePath = filepath.Join(gvs.SecretFilePath, "gvs")
	}

	if len(gvs.VaultSecretID) == 0 {
		gvs.VaultSecretID = "/run/secret/secret_id"
	}

	if len(gvs.SecretAvailabletime) == 0 {
		gvs.SecretAvailabletime = "60"
	}

	if len(gvs.VaultRoleID) == 0 {
		gvs.VaultRoleID = "/run/secret/role_id"
	}

	if len(gvs.OutputFormat) == 0 {
		gvs.OutputFormat = "yaml"
	}

	// get Vault App Role credentials
	gvs.VaultCredentials.RoleID, err = getSecretFromFile(gvs.VaultRoleID)
	if err != nil {
		log.Fatal("Error reading role_id docker secret: ", err)
	}

	gvs.VaultCredentials.SecretID, err = getSecretFromFile(gvs.VaultSecretID)
	if err != nil {
		log.Fatal("Error reading secret_id docker secret: ", err)
	}

	// get Vault App Role token
	err = gvs.getVaultAppRoleToken()
	if err != nil {
		log.Fatal("Vault auth error: ", err)
	}

	// get Vault kv version
	if len(gvs.VaultSecretPath) == 0 {
		// default to Vault kv v2, default secret/data path
		gvs.VaultSecretPath = filepath.Join("secret/data", gvs.AppName, gvs.AppEnv)
	} else {
		secretParam := strings.Split(gvs.VaultSecretPath, "/")
		if len(secretParam) == 1 {
			// we get a name, not a full secret path -> default to Vault kv v2, default secret/data path
			gvs.VaultSecretPath = filepath.Join("secret/data", secretParam[0])
		} else {
			// assume we got a full secret path
			// check kv version
			err = gvs.getKVVersion(secretParam[0] + "/")
			if err != nil {
				log.Fatal(err)
			}
			// Get the list of secrets from ENV
			gvs.SecretList = strings.Split(os.Getenv("GVS_SECRETLIST"), ",")
		}
	}
}

func main() {
	var err error
	//fmt.Println("Checking secretfileok")
	secretFileOK, errSecretFile := gvs.isSecretFilePathOK()
	if errSecretFile != nil {
		log.Fatal(errors.WithStack(errSecretFile))
	}
	if secretFileOK {
		// read Vault Secrets write them in kv file
		err = gvs.publishVaultSecret()
		if err != nil {
			//log.Fatalf("%+v", errors.WithStack(err))
			log.Fatal(errors.WithStack(err))
		}
	}
	_ = destroySecretFile(gvs.SecretFilePath, gvs.SecretAvailabletime)
}

func getSecretFromFile(path string) (secret string, err error) {
	// read from docker secret
	dat, err := ioutil.ReadFile(filepath.Join(path))
	if err != nil {
		return "", errors.Wrap(err, errInfo())
	}
	return string(dat), nil
}

func (a *gvsConfig) isSecretFilePathOK() (isOK bool, err error) {
	testFile := a.SecretFilePath + ".tmp"
	// create tmp test file
	f, err := os.Create(testFile)
	if err != nil {
		return false, errors.Wrap(err, errInfo())
	}
	defer f.Close()

	_, err = f.WriteString("test\n")
	if err != nil {
		return false, errors.Wrap(err, errInfo())
	}

	f.Sync()

	// remove tmp test file
	err = destroySecretFile(testFile, "0")
	if err != nil {
		return false, errors.Wrap(err, errInfo())
	}
	return true, nil

}

func destroySecretFile(path, delay string) error {
	rmPath, err := exec.LookPath("rm")
	if err != nil {
		return errors.Wrap(err, errInfo())
	}

	sleepPath, err := exec.LookPath("sleep")
	if err != nil {
		return errors.Wrap(err, errInfo())
	}

	cmdUmount := exec.Command("/bin/sh", "-c", sleepPath+" "+delay+" && "+rmPath+" "+path)
	err = cmdUmount.Start()
	if err != nil {
		return errors.Wrap(err, errInfo())
	}

	return nil
}
