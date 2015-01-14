package path

import (
	"os"
	"path/filepath"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&PathTests{})

type PathTests struct {
	dir string
}

func (t *PathTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	err := os.Chdir(t.dir)
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
	is, err := IsBelow("//foo/a", "/foo")
	c.Assert(err, IsNil)
	c.Assert(is, Equals, true)
}
