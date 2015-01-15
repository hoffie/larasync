package main

import (
	"bytes"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

type RegisterTests struct {
	dir   string
	oldWd string
	err   *bytes.Buffer
	out   *bytes.Buffer
	in    *bytes.Buffer
	d     *Dispatcher
	ts    *TestServer
}

var _ = Suite(&RegisterTests{})

func (t *RegisterTests) SetUpTest(c *C) {
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

func (t *RegisterTests) TearDownTest(c *C) {
	t.ts.Close()
	os.Chdir(t.oldWd)
}

func (t *RegisterTests) TestRegisterNoArgs(c *C) {
	c.Assert(t.d.run([]string{"register"}), Equals, 1)
}

func (t *RegisterTests) TestRegisterOnlyURL(c *C) {
	url := "http://127.0.0.1:14124"
	c.Assert(t.d.run([]string{"register", url}), Equals, 1)
}

func (t *RegisterTests) TestRegisterNoRepo(c *C) {
	c.Assert(t.d.run([]string{"register", "127.0.0.1:0", "unused"}), Equals, 1)
}

func (t *RegisterTests) TestRegister(c *C) {
	repoDir := "repo"
	c.Assert(t.d.run([]string{"init", repoDir}), Equals, 0)
	err := os.Chdir(repoDir)
	c.Assert(err, IsNil)

	repoName := "example"
	url := t.ts.hostAndPort
	t.in.Write(t.ts.adminSecret)
	t.in.WriteString("\n")
	c.Assert(t.d.run([]string{"register", url, repoName}), Equals, 0)

	// check the server dir:
	serverKeyFile := filepath.Join(t.ts.basePath, repoName, ".lara",
		"keys", "signing.pub")
	stat, err := os.Stat(serverKeyFile)
	c.Assert(err, IsNil)
	c.Assert(stat.IsDir(), Equals, false)
}
