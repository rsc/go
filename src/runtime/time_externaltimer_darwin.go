// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

package runtime

// External timer notifier for darwin.
//
// The notifier uses a dedicated kqueue with an EVFILT_TIMER armed against the
// mach continuous clock (NOTE_MACH_CONTINUOUS_TIME), which continues to advance
// while the system is asleep, matching nanotimeExternal. An EVFILT_USER event is
// used by externalTimerKick to interrupt a blocked kevent when a sooner timer is
// added. See time_externaltimer.go for how these are used.

// Constants from <sys/event.h> not otherwise defined for the runtime.
const (
	_EVFILT_TIMER = -0x7

	_EV_ONESHOT = 0x10

	_NOTE_NSECONDS             = 0x4
	_NOTE_ABSOLUTE             = 0x8
	_NOTE_MACH_CONTINUOUS_TIME = 0x80
)

const haveExternalTimer = true

// externalKq is the kqueue used to wait for external timers.
var externalKq int32 = -1

func externalTimerInit() {
	externalKq = kqueue()
	if externalKq < 0 {
		println("runtime: kqueue failed with", -externalKq)
		throw("runtime: external timer init failed")
	}
	closeonexec(externalKq)
	// Register a user event used to interrupt a blocked kevent (see
	// externalTimerKick).
	var ev keventt
	ev.ident = 0
	ev.filter = _EVFILT_USER
	ev.flags = _EV_ADD | _EV_CLEAR
	if n := kevent(externalKq, &ev, 1, nil, 0, nil); n < 0 {
		println("runtime: external timer kevent(EVFILT_USER) failed with", -n)
		throw("runtime: external timer init failed")
	}
}

// externalTimerArm arms the timer to fire when the external clock
// (nanotimeExternal) reaches when nanoseconds. If when <= 0 the timer is
// disarmed.
func externalTimerArm(when int64) {
	var ev keventt
	ev.ident = 1
	ev.filter = _EVFILT_TIMER
	if when <= 0 {
		ev.flags = _EV_DELETE
		// Ignore errors: the timer may not currently be registered.
		kevent(externalKq, &ev, 1, nil, 0, nil)
		return
	}
	ev.flags = _EV_ADD | _EV_ONESHOT
	ev.fflags = _NOTE_MACH_CONTINUOUS_TIME | _NOTE_ABSOLUTE | _NOTE_NSECONDS
	ev.data = when
	kevent(externalKq, &ev, 1, nil, 0, nil)
}

// externalTimerBlock blocks until the armed timer fires or externalTimerKick is
// called. Spurious wakeups are fine: the caller re-checks and re-arms.
func externalTimerBlock() {
	var events [8]keventt
	entersyscallblock()
	kevent(externalKq, nil, 0, &events[0], int32(len(events)), nil)
	exitsyscall()
}

// externalTimerKick interrupts a blocked externalTimerBlock by triggering the
// user event.
func externalTimerKick() {
	var ev keventt
	ev.ident = 0
	ev.filter = _EVFILT_USER
	ev.fflags = _NOTE_TRIGGER
	kevent(externalKq, &ev, 1, nil, 0, nil)
}
