package main

import (
	"io/ioutil"
	"path/filepath"

	. "gopkg.in/check.v1"

	"github.com/hoffie/larasync/helpers/test"
)

type SyncTests struct {
	BaseTests
}

var _ = Suite(&SyncTests{BaseTests{}})

func (t *SyncTests) TestTooManyArgs(c *C) {
	c.Assert(t.d.run([]string{"push", "foo"}), Equals, 1)
}

func (t *SyncTests) TestSync(c *C) {
	t.initRepo(c)
	t.registerServerInRepo(c)
	repoName := "example"

	uploadedTestFile := "foo2.txt"
	err := ioutil.WriteFile(uploadedTestFile, []byte("Sync works downwards"), 0600)
	c.Assert(err, IsNil)
	c.Assert(t.d.run([]string{"add", uploadedTestFile}), Equals, 0)
	c.Assert(t.d.run([]string{"push"}), Equals, 0)

	err = removeFilesInDir(filepath.Join(".lara", "objects"))
	c.Assert(err, IsNil)

	err = removeFilesInDir(filepath.Join(".lara", "nibs"))
	c.Assert(err, IsNil)

	testFile := "foo.txt"
	err = ioutil.WriteFile(testFile, []byte("Sync works upwards"), 0600)
	c.Assert(err, IsNil)

	c.Assert(t.d.run([]string{"add", testFile}), Equals, 0)

	res := t.d.run([]string{"sync"})
	if res != 0 {
		data, _ := ioutil.ReadAll(t.err)
		c.Errorf(string(data))
		return
	}

	num, err := test.NumFilesInDir(filepath.Join(t.ts.basePath,
		repoName, ".lara", "nibs"))
	c.Assert(err, IsNil)
	c.Assert(num, Equals, 2)

	num, err = test.NumFilesInDir(filepath.Join(t.ts.basePath,
		repoName, ".lara", "objects"))
	c.Assert(err, IsNil)
	c.Assert(num, Equals, 4)
}
