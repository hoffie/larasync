package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"

	"github.com/hoffie/larasync/helpers/test"
)

type PushTests struct {
	dir   string
	oldWd string
	err   *bytes.Buffer
	out   *bytes.Buffer
	in    *bytes.Buffer
	d     *Dispatcher
	ts    *TestServer
}

var _ = Suite(&PushTests{})

func (t *PushTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	wd, err := os.Getwd()
	c.Assert(err, IsNil)
	t.oldWd = wd

	err = os.Chdir(t.dir)
	c.Assert(err, IsNil)

	t.err = new(bytes.Buffer)
	t.out = new(bytes.Buffer)
	t.in = new(bytes.Buffer)
	t.d = &Dispatcher{stderr: t.err, stdout: t.out, stdin: t.in}

	ts, err := NewTestServer()
	c.Assert(err, IsNil)
	t.ts = ts
}

func (t *PushTests) TearDownTest(c *C) {
	t.ts.Close()
	os.Chdir(t.oldWd)
}

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
