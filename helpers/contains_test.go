package helpers

import (
	"fmt"

	. "gopkg.in/check.v1"
)

type ContainsTests struct {
	testStringSlice []string
}

var _ = Suite(ContainsTests{})

func (t *ContainsTests) SetUpSuite(c *C) {
	t.testStringSlice = []string{"a", "b", "c"}

}

func (t *ContainsTests) TestStringPositive(c *C) {
	for _, str := range t.testStringSlice {
		c.Assert(
			SliceContainsString(t.testStringSlice, str),
			Equals,
			true,
			fmt.Sprintf("Missing string %s", str),
		)
	}
}

func (t *ContainsTests) TestStringNegative(c *C) {
	c.Assert(
		SliceContainsString(t.testStringSlice, "d"),
		Equals,
		false,
	)
}
