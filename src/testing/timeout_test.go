// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testing_test

import (
	"internal/testenv"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
)

// TestPerTestTimeout exercises the per-test timeout (go.dev/issue/48157).
// Because a per-test timeout kills the whole test binary, each scenario runs
// in a helper subprocess and this test inspects the resulting output.
func TestPerTestTimeout(t *testing.T) {
	testenv.MustHaveExec(t)
	t.Parallel()

	testCases := []struct {
		desc        string
		test        string // name of the helper test to run in the subprocess
		wantTimeout bool   // whether the subprocess is expected to time out
		wantMatch   string // regexp the output must match when wantTimeout
	}{
		{
			desc:        "SetTimeout then block",
			test:        "TestTimeoutBasicHelper",
			wantTimeout: true,
			wantMatch:   `panic: test TestTimeoutBasicHelper timed out after`,
		},
		{
			desc:        "SetTimeout smaller than elapsed times out immediately",
			test:        "TestTimeoutShrinkHelper",
			wantTimeout: true,
			wantMatch:   `panic: test TestTimeoutShrinkHelper timed out after`,
		},
		{
			desc:        "subtest inherits timeout and is named",
			test:        "TestTimeoutSubtestHelper",
			wantTimeout: true,
			wantMatch:   `panic: test TestTimeoutSubtestHelper/slow timed out after`,
		},
		{
			desc:        "timer paused while blocked in Run",
			test:        "TestTimeoutPauseRunHelper",
			wantTimeout: false,
		},
		{
			desc:        "SetTimeout(0) disables the timeout",
			test:        "TestTimeoutDisabledHelper",
			wantTimeout: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command(testenv.Executable(t), "-test.run=^"+tc.test+"$", "-test.v")
			cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
			b, err := cmd.CombinedOutput()
			got := string(b)
			if tc.wantTimeout {
				if err == nil {
					t.Fatalf("helper %s unexpectedly succeeded; output:\n%s", tc.test, got)
				}
				if ok, matchErr := regexp.MatchString(tc.wantMatch, got); matchErr != nil || !ok {
					t.Errorf("helper %s output:\ngot:\n%s\nwant match:\n%s", tc.test, got, tc.wantMatch)
				}
			} else {
				if err != nil {
					t.Fatalf("helper %s unexpectedly failed (%v); output:\n%s", tc.test, err, got)
				}
				if strings.Contains(got, "timed out") {
					t.Errorf("helper %s reported an unexpected timeout; output:\n%s", tc.test, got)
				}
			}
		})
	}
}

func TestTimeoutBasicHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	t.SetTimeout(50 * time.Millisecond)
	time.Sleep(time.Hour)
}

func TestTimeoutShrinkHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// Start a generous timer and consume running time against it, then shrink the
	// budget below the elapsed time: SetTimeout must not reset the elapsed time,
	// so this triggers an immediate timeout.
	t.SetTimeout(time.Hour)
	time.Sleep(30 * time.Millisecond)
	t.SetTimeout(10 * time.Millisecond)
	time.Sleep(time.Hour)
}

func TestTimeoutSubtestHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// The subtest inherits the parent's timeout and gets its own timer; the
	// parent's timer is paused while it waits in Run, so the subtest (not the
	// parent) is the one that times out.
	t.SetTimeout(50 * time.Millisecond)
	t.Run("slow", func(t *testing.T) {
		time.Sleep(time.Hour)
	})
}

func TestTimeoutPauseRunHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// The parent's timer must be paused while it is blocked in Run. A slow
	// subtest that has disabled its own timeout must not trip the parent's
	// short timer.
	t.SetTimeout(50 * time.Millisecond)
	t.Run("slow", func(t *testing.T) {
		t.SetTimeout(0)
		time.Sleep(300 * time.Millisecond)
	})
}

func TestTimeoutDisabledHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	t.SetTimeout(0)
	time.Sleep(100 * time.Millisecond)
}
