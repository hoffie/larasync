package config

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/hoffie/larasync/api"
)

// ErrAdminPubkeyMissing is returned if no admin pubkey is specified.
var ErrAdminPubkeyMissing = errors.New("empty admin pubkey")

// ErrInvalidAdminPubkey is returned if decoding the admin pubkey fails.
var ErrInvalidAdminPubkey = errors.New("invalid admin pubkey")

// ErrTruncatedAdminPubkey is returned if the given admin pubkey is too short.
var ErrTruncatedAdminPubkey = errors.New("admin pubkey too short")

// ErrMissingBasePath is returned if no base path is configured.
var ErrMissingBasePath = errors.New("missing basepath")

// ErrBadBasePath is returned if the configured base path is not accessible.
var ErrBadBasePath = errors.New("unaccessible basepath")

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
		return ErrMissingBasePath
	}
	_, err = ioutil.ReadDir(c.Repository.BasePath)
	if err != nil {
		Log.Error("unable to open repository base path configured; refusing to run (%s)", err)
		return ErrBadBasePath
	}
	if c.Signatures.MaxAge == 0 {
		c.Signatures.MaxAge = 5 * time.Second
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
	if len(dec) != api.PublicKeySize {
		return ErrTruncatedAdminPubkey
	}
	c.Signatures.AdminPubkeyBinary = new([api.PublicKeySize]byte)
	copy(c.Signatures.AdminPubkeyBinary[:], dec)
	return nil
}
