package tracker

import (
	"io/ioutil"
	"path/filepath"

	. "gopkg.in/check.v1"
)

var _ = Suite(&NIBSearchResponseTest{})

type NIBSearchResponseTest struct {
	dirPath string
}

func (t *NIBSearchResponseTest) SetUpTest(c *C) {
	t.dirPath = c.MkDir()
}

func (t *NIBSearchResponseTest) newResponse(NIBID string, path string) *NIBSearchResponse {
	return NewNIBSearchResponse(NIBID, path, t.dirPath)
}

func (t *NIBSearchResponseTest) TestFileExistsNegative(c *C) {
	resp := t.newResponse("asdf", "test")
	c.Assert(resp.FileExists(), Equals, false)
}

func (t *NIBSearchResponseTest) TestFileExistsPositive(c *C) {
	resp := t.newResponse("asdf", "test")
	err := ioutil.WriteFile(resp.AbsPath(), []byte("test"), 0700)
	c.Assert(err, IsNil)
	c.Assert(resp.FileExists(), Equals, true)
}

func (t *NIBSearchResponseTest) TestAbsPath(c *C) {
	filePath := "test"
	resp := t.newResponse("asdf", filePath)
	err := ioutil.WriteFile(resp.AbsPath(), []byte("test"), 0700)
	c.Assert(err, IsNil)
	expectedPath := filepath.Join(t.dirPath, filePath)
	expectedPath, err = filepath.EvalSymlinks(expectedPath)
	c.Assert(err, IsNil)
	c.Assert(resp.AbsPath(), Equals, expectedPath)
}
