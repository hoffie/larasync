package config

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"time"

	apicommon "github.com/hoffie/larasync/api/common"
	"github.com/hoffie/larasync/api/server"
)

// ServerConfig contains all settings for our server mode.
type ServerConfig struct {
	Server struct {
		Listen string
	}
	Signatures struct {
		AdminPubkey       string
		AdminPubkeyBinary *[apicommon.PublicKeySize]byte
		MaxAge            time.Duration
	}
	Repository struct {
		BasePath string
	}
}

// Sanitize populates all zero values with sane defaults and ensures that any
// required options are set to sane values.
func (c *ServerConfig) Sanitize() error {
	if c.Server.Listen == "" {
		c.Server.Listen = fmt.Sprintf("127.0.0.1:%d", server.DefaultPort)
	}
	err := c.decodeAdminPubkey()
	if err != nil {
		Log.Error("no valid admin pubkey configured; refusing to run")
		return err
	}
	if len(c.Repository.BasePath) == 0 {
		Log.Error("no repository base path configured; refusing to run")
		return ErrMissingBasePath
	}
	_, err = ioutil.ReadDir(c.Repository.BasePath)
	if err != nil {
		Log.Error(fmt.Sprintf("unable to open repository base path configured; "+
			"refusing to run (%s)", err))
		return ErrBadBasePath
	}
	if c.Signatures.MaxAge == 0 {
		c.Signatures.MaxAge = 10 * time.Second
	}
	return nil
}

// decodeAdminPubkey reads AdminPubkey, hex-decodes it and performs validation steps.
func (c *ServerConfig) decodeAdminPubkey() error {
	if c.Signatures.AdminPubkey == "" {
		return ErrAdminPubkeyMissing
	}
	dec, err := hex.DecodeString(c.Signatures.AdminPubkey)
	if err != nil {
		return ErrInvalidAdminPubkey
	}
	if len(dec) != apicommon.PublicKeySize {
		return ErrTruncatedAdminPubkey
	}
	c.Signatures.AdminPubkeyBinary = new([apicommon.PublicKeySize]byte)
	copy(c.Signatures.AdminPubkeyBinary[:], dec)
	return nil
}
