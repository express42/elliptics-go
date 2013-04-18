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

type Streamer interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Writer
	io.WriterAt
	io.Seeker
}

type streamer struct {
	session *Session
	key     *Key
	offset  uint64
	size    uint64
}

func (s *streamer) Close() (err error) {
	return
}

func (s *streamer) Read(b []byte) (n int, err error) {
	offset := atomic.LoadUint64(&s.offset)
	n, err = s.ReadAt(b, int64(offset))
	atomic.AddUint64(&s.offset, uint64(n))
	return
}

func (s *streamer) ReadAt(b []byte, offset int64) (n int, err error) {
	data, dataSize, err := s.session.read(s.key, uint64(offset), uint64(len(b)))
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

func (s *streamer) Write(b []byte) (n int, err error) {
	offset := atomic.LoadUint64(&s.offset)
	n, err = s.WriteAt(b, int64(offset))
	atomic.AddUint64(&s.offset, uint64(n))
	return
}

func (s *streamer) WriteAt(b []byte, offset int64) (n int, err error) {
	err = s.session.Write(s.key, uint64(offset), b)
	if err == nil {
		n = len(b)
	}
	return
}

func (s *streamer) Seek(offset int64, whence int) (ret int64, err error) {
	switch whence {
	case 0:
		ret = offset
		atomic.StoreUint64(&s.offset, uint64(ret))
	case 1:
		ret = int64(atomic.AddUint64(&s.offset, uint64(offset)))
	case 2:
		ret = int64(s.size) + offset
		atomic.StoreUint64(&s.offset, uint64(ret))
	default:
		err = fmt.Errorf("elliptics-go: invalid whence value %d", whence)
	}

	if ret < 0 {
		err = fmt.Errorf("elliptics-go: seek resulted in negative offset %d", ret)
		ret = 0
	} else if uint64(ret) > s.size {
		err = fmt.Errorf("elliptics-go: seek resulted in offset %d > size %d", ret, s.size)
		ret = 0
	}
	return
}
