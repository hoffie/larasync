package repository

import (
	"github.com/inconshreveable/log15"
)

var Log = log15.New("module", "repository")

func init() {
	Log.SetHandler(log15.DiscardHandler())
}
