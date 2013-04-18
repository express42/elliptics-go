// +build linux,cgo

package elliptics

/*
#include "elliptics.h"
*/
import "C"

import (
	"errors"
	"sync/atomic"
	"unsafe"
)

var (
	readOffset = unsafe.Sizeof(C.struct_dnet_io_attr{})

	ErrZeroWrite = errors.New("elliptics-go: attempt to write 0 bytes")
)

type Session struct {
	Node    *Node
	session *C.struct_dnet_session
}

func (n *Node) NewSession(groups []uint32) (s *Session, err error) {
	// TODO runtime.LockOSThread() ?
	s = &Session{n, C.dnet_session_create(n.node)}

	gr := make([]C.int, len(groups))
	for i, g := range groups {
		gr[i] = C.int(g)
	}
	err = Error(C.dnet_session_set_groups(s.session, &gr[0], C.int(len(gr))))

	if err == nil {
		atomic.AddInt64(&cSessions, 1)
	}

	return
}

func (s *Session) Delete() {
	C.dnet_session_destroy(s.session)
	s.session = nil
	atomic.AddInt64(&cSessions, -1)
	// runtime.UnlockOSThread()
}

// Read object from Elliptics by key, offset and size. Session's groups are used (somehow), key's group is ignored.
// data may be nil, otherwise it should be C.free'd by caller.
func (s *Session) read(key *Key, offset uint64, size uint64) (data unsafe.Pointer, dataSize uint64, err error) {
	atomic.AddUint64(&cReads, 1)

	io_attr := C.struct_dnet_io_attr{
		parent: key.id.id, id: key.id.id, _type: key.id._type,
		offset: C.uint64_t(offset), size: C.uint64_t(size),
	}

	var cflags C.uint64_t
	var errp C.int
	data = C.dnet_read_data_wait(s.session, &key.id, &io_attr, cflags, &errp)
	dataSize = uint64(io_attr.size)
	if errp != 0 {
		err = Error(errp)
	}
	return
}

// Read object from Elliptics by key, offset and size. Session's groups are used (somehow), key's group is ignored.
func (s *Session) Read(key *Key, offset uint64, size uint64) (b []byte, err error) {
	// TODO use reflect.SliceHeader and manage data ourselves?
	data, dataSize, err := s.read(key, offset, size)
	if data == nil {
		return
	}
	defer C.free(data)

	b = C.GoBytes(unsafe.Pointer(uintptr(data)+readOffset), C.int(dataSize)-C.int(readOffset))
	return
}

func (s *Session) Streamer(key *Key, size uint64) Streamer {
	return &streamer{session: s, key: key, size: size}
}

// Write object to Elliptics by key. len(b) must be > 0. Session's groups are used (somehow), key's group is ignored.
func (s *Session) Write(key *Key, offset uint64, b []byte) (err error) {
	l := C.uint64_t(len(b))
	if l == 0 {
		return ErrZeroWrite
	}

	atomic.AddUint64(&cWrites, 1)

	io_attr := C.struct_dnet_io_attr{
		parent: key.id.id, id: key.id.id, _type: key.id._type,
		offset: C.uint64_t(offset), size: l,
	}
	io_control := C.struct_dnet_io_control{id: key.id, io: io_attr, data: unsafe.Pointer(&b[0]), fd: -1}

	var result unsafe.Pointer
	size := C.dnet_write_data_wait(s.session, &io_control, &result)
	err = Error(size)
	if err == nil {
		C.free(result)
	}

	return
}

// Remove object from Elliptics by key. Session's groups are ingored, key's group is used.
func (s *Session) Remove(key *Key) (err error) {
	atomic.AddUint64(&cDeletes, 1)

	var cflags C.uint64_t
	var ioflags C.uint64_t
	err = Error(C.dnet_remove_object_now(s.session, &key.id, cflags, ioflags))
	return
}
