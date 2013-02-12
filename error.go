// +build linux,cgo

package elliptics

import "C"

import (
	"syscall"
)

// TODO add context to error

func Error(n C.int) (e error) {
	if n < 0 {
		e = syscall.Errno(-n)
	}
	return
}
