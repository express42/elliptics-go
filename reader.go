// +build linux,cgo

package elliptics

/*
#include "elliptics.h"
*/
import "C"

import (
	"io"
	"reflect"
	"sync/atomic"
	"syscall"
	"unsafe"
)

type reader struct {
	session *Session
	key     *Key
	offset  uint64
}

// check interface
var (
	_ io.Reader = &reader{}
)

func (r *reader) Read(b []byte) (n int, err error) {
	offset := atomic.LoadUint64(&r.offset)

	data, dataSize, err := r.session.read(r.key, offset, uint64(len(b)))
	if err == syscall.E2BIG {
		err = io.EOF
	}
	if data == nil {
		return
	}
	defer C.free(data)

	l := int(dataSize - uint64(readOffset))
	header := reflect.SliceHeader{Data: uintptr(data) + readOffset, Len: l, Cap: l}
	n = copy(b, *(*[]byte)(unsafe.Pointer(&header)))
	if n < len(b) {
		err = io.EOF
	}

	atomic.AddUint64(&r.offset, uint64(n))
	return
}
