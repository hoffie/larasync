package config

import (
	"github.com/inconshreveable/log15"
)

var Log = log15.New("module", "config")

func init() {
	Log.SetHandler(log15.DiscardHandler())
}
