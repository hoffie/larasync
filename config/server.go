package config

import (
	"fmt"
	"log"
	"time"

	"github.com/hoffie/larasync/api"
)

// ServerConfig contains all settings for our server mode.
type ServerConfig struct {
	Server struct {
		Listen string
	}
	Signatures struct {
		AdminSecret string
		MaxAge      time.Duration
	}
	Repository struct {
		BasePath string
	}
}

// Sanitize populates all zero values with sane defaults and ensures that any
// required options are set to sane values.
func (c *ServerConfig) Sanitize() {
	if c.Server.Listen == "" {
		c.Server.Listen = fmt.Sprintf("127.0.0.1:%d", api.DefaultPort)
	}
	if len(c.Signatures.AdminSecret) == 0 {
		log.Fatal("no admin secret configured; refusing to run")
	}
	if len(c.Repository.BasePath) == 0 {
		log.Fatal("no repository base path configured; refusing to run")
	}
	if c.Signatures.MaxAge == 0 {
		c.Signatures.MaxAge = 5 * time.Second
	}
}
