// +build linux,cgo

package elliptics_test

import (
	. "."
	"fmt"
	"io"
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
	Log = func(level int, message string) {}

	t := time.Now().UnixNano()
	e.key = fmt.Sprintln(t)
	e.data = make([]byte, 256)
	for i := range e.data {
		e.data[i] = byte(i)
	}
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

	data12, err := session.Read(k, 0, 0)
	c.Assert(err, IsNil)
	c.Assert(data12, DeepEquals, e.data)

	err = session.Remove(k)
	c.Assert(err, IsNil)

	err = session.Remove(k)
	c.Assert(err, Equals, syscall.ENOENT)

	_, err = session.Read(k, 0, 0)
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

	data1, err := session.Read(k, 0, 0)
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
	c.Check(err, Equals, ErrZeroWrite)

	err = session.Write(k, []byte{})
	c.Check(err, Equals, ErrZeroWrite)
}

func (e *E) TestReadOffsets(c *C) {
	node := NewNode(timeout)
	defer node.Delete()
	node.Connect("127.0.0.1", 1025)

	session, err := node.NewSession([]uint32{1})
	c.Assert(err, IsNil)
	defer session.Delete()

	k := NewKey(e.key)
	err = session.Write(k, e.data)
	c.Assert(err, IsNil)
	defer session.Remove(k)

	data, err := session.Read(k, 0, 3)
	c.Assert(err, IsNil)
	c.Assert(data, DeepEquals, []byte{0, 1, 2})

	data, err = session.Read(k, 1, 3)
	c.Assert(err, IsNil)
	c.Assert(data, DeepEquals, []byte{1, 2, 3})

	data, err = session.Read(k, 254, 3)
	c.Assert(err, IsNil)
	c.Assert(data, DeepEquals, []byte{254, 255})

	data, err = session.Read(k, 254, 0)
	c.Assert(err, IsNil)
	c.Assert(data, DeepEquals, []byte{254, 255})

	data, err = session.Read(k, 256, 0)
	c.Assert(err, Equals, syscall.E2BIG)
	c.Assert(data, IsNil)

	data, err = session.Read(k, 256, 10)
	c.Assert(err, Equals, syscall.E2BIG)
	c.Assert(data, IsNil)
}

func (e *E) TestReaderSeekRead(c *C) {
	node := NewNode(timeout)
	defer node.Delete()
	node.Connect("127.0.0.1", 1025)

	session, err := node.NewSession([]uint32{1})
	c.Assert(err, IsNil)
	defer session.Delete()

	k := NewKey(e.key)
	err = session.Write(k, e.data)
	c.Assert(err, IsNil)
	defer session.Remove(k)

	r := session.Reader(k)
	buf := make([]byte, 150)

	s, err := r.Seek(50, 0)
	c.Check(err, IsNil)
	c.Check(s, Equals, int64(50))

	s, err = r.Seek(5, 0)
	c.Check(err, IsNil)
	c.Check(s, Equals, int64(5))

	s, err = r.Seek(5, 1)
	c.Check(err, IsNil)
	c.Check(s, Equals, int64(10))

	n, err := r.Read(buf)
	c.Check(err, IsNil)
	c.Check(n, Equals, 150)
	c.Check(buf[0:3], DeepEquals, []byte{10, 11, 12})
	c.Check(buf[147:150], DeepEquals, []byte{157, 158, 159})

	n, err = r.Read(buf)
	c.Check(err, Equals, io.EOF)
	c.Check(n, Equals, 96)
	c.Check(buf[0:3], DeepEquals, []byte{160, 161, 162})
	c.Check(buf[93:96], DeepEquals, []byte{253, 254, 255})

	n, err = r.Read(buf)
	c.Check(err, Equals, io.EOF)
	c.Check(n, Equals, 0)
}
