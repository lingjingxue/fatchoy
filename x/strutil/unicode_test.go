// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

//go:build !ignore

package strutil

import (
	"testing"
)

func TestWordCount(t *testing.T) {
	var cases = map[string]int{
		"one word: λ":             3,
		"中文":                      0,
		"你好，sekai！":               1,
		"oh, it's super-fancy!!a": 4,
		"":                        0,
		"-":                       0,
		"it's-'s":                 1,
	}
	for str, cnt := range cases {
		var n = WordCount(str)
		if n != cnt {
			t.Fatalf("%s is not %d length", str, n)
		}
	}
}

func TestRuneWidth(t *testing.T) {
	var cases = map[string]int{
		"a":    1,
		"中":    2,
		"\x11": 0,
	}
	for r, cnt := range cases {
		var n = RuneWidth([]rune(r)[0])
		if n != cnt {
			t.Fatalf("%s is not %d length", r, n)
		}
	}
}
