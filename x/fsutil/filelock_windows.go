// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

//go:build windows

package fsutil

import (
	"os"
)

// Windows byte-range locks are enforced (also referred to as mandatory locks) by the file systems
// more details, see https://en.wikipedia.org/wiki/File_locking

func LockFile(f *os.File) error {
	return nil
}

func UnlockFile(f *os.File) error {
	return nil
}
