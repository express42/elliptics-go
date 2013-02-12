// +build linux,cgo

package elliptics_test

import (
	. "."
	"fmt"
	. "launchpad.net/gocheck"
	"syscall"
)

type Err struct{}

var _ = Suite(&Err{})

func (*Err) TestError(c *C) {
	c.Check(Error(0), IsNil)
	c.Check(Error(1), IsNil)
	c.Check(Error(-42), Equals, syscall.ENOMSG)
	c.Check(Error(-42), ErrorMatches, `no message of desired type`)
	c.Check(fmt.Sprint(Error(-42)), Equals, `no message of desired type`)
	c.Check(fmt.Sprintf("%s", Error(-42)), Equals, `no message of desired type`)
	c.Check(fmt.Sprintf("%v", Error(-42)), Equals, `no message of desired type`)
	c.Check(fmt.Sprintf("%#v", Error(-42)), Equals, `0x2a`)
}
