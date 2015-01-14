package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	. "gopkg.in/check.v1"
)

type CloneTests struct {
	BaseTests
}

var _ = Suite(&CloneTests{BaseTests{}})
var authURLRegex = regexp.MustCompile(`(?m)^https?://.*$`)

func (t *CloneTests) TestClone(c *C) {
	testFileName := "foo.txt"
	testFileContent := []byte("test content")
	cloneName := "test-clone"

	t.initRepo(c)
	t.registerServerInRepo(c)

	c.Assert(t.d.run([]string{"authorize-new-client"}), Equals, 0)
	url := authURLRegex.FindString(t.out.String())
	c.Assert(strings.HasPrefix(url, "http"), Equals, true)

	err := ioutil.WriteFile(testFileName, testFileContent, 0600)
	c.Assert(err, IsNil)
	c.Assert(t.d.run([]string{"add", testFileName}), Equals, 0)
	c.Assert(t.d.run([]string{"sync"}), Equals, 0)

	err = os.Chdir(t.dir)
	c.Assert(err, IsNil)

	// we are testing with absolute paths here as "clone" will
	// change the cwd!
	clonePath := filepath.Join(t.dir, cloneName)
	c.Assert(t.d.run([]string{"clone", clonePath, url}), Equals, 0)
	err = os.Chdir(clonePath)
	c.Assert(err, IsNil)

	gotContent, err := ioutil.ReadFile(testFileName)
	c.Assert(err, IsNil)
	c.Assert(gotContent, DeepEquals, testFileContent)
}
