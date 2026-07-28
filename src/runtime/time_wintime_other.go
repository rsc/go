// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !windows

package runtime

// setwintime applies the GODEBUG=wintime setting, which only affects Windows.
// On all other systems it is a no-op. See time_wintime_windows.go and
// go.dev/issue/36141.
func setwintime(value string) {}
