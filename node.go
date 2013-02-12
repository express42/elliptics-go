// +build linux,cgo

package elliptics

/*
#include "log.h"
*/
import "C"

import (
	"fmt"
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

func NewNode() (node *Node) {
	logger := C.log_create(C.DNET_LOG_DEBUG)

	config := C.struct_dnet_config{sock_type: C.SOCK_STREAM, proto: C.IPPROTO_TCP, family: C.AF_INET}
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
	config := C.struct_dnet_config{sock_type: C.SOCK_STREAM, proto: C.IPPROTO_TCP, family: C.AF_INET}
	for _, f := range flags {
		config.flags |= C.int(f)
	}

	h := []byte(host)
	p := []byte(fmt.Sprintf("%d", port))
	copy(config.addr[:], *(*[]C.char)(unsafe.Pointer(&h)))
	copy(config.port[:], *(*[]C.char)(unsafe.Pointer(&p)))
	err = Error(C.dnet_add_state(n.node, &config))
	return
}
