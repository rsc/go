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
