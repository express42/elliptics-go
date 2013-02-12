// +build linux,cgo

package elliptics

/*
#include "log.h"
*/
import "C"

import (
	"expvar"
	"log"
	"os"
)

var Levels = map[int]string{
	C.DNET_LOG_DATA:   "DATA",
	C.DNET_LOG_ERROR:  "ERROR",
	C.DNET_LOG_INFO:   "INFO",
	C.DNET_LOG_NOTICE: "NOTICE",
	C.DNET_LOG_DEBUG:  "DEBUG",
}

type LogFunc func(level int, message string)

var logger *log.Logger

var Log LogFunc = func(level int, message string) {
	logger.Printf("[%6s] %s", Levels[level], message)
}

func init() {
	expvar.Publish("elliptics.log_queue", expvar.Func(func() interface{} {
		return C.log_queue_length()
	}))

	logger = log.New(os.Stderr, "[elliptics] ", log.Ldate|log.Lmicroseconds)

	go func() {
		for {
			data := C.log_dequeue()
			Log(int(data.level), C.GoString(data.msg))
			C.log_data_free(data)
		}
	}()
}
