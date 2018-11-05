package main

import (
	"fmt"
	"testing"
	//. "github.com/mch1307/gvs"
)

func Test_gvsConfig_getVaultAppRoleToken(t *testing.T) {
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
		VaultCredentials    VaultAppRoleCredentials
		VaultKvVersion      string
		OutputFormat        string
		LogLevel            string
	}
	tests := []struct {
		name     string
		fields   fields
		expected string
		wantErr  bool
	}{
		// TODO: Add test cases.
		{"AuthOK", fields{VaultURL: "http://localhost:8500", VaultCredentials: VaultAppRoleCredentials{RoleID: "ok"}}, "ok", false},
		{"AuthKO", fields{VaultURL: "http://localhost:8500", VaultCredentials: VaultAppRoleCredentials{RoleID: "ko"}}, "ko", false},
		{"HTTP5xx", fields{VaultURL: "http://localhost:8500", VaultCredentials: VaultAppRoleCredentials{RoleID: "5xx"}}, "5xx", true},
		//"JSONparseError", fields{VaultURL: "http://localhost:8500", VaultCredentials: VaultAppRoleCredentials{RoleID: "jsonKO"}}, "ko", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &gvsConfig{
				AppName:             tt.fields.AppName,
				AppEnv:              tt.fields.AppEnv,
				VaultURL:            tt.fields.VaultURL,
				VaultSecretPath:     tt.fields.VaultSecretPath,
				VaultRoleID:         tt.fields.VaultRoleID,
				VaultSecretID:       tt.fields.VaultSecretID,
				SecretFilePath:      tt.fields.SecretFilePath,
				SecretAvailabletime: tt.fields.SecretAvailabletime,
				SecretList:          tt.fields.SecretList,
				VaultToken:          tt.fields.VaultToken,
				VaultCredentials:    tt.fields.VaultCredentials,
				VaultKvVersion:      tt.fields.VaultKvVersion,
				OutputFormat:        tt.fields.OutputFormat,
				LogLevel:            tt.fields.LogLevel,
			}

			err := a.getVaultAppRoleToken()
			fmt.Println(err)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("unexpected error: gvsConfig.getVaultAppRoleToken() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else {
				if a.VaultToken != tt.expected {
					t.Errorf("gvsConfig.getVaultAppRoleToken() got %v, expected %v", a.VaultToken, tt.expected)
				}
			}

		})
	}
}

func Test_gvsConfig_getKVVersion(t *testing.T) {
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
		VaultCredentials    VaultAppRoleCredentials
		VaultKvVersion      string
		OutputFormat        string
		LogLevel            string
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args
		expected string
		wantErr  bool
	}{
		// TODO: Add test cases.
		{"kvV1", fields{VaultURL: "http://localhost:8500", VaultToken: "goodToken"}, args{name: "kvV1/"}, "1", false},
		{"kvV2", fields{VaultURL: "http://localhost:8500", VaultToken: "goodToken"}, args{name: "kvV2/"}, "2", false},
		{"noauth", fields{VaultURL: "http://localhost:8500", VaultToken: "noauth"}, args{name: "kvV2/"}, "0", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &gvsConfig{
				AppName:             tt.fields.AppName,
				AppEnv:              tt.fields.AppEnv,
				VaultURL:            tt.fields.VaultURL,
				VaultSecretPath:     tt.fields.VaultSecretPath,
				VaultRoleID:         tt.fields.VaultRoleID,
				VaultSecretID:       tt.fields.VaultSecretID,
				SecretFilePath:      tt.fields.SecretFilePath,
				SecretAvailabletime: tt.fields.SecretAvailabletime,
				SecretList:          tt.fields.SecretList,
				VaultToken:          tt.fields.VaultToken,
				VaultCredentials:    tt.fields.VaultCredentials,
				VaultKvVersion:      tt.fields.VaultKvVersion,
				OutputFormat:        tt.fields.OutputFormat,
				LogLevel:            tt.fields.LogLevel,
			}
			if err := a.getKVVersion(tt.args.name); err != nil {
				if !tt.wantErr {
					t.Errorf("gvsConfig.getKVVersion() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else {
				if a.VaultKvVersion != tt.expected {
					t.Errorf("gvsConfig.getKVVersion() got %v, expected %v", a.VaultKvVersion, tt.expected)
				}
			}
		})
	}
}
