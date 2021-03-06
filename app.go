package main

import (
	//"errors"

	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	vault "github.com/mch1307/vaultlib"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

const envVaultAddr = "GVS_VAULTURL"
const envAppName = "GVS_APPNAME"
const envAppEnv = "GVS_APPENV"
const envSecretPath = "GVS_SECRETPATH"
const envSecretFilePath = "GVS_SECRETTARGETPATH"
const envSecretAvailableTime = "GVS_SECRETAVAILABLETIME"
const envVaultRoleID = "GVS_VAULTROLEID"
const envVaultSecretID = "GVS_VAULTSECRETID"
const envOutputFormat = "GVS_OUTPUTFORMAT"
const envLogLevel = "GVS_LOGLEVEL"

// holds our config
type gvs struct {
	AppName             string
	AppEnv              string
	VaultURL            string
	VaultSecretPath     string
	VaultRoleID         string
	VaultSecretID       string
	SecretFilePath      string
	SecretAvailabletime string
	VaultToken          string
	OutputFormat        string
	LogLevel            string
	VaultConfig         *vault.Config
	VaultCli            *vault.Client
}

func errInfo() (info string) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.Function + ":" + strconv.Itoa(frame.Line)
}

//var gvs gvs
var version string

// Init read env and initialize app config
func newGVS() (*gvs, error) {

	gvs := new(gvs)

	// get config from env
	gvs.AppName = os.Getenv(envAppName)
	gvs.AppEnv = os.Getenv(envAppEnv)
	gvs.VaultURL = os.Getenv(envVaultAddr)
	gvs.VaultSecretPath = strings.TrimSuffix(strings.TrimPrefix(os.Getenv(envSecretPath), "/"), "/")
	gvs.SecretAvailabletime = os.Getenv(envSecretAvailableTime)
	gvs.VaultRoleID = os.Getenv(envVaultRoleID)
	gvs.VaultSecretID = os.Getenv(envVaultSecretID)
	gvs.SecretFilePath = os.Getenv(envSecretFilePath)
	gvs.OutputFormat = os.Getenv(envOutputFormat)
	gvs.LogLevel = os.Getenv(envLogLevel)

	// initialize logger
	log.SetFormatter(&log.TextFormatter{})
	if strings.ToUpper(gvs.LogLevel) != "DEBUG" {
		gvs.LogLevel = "INFO"
	}
	logLevel, _ := log.ParseLevel(gvs.LogLevel)
	log.SetOutput(os.Stdout)
	log.SetLevel(logLevel)
	log.Infof("Starting gvs version %v", version)

	// replace nil values with defaults when applicable
	if len(gvs.SecretFilePath) == 0 {
		gvs.SecretFilePath = "/dev/shm/gvs"
	} else {
		gvs.SecretFilePath = filepath.Join(gvs.SecretFilePath, "gvs")
	}

	if len(gvs.VaultSecretID) == 0 {
		gvs.VaultSecretID = "/run/secrets/secret_id"
	}

	if len(gvs.SecretAvailabletime) == 0 {
		gvs.SecretAvailabletime = "60"
	} else {
		numSec, err := strconv.Atoi(gvs.SecretAvailabletime)
		if err != nil || numSec > 180 {
			gvs.SecretAvailabletime = "180"
		}
	}

	if len(gvs.VaultRoleID) == 0 {
		gvs.VaultRoleID = "/run/secrets/role_id"
	}

	if len(gvs.OutputFormat) == 0 {
		gvs.OutputFormat = "yaml"
	}

	// get Vault App Role credentials
	vaultRoleID, err := getSecretFromFile(gvs.VaultRoleID)
	if err != nil {
		return gvs, errors.New("Error reading role_id secret: " + err.Error())
	}

	vaultSecretID, err := getSecretFromFile(gvs.VaultSecretID)
	if err != nil {
		return gvs, errors.New("Error reading secret_id secret: " + err.Error())
	}

	gvs.VaultConfig = vault.NewConfig()
	gvs.VaultConfig.Address = gvs.VaultURL
	gvs.VaultConfig.AppRoleCredentials = &vault.AppRoleCredentials{
		RoleID:   vaultRoleID,
		SecretID: vaultSecretID,
	}

	gvs.VaultCli, err = vault.NewClient(gvs.VaultConfig)
	if err != nil {
		return gvs, errors.New("Error creating new Vault client: " + err.Error())
	}

	log.Debugf("gvs config: %+v", gvs.AppName)

	return gvs, nil
}

func main() {
	if err := publishVaultSecret(); err != nil {
		log.Fatalf("Error publishing secret: %v", err)
	}
}

// getVaultSecret read secret kv at given path
// Returns a key value list
func (g *gvs) getVaultSecret(path string) (kv map[string]string, err error) {
	kvMap, err := g.VaultCli.GetSecret(path)
	if err != nil {
		return kv, err
	}

	return kvMap.KV, nil
}

func publishVaultSecret() error {
	g, err := newGVS()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), errInfo())
	}
	//Checking if secret file is writeable and deleteable
	secretFileOK, errSecretFile := g.isSecretFilePathOK()
	if errSecretFile != nil {
		return errors.WithStack(errSecretFile)
	}
	if secretFileOK {
		secretsList := make(map[string]string)
		kvMap, err := g.getVaultSecret(g.VaultSecretPath)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), errInfo())
		}
		for k, v := range kvMap {
			secretsList[k] = v
		}

		// add GVS_APPNAME & GVS_APPENV to secretfile
		secretsList["GVS_APPNAME"] = g.AppName
		secretsList["GVS_APPENV"] = g.AppEnv

		for kd, vd := range secretsList {
			log.Debugf("Populated secret: %v = %v (value hidden)", kd, generateRandomString(len(vd)))
		}
		// create the secret file
		err = g.writeSecret(secretsList)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), errInfo())
		}
		_ = destroySecretFile(g.SecretFilePath, g.SecretAvailabletime)
	}
	log.Infof("Secret file: %v, will be removed in %v seconds", g.SecretFilePath, g.SecretAvailabletime)
	return nil
}

func (g *gvs) writeSecret(kv map[string]string) error {
	f, err := os.Create(g.SecretFilePath)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), errInfo())
	}
	defer f.Close()
	if g.OutputFormat == "yaml" {
		output, _ := yaml.Marshal(&kv)
		_, _ = f.Write(output)
	} else {
		for k, v := range kv {
			_, _ = f.WriteString(k + "=" + v + "\n")
		}
	}
	f.Sync()
	return err
}

func getSecretFromFile(path string) (secret string, err error) {
	// read from docker secret
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.Wrap(err, errInfo())
	}
	return string(dat), nil
}

func (g *gvs) isSecretFilePathOK() (isOK bool, err error) {
	testFile := g.SecretFilePath + ".tmp"
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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func generateRandomString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
