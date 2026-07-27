// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package windows

import (
	"syscall"
	"unsafe"
)

// GetObjectName returns the NT object name of the object referenced by h,
// as reported by NtQueryObject with the ObjectNameInformation class.
//
// For a Winsock socket the name is `\Device\Afd`; for a file it is the full
// NT device path (for example `\Device\HarddiskVolume4\Users\gopher\foo`).
// The name may be empty for objects that are not named.
func GetObjectName(h syscall.Handle) (string, error) {
	// The returned OBJECT_NAME_INFORMATION is followed by the UTF-16 name
	// it points to, so allocate a single buffer for both. A socket name is
	// tiny, but file paths can be long, so grow the buffer if necessary.
	b := make([]byte, 512)
	for tries := 0; ; tries++ {
		var retLen uint32
		err := NtQueryObject(h, ObjectNameInformation, unsafe.Pointer(&b[0]), uint32(len(b)), &retLen)
		if err == nil {
			oni := (*OBJECT_NAME_INFORMATION)(unsafe.Pointer(&b[0]))
			if oni.Name.Buffer == nil || oni.Name.Length == 0 {
				return "", nil
			}
			return syscall.UTF16ToString(unsafe.Slice(oni.Name.Buffer, oni.Name.Length/2)), nil
		}
		if s, ok := err.(NTStatus); ok && s == STATUS_INFO_LENGTH_MISMATCH && int(retLen) > len(b) && tries == 0 {
			b = make([]byte, retLen)
			continue
		}
		return "", err
	}
}
