package main

import (
	"io/ioutil"
	"path/filepath"

	. "gopkg.in/check.v1"

	"github.com/hoffie/larasync/helpers/path"
	"github.com/hoffie/larasync/repository"
)

type SyncTests struct {
	BaseTests
	repoName string
}

var _ = Suite(&SyncTests{
	BaseTests: BaseTests{},
})

func (t *SyncTests) SetUpTest(c *C) {
	t.BaseTests.SetUpTest(c)
	t.repoName = "example"
}

func (t *SyncTests) TestTooManyArgs(c *C) {
	c.Assert(t.d.run([]string{"sync", "foo"}), Equals, 1)
}

func (t *SyncTests) prepareForSync(c *C) {
	t.initRepo(c)
	t.registerServerInRepo(c)

	uploadedTestFile := "foo2.txt"
	err := ioutil.WriteFile(uploadedTestFile, []byte("Sync works downwards"), 0600)
	c.Assert(err, IsNil)
	t.runAndExpectCode(c, []string{"add", uploadedTestFile}, 0)
	t.runAndExpectCode(c, []string{"push"}, 0)

	err = removeFilesInDir(filepath.Join(".lara", "objects"))
	c.Assert(err, IsNil)

	err = removeFilesInDir(filepath.Join(".lara", "nibs"))
	c.Assert(err, IsNil)

	testFile := "foo.txt"
	err = ioutil.WriteFile(testFile, []byte("Sync works upwards"), 0600)
	c.Assert(err, IsNil)
}

func (t *SyncTests) verifyAfterSync(c *C) {
	repoName := t.repoName
	num, err := path.NumFilesInDir(filepath.Join(t.ts.basePath,
		repoName, ".lara", "nibs"))
	c.Assert(err, IsNil)
	c.Assert(num, Equals, 2)

	num, err = path.NumFilesInDir(filepath.Join(t.ts.basePath,
		repoName, ".lara", "objects"))
	c.Assert(err, IsNil)
	c.Assert(num, Equals, 4)
}

func (t *SyncTests) TestFullSync(c *C) {
	t.prepareForSync(c)
	t.runAndExpectCode(c, []string{"sync", "--full"}, 0)
	t.verifyAfterSync(c)
}

func (t *SyncTests) TestSync(c *C) {
	t.prepareForSync(c)
	t.runAndExpectCode(c, []string{"sync"}, 0)
	t.verifyAfterSync(c)

}

func (t *SyncTests) breakLocalFingerprint(c *C) {
	scPath := filepath.Join(".lara", "state.json")
	sc := &repository.StateConfig{Path: scPath}
	err := sc.Load()
	c.Assert(err, IsNil)
	sc.DefaultServer.Fingerprint = "broken"
	err = sc.Save()
	c.Assert(err, IsNil)
}

func (t *SyncTests) TestSyncFingerprintFail(c *C) {
	t.TestSync(c)
	t.runAndExpectCode(c, []string{"sync"}, 0)
	t.breakLocalFingerprint(c)
	t.runAndExpectCode(c, []string{"sync"}, 1)
}
