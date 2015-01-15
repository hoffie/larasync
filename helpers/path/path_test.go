package path

import (
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

var _ = Suite(&PathTests{})

type PathTests struct {
	dir string
}

func (t *PathTests) SetUpTest(c *C) {
	var err error
	// Fix for OSX systems. Temporary folder lies in a symlink directory
	// /var/folders which is actually at /private/var/folders
	t.dir, err = filepath.EvalSymlinks(c.MkDir())
	c.Assert(err, IsNil)
	err = os.Chdir(t.dir)
	c.Assert(err, IsNil)
}

func (t *PathTests) TestNormalizeAbs(c *C) {
	err := os.Chdir(filepath.Join(t.dir, ".."))
	c.Assert(err, IsNil)
	base := filepath.Base(t.dir)
	n, err := Normalize(base)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, t.dir)
}

func (t *PathTests) TestNormalizeRedundantChar(c *C) {
	n, err := Normalize(string(filepath.Separator) + t.dir)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, t.dir)
}

func (t *PathTests) TestIsBelow(c *C) {
	basePath := "//foo/a"
	belowPath := "/foo"
	
	is, err := IsBelow(basePath, belowPath)
	c.Assert(err, IsNil)
	c.Assert(is, Equals, true)
}
