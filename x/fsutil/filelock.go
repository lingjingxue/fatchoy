// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

//go:build linux || darwin

package fsutil

import (
	"log"
	"os"
	"syscall"
)

func LockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

func UnlockFile(f *os.File) error {
	var err = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	if err != nil {
		log.Printf("UnlockFile: %v", err)
	}
	return err
}
