package tracker

import (
	"os"
	"path/filepath"
	"strings"

	. "gopkg.in/check.v1"
)

var _ = Suite(&DatabaseNIBTrackerTests{})

type DatabaseNIBTrackerTests struct {
	dirName      string
	databasePath string
}

func (t *DatabaseNIBTrackerTests) SetUpTest(c *C) {
	t.dirName = c.MkDir()
	t.databasePath = filepath.Join(t.dirName, "test.db")
}

func (t *DatabaseNIBTrackerTests) getTracker() (NIBTracker, error) {
	return NewDatabaseNIBTracker(t.databasePath, t.dirName)
}

func (t *DatabaseNIBTrackerTests) getVerifiedTracker(c *C) NIBTracker {
	tracker, err := t.getTracker()
	c.Assert(err, IsNil)
	return tracker
}

// It should create a database if none has exists yet.
func (t *DatabaseNIBTrackerTests) TestDatabaseCreation(c *C) {
	_, err := t.getTracker()
	c.Assert(err, IsNil)
	_, err = os.Stat(t.databasePath)
	c.Assert(err, IsNil)
}

func (t *DatabaseNIBTrackerTests) TestDatabaseCreationNonDirectory(c *C) {
	err := os.Remove(t.dirName)
	c.Assert(err, IsNil)
	_, err = t.getTracker()
	c.Assert(err, NotNil)
}

func (t *DatabaseNIBTrackerTests) TestAdd(c *C) {
	tracker := t.getVerifiedTracker(c)
	err := tracker.Add("/"+strings.Repeat("5", 4095), "123")
	c.Assert(err, IsNil)
}

func (t *DatabaseNIBTrackerTests) TestDoubleAdd(c *C) {
	tracker := t.getVerifiedTracker(c)
	err := tracker.Add("/test", "123")
	c.Assert(err, IsNil)
	err = tracker.Add("/test", "123")
	c.Assert(err, IsNil)

	responses, err := tracker.SearchPrefix("/test")
	c.Assert(err, IsNil)
	c.Assert(len(responses), Equals, 1)
}

func (t *DatabaseNIBTrackerTests) TestAddOverlength(c *C) {
	tracker := t.getVerifiedTracker(c)
	err := tracker.Add("/"+strings.Repeat("5", 8000), "123")
	c.Assert(err, NotNil)
}

func (t *DatabaseNIBTrackerTests) TestAddExisting(c *C) {
	tracker := t.getVerifiedTracker(c)
	err := tracker.Add("/test", "123")
	c.Assert(err, IsNil)
	err = tracker.Add("/test", "456")
	c.Assert(err, IsNil)
	resp, err := tracker.Get("/test")
	c.Assert(err, IsNil)
	c.Assert(resp.NIBID, Equals, "456")
}

func (t *DatabaseNIBTrackerTests) TestGet(c *C) {
	tracker := t.getVerifiedTracker(c)
	err := tracker.Add("/test", "123")
	c.Assert(err, IsNil)
	resp, err := tracker.Get("/test")
	c.Assert(err, IsNil)
	c.Assert(resp.NIBID, Equals, "123")
}

func (t *DatabaseNIBTrackerTests) TestGetNotExists(c *C) {
	tracker := t.getVerifiedTracker(c)
	resp, err := tracker.Get("/test")
	c.Assert(err, NotNil)
	c.Assert(resp, IsNil)
}

func (t *DatabaseNIBTrackerTests) TestSearchPrefixEmpty(c *C) {
	tracker := t.getVerifiedTracker(c)
	resp, err := tracker.SearchPrefix("/test")
	c.Assert(err, IsNil)
	c.Assert(len(resp), Equals, 0)
}

func (t *DatabaseNIBTrackerTests) TestSearchPrefix(c *C) {
	tracker := t.getVerifiedTracker(c)
	add := func(path string, nibID string) {
		err := tracker.Add(path, nibID)
		c.Assert(err, IsNil)
	}

	add("/test", "123")
	add("/test/sub", "234")
	add("/test2/sub", "456")

	resp, err := tracker.SearchPrefix("/test")
	c.Assert(err, IsNil)
	c.Assert(len(resp), Equals, 2)

	for _, entry := range resp {
		c.Assert(entry.Path, Not(Equals), "/test2/sub")
	}
}

func (t *DatabaseNIBTrackerTests) TestRemove(c *C) {
	tracker := t.getVerifiedTracker(c)
	add := func(path string, nibID string) {
		err := tracker.Add(path, nibID)
		c.Assert(err, IsNil)
	}

	add("/test", "123")
	err := tracker.Remove("/test")
	c.Assert(err, IsNil)

	_, err = tracker.Get("/test")
	c.Assert(err, NotNil)
}

func (t *DatabaseNIBTrackerTests) TestRemoveNotExisting(c *C) {
	tracker := t.getVerifiedTracker(c)
	err := tracker.Remove("/test")
	c.Assert(err, NotNil)
}
