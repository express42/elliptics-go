// +build linux,cgo

package elliptics_test

import (
	. "."
	"fmt"
	. "launchpad.net/gocheck"
	"syscall"
	"time"
)

type E struct {
	key  string
	data []byte
}

var _ = Suite(&E{})

var (
	timeout = 5 * time.Second
)

func (e *E) SetUpTest(c *C) {
	t := time.Now().UnixNano()
	e.key = fmt.Sprintln(t)
	e.data = []byte(e.key)
}

func (e *E) TestReadWriteRemove(c *C) {
	node := NewNode(timeout)
	defer node.Delete()
	node.Connect("127.0.0.1", 1025)

	session, err := node.NewSession([]uint32{1})
	c.Assert(err, IsNil)
	defer session.Delete()

	k := NewKey(e.key)
	err = session.Write(k, e.data)
	c.Assert(err, IsNil)

	data12, err := session.Read(k)
	c.Assert(err, IsNil)
	c.Assert(data12, DeepEquals, e.data)

	err = session.Remove(k)
	c.Assert(err, IsNil)

	err = session.Remove(k)
	c.Assert(err, Equals, syscall.ENOENT)

	_, err = session.Read(k)
	c.Assert(err, Equals, syscall.ENOENT)
}

func (e *E) TestBadConnect(c *C) {
	node := NewNode(timeout)
	defer node.Delete()
	node.Connect("127.0.0.1", 1024)

	session, err := node.NewSession([]uint32{1})
	c.Assert(err, IsNil)
	defer session.Delete()

	k := NewKey(e.key)
	err = session.Write(k, e.data)
	c.Assert(err, Equals, syscall.ENOENT)

	data1, err := session.Read(k)
	c.Assert(err, Equals, syscall.ENOENT)
	c.Assert(len(data1), Equals, 0)

	err = session.Remove(k)
	c.Assert(err, IsNil)
}

func (e *E) TestWriteEmpty(c *C) {
	node := NewNode(timeout)
	defer node.Delete()
	node.Connect("127.0.0.1", 1025)

	session, err := node.NewSession([]uint32{1})
	c.Assert(err, IsNil)
	defer session.Delete()

	k := NewKey(e.key)
	var b []byte
	err = session.Write(k, b)
	c.Check(err, Not(Equals), nil)
	c.Check(err, Not(Equals), syscall.ENOENT)

	err = session.Write(k, []byte{})
	c.Check(err, Not(Equals), nil)
	c.Check(err, Not(Equals), syscall.ENOENT)
}
