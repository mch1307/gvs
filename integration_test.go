package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"

	"github.com/pkg/errors"
)

var testVaultRoleID, testVaultSecretID string

var vaultVersion string
var mySecret = make(map[string]string)
var defaultEnv = map[string]string{
	envVaultAddr:           "http://localhost:8200",
	envAppName:             "my-app",
	envAppEnv:              "dev",
	envSecretPath:          "kv_v2/my-app-dev",
	envSecretFilePath:      "/dev/shm/",
	envSecretAvailableTime: "30",
	envVaultRoleID:         "/tmp/role_id",
	envVaultSecretID:       "/tmp/secret_id",
	envOutputFormat:        "yaml",
	envLogLevel:            "INFO",
}

func init() {
	flag.StringVar(&vaultVersion, "vaultVersion", "1.0.1", "provide vault version to be tested against")
	flag.Parse()
	mySecret["my-first-secret"] = "my-first-secret-value"
	mySecret["my-second-secret"] = "my-second-secret-value"
}
func TestMain(m *testing.M) {
	fmt.Println("Testing with Vault version", vaultVersion)
	fmt.Println("TestMain: Preparing Vault server")
	prepareVault()
	err := writeCredentials(testVaultRoleID, testVaultSecretID)
	if err != nil {
		log.Fatalf("Error creating credentials file: %v", err)
	}
	ret := m.Run()
	os.Exit(ret)
}

func prepareVault() {
	err := startVault(vaultVersion)
	if err != nil {
		log.Fatalf("Error in initVaultDev.sh %v", err)
	}
	cmd := exec.Command("./vault", "read", "-field=role_id", "auth/approle/role/my-role/role-id")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "VAULT_TOKEN=my-dev-root-vault-token")
	cmd.Env = append(cmd.Env, "VAULT_ADDR=http://localhost:8200")

	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("error getting role id %v %v", err, out)
	}
	testVaultRoleID = string(out)

	cmd = exec.Command("./vault", "write", "-field=secret_id", "-f", "auth/approle/role/my-role/secret-id")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "VAULT_TOKEN=my-dev-root-vault-token")
	cmd.Env = append(cmd.Env, "VAULT_ADDR=http://localhost:8200")
	out, err = cmd.Output()
	if err != nil {
		log.Fatalf("error getting secret id %v", err)
	}
	testVaultSecretID = string(out)
	os.Unsetenv("VAULT_TOKEN")

}

func writeCredentials(roleID, secretID string) error {
	f, err := os.Create("/tmp/role_id")
	if err != nil {
		return errors.Wrap(errors.WithStack(err), errInfo())
	}
	defer f.Close()
	_, _ = f.Write([]byte(roleID))
	f.Sync()

	f, err = os.Create("/tmp/secret_id")
	if err != nil {
		return errors.Wrap(errors.WithStack(err), errInfo())
	}
	defer f.Close()
	_, _ = f.Write([]byte(secretID))
	f.Sync()

	return err
}

func startVault(version string) error {
	cmd := exec.Command("bash", "./test-files/initVaultDev.sh", version)
	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil

}
