// +build linux,cgo

package elliptics

/*
#include "log.h"
*/
import "C"

import (
	"time"
	"unsafe"
)

type Node struct {
	config C.struct_dnet_config
	node   *C.struct_dnet_node
}

type ConnectFlags int

const (
	DontDownloadRoutes ConnectFlags = C.DNET_CFG_NO_ROUTE_LIST
)

func NewNode(timeout time.Duration) (node *Node) {
	logger := C.log_create(C.DNET_LOG_DEBUG)

	config := C.struct_dnet_config{family: C.AF_INET, wait_timeout: C.uint(timeout.Seconds())}
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

	err = Error(C.dnet_add_state(n.node, h, C.int(port), C.int(C.AF_INET), C.int(fl)))
	return
}
