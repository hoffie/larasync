package lock

import (
	"runtime"
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
	sleepTime := 10 * time.Millisecond

	var (
		unlockTime time.Time
		relockTime time.Time
	)
	go func() {
		sameLocker := t.manager.Get(t.repositoryPath, "lock")
		sameLocker.Lock()
		// On windows the internal Clock has some issues with very
		// short time spans. That is why we are waiting here another
		// amount of time before writing down the lock time.
		if runtime.GOOS == "windows" {
			time.Sleep(200 * time.Millisecond)
		}
		relockTime = time.Now()
		sameLocker.Unlock()
	}()

	time.Sleep(sleepTime)
	unlockTime = time.Now()
	locker.Unlock()
	time.Sleep(sleepTime)
	locker.Lock()
	c.Assert(relockTime, NotNil)
	delta := relockTime.Sub(unlockTime)
	c.Assert(
		delta.Nanoseconds() > 0, Equals, true,
		Commentf("Seconds passed %.2f must be positive", delta.Seconds()))
	locker.Unlock()
}
