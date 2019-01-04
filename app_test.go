package main

import (
	//	"log"
	"fmt"
	"os"
	"testing"

	vault "github.com/mch1307/vaultlib"
	log "github.com/sirupsen/logrus"
)

func Test_generateRandomString(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		n    int
	}{
		{"simple test", 123},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = generateRandomString(tt.n)
		})
	}
}

// func Test_gvsConfig_isSecretFilePathOK(t *testing.T) {
// 	type fields struct {
// 		AppName             string
// 		AppEnv              string
// 		VaultURL            string
// 		VaultSecretPath     string
// 		VaultRoleID         string
// 		VaultSecretID       string
// 		SecretFilePath      string
// 		SecretAvailabletime string
// 		SecretList          []string
// 		VaultToken          string
// 		VaultCredentials    VaultAppRoleCredentials
// 		VaultKvVersion      string
// 		OutputFormat        string
// 		LogLevel            string
// 	}
// 	tests := []struct {
// 		name     string
// 		fields   fields
// 		wantIsOK bool
// 		wantErr  bool
// 	}{
// 		{"fileOK", fields{SecretFilePath: "/dev/shm/test", SecretAvailabletime: "2"}, true, false},
// 		{"fileKO", fields{SecretFilePath: "/dev1/shm1/test", SecretAvailabletime: "2"}, false, true},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			a := &gvsConfig{
// 				AppName:             tt.fields.AppName,
// 				AppEnv:              tt.fields.AppEnv,
// 				VaultURL:            tt.fields.VaultURL,
// 				VaultSecretPath:     tt.fields.VaultSecretPath,
// 				VaultRoleID:         tt.fields.VaultRoleID,
// 				VaultSecretID:       tt.fields.VaultSecretID,
// 				SecretFilePath:      tt.fields.SecretFilePath,
// 				SecretAvailabletime: tt.fields.SecretAvailabletime,
// 				SecretList:          tt.fields.SecretList,
// 				VaultToken:          tt.fields.VaultToken,
// 				VaultCredentials:    tt.fields.VaultCredentials,
// 				VaultKvVersion:      tt.fields.VaultKvVersion,
// 				OutputFormat:        tt.fields.OutputFormat,
// 				LogLevel:            tt.fields.LogLevel,
// 			}
// 			gotIsOK, err := a.isSecretFilePathOK()
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("gvsConfig.isSecretFilePathOK() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if gotIsOK != tt.wantIsOK {
// 				t.Errorf("gvsConfig.isSecretFilePathOK() = %v, want %v", gotIsOK, tt.wantIsOK)
// 			}
// 		})
// 	}
// }

func Test_getSecretFromFile(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name       string
		args       args
		wantSecret string
		wantErr    bool
	}{
		{"fileOK", args{path: "tmpSecret"}, "test", false},
		{"fileNotFound", args{path: "tmpNoSecret"}, "", true},
	}
	filePath := "tmpSecret"
	f, _ := os.Create(filePath)
	defer f.Close()
	_, _ = f.WriteString("test")
	f.Sync()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSecret, err := getSecretFromFile(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSecretFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotSecret != tt.wantSecret {
				t.Errorf("getSecretFromFile() = %v, want %v", gotSecret, tt.wantSecret)
			}
		})
	}
	_ = destroySecretFile(filePath, "1")
}

func Test_gvsConfig_isSecretFilePathOK(t *testing.T) {
	type fields struct {
		AppName             string
		AppEnv              string
		VaultURL            string
		VaultSecretPath     string
		VaultRoleID         string
		VaultSecretID       string
		SecretFilePath      string
		SecretAvailabletime string
		SecretList          []string
		VaultToken          string
		OutputFormat        string
		LogLevel            string
		VaultConfig         *vault.Config
		VaultCli            *vault.VaultClient
	}
	tests := []struct {
		name     string
		fields   fields
		wantIsOK bool
		wantErr  bool
	}{
		{"ok", fields{SecretFilePath: "./test"}, true, false},
		{"ko", fields{SecretFilePath: "/notexist"}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &gvs{
				AppName:             tt.fields.AppName,
				AppEnv:              tt.fields.AppEnv,
				VaultURL:            tt.fields.VaultURL,
				VaultSecretPath:     tt.fields.VaultSecretPath,
				VaultRoleID:         tt.fields.VaultRoleID,
				VaultSecretID:       tt.fields.VaultSecretID,
				SecretFilePath:      tt.fields.SecretFilePath,
				SecretAvailabletime: tt.fields.SecretAvailabletime,
				//				SecretList:          tt.fields.SecretList,
				VaultToken:   tt.fields.VaultToken,
				OutputFormat: tt.fields.OutputFormat,
				LogLevel:     tt.fields.LogLevel,
				VaultConfig:  tt.fields.VaultConfig,
				VaultCli:     tt.fields.VaultCli,
			}
			gotIsOK, err := a.isSecretFilePathOK()
			if (err != nil) != tt.wantErr {
				t.Errorf("gvsConfig.isSecretFilePathOK() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotIsOK != tt.wantIsOK {
				t.Errorf("gvsConfig.isSecretFilePathOK() = %v, want %v", gotIsOK, tt.wantIsOK)
			}
		})
	}
}

