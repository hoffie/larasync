package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"

	"github.com/hoffie/larasync/helpers/test"
)

type PushTests struct {
	BaseTests
}

var _ = Suite(&PushTests{BaseTests{}})

func (t *PushTests) TestTooManyArgs(c *C) {
	c.Assert(t.d.run([]string{"push", "foo"}), Equals, 1)
}

func (t *PushTests) TestPush(c *C) {
	repoDir := "repo"
	c.Assert(t.d.run([]string{"init", repoDir}), Equals, 0)
	err := os.Chdir(repoDir)
	c.Assert(err, IsNil)

	repoName := "example"
	url := t.ts.hostAndPort
	t.in.Write(t.ts.adminSecret)
	t.in.WriteString("\n")
	c.Assert(t.d.run([]string{"register", url, repoName}), Equals, 0)

	testFile := "foo.txt"
	err = ioutil.WriteFile(testFile, []byte("Sync works"), 0600)
	c.Assert(err, IsNil)

	c.Assert(t.d.run([]string{"add", testFile}), Equals, 0)

	num, err := test.NumFilesInDir(filepath.Join(t.ts.basePath,
		repoName, ".lara", "nibs"))
	c.Assert(err, IsNil)
	c.Assert(num, Equals, 0)

	num, err = test.NumFilesInDir(filepath.Join(t.ts.basePath,
		repoName, ".lara", "objects"))
	c.Assert(err, IsNil)
	c.Assert(num, Equals, 0)
	c.Assert(t.d.run([]string{"push"}), Equals, 0)

	num, err = test.NumFilesInDir(filepath.Join(t.ts.basePath,
		repoName, ".lara", "nibs"))
	c.Assert(err, IsNil)
	c.Assert(num, Equals, 1)

	num, err = test.NumFilesInDir(filepath.Join(t.ts.basePath,
		repoName, ".lara", "objects"))
	c.Assert(err, IsNil)
	c.Assert(num, Equals, 2)
}
