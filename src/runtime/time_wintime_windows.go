// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows

package runtime

// wintimeExternal selects which interrupt-time clock nanotime and time.Now use
// on Windows, controlled by GODEBUG=wintime (see setwintime and
// go.dev/issue/36141). It is read directly by the assembly implementations of
// nanotime1 and time·now in sys_windows_GOARCH.s / time_windows_GOARCH.s, so it
// must be a plain byte at a stable symbol.
//
//	0 (wintime=internal, the default): nanotime reads the unbiased interrupt
//	  time ("program time"), which stops while the system is asleep, matching
//	  the Unix CLOCK_MONOTONIC behavior.
//	1 (wintime=external): nanotime reads the biased interrupt time, which keeps
//	  advancing during sleep. This restores the classic pre-Go 1.28 behavior.
//
// nanotimeExternal always reads the biased interrupt time regardless of this
// setting.
var wintimeExternal uint8

// setwintime applies the GODEBUG=wintime setting. It is called from
// parsegodebug at startup. The default (any value other than "external",
// including unset) is the new unbiased "program time" behavior.
func setwintime(value string) {
	if value == "external" {
		wintimeExternal = 1
	} else {
		wintimeExternal = 0
	}
}
