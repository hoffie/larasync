package repository

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hoffie/larasync/helpers"
	"github.com/hoffie/larasync/repository/nib"

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

// It should throw an error if a content id references in the nib
// is not existing yet.
func (t *RepositoryAddItemTests) TestAddNibContentObjectIDsMissing(c *C) {
	n := &nib.NIB{
		ID: "asdf",
		Revisions: []*nib.Revision{
			&nib.Revision{
				MetadataID: "not-existing",
				ContentIDs: []string{},
			},
		},
	}
	r := New(t.dir)
	err := r.CreateManagementDir()
	c.Assert(err, IsNil)

	err = r.nibStore.Add(n)
	c.Assert(err, IsNil)
	data, err := r.nibStore.GetBytes(n.ID)
	c.Assert(err, IsNil)

	buffer := bytes.NewBuffer(data)

	err = r.AddNIBContent(buffer)
	c.Assert(IsNIBContentMissing(err), Equals, true)
}

func (t *RepositoryAddItemTests) TestAddNIBContentConflict(c *C) {
	n := &nib.NIB{
		ID: "asdf",
		Revisions: []*nib.Revision{
			&nib.Revision{
				MetadataID: "metadata123",
				ContentIDs: []string{},
			},
		},
	}
	r := New(t.dir)
	err := r.CreateManagementDir()
	c.Assert(err, IsNil)

	err = r.AddObject("metadata123", bytes.NewBufferString("x"))
	c.Assert(err, IsNil)

	err = r.AddObject("metadata456", bytes.NewBufferString("y"))
	c.Assert(err, IsNil)

	err = r.nibStore.Add(n)
	c.Assert(err, IsNil)
	data1, err := r.nibStore.GetBytes(n.ID)
	c.Assert(err, IsNil)

	n.AppendRevision(&nib.Revision{MetadataID: "metadata456"})

	err = r.nibStore.Add(n)
	c.Assert(err, IsNil)
	data2, err := r.nibStore.GetBytes(n.ID)
	c.Assert(err, IsNil)

	buffer1 := bytes.NewBuffer(data1)
	buffer2 := bytes.NewBuffer(data2)
	err = r.AddNIBContent(buffer2)
	c.Assert(err, IsNil)
	err = r.AddNIBContent(buffer1)
	c.Assert(err, Equals, ErrNIBConflict)
}

func (t *RepositoryAddItemTests) TestAddNIBContentObjNotExisting(c *C) {
	n := &nib.NIB{
		ID: "asdf",
		Revisions: []*nib.Revision{
			&nib.Revision{
				MetadataID: "metadata1",
				ContentIDs: []string{"content1", "content2", "content3"},
			},
		},
	}
	r := New(t.dir)
	err := r.CreateManagementDir()
	c.Assert(err, IsNil)

	err = r.AddObject("content2", bytes.NewBufferString("x"))
	c.Assert(err, IsNil)

	err = r.nibStore.Add(n)
	c.Assert(err, IsNil)

	reader, err := r.nibStore.getReader(n.ID)
	c.Assert(err, IsNil)

	err = r.AddNIBContent(reader)
	c.Assert(err, NotNil)

	c.Assert(IsNIBContentMissing(err), Equals, true)
	nibContentMissing := err.(*NIBContentMissing)
	missingIDs := nibContentMissing.MissingContentIDs()

	c.Assert(helpers.SliceContainsString(missingIDs, "metadata1"), Equals, true)
	c.Assert(helpers.SliceContainsString(missingIDs, "content1"), Equals, true)
	c.Assert(helpers.SliceContainsString(missingIDs, "content2"), Equals, false)
	c.Assert(helpers.SliceContainsString(missingIDs, "content3"), Equals, true)
}
