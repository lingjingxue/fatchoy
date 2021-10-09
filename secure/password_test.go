// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package secure

import (
	"math/rand"
	"testing"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func TestGeneratePasswordHash(t *testing.T) {
	var methods = []string{"plain", "default"}
	for _, method := range methods {
		for i := 0; i < 20; i++ {
			var password = RandString(12)
			var hashText = GeneratePasswordHash(password, method)
			var ok = VerifyPasswordHash(hashText, password)
			if !ok {
				t.Fatalf("password mismatch: %s, %s", password, hashText)
			}
		}
	}
}

func BenchmarkGeneratePasswordHash(b *testing.B) {
	b.StopTimer()
	var password = RandString(12)
	b.StartTimer()
	var hashText = GeneratePasswordHash(password, "")
	var ok = VerifyPasswordHash(hashText, password)
	if !ok {
		b.Fatalf("password mismatch: %s, %s", password, hashText)
	}
}
