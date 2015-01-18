package atomic

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

const (
	// default permissions
	defaultFilePerms = 0600
)

type WriterTests struct {
	dir string
}

var _ = Suite(&WriterTests{})

func (t *WriterTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
}

func (t *WriterTests) TearDownTest(c *C) {
	writerErrorHook = nil
}

func (t *WriterTests) TestTmpFileCreation(c *C) {
	testFilePath := filepath.Join(t.dir, "testfile")
	writer, err := NewWriter(testFilePath, "testprefix", defaultFilePerms)
	c.Assert(err, IsNil)

	_, err = os.Stat(writer.tmpPath())
	c.Assert(err, IsNil)

	_, err = os.Stat(testFilePath)
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (t *WriterTests) TestErrorOnInit(c *C) {
	testFilePath := filepath.Join(t.dir, "testfile")
	writerErrorHook = errors.New("Test")

	_, err := NewWriter(testFilePath, "testprefix", defaultFilePerms)
	c.Assert(err, NotNil)
}

func (t *WriterTests) TestClose(c *C) {
	testFilePath := filepath.Join(t.dir, "testfile")
	writer, err := NewWriter(testFilePath, "testprefix", defaultFilePerms)
	c.Assert(err, IsNil)

	err = writer.Close()
	c.Assert(err, IsNil)

	_, err = os.Stat(writer.tmpPath())
	c.Assert(os.IsNotExist(err), Equals, true)

	_, err = os.Stat(testFilePath)
	c.Assert(err, IsNil)
}

func (t *WriterTests) TestWrite(c *C) {
	testFilePath := filepath.Join(t.dir, "testfile")
	writer, err := NewWriter(testFilePath, "testprefix", defaultFilePerms)
	c.Assert(err, IsNil)

	testBytes := []byte("This is a small test")
	_, err = writer.Write(testBytes)
	c.Assert(err, IsNil)

	err = writer.Close()
	c.Assert(err, IsNil)

	data, err := ioutil.ReadFile(testFilePath)
	c.Assert(err, IsNil)

	c.Assert(data, DeepEquals, testBytes)
}

func (t *WriterTests) TestAbort(c *C) {
	testFilePath := filepath.Join(t.dir, "testfile")
	writer, err := NewStandardWriter(testFilePath, 0770)
	c.Assert(err, IsNil)
	writer.Write([]byte("this is a testfile"))
	writer.Abort()
	err = writer.Close()
	c.Assert(err, IsNil)

	_, err = os.Stat(testFilePath)
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (t *WriterTests) TestAbortNoOverwrite(c *C) {
	testFilePath := filepath.Join(t.dir, "testfile")
	oldFileData := []byte("oldfiledata")
	err := ioutil.WriteFile(testFilePath, oldFileData, defaultFilePerms)
	c.Assert(err, IsNil)
	writer, err := NewStandardWriter(testFilePath, 0770)
	c.Assert(err, IsNil)

	writer.Write([]byte("newfiledata"))
	writer.Abort()
	writer.Close()
	d, err := ioutil.ReadFile(testFilePath)
	c.Assert(d, DeepEquals, oldFileData)
}

func (t *WriterTests) TestPlatformHookError(c *C) {
	testFilePath := filepath.Join(t.dir, "testfile")
	writer, err := NewStandardWriter(testFilePath, 0770)
	c.Assert(err, IsNil)
	writerErrorHook = errors.New("Test")
	err = writer.Close()
	c.Assert(err, NotNil)

}
