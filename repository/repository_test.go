package repository

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

type RepositoryTests struct {
	dir string
}

var _ = Suite(&RepositoryTests{})

func (t *RepositoryTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
}

func (t *RepositoryTests) TestGetRepoRelativePath(c *C) {
	r := New(filepath.Join(t.dir, "foo"))
	err := r.Create()
	c.Assert(err, IsNil)
	in := filepath.Join(t.dir, "foo", "test", "bar")
	out, err := r.getRepoRelativePath(in)
	c.Assert(err, IsNil)
	c.Assert(out, Equals, filepath.Join("test", "bar"))
}

func (t *RepositoryTests) TestGetRepoRelativePathFail(c *C) {
	r := New(filepath.Join(t.dir, "foo"))
	err := r.Create()
	c.Assert(err, IsNil)
	in := t.dir
	out, err := r.getRepoRelativePath(in)
	c.Assert(err, NotNil)
	c.Assert(out, Equals, "")
}

func (t *RepositoryTests) TestCreateManagementDir(c *C) {
	r := New(t.dir)
	err := r.CreateManagementDir()
	c.Assert(err, IsNil)

	s, err := os.Stat(filepath.Join(t.dir, ".lara"))
	c.Assert(err, IsNil)
	c.Assert(s.IsDir(), Equals, true)

	s, err = os.Stat(filepath.Join(t.dir, ".lara", "objects"))
	c.Assert(err, IsNil)
	c.Assert(s.IsDir(), Equals, true)

	s, err = os.Stat(filepath.Join(t.dir, ".lara", "nibs"))
	c.Assert(err, IsNil)
	c.Assert(s.IsDir(), Equals, true)

}

func (t *RepositoryTests) TestAddObject(c *C) {
	r := New(t.dir)
	err := r.CreateManagementDir()
	c.Assert(err, IsNil)
	objectID := "1234567890"
	objectReader := bytes.NewReader([]byte("Test data"))

	err = r.AddObject(objectID, objectReader)
	c.Assert(err, IsNil)
}

func (t *RepositoryTests) TestGetObject(c *C) {
	r := New(t.dir)
	err := r.CreateManagementDir()
	c.Assert(err, IsNil)
	objectID := "1234567890"
	objectData := []byte("Test data")
	objectReader := bytes.NewReader(objectData)

	r.AddObject(objectID, objectReader)

	reader, err := r.GetObjectData(objectID)
	c.Assert(err, IsNil)

	data, err := ioutil.ReadAll(reader)

	err = reader.Close()
	c.Assert(err, IsNil)

	c.Assert(objectData, DeepEquals, data)
}

func (t *RepositoryTests) TestPathToNIBID(c *C) {
	r := New(t.dir)
	err := r.CreateManagementDir()
	c.Assert(err, IsNil)

	err = r.keys.CreateHashingKey()
	c.Assert(err, IsNil)

	path := "foo/bar.txt"
	id, err := r.pathToNIBID(path)
	c.Assert(err, IsNil)
	c.Assert(id, Not(Equals), "")

	id2, err := r.pathToNIBID(path)
	c.Assert(err, IsNil)
	c.Assert(id2, Equals, id)
}

func (t *RepositoryTests) TestGetFileChunkIDs(c *C) {
	r := New(t.dir)
	err := r.CreateManagementDir()
	c.Assert(err, IsNil)

	err = r.keys.CreateHashingKey()
	c.Assert(err, IsNil)

	path := filepath.Join(t.dir, "foo.txt")
	err = ioutil.WriteFile(path, []byte("test"), 0600)
	c.Assert(err, IsNil)

	ids, err := r.getFileChunkIDs(path)
	c.Assert(err, IsNil)
	c.Assert(len(ids), Equals, 1)
	c.Assert(len(ids[0]), Not(Equals), 0)

	ids2, err := r.getFileChunkIDs(path)
	c.Assert(err, IsNil)
	c.Assert(ids2, DeepEquals, ids)
}

func (t *RepositoryTests) TestStateConfig(c *C) {
	exp := "example.org:14124"

	r := New(t.dir)
	err := r.CreateManagementDir()
	c.Assert(err, IsNil)

	sc, err := r.StateConfig()
	c.Assert(err, IsNil)
	sc.DefaultServer = exp
	sc.Save()

	r2 := New(t.dir)
	sc2, err := r2.StateConfig()
	c.Assert(err, IsNil)
	c.Assert(sc2.DefaultServer, Equals, exp)
}

func numFilesInDir(path string) (int, error) {
	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return 0, err
	}
	return len(entries), nil
}
