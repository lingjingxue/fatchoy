// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fsutil

import (
	"os"
	"testing"
)

func TestReadFileToLines(t *testing.T) {
	filename := "filelock.go"
	lines, err := ReadFileToLines(filename)
	if err != nil {
		t.Errorf("ReadFileToLines: %v", err)
	}
	for _, line := range lines {
		t.Logf("%s\n", line)
	}
}

func TestLockFile(t *testing.T) {
	filename := "filelock.go"
	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer f.Close()
	LockFile(f)
	defer UnlockFile(f)
}
