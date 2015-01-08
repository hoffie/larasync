package lock

import (
	"time"

	. "gopkg.in/check.v1"
)

type ProcessManagerTests struct {
	manager        *ProcessManager
	repositoryPath string
}

var _ = Suite(&ProcessManagerTests{})

func (t *ProcessManagerTests) SetUpTest(c *C) {
	t.manager = newProcessManager()
	t.repositoryPath = "/repository/path"
}

func (t *ProcessManagerTests) TestGetRole(c *C) {
	locker := t.manager.Get(t.repositoryPath, "lock")
	locker.Lock()
	locker.Unlock()
}

func (t *ProcessManagerTests) TestGetSame(c *C) {
	locker := t.manager.Get(t.repositoryPath, "lock")
	c.Assert(locker, Equals, t.manager.Get(t.repositoryPath, "lock"))
}

func (t *ProcessManagerTests) TestReset(c *C) {
	locker := t.manager.Get(t.repositoryPath, "lock")
	t.manager.reset()
	c.Assert(locker, Not(Equals), t.manager.Get(t.repositoryPath, "lock"))
}

func (t *ProcessManagerTests) TestDifferentRepositories(c *C) {
	locker := t.manager.Get(t.repositoryPath, "lock")
	locker2 := t.manager.Get("/repository/otherpath", "lock")
	c.Assert(locker, Not(Equals), locker2)
}

func (t *ProcessManagerTests) TestDifferentRoles(c *C) {
	locker := t.manager.Get(t.repositoryPath, "lock")
	locker2 := t.manager.Get(t.repositoryPath, "other_role")
	c.Assert(locker, Not(Equals), locker2)
}

func (t *ProcessManagerTests) TestLockInteraction(c *C) {
	locker := t.manager.Get(t.repositoryPath, "lock")
	locker.Lock()
	var (
		unlockTime time.Time
		relockTime time.Time
	)
	go func() {
		sameLocker := t.manager.Get(t.repositoryPath, "lock")
		sameLocker.Lock()
		relockTime = time.Now()
		sameLocker.Unlock()
	}()

	time.Sleep(time.Duration(10 * time.Millisecond))
	unlockTime = time.Now()
	locker.Unlock()
	time.Sleep(time.Duration(10 * time.Millisecond))
	locker.Lock()
	c.Assert(relockTime, NotNil)
	delta := relockTime.Sub(unlockTime)
	c.Assert(
		delta.Nanoseconds() > 0, Equals, true,
		Commentf("Seconds passed %.2f must be positive", delta.Seconds()))
	locker.Unlock()
}
