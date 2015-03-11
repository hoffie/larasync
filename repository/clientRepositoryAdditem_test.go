package repository

import (
	"io/ioutil"
	"path/filepath"
	"runtime"

	"github.com/hoffie/larasync/helpers/path"
	. "gopkg.in/check.v1"
)

var _ = Suite(&RepositoryAddItemTests{})

type RepositoryAddItemTests struct {
	dir string
	r   *ClientRepository
}

func (t *RepositoryAddItemTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	t.r = NewClient(t.dir)
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
	fullpath := filepath.Join(t.dir, "foo.txt")
	err := ioutil.WriteFile(fullpath, []byte("foo"), 0600)
	c.Assert(err, IsNil)
	numFiles, err := path.NumFilesInDir(filepath.Join(t.dir, ".lara", "objects"))
	c.Assert(err, IsNil)
	c.Assert(numFiles, Equals, 0)
	err = t.r.AddItem(fullpath)
	c.Assert(err, IsNil)
	numFiles, err = path.NumFilesInDir(filepath.Join(t.dir, ".lara", "objects"))
	c.Assert(err, IsNil)
	c.Assert(numFiles, Equals, 2)
}

func (t *RepositoryAddItemTests) TestAddtoNIBStore(c *C) {
	fullpath := filepath.Join(t.dir, "foo.txt")
	err := ioutil.WriteFile(fullpath, []byte("foo"), 0600)
	c.Assert(err, IsNil)
	err = t.r.AddItem(fullpath)
	c.Assert(err, IsNil)
	tracker, err := t.r.NIBTracker()
	c.Assert(err, IsNil)
	d, err := tracker.Get("foo.txt")
	c.Assert(err, IsNil)
	c.Assert(d.Path, Equals, "foo.txt")
	nibID, err := t.r.pathToNIBID("foo.txt")
	c.Assert(err, IsNil)
	c.Assert(d.NIBID, Equals, nibID)
}

// TestExistingFileNIBReuse ensures that pre-existing NIBs for a path are
// re-used upon updates.
func (t *RepositoryAddItemTests) TestExistingFileNIBReuse(c *C) {
	nibsPath := filepath.Join(t.dir, ".lara", "nibs")
	filename := "foo.txt"
	fullpath := filepath.Join(t.dir, filename)
	err := ioutil.WriteFile(fullpath, []byte("foo"), 0600)
	c.Assert(err, IsNil)

	numFiles, err := path.NumFilesInDir(nibsPath)
	c.Assert(err, IsNil)
	c.Assert(numFiles, Equals, 0)

	err = t.r.AddItem(fullpath)
	c.Assert(err, IsNil)

	numFiles, err = path.NumFilesInDir(nibsPath)
	c.Assert(err, IsNil)
	c.Assert(numFiles, Equals, 1)

	err = ioutil.WriteFile(fullpath, []byte("foo2"), 0600)
	c.Assert(err, IsNil)

	err = t.r.AddItem(fullpath)
	c.Assert(err, IsNil)

	numFiles, err = path.NumFilesInDir(nibsPath)
	c.Assert(err, IsNil)
	c.Assert(numFiles, Equals, 1)

	nibID, err := t.r.pathToNIBID(filename)
	c.Assert(err, IsNil)
	nib, err := t.r.nibStore.Get(nibID)
	c.Assert(err, IsNil)
	c.Assert(len(nib.Revisions), Equals, 2)
	c.Assert(nib.Revisions[0].UTCTimestamp, Not(Equals), int64(0))
	c.Assert(nib.Revisions[0].UTCTimestamp <= nib.Revisions[1].UTCTimestamp,
		Equals, true)
}

// TestExistingFileNoChange ensures that no unnecessary updates
// are recorded.
func (t *RepositoryAddItemTests) TestExistingFileNoChange(c *C) {
	nibsPath := filepath.Join(t.dir, ".lara", "nibs")
	filename := "foo.txt"
	fullpath := filepath.Join(t.dir, filename)
	err := ioutil.WriteFile(fullpath, []byte("foo"), 0600)
	c.Assert(err, IsNil)

	numFiles, err := path.NumFilesInDir(nibsPath)
	c.Assert(err, IsNil)
	c.Assert(numFiles, Equals, 0)

	err = t.r.AddItem(fullpath)
	c.Assert(err, IsNil)

	numFiles, err = path.NumFilesInDir(nibsPath)
	c.Assert(err, IsNil)
	c.Assert(numFiles, Equals, 1)

	err = t.r.AddItem(fullpath)
	c.Assert(err, IsNil)

	numFiles, err = path.NumFilesInDir(nibsPath)
	c.Assert(err, IsNil)
	c.Assert(numFiles, Equals, 1)

	nibID, err := t.r.pathToNIBID(filename)
	c.Assert(err, IsNil)
	nib, err := t.r.nibStore.Get(nibID)
	c.Assert(err, IsNil)
	c.Assert(len(nib.Revisions), Equals, 1)
}

func (t *RepositoryAddItemTests) TestAddDotLara(c *C) {
	err := t.r.AddItem(filepath.Join(t.r.Path, managementDirName))
	c.Assert(err, Equals, ErrRefusingWorkOnDotLara)
}

func (t *RepositoryAddItemTests) TestAddDotLaraModified(c *C) {
	path := string(filepath.Separator) + filepath.Join(t.r.Path, managementDirName)

	err := t.r.AddItem(path)
	if runtime.GOOS != "windows" {
		c.Assert(err, Equals, ErrRefusingWorkOnDotLara)
	} else {
		c.Assert(err, NotNil)
	}
}

func (t *RepositoryAddItemTests) TestAddDotLaraSubdir(c *C) {
	path := filepath.Join(t.r.Path, managementDirName, nibsDirName)
	err := t.r.AddItem(path)
	c.Assert(err, Equals, ErrRefusingWorkOnDotLara)
}
