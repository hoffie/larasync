package repository

import (
	"io/ioutil"
	"path/filepath"

	. "gopkg.in/check.v1"
)

var _ = Suite(&RepositoryAddItemTests{})

type RepositoryAddItemTests struct {
	dir string
	r   *Repository
}

func (t *RepositoryAddItemTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	t.r = New(t.dir)
	err := t.r.CreateManagementDir()
	c.Assert(err, IsNil)
	err = t.r.keys.CreateSigningKey()
	c.Assert(err, IsNil)

	err = t.r.keys.CreateEncryptionKey()
	c.Assert(err, IsNil)

	err = t.r.keys.CreateHashingKey()
	c.Assert(err, IsNil)
}

func (t *RepositoryAddItemTests) TestWriteFileToChunks(c *C) {
	path := filepath.Join(t.dir, "foo.txt")
	err := ioutil.WriteFile(path, []byte("foo"), 0600)
	c.Assert(err, IsNil)
	numFiles, err := numFilesInDir(filepath.Join(t.dir, ".lara", "objects"))
	c.Assert(err, IsNil)
	c.Assert(numFiles, Equals, 0)
	err = t.r.AddItem(path)
	c.Assert(err, IsNil)
	numFiles, err = numFilesInDir(filepath.Join(t.dir, ".lara", "objects"))
	c.Assert(err, IsNil)
	c.Assert(numFiles, Equals, 2)
}

// TestExistingFileNIBReuse ensures that pre-existing NIBs for a path are
// re-used upon updates.
func (t *RepositoryAddItemTests) TestExistingFileNIBReuse(c *C) {
	c.Skip("waiting for decision on #67")
	nibsPath := filepath.Join(t.dir, ".lara", "nibs")
	path := filepath.Join(t.dir, "foo.txt")
	err := ioutil.WriteFile(path, []byte("foo"), 0600)
	c.Assert(err, IsNil)

	numFiles, err := numFilesInDir(nibsPath)
	c.Assert(err, IsNil)
	c.Assert(numFiles, Equals, 0)

	err = t.r.AddItem(path)
	c.Assert(err, IsNil)

	numFiles, err = numFilesInDir(nibsPath)
	c.Assert(err, IsNil)
	c.Assert(numFiles, Equals, 1)

	err = t.r.AddItem(path)
	c.Assert(err, IsNil)

	numFiles, err = numFilesInDir(nibsPath)
	c.Assert(err, IsNil)
	c.Assert(numFiles, Equals, 1)

}
