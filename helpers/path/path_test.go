package path

import (
	"os"
	"path/filepath"
	"strings"

	. "gopkg.in/check.v1"
)

var _ = Suite(&PathTests{})

type PathTests struct {
	oldDir string
	dir    string
}

func (t *PathTests) SetUpTest(c *C) {
	var err error
	t.oldDir, err = os.Getwd()
	c.Assert(err, IsNil)
	// Fix for OSX systems. Temporary folder lies in a symlink directory
	// /var/folders which is actually at /private/var/folders
	t.dir, err = filepath.EvalSymlinks(c.MkDir())
	c.Assert(err, IsNil)
	err = os.Chdir(t.dir)
	c.Assert(err, IsNil)
}

func (t *PathTests) TearDownTest(c *C) {
	os.Chdir(t.oldDir)
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
	path := strings.Replace(t.dir, string(filepath.Separator),
		string(filepath.Separator)+string(filepath.Separator), -1)
	c.Assert(path, Not(Equals), t.dir)
	n, err := Normalize(path)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, t.dir)
}

func (t *PathTests) TestIsBelow(c *C) {
	basePath := t.dir
	belowPath := filepath.Dir(t.dir)

	is, err := IsBelow(basePath, belowPath)
	c.Assert(err, IsNil)
	c.Assert(is, Equals, true)
}
