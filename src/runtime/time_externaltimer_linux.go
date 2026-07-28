// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

package runtime

import (
	"internal/runtime/syscall/linux"
	"unsafe"
)

// External timer notifier for linux.
//
// The notifier uses a timerfd armed against CLOCK_BOOTTIME, which continues to
// advance while the system is asleep, matching nanotimeExternal. Arming the
// timer with timerfd_settime re-arms the fd atomically, so a read that is
// already blocked returns at the new (possibly sooner) expiration; no separate
// kick is required. See time_externaltimer.go for how these are used.

const haveExternalTimer = true

// externalItimerspec matches struct itimerspec for timerfd_settime. It is built
// from the runtime's per-arch timespec so that its layout matches the kernel on
// both 32- and 64-bit platforms.
type externalItimerspec struct {
	interval timespec
	value    timespec
}

// externalTimerFD is the timerfd used to wait for external timers.
var externalTimerFD int32 = -1

func externalTimerInit() {
	fd, errno := linux.TimerfdCreate(linux.CLOCK_BOOTTIME, linux.TFD_CLOEXEC)
	if errno != 0 {
		println("runtime: timerfd_create failed with", errno)
		throw("runtime: external timer init failed")
	}
	externalTimerFD = fd
}

// externalTimerArm arms the timerfd to fire when the external clock
// (nanotimeExternal) reaches when nanoseconds. If when <= 0 the timer is
// disarmed (a zero it_value disarms a timerfd).
func externalTimerArm(when int64) {
	var its externalItimerspec
	if when > 0 {
		its.value.setNsec(when)
	}
	errno := linux.TimerfdSettime(externalTimerFD, linux.TFD_TIMER_ABSTIME, unsafe.Pointer(&its), nil)
	if errno != 0 {
		println("runtime: timerfd_settime failed with", errno)
		throw("runtime: external timer arm failed")
	}
}

// externalTimerBlock blocks until the armed timerfd fires. A blocking read
// returns the number of expirations; spurious EINTR wakeups are retried.
func externalTimerBlock() {
	var buf uint64
	for {
		entersyscallblock()
		n := read(externalTimerFD, noescape(unsafe.Pointer(&buf)), int32(unsafe.Sizeof(buf)))
		exitsyscall()
		if n != -_EINTR {
			return
		}
	}
}

// externalTimerKick is a no-op on linux: re-arming the timerfd via
// externalTimerArm already causes a blocked externalTimerBlock to return at the
// new expiration.
func externalTimerKick() {}
