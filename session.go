// +build linux,cgo

package elliptics

/*
#include "elliptics.h"
*/
import "C"

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

var (
	readOffset = unsafe.Sizeof(C.struct_dnet_io_attr{})

	cSessions                 int64
	cReads, cWrites, cDeletes uint64
)

type Stats struct {
	Sessions               int64
	Reads, Writes, Deletes uint64
}

func (n *Node) Stats() *Stats {
	return &Stats{
		Sessions: atomic.LoadInt64(&cSessions),
		Reads:    atomic.LoadUint64(&cReads),
		Writes:   atomic.LoadUint64(&cWrites),
		Deletes:  atomic.LoadUint64(&cDeletes),
	}
}

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

// Read object from Elliptics by key. Session's groups are used (somehow), key's group is ignored.
func (s *Session) Read(key *Key) (b []byte, err error) {
	atomic.AddUint64(&cReads, 1)

	io_attr := C.struct_dnet_io_attr{parent: key.id.id, id: key.id.id, _type: key.id._type}

	var cflags C.uint64_t
	var errp C.int
	data := C.dnet_read_data_wait(s.session, &key.id, &io_attr, cflags, &errp)
	if data != nil {
		defer C.free(data)
	}
	if errp != 0 {
		err = Error(errp)
		return
	}

	// TODO use reflect.SliceHeader and manage data ourselves?
	b = C.GoBytes(unsafe.Pointer(uintptr(data)+readOffset), C.int(io_attr.size-C.uint64_t(readOffset)))
	return
}

// Write object to Elliptics by key. len(b) must be > 0. Session's groups are used (somehow), key's group is ignored.
func (s *Session) Write(key *Key, b []byte) (err error) {
	l := C.uint64_t(len(b))
	if l == 0 {
		return fmt.Errorf("Attempt to write 0 bytes to %s", key)
	}

	atomic.AddUint64(&cWrites, 1)

	io_attr := C.struct_dnet_io_attr{parent: key.id.id, id: key.id.id, _type: key.id._type, size: l}
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
