// +build linux,cgo

package elliptics

/*
#include "elliptics.h"
*/
import "C"

import (
	"crypto/sha512"
	"fmt"
	"unsafe"
)

type Key struct {
	id C.struct_dnet_id
}

func (k Key) Short() string {
	return fmt.Sprintf("%012x", k.id.id[:6])
}

func (k Key) String() string {
	return fmt.Sprintf("%0128x", k.id.id)
}

func (k Key) GoString() string {
	return fmt.Sprintf("%0128x (%#v)", k.id.id, k.id)
}

var _ fmt.Stringer = Key{}
var _ fmt.GoStringer = Key{}

// Make new Elliptics key (dnet_id) by hashing name with SHA-512.
// dnet_id's group_id and type are deprecated and set to 0.
func NewKey(name string) (key *Key) {
	h := sha512.New()
	h.Write([]byte(name))
	sum := h.Sum(nil)

	key = &Key{}
	copy(key.id.id[:], *(*[]C.uint8_t)(unsafe.Pointer(&sum)))
	return
}
