package main

import (
	"os"
	"path/filepath"

	"github.com/inconshreveable/log15"

	"github.com/hoffie/larasync/api/server"
	"github.com/hoffie/larasync/config"
	"github.com/hoffie/larasync/helpers/x509"
	"github.com/hoffie/larasync/repository"
)

const (
	certFileName = "larasync-server.crt"
	keyFileName  = "larasync-server.key"
)

// serverAction starts the server process.
func (d *Dispatcher) serverAction() int {
	cfg, err := d.loadServerConfig()
	if err != nil {
		log.Error("unable to load server config", log15.Ctx{"error": err})
		return 1
	}
	rm, err := repository.NewManager(cfg.Repository.BasePath)
	if err != nil {
		log.Error("repository.Manager creation failure", log15.Ctx{"error": err})
		return 1
	}
	err = d.needServerCert()
	if err != nil {
		log.Error("unable to load/generate keys", log15.Ctx{"error": err})
		return 1
	}
	certFile, keyFile := d.serverCertFilePaths()
	s, err := server.New(*cfg.Signatures.AdminPubkeyBinary,
		cfg.Signatures.MaxAge, rm, certFile, keyFile)
	if err != nil {
		log.Error("unable to initialize server", log15.Ctx{"error": err})
		return 1
	}
	log.Info("Listening", log15.Ctx{"address": cfg.Server.Listen})
	log.Error("Error", log15.Ctx{"code": s.ListenAndServe()})
	return 1
}

// needServerCert checks whether both required certificate files exist;
// if they don't, an appropriate certificate is generated
func (d *Dispatcher) needServerCert() error {
	haveKeys, err := d.haveServerCert()
	if err != nil {
		return err
	}
	if haveKeys {
		return nil
	}
	certFile, keyFile := d.serverCertFilePaths()
	log.Info("no server certificate found; generating one")
	return x509.GenerateServerCertFiles(certFile, keyFile)
}

// serverCertFilePaths returns the server certificate file paths
func (d *Dispatcher) serverCertFilePaths() (string, string) {
	cfgDir := filepath.Dir(d.serverCfgPath)
	certFile := filepath.Join(cfgDir, certFileName)
	keyFile := filepath.Join(cfgDir, keyFileName)
	return certFile, keyFile
}

// haveServerCert returns whether both required files are present.
func (d *Dispatcher) haveServerCert() (bool, error) {
	certFile, keyFile := d.serverCertFilePaths()
	for _, file := range []string{certFile, keyFile} {
		_, err := os.Stat(file)
		if os.IsNotExist(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
	}
	return true, nil
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

// loadServerConfig attempts to load the server config file
func (d *Dispatcher) loadServerConfig() (*config.ServerConfig, error) {
	var err error
	d.serverCfgPath, err = d.getServerConfigPath()
	if err != nil {
		return nil, err
	}
	cfg, err := getServerConfig(d.serverCfgPath)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
