package api

import (
	"github.com/inconshreveable/log15"
)

var Log = log15.New("module", "api")

func init() {
	Log.SetHandler(log15.DiscardHandler())
}
