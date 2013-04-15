// +build linux,cgo

package elliptics

/*
#include "elliptics.h"
*/
import "C"

import (
	"fmt"
	"io"
	"reflect"
	"sync/atomic"
	"syscall"
	"unsafe"
)

type ReadSeekCloser interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
}

type reader struct {
	session *Session
	key     *Key
	offset  uint64
	size    uint64
}

func (r *reader) Close() (err error) {
	return
}

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
		ret = offset
		atomic.StoreUint64(&r.offset, uint64(ret))
	case 1:
		ret = int64(atomic.AddUint64(&r.offset, uint64(offset)))
	case 2:
		ret = int64(r.size) - offset
		atomic.StoreUint64(&r.offset, uint64(ret))
	default:
		err = fmt.Errorf("elliptics-go: invalid whence value %d", whence)
	}

	if ret < 0 {
		err = fmt.Errorf("elliptics-go: seek resulted in negative offset %d", ret)
		ret = 0
	} else if uint64(ret) > r.size {
		err = fmt.Errorf("elliptics-go: seek resulted in offset %d > size %d", ret, r.size)
		ret = 0
	}
	return
}
