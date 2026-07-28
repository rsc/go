// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/runtime/atomic"
)

// External ("real time") timers (see go.dev/issue/36141).
//
// External timers fire relative to the external clock read by nanotimeExternal,
// which continues to advance while the system is asleep, unlike the internal
// ("program time") clock read by nanotime. They are used by time.ExternalSleep,
// time.ExternalNewTimer, time.ExternalAfterFunc, and time.ExternalNewTicker.
//
// Rather than adding a second timer heap to every P and threading it through the
// scheduler's hot paths, all external timers live in a single global heap,
// externalTimers, serviced by a dedicated goroutine, externalTimerLoop. External
// timers are opt-in and comparatively rare, so a global heap is sufficient and
// keeps the change entirely out of findRunnable, work stealing, and sysmon. The
// manager services the heap exactly the way a synctest bubble services its own
// timers (see synctestRun): it runs due timers on the system stack and then
// blocks on a per-OS notifier armed against the external clock.
//
// On platforms where haveExternalTimer is false there is no external clock
// notifier, the manager is never started, and package time routes the External*
// functions to their internal equivalents instead (see package time).

// externalTimers is the global heap of external timers.
var externalTimers timers

// externalTimerStarted reports whether the external timer machinery has been
// initialized and externalTimerLoop has been started. externalTimerStarting is
// used to ensure that startup happens exactly once.
var (
	externalTimerStarted  atomic.Bool
	externalTimerStarting atomic.Uint32
)

// externalTimerEnsureStarted lazily initializes the external timer machinery and
// starts externalTimerLoop the first time an external timer is created. It must
// be called from a normal goroutine (not the system stack), because it may start
// a new goroutine.
func externalTimerEnsureStarted() {
	if externalTimerStarted.Load() {
		return
	}
	if externalTimerStarting.CompareAndSwap(0, 1) {
		lockInit(&externalTimers.mu, lockRankTimers)
		externalTimerInit()
		externalTimerStarted.Store(true) // publish after init completes
		go externalTimerLoop()
		return
	}
	// Another goroutine is starting the machinery; wait until it is ready so
	// that externalTimers.mu is initialized before the caller uses the heap.
	for !externalTimerStarted.Load() {
		osyield()
	}
}

// externalTimerLoop runs due external timers and then blocks until the next one
// is due (or a newly added timer wakes it). It never returns.
func externalTimerLoop() {
	for {
		// Run any due external timers. Clear m.curg while running them so that
		// timer goroutines inherit their race context from g0, matching the
		// synctest bubble scheduler.
		systemstack(func() {
			gp := getg()
			curg := gp.m.curg
			gp.m.curg = nil
			externalTimers.check(nanotimeExternal(), nil)
			gp.m.curg = curg
		})

		// Arm the OS notifier for the soonest external timer and block.
		lock(&externalTimers.mu)
		armExternalLocked()
		unlock(&externalTimers.mu)
		externalTimerBlock()
	}
}

// armExternalLocked arms the OS notifier to fire when the soonest external timer
// is due. externalTimers.mu must be held to serialize arming so that the
// notifier is never left set to a stale, later deadline.
func armExternalLocked() {
	externalTimerArm(externalTimers.wakeTime())
}

// externalTimerWakeup re-arms the OS notifier and wakes externalTimerLoop after
// an external timer that is sooner than the currently armed deadline has been
// added or modified.
func externalTimerWakeup() {
	if !externalTimerStarted.Load() {
		// The manager has not started yet; whoever starts it will arm the
		// notifier against the current heap.
		return
	}
	lock(&externalTimers.mu)
	armExternalLocked()
	unlock(&externalTimers.mu)
	externalTimerKick()
}
