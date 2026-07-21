// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build openbsd

package runtime

// On OpenBSD the external ("real time") monotonic clock is CLOCK_BOOTTIME,
// which, unlike CLOCK_MONOTONIC (used by nanotime), continues to advance while
// the system is asleep. See go.dev/issue/36141.

// haveExternalTime reports whether nanotimeExternal reads a real external
// clock distinct from nanotime on this platform.
const haveExternalTime = true

// nanotimeExternal returns the CLOCK_BOOTTIME reading in nanoseconds.
//
//go:nosplit
func nanotimeExternal() int64 {
	return nanotimeExternal1()
}
