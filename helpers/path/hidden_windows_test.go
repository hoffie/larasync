package path

import (
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

var _ = Suite(&HiddenWindowsTests{})

type HiddenWindowsTests struct {
	dir string
}

func (t *HiddenWindowsTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	c.Assert(os.Mkdir(t.getDir(), 0700), IsNil)
}

func (t *HiddenWindowsTests) getDir() string {
	return filepath.Join(t.dir, "test")
}

func (t *HiddenWindowsTests) TestDirectoryHidden(c *C) {
	res, err := IsHidden(t.getDir())
	c.Assert(err, IsNil)
	c.Assert(res, Equals, false)
}

func (t *HiddenWindowsTests) TestHideDirectory(c *C) {
	c.Assert(Hide(t.getDir()), IsNil)
	res, err := IsHidden(t.getDir())
	c.Assert(err, IsNil)
	c.Assert(res, Equals, true)
}

func (t *HiddenWindowsTests) TestError(c *C) {
	c.Assert(os.Remove(t.getDir()), IsNil)
	c.Assert(Hide(t.getDir()), NotNil)
}

func (t *HiddenWindowsTests) TestCheckError(c *C) {
	c.Assert(os.Remove(t.getDir()), IsNil)
	_, err := IsHidden(t.getDir())
	c.Assert(err, NotNil)
}
