package lock

import (
	. "gopkg.in/check.v1"
)

type LockTests struct {
}

var _ = Suite(&LockTests{})

func (t *LockTests) TestManagerReturn(c *C) {
	c.Assert(CurrentManager(), NotNil)
}
