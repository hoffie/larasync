package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"

	"github.com/hoffie/larasync/helpers/path"
)

type PushTests struct {
	BaseTests
}

var _ = Suite(&PushTests{BaseTests{}})

func (t *PushTests) TestTooManyArgs(c *C) {
	t.runAndExpectCode(c, []string{"push", "foo"}, 1)
}

func (t *PushTests) TestPush(c *C) {
	repoDir := "repo"
	t.runAndExpectCode(c, []string{"init", repoDir}, 0)
	err := os.Chdir(repoDir)
	c.Assert(err, IsNil)

	repoName := "example"
	url := t.ts.hostAndPort
	t.in.Write(t.ts.adminSecret)
	t.in.WriteString("\n")
	t.runAndExpectCode(c, []string{"register", url, repoName}, 0)

	testFile := "foo.txt"
	err = ioutil.WriteFile(testFile, []byte("Sync works"), 0600)
	c.Assert(err, IsNil)

	t.runAndExpectCode(c, []string{"add", testFile}, 0)

	num, err := path.NumFilesInDir(filepath.Join(t.ts.basePath,
		repoName, ".lara", "nibs"))
	c.Assert(err, IsNil)
	c.Assert(num, Equals, 0)

	num, err = path.NumFilesInDir(filepath.Join(t.ts.basePath,
		repoName, ".lara", "objects"))
	c.Assert(err, IsNil)
	c.Assert(num, Equals, 0)
	t.runAndExpectCode(c, []string{"push"}, 0)

	num, err = path.NumFilesInDir(filepath.Join(t.ts.basePath,
		repoName, ".lara", "nibs"))
	c.Assert(err, IsNil)
	c.Assert(num, Equals, 1)

	num, err = path.NumFilesInDir(filepath.Join(t.ts.basePath,
		repoName, ".lara", "objects"))
	c.Assert(err, IsNil)
	c.Assert(num, Equals, 2)
}
