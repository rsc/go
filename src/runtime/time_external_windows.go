// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows

package runtime

// On Windows the external ("real time") monotonic clock is the biased
// interrupt time (KUSER_SHARED_DATA.InterruptTime), which continues to advance
// while the system is asleep. This is the clock Go historically used for
// nanotime and time.Now's monotonic reading on Windows.
//
// As of go.dev/issue/36141 the default (GODEBUG=wintime=internal) instead reads
// the unbiased interrupt time for nanotime and time.Now, so that program time
// stops during sleep, matching the Unix CLOCK_MONOTONIC behavior. The biased
// reading remains available as the external clock (and, for compatibility, as
// the default when GODEBUG=wintime=external). See time_windows.h and
// nanotime1/nanotimeExternal1 in sys_windows_GOARCH.s.

// haveExternalTime reports whether nanotimeExternal reads a real external
// clock distinct from nanotime on this platform.
const haveExternalTime = true

// nanotimeExternal returns the biased interrupt-time reading in nanoseconds.
//
//go:nosplit
func nanotimeExternal() int64 {
	return nanotimeExternal1()
}

// nanotimeExternal1 reads the biased interrupt time (always real time,
// regardless of the wintime setting). It is implemented in assembly in each
// sys_windows_GOARCH.s, mirroring nanotime1 but without the bias subtraction.
func nanotimeExternal1() int64
