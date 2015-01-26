package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

type PullTests struct {
	BaseTests
	repoDir     string
	repoName    string
	testFile    string
	testContent []byte
}

var _ = Suite(&PullTests{BaseTests: BaseTests{}})

func (t *PullTests) TestTooManyArgs(c *C) {
	c.Assert(t.d.run([]string{"pull", "foo"}), Equals, 1)
}

func (t *PullTests) initializeRepository(c *C) {
	t.repoDir = "repo"
	t.runAndExpectCode(c, []string{"init", t.repoDir}, 0)
	err := os.Chdir(t.repoDir)
	c.Assert(err, IsNil)

	t.repoName = "example"
	url := t.ts.hostAndPort
	t.in.Write(t.ts.adminSecret)
	t.in.WriteString("\n")
	t.in.WriteString("y\n") // accept fingerprint
	t.runAndExpectCode(c, []string{"register", url, t.repoName}, 0)

	t.testFile = "foo.txt"
	t.testContent = []byte("Sync works")
	err = ioutil.WriteFile(t.testFile, t.testContent, 0600)
	c.Assert(err, IsNil)

	t.runAndExpectCode(c, []string{"add", t.testFile}, 0)

	t.runAndExpectCode(c, []string{"push"}, 0)
	err = os.Remove(t.testFile)
	c.Assert(err, IsNil)

	err = removeFilesInDir(filepath.Join(".lara", "objects"))
	c.Assert(err, IsNil)

	err = removeFilesInDir(filepath.Join(".lara", "nibs"))
	c.Assert(err, IsNil)

	t.runAndExpectCode(c, []string{"checkout"}, 0)
	_, err = os.Stat(t.testFile)
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (t *PullTests) verifyExpectedDataStructure(c *C) {
	_, err := os.Stat(t.testFile)
	c.Assert(os.IsNotExist(err), Equals, true)

	t.runAndExpectCode(c, []string{"checkout"}, 0)
	content, err := ioutil.ReadFile(t.testFile)
	c.Assert(err, IsNil)
	c.Assert(content, DeepEquals, t.testContent)
}

func (t *PullTests) TestPull(c *C) {
	t.initializeRepository(c)

	t.d.run([]string{"pull"})
	d, _ := ioutil.ReadAll(t.err)
	fmt.Println(string(d))
	t.runAndExpectCode(c, []string{"pull"}, 0)

	t.verifyExpectedDataStructure(c)
}

func (t *PullTests) TestPullFull(c *C) {
	t.initializeRepository(c)
	t.runAndExpectCode(c, []string{"pull", "--full"}, 0)
	t.verifyExpectedDataStructure(c)
}
