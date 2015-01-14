package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

type BaseTests struct {
	dir   string
	oldWd string
	err   *bytes.Buffer
	out   *bytes.Buffer
	in    *bytes.Buffer
	d     *Dispatcher
	ts    *TestServer
}

func (t *BaseTests) SetUpTest(c *C) {
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

func (t *BaseTests) TearDownTest(c *C) {
	t.ts.Close()
	os.Chdir(t.oldWd)
}

func (t *BaseTests) initRepo(c *C) {
	repoDir := "repo"
	c.Assert(t.d.run([]string{"init", repoDir}), Equals, 0)
	err := os.Chdir(repoDir)
	c.Assert(err, IsNil)
}

func (t *BaseTests) registerServerInRepo(c *C) {
	repoName := "example"
	url := t.ts.hostAndPort
	t.in.Write(t.ts.adminSecret)
	t.in.WriteString("\n")
	c.Assert(t.d.run([]string{"register", url, repoName}), Equals, 0)
}

func (t *BaseTests) serverRepoPath() string {
	return filepath.Join(t.ts.basePath, "example", ".lara")
}

func (t *BaseTests) runAndExpectCode(c *C, args []string, expectedReturnCode int) {
	rCode := t.d.run(args)
	if rCode != expectedReturnCode {
		data, _ := ioutil.ReadAll(t.err)
		c.Error(string(data))
		c.FailNow()
	}
}
