package repository

import (
	"encoding/json"
	"io/ioutil"
)

type StateConfig struct {
	Path          string `json:"-"`
	DefaultServer string
}

func (sc *StateConfig) Load() error {
	data, err := ioutil.ReadFile(sc.Path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, sc)
	return nil
}

func (sc *StateConfig) Save() error {
	data, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(sc.Path, data, defaultFilePerms)
	return err
}
