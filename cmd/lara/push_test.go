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
	repoName string
}

var _ = Suite(&PushTests{BaseTests: BaseTests{}})

func (t *PushTests) TestTooManyArgs(c *C) {
	t.runAndExpectCode(c, []string{"push", "foo"}, 1)
}

func (t *PushTests) initializeRepository(c *C) {
	repoDir := "repo"
	t.runAndExpectCode(c, []string{"init", repoDir}, 0)
	err := os.Chdir(repoDir)
	c.Assert(err, IsNil)

	repoName := "example"
	t.repoName = repoName
	url := t.ts.hostAndPort
	t.in.Write(t.ts.adminSecret)
	t.in.WriteString("\n")
	t.in.WriteString("y\n")
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
}

func (t *PushTests) verifyRepository(c *C) {
	repoName := t.repoName
	num, err := path.NumFilesInDir(filepath.Join(t.ts.basePath,
		repoName, ".lara", "nibs"))
	c.Assert(err, IsNil)
	c.Assert(num, Equals, 1)

	num, err = path.NumFilesInDir(filepath.Join(t.ts.basePath,
		repoName, ".lara", "objects"))
	c.Assert(err, IsNil)
	c.Assert(num, Equals, 2)
}

func (t *PushTests) TestPush(c *C) {
	t.initializeRepository(c)

	t.runAndExpectCode(c, []string{"push"}, 0)

	t.verifyRepository(c)
}

func (t *PushTests) TestPushFull(c *C) {
	t.initializeRepository(c)

	t.runAndExpectCode(c, []string{"push", "--full"}, 0)

	t.verifyRepository(c)
}

func (t *PushTests) TestDoublePush(c *C) {
	t.initializeRepository(c)

	t.runAndExpectCode(c, []string{"push"}, 0)
	t.verifyRepository(c)

	t.runAndExpectCode(c, []string{"push"}, 0)
	t.verifyRepository(c)
}
