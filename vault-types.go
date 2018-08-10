package main

import (
	"encoding/json"
	"time"
)

// VaultSecretMount holds vault mount info
type VaultSecretMount struct {
	Name     string `json:"??,string"`
	Accessor string `json:"accessor"`
	Config   struct {
		DefaultLeaseTTL int    `json:"default_lease_ttl"`
		ForceNoCache    bool   `json:"force_no_cache"`
		MaxLeaseTTL     int    `json:"max_lease_ttl"`
		PluginName      string `json:"plugin_name"`
	} `json:"config"`
	Description string                 `json:"description"`
	Local       bool                   `json:"local"`
	Options     map[string]interface{} `json:"options"`
	SealWrap    bool                   `json:"seal_wrap"`
	Type        string                 `json:"type"`
}

// VaultSecretMounts string map of VaultSecretMount
// used to parse the Vault JSON response to /v1/sys/internal/ui/mounts
type VaultSecretMounts map[string]*VaultSecretMount

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

// VaultMountListRespone holds the Vault Response to mounts info
type VaultMountListRespone struct {
	RequestID     string `json:"request_id"`
	LeaseID       string `json:"lease_id"`
	Renewable     bool   `json:"renewable"`
	LeaseDuration int    `json:"lease_duration"`
	Data          struct {
		Auth struct {
		} `json:"auth"`
		Secret json.RawMessage `json:"secret"`
	} `json:"data"`
	WrapInfo interface{} `json:"wrap_info"`
	Warnings interface{} `json:"warnings"`
	Auth     interface{} `json:"auth"`
}

// VaultSecretv2 holds the Vault secret (kv v2)
type VaultSecretv2 struct {
	RequestID     string `json:"request_id"`
	LeaseID       string `json:"lease_id"`
	Renewable     bool   `json:"renewable"`
	LeaseDuration int    `json:"lease_duration"`
	Data          struct {
		Data     map[string]string `json:"data"`
		Metadata struct {
			CreatedTime  time.Time `json:"created_time"`
			DeletionTime string    `json:"deletion_time"`
			Destroyed    bool      `json:"destroyed"`
			Version      int       `json:"version"`
		} `json:"metadata"`
	} `json:"data"`
	WrapInfo interface{} `json:"wrap_info"`
	Warnings interface{} `json:"warnings"`
	Auth     interface{} `json:"auth"`
}

// VaultSecret holds the Vault secret (kv v1)
type VaultSecret struct {
	RequestID     string            `json:"request_id"`
	LeaseID       string            `json:"lease_id"`
	Renewable     bool              `json:"renewable"`
	LeaseDuration int               `json:"lease_duration"`
	Data          map[string]string `json:"data"`
	WrapInfo      interface{}       `json:"wrap_info"`
	Warnings      interface{}       `json:"warnings"`
	Auth          interface{}       `json:"auth"`
}

// VaultAppRoleCredentials holds the role and secret id for Vault approle auth
type VaultAppRoleCredentials struct {
	RoleID   string `json:"role_id"`
	SecretID string `json:"secret_id"`
}

// VaultAuthResponse holds the Vault auth response, used to get the Clienttoken
type VaultAuthResponse struct {
	RequestID     string      `json:"request_id"`
	LeaseID       string      `json:"lease_id"`
	Renewable     bool        `json:"renewable"`
	LeaseDuration int         `json:"lease_duration"`
	Data          interface{} `json:"data"`
	WrapInfo      interface{} `json:"wrap_info"`
	Warnings      interface{} `json:"warnings"`
	Auth          struct {
		ClientToken string   `json:"client_token"`
		Accessor    string   `json:"accessor"`
		Policies    []string `json:"policies"`
		Metadata    struct {
			RoleName string `json:"role_name"`
		} `json:"metadata"`
		LeaseDuration int    `json:"lease_duration"`
		Renewable     bool   `json:"renewable"`
		EntityID      string `json:"entity_id"`
	} `json:"auth"`
}
