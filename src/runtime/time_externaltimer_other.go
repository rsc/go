// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !linux && !darwin && !windows

package runtime

// This file provides the fallback (no-op) external timer notifier for platforms
// that do not have a way to block until a deadline on the external ("real time")
// clock. On these platforms haveExternalTimer is false, so the external timer
// manager is never started and these functions are never called; package time
// routes the External* functions to their internal equivalents instead.

// haveExternalTimer reports whether this platform can block until a deadline
// measured against the external clock read by nanotimeExternal.
const haveExternalTimer = false

func externalTimerInit()          {}
func externalTimerArm(when int64) {}
func externalTimerBlock()         {}
func externalTimerKick()          {}
