package watcher

import (
	"os"
	"io/ioutil"
	"path/filepath"

	"github.com/hoffie/larasync/helpers"

	. "gopkg.in/check.v1"
)

const (
	// default permissions
	defaultFilePerms = 0600
	defaultDirPerms  = 0700
)

type WatcherTest struct {
	Dir             string
	Watcher         *Watcher
	RepositoryCheck *RepositoryCheck
}

var _ = Suite(&WatcherTest{})

type RepositoryCheck struct {
	c            *C
	addedItems   []string
	removedItems []string
}

// AddItem adds the item to the internal state repository
// which resides in the passed absolute path.
func (rc *RepositoryCheck) AddItem(absPath string) error {
	rc.addedItems = append(rc.addedItems, absPath)
	return nil
}

// DeleteItem marks the item accessible through the passed
// absolute path as deleted in the internal repository state.
func (rc *RepositoryCheck) DeleteItem(absPath string) error {
	rc.removedItems = append(rc.removedItems, absPath)
	return nil
}

func (rc *RepositoryCheck) HasItemAdded(absPath string) bool {
	return helpers.SliceContainsString(rc.addedItems, absPath)
}

func (rc *RepositoryCheck) ShouldHaveItemAdded(absPath string) {
	rc.c.Check(rc.HasItemAdded(absPath), Equals, true)
}

func (rc *RepositoryCheck) HasItemRemoved(absPath string) bool {
	return helpers.SliceContainsString(rc.removedItems, absPath)
}

func (rc *RepositoryCheck) ShouldHaveItemRemoved(absPath string) {
	rc.c.Check(rc.HasItemRemoved(absPath), Equals, true)
}

func (t *WatcherTest) SetUpTest(c *C) {
	t.Dir = c.MkDir()
	t.RepositoryCheck = &RepositoryCheck{
		c:            c,
		addedItems:   []string{},
		removedItems: []string{},
	}
	watcher, err := New(t.Dir, t.RepositoryCheck)
	c.Assert(err, IsNil)
	t.Watcher = watcher
}

func (t *WatcherTest) TearDownTest(c *C) {
	t.Watcher.Stop()
}

func (t *WatcherTest) TestFileAdd(c *C) {
	err := t.Watcher.Start()
	c.Assert(err, IsNil)
	filePath := filepath.Join(t.Dir, "test.txt")
	err = ioutil.WriteFile(filePath, []byte("Hello World"), defaultFilePerms)
	c.Assert(err, IsNil)

	t.RepositoryCheck.ShouldHaveItemAdded(filePath)
}

func (t *WatcherTest) TestFileRemoved(c *C) {
	filePath := filepath.Join(t.Dir, "test.txt")
	err := ioutil.WriteFile(filePath, []byte("Hello World"), defaultFilePerms)
	c.Assert(err, IsNil)

	err = t.Watcher.Start()
	c.Assert(err, IsNil)

	err = os.Remove(filePath)
	c.Assert(err, IsNil)
	t.RepositoryCheck.ShouldHaveItemRemoved(filePath)
}

func (t *WatcherTest) TestFileAddSubDirectory(c *C) {
	err := t.Watcher.Start()
	c.Assert(err, IsNil)
	dir := filepath.Join(t.Dir, "test")
	err = os.Mkdir(dir, defaultDirPerms)
	c.Assert(err, IsNil)
	filePath := filepath.Join(dir, "test.txt")
	err = ioutil.WriteFile(filePath, []byte("Hello World"), defaultFilePerms)
	c.Assert(err, IsNil)

	t.RepositoryCheck.ShouldHaveItemAdded(filePath)
}

func (t *WatcherTest) TestFileRemoveSubDirectory(c *C) {
	dir := filepath.Join(t.Dir, "test")
	err := os.Mkdir(dir, defaultDirPerms)
	c.Assert(err, IsNil)
	filePath := filepath.Join(dir, "test.txt")
	err = ioutil.WriteFile(filePath, []byte("Hello World"), defaultFilePerms)
	c.Assert(err, IsNil)
	err = t.Watcher.Start()
	c.Assert(err, IsNil)

	err = os.Remove(filePath)
	c.Assert(err, IsNil)
	t.RepositoryCheck.ShouldHaveItemRemoved(filePath)
}
