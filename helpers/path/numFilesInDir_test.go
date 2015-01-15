package path

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

func (t *PathTests) TestNumFilesInDir(c *C) {
	num, err := NumFilesInDir(t.dir)
	c.Assert(err, IsNil)
	c.Assert(num, Equals, 0)
}

func (t *PathTests) TestNumFilesInDirErr(c *C) {
	num, err := NumFilesInDir(filepath.Join(t.dir, "non-existing"))
	c.Assert(os.IsNotExist(err), Equals, true)
	c.Assert(num, Equals, 0)
}

func (t *PathTests) TestNumFilesInDirOne(c *C) {
	err := ioutil.WriteFile(filepath.Join(t.dir, "foo.txt"), []byte{}, 0600)
	c.Assert(err, IsNil)
	num, err := NumFilesInDir(t.dir)
	c.Assert(err, IsNil)
	c.Assert(num, Equals, 1)
}
