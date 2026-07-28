// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time_test

import (
	"testing"
	. "time"
)

// TestExternalNowMonotonic checks that ExternalNow carries a monotonic
// reading and that it is tagged external exactly when the platform supports it.
func TestExternalNowMonotonic(t *testing.T) {
	tm := ExternalNow()
	if GetMono(&tm) == 0 {
		t.Fatalf("ExternalNow() has no monotonic reading")
	}
	if got, want := IsExternal(&tm), HaveExternalTime(); got != want {
		t.Errorf("IsExternal(ExternalNow()) = %v, want %v (HaveExternalTime)", got, want)
	}
}

// TestExternalMonotonicOrdering checks that two external readings are ordered
// and subtract using the monotonic path.
func TestExternalMonotonicOrdering(t *testing.T) {
	t1 := ExternalNow()
	t2 := ExternalNow()
	if t2.Before(t1) {
		t.Errorf("second ExternalNow() is before the first: %v < %v", t2, t1)
	}
	if d := t2.Sub(t1); d < 0 {
		t.Errorf("t2.Sub(t1) = %v, want >= 0", d)
	}
	// Since/Until on an external time must not be wildly off.
	if d := Since(t1); d < 0 || d > Hour {
		t.Errorf("Since(externalNow) = %v, want small non-negative", d)
	}
	if d := Until(t1); d > 0 || d < -Hour {
		t.Errorf("Until(externalNow) = %v, want small non-positive", d)
	}
}

// TestExternalInternalMixedSub checks that mixing external and internal
// monotonic times in Sub does not panic and falls back to wall-clock
// subtraction (yielding a small magnitude, since both were read ~now).
func TestExternalInternalMixedSub(t *testing.T) {
	ext := ExternalNow()
	in := Now()
	// Mixed subtraction must not panic and should be near zero.
	d1 := in.Sub(ext)
	d2 := ext.Sub(in)
	if d1 < -Minute || d1 > Minute {
		t.Errorf("Now().Sub(ExternalNow()) = %v, want near zero", d1)
	}
	if d2 < -Minute || d2 > Minute {
		t.Errorf("ExternalNow().Sub(Now()) = %v, want near zero", d2)
	}
	// Comparison must also not use the (incompatible) raw ext values.
	_ = ext.Before(in)
	_ = ext.After(in)
	_ = ext.Equal(in)
	_ = ext.Compare(in)
}

func TestExternalNewTimer(t *testing.T) {
	const delay = 50 * Millisecond
	start := ExternalNow()
	tm := ExternalNewTimer(delay)
	defer tm.Stop()
	var tick Time
	select {
	case tick = <-tm.C:
	case <-After(10 * Second):
		t.Fatal("ExternalNewTimer did not fire within 10s")
	}
	if elapsed := ExternalNow().Sub(start); elapsed < delay {
		t.Errorf("ExternalNewTimer(%v) fired after %v, want >= %v", delay, elapsed, delay)
	}
	// The value delivered on the channel should carry a monotonic reading, and
	// be tagged external exactly when the platform has external timers. (This is
	// HaveExternalTimer, not HaveExternalTime: on platforms with a real external
	// clock but no external timer notifier, such as Windows, external timers
	// fall back to the internal machinery and deliver internal readings.)
	if GetMono(&tick) == 0 {
		t.Errorf("tick from external timer has no monotonic reading")
	}
	if got, want := IsExternal(&tick), HaveExternalTimer(); got != want {
		t.Errorf("IsExternal(external tick) = %v, want %v", got, want)
	}
	if d := tick.Sub(start); d < 0 {
		t.Errorf("tick.Sub(start) = %v, want >= 0", d)
	}
}

func TestExternalAfterFunc(t *testing.T) {
	done := make(chan bool, 1)
	ExternalAfterFunc(20*Millisecond, func() { done <- true })
	select {
	case <-done:
	case <-After(10 * Second):
		t.Fatal("ExternalAfterFunc did not fire within 10s")
	}
}

func TestExternalSleep(t *testing.T) {
	const delay = 50 * Millisecond
	start := ExternalNow()
	ExternalSleep(delay)
	if elapsed := ExternalNow().Sub(start); elapsed < delay {
		t.Errorf("ExternalSleep(%v) slept only %v", delay, elapsed)
	}
	// Non-positive durations return immediately.
	ExternalSleep(0)
	ExternalSleep(-1)
}

func TestExternalNewTicker(t *testing.T) {
	const period = 30 * Millisecond
	tk := ExternalNewTicker(period)
	defer tk.Stop()
	start := ExternalNow()
	const want = 3
	for i := 0; i < want; i++ {
		select {
		case <-tk.C:
		case <-After(10 * Second):
			t.Fatalf("ExternalNewTicker only produced %d of %d ticks", i, want)
		}
	}
	if elapsed := ExternalNow().Sub(start); elapsed < want*period {
		t.Errorf("%d ticks took %v, want >= %v", want, elapsed, want*period)
	}
	// Reset to a new period and confirm it keeps ticking.
	tk.Reset(15 * Millisecond)
	select {
	case <-tk.C:
	case <-After(10 * Second):
		t.Fatal("ExternalNewTicker did not tick after Reset")
	}
}

func TestExternalTimerReset(t *testing.T) {
	tm := ExternalNewTimer(Hour)
	defer tm.Stop()
	if !tm.Reset(20 * Millisecond) {
		t.Error("Reset returned false for active timer")
	}
	select {
	case <-tm.C:
	case <-After(10 * Second):
		t.Fatal("external timer did not fire within 10s after Reset")
	}
}

func TestExternalTimerStop(t *testing.T) {
	tm := ExternalNewTimer(20 * Millisecond)
	if !tm.Stop() {
		t.Error("Stop returned false for active timer")
	}
	select {
	case <-tm.C:
		t.Error("external timer fired after Stop")
	case <-After(100 * Millisecond):
	}
}

func TestExternalTick(t *testing.T) {
	if ExternalTick(-1) != nil {
		t.Error("ExternalTick(-1) != nil")
	}
	c := ExternalTick(20 * Millisecond)
	select {
	case <-c:
	case <-After(10 * Second):
		t.Fatal("ExternalTick did not tick within 10s")
	}
}
