package main

import (
	"code.google.com/p/gcfg"

	"github.com/hoffie/larasync/config"
)

const defaultServerConfigPath = "larasync-server.gcfg"

// getServerConfig reads the best-matching config file, sanitizes it
// and returns the resulting config object.
func getServerConfig() *config.ServerConfig {
	cfg := &config.ServerConfig{}
	if configPath == "" {
		configPath = defaultServerConfigPath
	}
	gcfg.ReadFileInto(cfg, configPath)
	cfg.Sanitize()
	return cfg
}
