package main

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"

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
		VaultCli            *vault.Client
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
	tmpKV := make(map[string]string)
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
		VaultCli            *vault.Client
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
			if !tt.wantErr {
				data, err := ioutil.ReadFile(a.SecretFilePath)
				if err != nil {
					t.Errorf("Could not read result file %v", err)
				}
				if a.OutputFormat == "yaml" {
					err := yaml.Unmarshal(data, tmpKV)
					if err != nil {
						t.Errorf("Could not unmarshal result yaml %v", err)
					}
					if !reflect.DeepEqual(tmpKV, tt.args.kv) {
						t.Errorf("gvs writesecret failed: got %v expected %v", tmpKV, tt.args.kv)
					}

				}

			}
		})
	}
	_ = destroySecretFile("./test.kv", "0")
}

func setEnv(kv map[string]string) {
	for k, v := range kv {
		os.Setenv(k, v)
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
	tests := []struct {
		name    string
		wantGvs *gvs
		env     map[string]string
		wantErr bool
	}{
		{"Standard", &gvs{AppName: "my-app",
			AppEnv:              "dev",
			VaultURL:            vCfg.Address,
			VaultSecretPath:     "kv_v2/my-app-dev",
			VaultRoleID:         "/tmp/role_id",
			VaultSecretID:       "/tmp/secret_id",
			SecretFilePath:      "/dev/shm/gvs",
			SecretAvailabletime: "60",
			VaultToken:          "",
			OutputFormat:        "yaml",
			LogLevel:            "INFO",
			VaultConfig:         vCfg,
			VaultCli:            vCli,
		}, map[string]string{
			"GVS_SECRETAVAILABLETIME": "60",
		},
			false},
		{"maxAvailableTime", &gvs{AppName: "my-app",
			AppEnv:              "dev",
			VaultURL:            vCfg.Address,
			VaultSecretPath:     "kv_v2/my-app-dev",
			VaultRoleID:         "/tmp/role_id",
			VaultSecretID:       "/tmp/secret_id",
			SecretFilePath:      "/dev/shm/gvs",
			SecretAvailabletime: "180",
			VaultToken:          "",
			OutputFormat:        "yaml",
			LogLevel:            "INFO",
			VaultConfig:         vCfg,
			VaultCli:            vCli,
		}, map[string]string{
			"GVS_SECRETAVAILABLETIME": "190",
		},
			false},
		{"noAvailableTime", &gvs{AppName: "my-app",
			AppEnv:              "dev",
			VaultURL:            vCfg.Address,
			VaultSecretPath:     "kv_v2/my-app-dev",
			VaultRoleID:         "/tmp/role_id",
			VaultSecretID:       "/tmp/secret_id",
			SecretFilePath:      "/dev/shm/gvs",
			SecretAvailabletime: "60",
			VaultToken:          "",
			OutputFormat:        "yaml",
			LogLevel:            "INFO",
			VaultConfig:         vCfg,
			VaultCli:            vCli,
		}, map[string]string{
			"GVS_SECRETAVAILABLETIME": "",
		},
			false},
		{"defaults", &gvs{AppName: "my-app",
			AppEnv:              "dev",
			VaultURL:            vCfg.Address,
			VaultSecretPath:     "kv_v2/my-app-dev",
			VaultRoleID:         "/tmp/role_id",
			VaultSecretID:       "/tmp/secret_id",
			SecretFilePath:      "/dev/shm/gvs",
			SecretAvailabletime: "60",
			VaultToken:          "",
			OutputFormat:        "yaml",
			LogLevel:            "INFO",
			VaultConfig:         vCfg,
			VaultCli:            vCli,
		}, map[string]string{
			envSecretAvailableTime: "",
			envSecretFilePath:      "",
			envOutputFormat:        "",
			envLogLevel:            "",
		},
			false},
		{"default-err", &gvs{AppName: "my-app",
			AppEnv:              "dev",
			VaultURL:            vCfg.Address,
			VaultSecretPath:     "kv_v2/my-app-dev",
			VaultRoleID:         "/run/secrets/role_id",
			VaultSecretID:       "/run/secrets/secret_id",
			SecretFilePath:      "/dev/shm/gvs",
			SecretAvailabletime: "60",
			VaultToken:          "",
			OutputFormat:        "yaml",
			LogLevel:            "INFO",
			VaultConfig:         nil,
			VaultCli:            nil,
		}, map[string]string{
			envSecretAvailableTime: "",
			envSecretFilePath:      "",
			envOutputFormat:        "",
			envLogLevel:            "",
			envVaultRoleID:         "",
			envVaultSecretID:       "",
		},
			true},
		{"NoRoleID-err", &gvs{AppName: "my-app",
			AppEnv:              "dev",
			VaultURL:            vCfg.Address,
			VaultSecretPath:     "kv_v2/my-app-dev",
			VaultRoleID:         "/tmp/role_id",
			VaultSecretID:       "/run/secrets/secret_id",
			SecretFilePath:      "/dev/shm/gvs",
			SecretAvailabletime: "60",
			VaultToken:          "",
			OutputFormat:        "yaml",
			LogLevel:            "INFO",
			VaultConfig:         nil,
			VaultCli:            nil,
		}, map[string]string{
			envSecretAvailableTime: "",
			envSecretFilePath:      "",
			envOutputFormat:        "",
			envLogLevel:            "",
			envVaultSecretID:       "",
		},
			true},
		{"wrongVaultURL", &gvs{AppName: "my-app",
			AppEnv:              "dev",
			VaultURL:            "ht@ps/:/wronghost",
			VaultSecretPath:     "kv_v2/my-app-dev",
			VaultRoleID:         "/tmp/role_id",
			VaultSecretID:       "/tmp/secret_id",
			SecretFilePath:      "/dev/shm/gvs",
			SecretAvailabletime: "60",
			VaultToken:          "",
			OutputFormat:        "yaml",
			LogLevel:            "INFO",
			VaultConfig:         nil,
			VaultCli:            nil,
		}, map[string]string{
			"GVS_SECRETAVAILABLETIME": "60",
			envVaultAddr:              "ht@ps/:/wronghost",
		},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(defaultEnv)
			setEnv(tt.env)
			gotGvs, err := newGVS()
			if (err != nil) != tt.wantErr {
				t.Errorf("newGVS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (gotGvs.AppName != tt.wantGvs.AppName) ||
				(gotGvs.AppEnv != tt.wantGvs.AppEnv) ||
				(gotGvs.VaultURL != tt.wantGvs.VaultURL) ||
				(gotGvs.VaultSecretPath != tt.wantGvs.VaultSecretPath) ||
				(gotGvs.VaultRoleID != tt.wantGvs.VaultRoleID) ||
				(gotGvs.VaultSecretID != tt.wantGvs.VaultSecretID) ||
				(gotGvs.SecretFilePath != tt.wantGvs.SecretFilePath) ||
				(gotGvs.SecretAvailabletime != tt.wantGvs.SecretAvailabletime) {
				t.Errorf("newGVS() = %v, want %v", gotGvs, tt.wantGvs)
			}
		})
	}
}

func Test_gvs_publishVaultSecret(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr bool
	}{
		{"default",
			defaultEnv,
			false},
		{"secretFileKO", map[string]string{
			envSecretFilePath: "/no-valid-path"},
			true},
		{"secretPathKO", map[string]string{
			envSecretFilePath: defaultEnv[envSecretFilePath],
			envSecretPath:     "/no-valid-path"},
			true},
		{"newGvsKO", map[string]string{
			envSecretFilePath: defaultEnv[envSecretFilePath],
			envVaultRoleID:    "",
			envVaultSecretID:  ""},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(tt.env)
			//g, _ := newGVS()
			if err := publishVaultSecret(); (err != nil) != tt.wantErr {
				t.Errorf("gvs.publishVaultSecret() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_gvs_getVaultSecret(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		env     map[string]string
		args    args
		wantKv  map[string]string
		wantErr bool
	}{
		{"my-secretV2",
			defaultEnv,
			args{
				path: "kv_v2/my-app-dev"},
			mySecret,
			false,
		},
		{"my-secretV1",
			defaultEnv,
			args{
				path: "kv_v1/my-secret"},
			mySecret,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(tt.env)
			g, _ := newGVS()
			gotKv, err := g.getVaultSecret(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("gvs.getVaultSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotKv, tt.wantKv) {
				t.Errorf("gvs.getVaultSecret() = %v, want %v", gotKv, tt.wantKv)
			}
		})
	}
}
