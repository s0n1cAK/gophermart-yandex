package job

import (
	"sync/atomic"
	"time"
)

type Throttler struct {
	pausedUntil atomic.Value
}

func NewThrottler() *Throttler {
	t := &Throttler{}
	t.pausedUntil.Store(time.Time{})
	return t
}

func (t *Throttler) IsPaused() bool {
	until := t.pausedUntil.Load().(time.Time)
	return time.Now().Before(until)
}

func (t *Throttler) Pause(duration time.Duration) {
	until := time.Now().Add(duration)
	t.pausedUntil.Store(until)
}
