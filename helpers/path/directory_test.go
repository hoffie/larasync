package path

import (
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
	"io/ioutil"
)

var _ = Suite(&DirectoryTests{})

type DirectoryTests struct {
	dir string
}

func (t *DirectoryTests) SetUpTest(c *C) {
	var err error
	// Fix for OSX systems. Temporary folder lies in a symlink directory
	// /var/folders which is actually at /private/var/folders
	t.dir, err = filepath.EvalSymlinks(c.MkDir())
	c.Assert(err, IsNil)
}

func (t *DirectoryTests) TestDeleteEmpty(c *C) {
	c.Assert(os.Mkdir(filepath.Join(t.dir, "test"), 0700), IsNil)
	c.Assert(CleanUpEmptyDirs(t.dir), IsNil)

	_, err := os.Stat(t.dir)
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (t *DirectoryTests) TestNotDeleteUnempty(c *C) {
	dirName := filepath.Join(t.dir, "test")
	c.Assert(os.Mkdir(dirName, 0700), IsNil)

	fileName := filepath.Join(dirName, "test.txt")
	err := ioutil.WriteFile(fileName, []byte{}, 0600)
	c.Assert(err, IsNil)

	err = CleanUpEmptyDirs(t.dir)
	c.Assert(err, IsNil)

	stat, err := os.Stat(t.dir)
	c.Assert(err, IsNil)
	c.Assert(stat.IsDir(), Equals, true)

	stat, err = os.Stat(fileName)
	c.Assert(err, IsNil)
	c.Assert(stat.IsDir(), Equals, false)
}

func (t *DirectoryTests) TestDeleteHybrid(c *C) {
	dirName := filepath.Join(t.dir, "test")
	c.Assert(os.Mkdir(dirName, 0700), IsNil)

	fileName := filepath.Join(dirName, "test.txt")
	err := ioutil.WriteFile(fileName, []byte{}, 0600)
	c.Assert(err, IsNil)

	dirName2 := filepath.Join(t.dir, "todelete")
	c.Assert(os.Mkdir(dirName2, 0700), IsNil)

	err = CleanUpEmptyDirs(t.dir)
	c.Assert(err, IsNil)

	_, err = os.Stat(dirName2)
	c.Assert(os.IsNotExist(err), Equals, true)

	_, err = os.Stat(fileName)
	c.Assert(err, IsNil)
}
