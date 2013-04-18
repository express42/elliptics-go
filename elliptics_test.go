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
	err = session.Write(k, 0, e.data)
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
	err = session.Write(k, 0, e.data)
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
	err = session.Write(k, 0, b)
	c.Check(err, Equals, ErrZeroWrite)

	err = session.Write(k, 0, []byte{})
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
	err = session.Write(k, 0, e.data)
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

func (e *E) TestStreamer(c *C) {
	node := NewNode(timeout)
	defer node.Delete()
	node.Connect("127.0.0.1", 1025)

	session, err := node.NewSession([]uint32{1})
	c.Assert(err, IsNil)
	defer session.Delete()

	k := NewKey(e.key)
	defer session.Remove(k)

	s := session.Streamer(k, 12)
	defer func() {
		c.Check(s.Close(), IsNil)
	}()

	p, err := s.Seek(3, 0)
	c.Check(err, IsNil)
	c.Check(p, Equals, int64(3))

	n, err := s.Write(e.data[3:4])
	c.Check(err, IsNil)
	c.Check(n, Equals, 1)

	n, err = s.WriteAt(e.data[9:12], 9)
	c.Check(err, IsNil)
	c.Check(n, Equals, 3)

	n, err = s.Write(e.data[4:6])
	c.Check(err, IsNil)
	c.Check(n, Equals, 2)

	p, err = s.Seek(0, 0)
	c.Check(err, IsNil)
	c.Check(p, Equals, int64(0))

	buf := make([]byte, 12)
	n, err = s.Read(buf)
	c.Check(err, IsNil)
	c.Check(n, Equals, 12)
	c.Check(buf[:n], DeepEquals, []byte{0x0, 0x0, 0x0, 0x3, 0x4, 0x5, 0x0, 0x0, 0x0, 0x9, 0xa, 0xb})

	p, err = s.Seek(-6, 2)
	c.Check(err, IsNil)
	c.Check(p, Equals, int64(6))

	n, err = s.Read(buf)
	c.Check(err, Equals, io.EOF)
	c.Check(n, Equals, 6)
	c.Check(buf[:n], DeepEquals, []byte{0x0, 0x0, 0x0, 0x9, 0xa, 0xb})

	n, err = s.ReadAt(buf, 9)
	c.Check(err, Equals, io.EOF)
	c.Check(n, Equals, 3)
	c.Check(buf[:n], DeepEquals, []byte{0x9, 0xa, 0xb})

	n, err = s.ReadAt(buf, 100500)
	c.Check(err, Equals, io.EOF)
	c.Check(n, Equals, 0)

	n, err = s.Read(buf)
	c.Check(err, Equals, io.EOF)
	c.Check(n, Equals, 0)
}
