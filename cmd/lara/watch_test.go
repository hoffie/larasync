package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	. "gopkg.in/check.v1"
)

type WatchTests struct {
	dir     string
	out     *bytes.Buffer
	d       *Dispatcher
	repoDir string
}

var _ = Suite(&WatchTests{})

func (t *WatchTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	t.out = new(bytes.Buffer)
	t.d = &Dispatcher{stderr: t.out}
	t.repoDir = filepath.Join(t.dir, "repo")
	c.Assert(t.d.run([]string{"init", t.repoDir}), Equals, 0)
	err := os.Chdir(t.repoDir)
	c.Assert(err, IsNil)
}

func (t *WatchTests) waitForNIBExistence() {
	nibPath := filepath.Join(t.repoDir, ".lara", "nibs")
	i := 0
	for {
		time.Sleep(10 * time.Millisecond)
		files, err := ioutil.ReadDir(nibPath)
		if (err == nil && len(files) > 0) || i > 100 {
			break
		}
		i++
	}
}

func (t *WatchTests) TestWatchAddition(c *C) {
	file := filepath.Join(t.repoDir, "foo")
	watchStarted := make(chan bool)
	go func() {
		go func() {
			watchStarted <- true
		}()
		c.Assert(t.d.run([]string{"watch"}), Equals, 0)
	}()

	_ = <-watchStarted
	close(watchStarted)
	realContent := []byte("This is test")
	err := ioutil.WriteFile(file, realContent, 0600)
	c.Assert(err, IsNil)

	t.waitForNIBExistence()

	close(watchCancelChannel)

	os.Remove(file)
	c.Assert(t.d.run([]string{"checkout", file}), Equals, 0)

	content, err := ioutil.ReadFile(file)
	c.Assert(err, IsNil)
	c.Assert(content, DeepEquals, realContent)
}
