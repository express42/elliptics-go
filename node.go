// +build linux,cgo

package elliptics

/*
#include "log.h"
*/
import "C"
import "unsafe"

type Node struct {
	config C.struct_dnet_config
	node   *C.struct_dnet_node
}

type ConnectFlags int

const (
	DontDownloadRoutes ConnectFlags = C.DNET_CFG_NO_ROUTE_LIST

	WaitTimeout = 30
)

func NewNode() (node *Node) {
	logger := C.log_create(C.DNET_LOG_DEBUG)

	config := C.struct_dnet_config{
		family: C.AF_INET,
		wait_timeout: WaitTimeout,
	}
	config.log = &logger

	node = &Node{config: config}
	node.node = C.dnet_node_create(&node.config)

	return
}

func (n *Node) Delete() {
	C.dnet_node_destroy(n.node)
	n.node = nil
}

// Calls dnet_add_state.
func (n *Node) Connect(host string, port uint16, flags ...ConnectFlags) (err error) {
  var fl int
	for _, f := range flags {
		fl |= int(f)
	}

	h := C.CString(host)
	defer C.free(unsafe.Pointer(h))
	p := C.int(port)
	family := C.int(C.AF_INET)

	err = Error(C.dnet_add_state(n.node, h, p, family, C.int(fl)))
	return
}
