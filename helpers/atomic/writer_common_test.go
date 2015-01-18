// +build !windows
package atomic

import (
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

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
