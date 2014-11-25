package main

import (
	"code.google.com/p/gcfg"

	"github.com/hoffie/larasync/config"
)

const defaultServerConfigPath = "larasync-server.gcfg"

func getServerConfig() *config.ServerConfig {
	cfg := &config.ServerConfig{}
	if configPath == "" {
		configPath = defaultServerConfigPath
	}
	gcfg.ReadFileInto(cfg, configPath)
	cfg.Sanitize()
	return cfg
}
