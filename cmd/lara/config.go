package main

import (
	"code.google.com/p/gcfg"

	"github.com/hoffie/larasync/config"
)

const defaultServerConfigPath = "larasync-server.gcfg"

// getServerConfig reads the best-matching config file, sanitizes it
// and returns the resulting config object.
func getServerConfig() (*config.ServerConfig, error) {
	cfg := &config.ServerConfig{}
	if configPath == "" {
		configPath = defaultServerConfigPath
	}
	err := gcfg.ReadFileInto(cfg, configPath)
	if err != nil {
		return nil, err
	}
	err = cfg.Sanitize()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
