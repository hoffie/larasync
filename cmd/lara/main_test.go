package main

import (
	"strings"

	. "gopkg.in/check.v1"
)

type MainTests struct {
	BaseTests
}

var _ = Suite(&MainTests{})

func (t *MainTests) TestPrompt(c *C) {
	t.in.WriteString("test\n")
	res, err := t.d.promptPassword("Foo: ")
	c.Assert(t.out.Bytes(), DeepEquals, []byte("Foo: "))
	c.Assert(err, IsNil)
	c.Assert(res, DeepEquals, []byte("test"))
}

func (t *MainTests) TestPromptTwice(c *C) {
	t.in.WriteString("test\nfoo\n")
	res, err := t.d.promptPassword("Foo: ")
	c.Assert(t.out.Bytes(), DeepEquals, []byte("Foo: "))
	c.Assert(err, IsNil)
	c.Assert(res, DeepEquals, []byte("test"))

	res, err = t.d.promptPassword("Bar: ")
	c.Assert(t.out.Bytes(), DeepEquals, []byte("Foo: Bar: "))
	c.Assert(err, IsNil)
	c.Assert(res, DeepEquals, []byte("foo"))
}

func (t *MainTests) TestArgHandling(c *C) {
	c.Assert(t.d.run([]string{"help", "sync"}), Equals, 0)
	c.Assert(strings.Index(t.out.String(), "command sync"), Not(Equals), -1)
}

func (t *MainTests) TestNoArg(c *C) {
	c.Assert(t.d.run([]string{}), Equals, 1)
}

func (t *MainTests) TestHelpOnUnknownCommand(c *C) {
	c.Assert(t.d.run([]string{"help", "non-existant-command"}), Equals, 1)
}

func (t *MainTests) TestNonExistentCommand(c *C) {
	c.Assert(t.d.run([]string{"non-existant-command"}), Equals, 1)
}
