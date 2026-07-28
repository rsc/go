// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows

package runtime

import (
	"internal/runtime/atomic"
	"unsafe"
)

// External timer notifier for windows.
//
// The external ("real time") clock read by nanotimeExternal is the biased
// interrupt time, which advances while the system is asleep. Windows wait
// timeouts (WaitForSingleObject and friends) are driven by that same biased
// interrupt time, so a relative millisecond timeout naturally elapses in real
// time and survives sleep, matching nanotimeExternal.
//
// The notifier is an auto-reset event. externalTimerBlock waits on it with a
// timeout computed from the soonest external deadline; externalTimerKick signals
// it to wake the wait early when a sooner timer is added. See
// time_externaltimer.go for how these are used.

const haveExternalTimer = true

const (
	_INFINITE       = 0xffffffff
	_INFINITE_MINUS = _INFINITE - 1 // largest finite WaitForSingleObject timeout
)

// externalTimerEvent is an auto-reset event handle used to wake
// externalTimerBlock early (see externalTimerKick).
var externalTimerEvent uintptr

// externalTimerWhen is the absolute external-clock deadline (in nanoseconds, as
// read by nanotimeExternal) that externalTimerBlock should wait for, or <= 0 to
// wait indefinitely. It is set by externalTimerArm while externalTimers.mu is
// held and read by externalTimerBlock after the lock is released.
var externalTimerWhen atomic.Int64

func externalTimerInit() {
	// Auto-reset event (bManualReset = 0), initially non-signaled.
	var h uintptr
	systemstack(func() {
		h = stdcall(_CreateEventA, 0, 0, 0, 0)
	})
	if h == 0 {
		println("runtime: CreateEventA failed; errno=", getlasterror())
		throw("runtime: external timer init failed")
	}
	externalTimerEvent = h
}

// externalTimerArm records the soonest external deadline for externalTimerBlock.
// If when <= 0 the notifier waits indefinitely until kicked.
func externalTimerArm(when int64) {
	externalTimerWhen.Store(when)
}

// externalTimerBlock waits until the armed external deadline elapses or
// externalTimerKick signals the event. The timeout is measured against the
// biased interrupt time, so it advances during sleep, matching nanotimeExternal.
func externalTimerBlock() {
	when := externalTimerWhen.Load()
	var ms uint32
	if when <= 0 {
		ms = _INFINITE
	} else {
		delta := when - nanotimeExternal()
		if delta <= 0 {
			// Already due; return to run the due timers. Any pending kick stays
			// signaled and is consumed by the next wait, which is harmless.
			return
		}
		// Round the nanosecond delay up to whole milliseconds.
		msDelta := (delta + 1e6 - 1) / 1e6
		if msDelta > _INFINITE_MINUS {
			msDelta = _INFINITE_MINUS
		}
		ms = uint32(msDelta)
	}

	// Block using the cgocall path (as syscall.SyscallN does) so that the P is
	// released for the duration of the wait, rather than the runtime-internal
	// stdcall, which would hold the P and can deadlock at GOMAXPROCS=1. The call
	// parameters live in m.winsyscall because the stack may move during the call.
	c := &getg().m.winsyscall
	c.Fn = uintptr(unsafe.Pointer(_WaitForSingleObject))
	c.N = 2
	args := [2]uintptr{externalTimerEvent, uintptr(ms)}
	c.Args = uintptr(noescape(unsafe.Pointer(&args[0])))
	cgocall(asmstdcallAddr, unsafe.Pointer(c))
}

// externalTimerKick wakes a blocked externalTimerBlock by signaling the event.
func externalTimerKick() {
	var ok uintptr
	systemstack(func() {
		ok = stdcall(_SetEvent, externalTimerEvent)
	})
	if ok == 0 {
		println("runtime: external timer SetEvent failed; errno=", getlasterror())
		throw("runtime: external timer kick failed")
	}
}
