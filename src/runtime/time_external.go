// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !linux && !darwin && !openbsd

package runtime

// This file provides the default (fallback) implementation of the external
// ("real time") monotonic clock used by time.ExternalNow and the external
// timer functions (see go.dev/issue/36141).
//
// The external clock is a monotonic clock that continues to advance while the
// system is asleep, unlike the internal ("program time") clock read by
// nanotime, which stops during sleep. On platforms that do not provide such a
// clock, haveExternalTime is false and nanotimeExternal falls back to
// nanotime, so external times behave exactly like internal times.

// haveExternalTime reports whether nanotimeExternal reads a real external
// clock distinct from nanotime on this platform.
const haveExternalTime = false

// nanotimeExternal returns the external ("real time") monotonic clock reading
// in nanoseconds. See the file comment.
//
//go:nosplit
func nanotimeExternal() int64 {
	return nanotime()
}
