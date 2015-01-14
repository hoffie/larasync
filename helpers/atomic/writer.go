package atomic

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
)

// ReadCloserAbort provides an additional Abort method to the
// io.ReadCloser interface which tries to not do any modifications
// on the original file.
type ReadCloserAbort interface {
	io.ReadCloser
	// Abort ensures that the final file does not get
	// written.
	Abort()
}

// Writer implements the writer interface and is used to store
// data to the file system in an atomic manner.
type Writer struct {
	path      string
	tmpPrefix string
	filePerms os.FileMode
	tmpFile   *os.File
	aborted   bool
}

// NewStandardWriter initializes and returns a new AtomicWriter with a default
// prefix for temporary files.
func NewStandardWriter(path string, perm os.FileMode) (*Writer, error) {
	return NewWriter(path, ".lara.", perm)
}

// NewWriter initializes and returns a new AtomicWriter.
func NewWriter(path, tmpPrefix string, perm os.FileMode) (*Writer, error) {
	writer := &Writer{
		path:      path,
		tmpPrefix: tmpPrefix,
		filePerms: perm,
		aborted:   false,
	}
	err := writer.init()
	return writer, err
}

// getDirFileName splits the directory and the filename
// and returns the data entry.
func (aw *Writer) getDirFileName() (string, string) {
	return path.Split(aw.path)
}

// tmpFileNamePrefix returns the prefix which should be passed when
// creating a temporary file.
func (aw *Writer) tmpFileNamePrefix() string {
	_, fileName := aw.getDirFileName()
	return fmt.Sprintf("%s%s", aw.tmpPrefix, fileName)
}

// tmpPath returns the file path to the temporary created file.
func (aw *Writer) tmpPath() string {
	return aw.tmpFile.Name()
}

// init initializes the AtomicWriter and creates the underlying temporary file.
func (aw *Writer) init() error {
	dirName, _ := aw.getDirFileName()

	f, err := ioutil.TempFile(dirName, aw.tmpFileNamePrefix())

	err = f.Chmod(aw.filePerms)
	if err != nil {
		f.Close()
		return err
	}

	if err != nil {
		f.Close()
		return err
	}

	aw.tmpFile = f
	return nil
}

// Write implements the Write method of the Writer interface and adds the data
// to the underlying temporary file.
func (aw *Writer) Write(p []byte) (n int, err error) {
	return aw.tmpFile.Write(p)
}

// Abort cancels the atomic write. The file will not be written into its
// final floats.
func (aw *Writer) Abort() {
	aw.aborted = true
}

// Close implements the Close Method of the Closer. It finalizes the file stream
// and copies it to the final location.
func (aw *Writer) Close() error {
	err := aw.tmpFile.Close()
	if err != nil {
		return err
	}
	if aw.aborted {
		os.Remove(aw.tmpFile.Name())
		return nil
	}

	// now we know it's fine to (over)write the file;
	// sadly, there is a TOCTU race here, which seems kind of unavoidable
	// (our check is already done, yet the actual rename operation happens just now)
	return os.Rename(aw.tmpPath(), aw.path)
}
