// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

package runtime

// On Darwin the external ("real time") monotonic clock is mach_continuous_time,
// which, unlike mach_absolute_time (used by nanotime), continues to advance
// while the system is asleep. See go.dev/issue/36141.

// haveExternalTime reports whether nanotimeExternal reads a real external
// clock distinct from nanotime on this platform.
const haveExternalTime = true

// nanotimeExternal returns the mach_continuous_time reading in nanoseconds.
//
//go:nosplit
func nanotimeExternal() int64 {
	return nanotimeExternal1()
}
