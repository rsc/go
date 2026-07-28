// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import "unsafe"

// This file implements the "external" (real time) variants of the sleep and
// timer APIs. Unlike Sleep, NewTimer, and friends, which schedule against the
// internal ("program time") monotonic clock that stops while the system is
// asleep, these functions schedule against the external ("real time") monotonic
// clock read by ExternalNow, which continues to advance while the system is
// asleep. See go.dev/issue/36141 and the "Monotonic Clocks" section of the
// package documentation.
//
// On platforms that cannot schedule against the external clock (see
// externalTimersReal), each function falls back to its internal counterpart, so
// the external variants are always safe to call.

// runtimeExternalTimerReal reports whether this platform can schedule timers
// against the external clock. When false, the External functions below fall
// back to their internal counterparts.
//
//go:linkname runtimeExternalTimerReal
func runtimeExternalTimerReal() bool

// externalTimersReal reports whether external timers use the external clock on
// this platform (rather than falling back to the internal timer machinery).
var externalTimersReal = runtimeExternalTimerReal()

// whenExternal is like when, but for the external ("real time") clock.
func whenExternal(d Duration) int64 {
	if d <= 0 {
		return runtimeNanoExternal()
	}
	t := runtimeNanoExternal() + int64(d)
	if t < 0 {
		// N.B. runtimeNanoExternal() and d are always positive, so addition
		// (including overflow) will never result in t == 0.
		t = 1<<63 - 1 // math.MaxInt64
	}
	return t
}

// sendTimeExternal is like sendTime, but sends an external-clock time so that
// the value received from an external Timer or Ticker channel carries an
// external monotonic reading.
func sendTimeExternal(c any, seq uintptr, delta int64) {
	select {
	case c.(chan Time) <- ExternalNow().Add(Duration(-delta)):
	default:
	}
}

// ExternalSleep pauses the current goroutine for at least the duration d,
// measured against the external ("real time") clock, which continues to advance
// while the system is asleep. A negative or zero duration causes ExternalSleep
// to return immediately.
//
// On platforms without a distinct external clock, ExternalSleep behaves exactly
// like [Sleep].
func ExternalSleep(d Duration) {
	if d <= 0 {
		return
	}
	if !externalTimersReal {
		Sleep(d)
		return
	}
	<-ExternalNewTimer(d).C
}

// ExternalNewTimer creates a new Timer that will send the current time on its
// channel after at least duration d, measured against the external ("real
// time") clock, which continues to advance while the system is asleep. The
// value sent on the channel carries an external monotonic reading (see
// [ExternalNow]).
//
// On platforms without a distinct external clock, ExternalNewTimer behaves
// exactly like [NewTimer].
func ExternalNewTimer(d Duration) *Timer {
	if !externalTimersReal {
		return NewTimer(d)
	}
	c := make(chan Time, 1)
	t := newTimer(whenExternal(d), 0, sendTimeExternal, c, syncTimer(c), true)
	t.C = c
	t.external = true
	return t
}

// ExternalAfter waits for the duration to elapse on the external ("real time")
// clock and then sends the current time on the returned channel. It is
// equivalent to [ExternalNewTimer](d).C.
func ExternalAfter(d Duration) <-chan Time {
	return ExternalNewTimer(d).C
}

// ExternalAfterFunc waits for the duration to elapse on the external ("real
// time") clock and then calls f in its own goroutine. It returns a [Timer] that
// can be used to cancel the call using its Stop method. The returned Timer's C
// field is not used and will be nil.
//
// On platforms without a distinct external clock, ExternalAfterFunc behaves
// exactly like [AfterFunc].
func ExternalAfterFunc(d Duration, f func()) *Timer {
	if !externalTimersReal {
		return AfterFunc(d, f)
	}
	t := newTimer(whenExternal(d), 0, goFunc, f, nil, true)
	t.external = true
	return t
}

// ExternalNewTicker returns a new [Ticker] like [NewTicker], but its ticks are
// measured against the external ("real time") clock, which continues to advance
// while the system is asleep. The duration d must be greater than zero; if not,
// ExternalNewTicker will panic.
//
// On platforms without a distinct external clock, ExternalNewTicker behaves
// exactly like [NewTicker].
func ExternalNewTicker(d Duration) *Ticker {
	if d <= 0 {
		panic("non-positive interval for ExternalNewTicker")
	}
	if !externalTimersReal {
		return NewTicker(d)
	}
	c := make(chan Time, 1)
	t := (*Ticker)(unsafe.Pointer(newTimer(whenExternal(d), int64(d), sendTimeExternal, c, syncTimer(c), true)))
	t.C = c
	t.external = true
	return t
}

// ExternalTick is a convenience wrapper for [ExternalNewTicker] providing access
// to the ticking channel only. Unlike ExternalNewTicker, ExternalTick will
// return nil if d <= 0.
func ExternalTick(d Duration) <-chan Time {
	if d <= 0 {
		return nil
	}
	return ExternalNewTicker(d).C
}
