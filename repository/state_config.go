package repository

import (
	"encoding/json"
	"io/ioutil"
)

// StateConfig is used to keep track of state information
// which has be be read and written programatically.
type StateConfig struct {
	Path          string `json:"-"`
	DefaultServer string
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
