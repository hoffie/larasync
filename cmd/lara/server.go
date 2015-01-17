package main

import (
	"path/filepath"

	"github.com/inconshreveable/log15"

	"github.com/hoffie/larasync/api"
	"github.com/hoffie/larasync/repository"
)

// serverAction starts the server process.
func (d *Dispatcher) serverAction() int {
	d.setupLogging()
	cfgPath, err := d.getServerConfigPath()
	if err != nil {
		log.Error("unable to get absolute server config path", log15.Ctx{"error": err})
		return 1
	}
	cfg, err := getServerConfig(cfgPath)
	if err != nil {
		log.Error("unable to parse configuration", log15.Ctx{"error": err})
		return 1
	}
	rm, err := repository.NewManager(cfg.Repository.BasePath)
	if err != nil {
		log.Error("repository.Manager creation failure", log15.Ctx{"error": err})
		return 1
	}
	s := api.New(*cfg.Signatures.AdminPubkeyBinary,
		cfg.Signatures.MaxAge, rm)
	log.Info("Listening", log15.Ctx{"address": cfg.Server.Listen})
	log.Error("Error", log15.Ctx{"code": s.ListenAndServe()})
	return 1
}

// getServerConfigPath returns the absolute path of the server config file
func (d *Dispatcher) getServerConfigPath() (string, error) {
	path := d.context.String("config")
	if path == "" {
		path = defaultServerConfigPath
	}
	path, err := filepath.Abs(path)
	return path, err
}
