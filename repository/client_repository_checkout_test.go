package repository

import (
    "os"
    "io/ioutil"
    "path/filepath"

    "github.com/hoffie/larasync/repository/nib"
    . "gopkg.in/check.v1"
)

var _ = Suite(&RepositoryCheckoutTests{})

type RepositoryCheckoutTests struct {
    dir string
    r   *ClientRepository
    testData []byte
    fullPath string
}

func (t *RepositoryCheckoutTests) SetUpTest(c *C) {
    t.dir = c.MkDir()
    t.r = NewClient(t.dir)
    err := t.r.CreateManagementDir()
    c.Assert(err, IsNil)
    err = t.r.keys.CreateSigningKey()
    c.Assert(err, IsNil)

    err = t.r.keys.CreateEncryptionKey()
    c.Assert(err, IsNil)

    err = t.r.keys.CreateHashingKey()
    c.Assert(err, IsNil)

    t.testData = []byte("foo")
    t.fullPath = filepath.Join(t.dir, "foo.txt")
}

// TestRemoveFile verifies if empty content and metadata ids
// are being removed.
func (t *RepositoryCheckoutTests) TestRemoveFile(c *C) {
    t.addTestFile(c)
    fullPath := t.fullPath

    metadataID, err := t.r.writeMetadata(fullPath)
    c.Assert(err, IsNil)

    n := &nib.NIB{
        ID: "",
        Revisions: []*nib.Revision{
            &nib.Revision{
                MetadataID: metadataID,
                ContentIDs: []string{},
            },
        },
    }

    err = t.r.checkoutNIB(n)
    c.Assert(err, IsNil)

    _, err = os.Stat(fullPath)
    c.Assert(err, NotNil)
    c.Assert(os.IsNotExist(err), Equals, true)
}

func (t *RepositoryCheckoutTests) addTestFile(c *C) {
    err := ioutil.WriteFile(t.fullPath, t.testData, 0600)
    c.Assert(err, IsNil)
}

func (t *RepositoryCheckoutTests) TestAddFile(c *C) {
    t.addTestFile(c)
    fullPath := t.fullPath
    data := t.testData

    err := t.r.AddItem(fullPath)
    c.Assert(err, IsNil)

    err = os.Remove(fullPath)
    c.Assert(err, IsNil)

    err = t.r.CheckoutPath(fullPath)
    c.Assert(err, IsNil)

    readData, err := ioutil.ReadFile(fullPath)
    c.Assert(err, IsNil)

    c.Assert(data, DeepEquals, readData)
}

func (t *RepositoryCheckoutTests) TestModifyFileWorkdirConflict(c *C) {
    t.addTestFile(c)

    err := t.r.AddItem(t.fullPath)
    c.Assert(err, IsNil)

    err = ioutil.WriteFile(t.fullPath, []byte("overwrittenstuff"), 0600)
    c.Assert(err, IsNil)

    err = t.r.CheckoutPath(t.fullPath)
    c.Assert(err, Equals, ErrWorkDirConflict)
}