func Test_gvsConfig_writeSecret(t *testing.T) {
	mySecret := make(map[string]string)
	mySecret["secret"] = "value"
	type fields struct {
		AppName             string
		AppEnv              string
		VaultURL            string
		VaultSecretPath     string
		VaultRoleID         string
		VaultSecretID       string
		SecretFilePath      string
		SecretAvailabletime string
		SecretList          []string
		VaultToken          string
		OutputFormat        string
		LogLevel            string
		VaultConfig         *vault.Config
		VaultCli            *vault.VaultClient
	}
	type args struct {
		kv map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"ok", fields{SecretFilePath: "./test.kv"}, args{mySecret}, false},
		{"ko", fields{SecretFilePath: "/root/test.kv"}, args{mySecret}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &gvs{
				AppName:             tt.fields.AppName,
				AppEnv:              tt.fields.AppEnv,
				VaultURL:            tt.fields.VaultURL,
				VaultSecretPath:     tt.fields.VaultSecretPath,
				VaultRoleID:         tt.fields.VaultRoleID,
				VaultSecretID:       tt.fields.VaultSecretID,
				SecretFilePath:      tt.fields.SecretFilePath,
				SecretAvailabletime: tt.fields.SecretAvailabletime,
				//				SecretList:          tt.fields.SecretList,
				VaultToken:   tt.fields.VaultToken,
				OutputFormat: tt.fields.OutputFormat,
				LogLevel:     tt.fields.LogLevel,
				VaultConfig:  tt.fields.VaultConfig,
				VaultCli:     tt.fields.VaultCli,
			}
			if err := a.writeSecret(tt.args.kv); (err != nil) != tt.wantErr {
				t.Errorf("gvsConfig.writeSecret() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_newGVS(t *testing.T) {
	vCfg := vault.NewConfig()
	vCfg.Address = "http://localhost:8200"
	vCfg.AppRoleCredentials = &vault.AppRoleCredentials{RoleID: testVaultRoleID, SecretID: testVaultSecretID}
	vCli, err := vault.NewClient(vCfg)
	if err != nil {
		log.Fatalf("Error getting new vault client: %v", err)
	}
	//vaultConfig.AppRoleCredentials
	os.Setenv("GVS_APPNAME", "my-app")
	os.Setenv("GVS_APPENV", "dev")
	os.Setenv("GVS_VAULTURL", vCfg.Address)
	os.Setenv("GVS_VAULTROLEID", "/tmp/role_id")
	os.Setenv("GVS_VAULTSECRETID", "/tmp/secret_id")
	os.Setenv("GVS_SECRETPATH", "kv_v2/my-app-dev")
	tests := []struct {
		name    string
		wantGvs *gvs
		wantErr bool
	}{
		{"ok", &gvs{AppName: "my-app",
			AppEnv:              "dev",
			VaultURL:            vCfg.Address,
			VaultSecretPath:     "kv_v2/my-app-dev",
			VaultRoleID:         "/tmp/role_id",
			VaultSecretID:       "/tmp/secret_id",
			SecretFilePath:      "/dev/shm/gvs",
			SecretAvailabletime: "60",
			//			SecretList:          nil,
			VaultToken:   "",
			OutputFormat: "yaml",
			LogLevel:     "INFO",
			VaultConfig:  vCfg,
			VaultCli:     vCli,
		},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGvs, err := newGVS()
			if (err != nil) != tt.wantErr {
				t.Errorf("newGVS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//!reflect.DeepEqual(gotGvs.VaultCli.Status, tt.wantGvs.VaultCli.Status)

			if (gotGvs.AppName != tt.wantGvs.AppName) ||
				(gotGvs.AppEnv != tt.wantGvs.AppEnv) ||
				(gotGvs.VaultURL != tt.wantGvs.VaultURL) ||
				(gotGvs.VaultSecretPath != tt.wantGvs.VaultSecretPath) ||
				(gotGvs.VaultRoleID != tt.wantGvs.VaultRoleID) ||
				(gotGvs.VaultSecretID != tt.wantGvs.VaultSecretID) ||
				(gotGvs.SecretFilePath != tt.wantGvs.SecretFilePath) ||
				(gotGvs.VaultCli.Status != tt.wantGvs.VaultCli.Status) {
				fmt.Println("status: " + gotGvs.VaultCli.Status + " " + tt.wantGvs.VaultCli.Status)
				t.Errorf("newGVS() = %v, want %v", gotGvs, tt.wantGvs)
			}
		})
	}
}

func Test_gvs_publishVaultSecret(t *testing.T) {
	os.Setenv("GVS_APPNAME", "my-app")
	os.Setenv("GVS_APPENV", "dev")
	os.Setenv("GVS_VAULTURL", "http://localhost:8200")
	os.Setenv("GVS_VAULTROLEID", "/tmp/role_id")
	os.Setenv("GVS_VAULTSECRETID", "/tmp/secret_id")
	os.Setenv("GVS_SECRETPATH", "kv_v2/my-app-dev")
	goodCli, err := newGVS()
	if err != nil {
		t.Errorf("Error getting gvs %v", err)
	}
	os.Setenv("GVS_SECRETPATH", "invalid_secret/")
	badSecretCli, err := newGVS()
	if err != nil {
		log.Fatalf("Error getting gvs %v", err)
	}
	//os.Setenv("GVS_VAULTROLEID", "/tmp/role_id")
	// os.Setenv("GVS_VAULTSECRETID", "/no_secret_id")
	// badSecretIDCli, err := newGVS()
	// if err != nil {
	// 	t.Errorf("Error getting gvs %v", err)
	// }

	tests := []struct {
		name    string
		cli     *gvs
		wantErr bool
	}{
		{"ok", goodCli, false},
		{"badSecretPath", badSecretCli, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cli.publishVaultSecret(); (err != nil) != tt.wantErr {
				t.Errorf("gvs.publishVaultSecret() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
