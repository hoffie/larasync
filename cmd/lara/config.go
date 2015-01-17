package main

import (
	"code.google.com/p/gcfg"

	"github.com/hoffie/larasync/config"
)

const defaultServerConfigPath = "larasync-server.gcfg"

// getServerConfig reads the best-matching config file, sanitizes it
// and returns the resulting config object.
func getServerConfig(path string) (*config.ServerConfig, error) {
	cfg := &config.ServerConfig{}
	err := gcfg.ReadFileInto(cfg, path)
	if err != nil {
		return nil, err
	}
	err = cfg.Sanitize()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
