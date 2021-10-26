// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fsutil

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"testing"
)

func randBytes(length int) []byte {
	if length <= 0 {
		return nil
	}
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		ch := uint8(rand.Int31() % 0xFF)
		result[i] = ch
	}
	return result
}

func TestFileWriter(t *testing.T) {
	filename := fmt.Sprintf("tmp-%s.log", hex.EncodeToString(randBytes(4)))
	var writer = NewFileWriter(filename, 0, WriterAsync)
	defer os.Remove(filename)
	for i := 0; i < 10000; i++ {
		fmt.Fprintf(writer, "%s-%d\n", "a quick brown fox jumps over the lazy dog", i+1)
	}
	writer.Close()
}
