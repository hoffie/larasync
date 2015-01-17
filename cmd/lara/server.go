package main

import (
	"os"
	"path/filepath"

	"github.com/inconshreveable/log15"

	"github.com/hoffie/larasync/api"
	"github.com/hoffie/larasync/helpers/x509"
	"github.com/hoffie/larasync/repository"
)

const (
	certFileName = "larasync-server.crt"
	keyFileName  = "larasync-server.key"
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
	cfgDir := filepath.Dir(cfgPath)
	certFile := filepath.Join(cfgDir, certFileName)
	keyFile := filepath.Join(cfgDir, keyFileName)
	err = d.needServerCert(certFile, keyFile)
	if err != nil {
		log.Error("unable to load/generate keys", log15.Ctx{"error": err})
		return 1
	}
	s := api.New(*cfg.Signatures.AdminPubkeyBinary,
		cfg.Signatures.MaxAge, rm, certFile, keyFile)
	log.Info("Listening", log15.Ctx{"address": cfg.Server.Listen})
	log.Error("Error", log15.Ctx{"code": s.ListenAndServe()})
	return 1
}

// needServerCert checks whether both required certificate files exist;
// if they don't, an appropriate certificate is generated
func (d *Dispatcher) needServerCert(certFile, keyFile string) error {
	haveKeys, err := d.haveServerCert(certFile, keyFile)
	if err != nil {
		return err
	}
	if haveKeys {
		return nil
	}
	return x509.GenerateServerCertFiles(certFile, keyFile)
}

// haveServerCert returns whether both required files are present.
func (d *Dispatcher) haveServerCert(certFile, keyFile string) (bool, error) {
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
