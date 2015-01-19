package content

import (
	. "gopkg.in/check.v1"
)

type UUIDStorageTests struct {
	storage *UUIDStorage
	dir     string
}

var _ = Suite(&UUIDStorageTests{})

func (t *UUIDStorageTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	s := NewFileStorage(t.dir)
	t.storage = NewUUIDStorage(s)
}

func (t *UUIDStorageTests) TestFind(c *C) {
	uuid, err := t.storage.FindFreeUUID()
	c.Assert(err, IsNil)
	c.Assert(len(uuid) > 0, Equals, true)
}

func (t *UUIDStorageTests) TestHas(c *C) {
	res := t.storage.HasUUID([]byte("asdf"))
	c.Assert(res, Equals, false)
}
