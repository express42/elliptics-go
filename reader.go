// +build linux,cgo

package elliptics

/*
#include "elliptics.h"
*/
import "C"

import (
	"errors"
	"io"
	"reflect"
	"sync/atomic"
	"syscall"
	"unsafe"
)

type ReadAtSeeker interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

type reader struct {
	session *Session
	key     *Key
	offset  uint64
}

// check interface
var (
	_ io.Reader   = &reader{}
	_ io.ReaderAt = &reader{}
	_ io.Seeker   = &reader{}
)

func (r *reader) Read(b []byte) (n int, err error) {
	offset := atomic.LoadUint64(&r.offset)
	n, err = r.ReadAt(b, int64(offset))
	atomic.AddUint64(&r.offset, uint64(n))
	return
}

func (r *reader) ReadAt(b []byte, offset int64) (n int, err error) {
	data, dataSize, err := r.session.read(r.key, uint64(offset), uint64(len(b)))
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

	return
}

func (r *reader) Seek(offset int64, whence int) (ret int64, err error) {
	switch whence {
	case 0:
		atomic.StoreUint64(&r.offset, uint64(offset))
		ret = offset
	case 1:
		ret = int64(atomic.AddUint64(&r.offset, uint64(offset)))
	case 2:
		err = errors.New("whence 2 is not supported yet")
	}
	return
}
