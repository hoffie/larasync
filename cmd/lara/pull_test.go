package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

type PullTests struct {
	dir   string
	oldWd string
	err   *bytes.Buffer
	out   *bytes.Buffer
	in    *bytes.Buffer
	d     *Dispatcher
	ts    *TestServer
}

var _ = Suite(&PullTests{})

func (t *PullTests) SetUpTest(c *C) {
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

func (t *PullTests) TearDownTest(c *C) {
	t.ts.Close()
	os.Chdir(t.oldWd)
}

func (t *PullTests) TestTooManyArgs(c *C) {
	c.Assert(t.d.run([]string{"pull", "foo"}), Equals, 1)
}

func (t *PullTests) TestPull(c *C) {
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
	testContent := []byte("Sync works")
	err = ioutil.WriteFile(testFile, testContent, 0600)
	c.Assert(err, IsNil)

	c.Assert(t.d.run([]string{"add", testFile}), Equals, 0)
	c.Assert(t.d.run([]string{"push"}), Equals, 0)
	err = os.Remove(testFile)
	c.Assert(err, IsNil)

	err = removeFilesInDir(filepath.Join(".lara", "objects"))
	c.Assert(err, IsNil)

	err = removeFilesInDir(filepath.Join(".lara", "nibs"))
	c.Assert(err, IsNil)

	c.Assert(t.d.run([]string{"checkout"}), Equals, 0)
	_, err = os.Stat(testFile)
	c.Assert(os.IsNotExist(err), Equals, true)

	c.Assert(t.d.run([]string{"pull"}), Equals, 0)

	_, err = os.Stat(testFile)
	c.Assert(os.IsNotExist(err), Equals, true)

	c.Assert(t.d.run([]string{"checkout"}), Equals, 0)
	content, err := ioutil.ReadFile(testFile)
	c.Assert(err, IsNil)
	c.Assert(content, DeepEquals, testContent)
}
