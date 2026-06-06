package handler

import (
	"sync"
	"time"
)

type requestLimitState struct {
	windowStart int64
	count       int
	blockedTill int64
}

type requestLimiter struct {
	mu           sync.Mutex
	states       map[string]requestLimitState
	limit        int
	windowSecs   int64
	blockSecs    int64
	lastSweepSec int64
}

func newRequestLimiter(limit, windowSecs, blockSecs int) *requestLimiter {
	return &requestLimiter{
		states:     make(map[string]requestLimitState),
		limit:      limit,
		windowSecs: int64(windowSecs),
		blockSecs:  int64(blockSecs),
	}
}

func (l *requestLimiter) allow(key string, now time.Time) (bool, int64) {
	nowSec := now.Unix()

	l.mu.Lock()
	defer l.mu.Unlock()

	l.sweep(nowSec)

	state := l.states[key]
	if state.blockedTill > nowSec {
		return false, state.blockedTill - nowSec
	}
	if state.windowStart == 0 || nowSec-state.windowStart >= l.windowSecs {
		state.windowStart = nowSec
		state.count = 0
		state.blockedTill = 0
	}

	state.count++
	if state.count > l.limit {
		state.blockedTill = nowSec + l.blockSecs
		l.states[key] = state
		return false, l.blockSecs
	}

	l.states[key] = state
	return true, 0
}

func (l *requestLimiter) sweep(nowSec int64) {
	if nowSec-l.lastSweepSec < 60 {
		return
	}
	l.lastSweepSec = nowSec

	for key, state := range l.states {
		windowExpired := state.windowStart > 0 && nowSec-state.windowStart >= l.windowSecs
		unblocked := state.blockedTill > 0 && state.blockedTill <= nowSec
		if (state.blockedTill == 0 && windowExpired) || unblocked {
			delete(l.states, key)
		}
	}
}
