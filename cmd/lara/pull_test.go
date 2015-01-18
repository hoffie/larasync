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
}

var _ = Suite(&PullTests{BaseTests{}})

func (t *PullTests) TestTooManyArgs(c *C) {
	c.Assert(t.d.run([]string{"pull", "foo"}), Equals, 1)
}

func (t *PullTests) TestPull(c *C) {
	repoDir := "repo"
	t.runAndExpectCode(c, []string{"init", repoDir}, 0)
	err := os.Chdir(repoDir)
	c.Assert(err, IsNil)

	repoName := "example"
	url := t.ts.hostAndPort
	t.in.Write(t.ts.adminSecret)
	t.in.WriteString("\n")
	t.in.WriteString("y\n") // accept fingerprint
	t.runAndExpectCode(c, []string{"register", url, repoName}, 0)

	testFile := "foo.txt"
	testContent := []byte("Sync works")
	err = ioutil.WriteFile(testFile, testContent, 0600)
	c.Assert(err, IsNil)

	t.runAndExpectCode(c, []string{"add", testFile}, 0)
	t.runAndExpectCode(c, []string{"push"}, 0)
	err = os.Remove(testFile)
	c.Assert(err, IsNil)

	err = removeFilesInDir(filepath.Join(".lara", "objects"))
	c.Assert(err, IsNil)

	err = removeFilesInDir(filepath.Join(".lara", "nibs"))
	c.Assert(err, IsNil)

	t.runAndExpectCode(c, []string{"checkout"}, 0)
	_, err = os.Stat(testFile)
	c.Assert(os.IsNotExist(err), Equals, true)

	t.d.run([]string{"pull"})
	d, _ := ioutil.ReadAll(t.err)
	fmt.Println(string(d))
	t.runAndExpectCode(c, []string{"pull"}, 0)

	_, err = os.Stat(testFile)
	c.Assert(os.IsNotExist(err), Equals, true)

	t.runAndExpectCode(c, []string{"checkout"}, 0)
	content, err := ioutil.ReadFile(testFile)
	c.Assert(err, IsNil)
	c.Assert(content, DeepEquals, testContent)
}
