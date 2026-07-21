// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

package runtime

// On Linux the external ("real time") monotonic clock is CLOCK_BOOTTIME,
// which, unlike CLOCK_MONOTONIC (used by nanotime), continues to advance while
// the system is suspended. See go.dev/issue/36141.

// haveExternalTime reports whether nanotimeExternal reads a real external
// clock distinct from nanotime on this platform.
const haveExternalTime = true

// nanotimeExternal returns the CLOCK_BOOTTIME reading in nanoseconds.
//
//go:nosplit
func nanotimeExternal() int64 {
	return nanotimeExternal1()
}

// nanotimeExternal1 reads CLOCK_BOOTTIME. It is implemented in assembly in each
// sys_linux_GOARCH.s, mirroring nanotime1 but with a different clock id.
func nanotimeExternal1() int64
