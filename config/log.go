package config

import (
	"github.com/inconshreveable/log15"
)

// Log is our logger reference, available for external configuration.
var Log = log15.New("module", "config")

func init() {
	Log.SetHandler(log15.DiscardHandler())
}
