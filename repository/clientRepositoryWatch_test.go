package repository

import (
	"time"
	"io/ioutil"
	"path/filepath"

	. "gopkg.in/check.v1"

	"github.com/hoffie/larasync/repository/watcher"
)

type ClientRepositoryWatchTests struct {
	RepositoryTests
	repo *ClientRepository
	watcher *watcher.Watcher
}

var _ = Suite(&ClientRepositoryWatchTests{})

func (t *ClientRepositoryWatchTests) SetUpTest(c *C) {
	t.RepositoryTests.SetUpTest(c)
	var err error
	t.repo, err = NewClient(t.dir)
	c.Assert(err, IsNil)
	err = t.repo.Create()
	c.Assert(err, IsNil)
}

func (t *ClientRepositoryWatchTests) TearDownTest(c *C) {
	if t.watcher != nil {
		err := t.watcher.Stop()
		c.Assert(err, IsNil)
	}
}

func (t *ClientRepositoryWatchTests) TestWatcher(c *C) {
	watcher, err := t.repo.Watch()
	t.watcher = watcher
	c.Assert(err, IsNil)
	err = watcher.Start()
	c.Assert(err, IsNil)

	relPath := "test.txt"
	path := filepath.Join(t.dir, relPath)
	err = ioutil.WriteFile(path, []byte("this is a testfile"), defaultFilePerms)
	c.Assert(err, IsNil)

	// Give the process some time for the complete update process to run through.
	time.Sleep(100 * time.Millisecond)
	nibID, err := t.repo.pathToNIBID(relPath)
	c.Assert(err, IsNil)

	c.Assert(t.repo.HasNIB(nibID), Equals, true)
}
