package repository

import (
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

type StateConfigTests struct {
	dir string
}

var _ = Suite(&StateConfigTests{})

func (t *StateConfigTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
}

func (t *StateConfigTests) getStateConfigPath() string {
	return filepath.Join(t.dir, "config.json")
}

func (t *StateConfigTests) getStateConfig() *StateConfig {
	stateConfig := NewStateConfig(t.getStateConfigPath())
	stateConfig.DefaultServer = &ServerStateConfig{
		URL:                 "default_server",
		Fingerprint:         "fp",
		RemoteTransactionID: "remotetransid",
		LocalTransactionID:  "localtransid",
	}
	return stateConfig
}

func (t *StateConfigTests) TestSave(c *C) {
	stateConfig := t.getStateConfig()
	err := stateConfig.Save()
	c.Assert(err, IsNil)
}

func (t *StateConfigTests) TestSaveNotExists(c *C) {
	os.RemoveAll(t.dir)
	stateConfig := t.getStateConfig()
	err := stateConfig.Save()
	c.Assert(err, NotNil)
}

func (t *StateConfigTests) storeStateConfig(c *C) {
	stateConfig := t.getStateConfig()
	err := stateConfig.Save()
	c.Assert(err, IsNil)
}

func (t *StateConfigTests) TestLoad(c *C) {
	t.storeStateConfig(c)

	stateConfig := &StateConfig{
		Path: t.getStateConfigPath(),
	}
	err := stateConfig.Load()
	c.Assert(err, IsNil)

	defaultServer := stateConfig.DefaultServer
	c.Assert(defaultServer.URL, Equals, "default_server")
	c.Assert(defaultServer.Fingerprint, Equals, "fp")
	c.Assert(defaultServer.LocalTransactionID, Equals, "localtransid")
	c.Assert(defaultServer.RemoteTransactionID, Equals, "remotetransid")
}

func (t *StateConfigTests) TestLoadNotExists(c *C) {
	stateConfig := t.getStateConfig()
	err := stateConfig.Load()
	c.Assert(err, NotNil)
}
