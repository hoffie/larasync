package repository

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

var _ = Suite(&RepositoryDeleteItemTests{})

type RepositoryDeleteItemTests struct {
	dir string
	r   *ClientRepository
}

func (t *RepositoryDeleteItemTests) SetUpTest(c *C) {
	var err error
	t.dir, err = filepath.EvalSymlinks(c.MkDir())
	c.Assert(err, IsNil)
	t.r, err = NewClient(t.dir)
	c.Assert(err, IsNil)
	err = t.r.Create()
	c.Assert(err, IsNil)
}

func (t *RepositoryDeleteItemTests) TestFileDeletion(c *C) {
	filename := "foo.txt"
	fullpath := filepath.Join(t.dir, filename)
	err := ioutil.WriteFile(fullpath, []byte("foo"), 0600)
	c.Assert(err, IsNil)

	err = t.r.AddItem(fullpath)
	c.Assert(err, IsNil)

	err = t.r.DeleteItem(fullpath)
	c.Assert(err, IsNil)

	_, err = os.Stat(fullpath)
	c.Assert(os.IsNotExist(err), Equals, true)

	nibID, err := t.r.pathToNIBID(filename)
	c.Assert(err, IsNil)

	nib, err := t.r.nibStore.Get(nibID)
	c.Assert(err, IsNil)

	rev, err := nib.LatestRevision()
	c.Assert(err, IsNil)

	c.Assert(rev.IsDeletion(), Equals, true)
}

func (t *RepositoryDeleteItemTests) TestFileDeletionModified(c *C) {
	filename := "foo.txt"
	fullpath := filepath.Join(t.dir, filename)
	err := ioutil.WriteFile(fullpath, []byte("foo"), 0600)
	c.Assert(err, IsNil)

	err = t.r.AddItem(fullpath)
	c.Assert(err, IsNil)

	err = ioutil.WriteFile(fullpath, []byte("bar"), 0600)
	c.Assert(err, IsNil)

	err = t.r.DeleteItem(fullpath)
	c.Assert(err, IsNil)

	data, err := ioutil.ReadFile(fullpath)
	c.Assert(err, IsNil)

	c.Assert(data, DeepEquals, []byte("bar"))
}

func (t *RepositoryDeleteItemTests) TestFileAlreadyDeleted(c *C) {
	filename := "foo.txt"
	fullpath := filepath.Join(t.dir, filename)
	err := ioutil.WriteFile(fullpath, []byte("foo"), 0600)
	c.Assert(err, IsNil)

	err = t.r.AddItem(fullpath)
	c.Assert(err, IsNil)

	err = os.Remove(fullpath)
	c.Assert(err, IsNil)

	err = t.r.DeleteItem(fullpath)
	c.Assert(err, IsNil)
}

func (t *RepositoryDeleteItemTests) TestDirectory(c *C) {
	filename := "foo.txt"
	directory := "foo"
	nestedDir := "bar"
	dirPath := filepath.Join(t.dir, directory)
	nestedDirPath := filepath.Join(dirPath, nestedDir)
	fullpath := filepath.Join(nestedDirPath, filename)
	err := os.MkdirAll(nestedDirPath, 0700)
	c.Assert(err, IsNil)
	err = ioutil.WriteFile(fullpath, []byte("foo"), 0600)
	c.Assert(err, IsNil)

	err = t.r.AddItem(dirPath)
	c.Assert(err, IsNil)

	err = t.r.DeleteItem(dirPath)
	c.Assert(err, IsNil)

	_, err = os.Stat(dirPath)
	c.Assert(os.IsNotExist(err), Equals, true)
}
