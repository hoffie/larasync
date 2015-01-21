package server

import (
	"fmt"

	. "gopkg.in/check.v1"
)

type NIBItemTest struct {
	NIBTest
}

func getNIBItemTest() NIBItemTest {
	return NIBItemTest{getNIBTest()}
}

func (t *NIBItemTest) SetUpTest(c *C) {
	t.NIBTest.SetUpTest(c)

	origGetURL := t.getURL
	t.getURL = func() string {
		return fmt.Sprintf(
			"%s/%s",
			origGetURL(),
			t.nibID,
		)
	}
	t.req = t.requestEmptyBody(c)
}
