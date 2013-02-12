// +build linux,cgo

package elliptics

// TODO: #cgo pkg-config: elliptics-client

/*
#cgo LDFLAGS: -lelliptics_client
*/
import "C"

type CheckSum [64]byte
