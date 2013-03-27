package elliptics

import (
	"sync/atomic"
)

var (
	cSessions                 int64
	cReads, cWrites, cDeletes uint64
)

type Stats struct {
	Sessions               int64
	Reads, Writes, Deletes uint64
}

func GetStats() *Stats {
	return &Stats{
		Sessions: atomic.LoadInt64(&cSessions),
		Reads:    atomic.LoadUint64(&cReads),
		Writes:   atomic.LoadUint64(&cWrites),
		Deletes:  atomic.LoadUint64(&cDeletes),
	}
}
