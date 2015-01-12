package atomic

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// AtomicWriter implements the writer interface and is used to store
// data to the file system in an atomic manner.
type AtomicWriter struct {
	path      string
	tmpPrefix string
	filePerms os.FileMode
	tmpFile   *os.File
}

// NewStandardWriter initializes and returns a new AtomicWriter with a default
// prefix for temporary files.
func NewStandardWriter(path string, perm os.FileMode) (*AtomicWriter, error) {
	return NewWriter(path, ".lara.", perm)
}

// NewWriter initializes and returns a new AtomicWriter.
func NewWriter(path, tmpPrefix string, perm os.FileMode) (*AtomicWriter, error) {
	writer := &AtomicWriter{
		path:      path,
		tmpPrefix: tmpPrefix,
		filePerms: perm,
	}
	err := writer.init()
	return writer, err
}

// getDirFileName splits the directory and the filename
// and returns the data entry.
func (aw *AtomicWriter) getDirFileName() (string, string) {
	return path.Split(aw.path)
}

// tmpFileNamePrefix returns the prefix which should be passed when
// creating a temporary file.
func (aw *AtomicWriter) tmpFileNamePrefix() string {
	_, fileName := aw.getDirFileName()
	return fmt.Sprintf("%s%s", aw.tmpPrefix, fileName)
}

// tmpPath returns the file path to the temporary created file.
func (aw *AtomicWriter) tmpPath() string {
	return aw.tmpFile.Name()
}

// init initializes the AtomicWriter and creates the underlying temporary file.
func (aw *AtomicWriter) init() error {
	dirName, _ := aw.getDirFileName()

	f, err := ioutil.TempFile(dirName, aw.tmpFileNamePrefix())

	if err != nil {
		f.Close()
		return err
	}

	aw.tmpFile = f
	return nil
}

// Write implements the Write method of the Writer interface and adds the data
// to the underlying temporary file.
func (aw *AtomicWriter) Write(p []byte) (n int, err error) {
	return aw.tmpFile.Write(p)
}

// Close implements the Close Method of the Closer. It finalizes the file stream
// and copies it to the final location.
func (aw *AtomicWriter) Close() error {
	err := aw.tmpFile.Close()
	if err != nil {
		return err
	}

	err = os.Chmod(aw.tmpPath(), aw.filePerms)
	if err != nil {
		return err
	}

	// now we know it's fine to (over)write the file;
	// sadly, there is a TOCTU race here, which seems kind of unavoidable
	// (our check is already done, yet the actual rename operation happens just now)
	return os.Rename(aw.tmpPath(), aw.path)
}
