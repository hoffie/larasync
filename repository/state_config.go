package repository

import (
	"encoding/json"
	"io/ioutil"
)

// StateConfig is used to keep track of state information
// which has be be read and written programatically.
type StateConfig struct {
	Path          string             `json:"-"`
	DefaultServer *ServerStateConfig `json:"default_server"`
}

// ServerStateConfig is a substruct which stores the state
// established between the client and the remote server.
type ServerStateConfig struct {
	URL                 string `json:"url"`
	Fingerprint         string `json:"fingerprint"`
	RemoteTransactionID string `json:"remote_transaction_id"`
	LocalTransactionID  string `json:"local_transaction_id"`
}

func NewStateConfig(path string) *StateConfig {
	return &StateConfig{
		Path:          path,
		DefaultServer: &ServerStateConfig{},
	}
}

// Load attempts to load previous state config from disk.
func (sc *StateConfig) Load() error {
	data, err := ioutil.ReadFile(sc.Path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, sc)
	return nil
}

// Save serializes the current StateConfig to disk.
func (sc *StateConfig) Save() error {
	data, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(sc.Path, data, defaultFilePerms)
	return err
}
