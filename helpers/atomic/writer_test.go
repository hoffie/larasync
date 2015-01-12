package atomic

import (
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

func (t *WriterTests) TestTmpFileCreation(c *C) {
	testFilePath := filepath.Join(t.dir, "testfile")
	writer, err := NewWriter(testFilePath, "testprefix", defaultFilePerms)
	c.Assert(err, IsNil)

	_, err = os.Stat(writer.tmpPath())
	c.Assert(err, IsNil)

	_, err = os.Stat(testFilePath)
	c.Assert(os.IsNotExist(err), Equals, true)
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

func (t *WriterTests) TestFileModeWrite(c *C) {
	testFilePath := filepath.Join(t.dir, "testfile")
	writer, err := NewStandardWriter(testFilePath, 0770)

	writer.Close()

	stat, err := os.Stat(testFilePath)
	c.Assert(err, IsNil)
	var fileMode os.FileMode
	fileMode = 0770

	c.Assert(stat.Mode(), Equals, fileMode)
}
