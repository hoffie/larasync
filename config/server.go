package config

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/hoffie/larasync/api"
)

// ServerConfig contains all settings for our server mode.
type ServerConfig struct {
	Server struct {
		Listen string
	}
	Signatures struct {
		AdminPubkey       string
		AdminPubkeyBinary *[api.PublicKeySize]byte
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
		c.Server.Listen = fmt.Sprintf("127.0.0.1:%d", api.DefaultPort)
	}
	err := c.decodeAdminPubkey()
	if err != nil {
		Log.Error("no valid admin pubkey configured; refusing to run")
		return err
	}
	if len(c.Repository.BasePath) == 0 {
		Log.Error("no repository base path configured; refusing to run")
		return errors.New("missing basepath")
	}
	if c.Signatures.MaxAge == 0 {
		c.Signatures.MaxAge = 5 * time.Second
	}
	return nil
}

// decodeAdminPubkey reads AdminPubkey, hex-decodes it and performs validation steps.
func (c *ServerConfig) decodeAdminPubkey() error {
	if c.Signatures.AdminPubkey == "" {
		return errors.New("empty admin pubkey")
	}
	dec, err := hex.DecodeString(c.Signatures.AdminPubkey)
	if err != nil {
		return err
	}
	if len(dec) != api.PublicKeySize {
		return errors.New("admin pubkey too short")
	}
	c.Signatures.AdminPubkeyBinary = new([api.PublicKeySize]byte)
	copy(c.Signatures.AdminPubkeyBinary[:], dec)
	return nil
}
