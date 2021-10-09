// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package strutil

import (
	"bytes"
	"testing"
)

func checkStrEqual(t *testing.T, s1, s2 string) {
	if s1 != s2 {
		t.Fatalf("string not equal, %s != %s", s1, s2)
	}
}

func checkBytesEqual(t *testing.T, b1, b2 []byte) {
	if !bytes.Equal(b1, b2) {
		t.Fatalf("bytes not equal, %v != %v", b1, b2)
	}
}

func TestFastBytesToString(t *testing.T) {
	var rawbytes = RandBytes(1024)
	checkStrEqual(t, string(rawbytes), BytesAsString(rawbytes))
}

func TestFastStringToBytes(t *testing.T) {
	var text = RandString(1024)
	checkBytesEqual(t, []byte(text), StringAsBytes(text))
}

func TestFindString(t *testing.T) {
	tests := []struct {
		input    []string
		target   string
		expected int
	}{
		{[]string{}, "", -1},
		{[]string{"1", "2", "3", "4"}, "4", 3},
		{[]string{"1", "2", "3", "4"}, "", -1},
	}
	for i, test := range tests {
		output := FindStringInArray(test.input, test.target)
		if test.expected != output {
			t.Fatalf("Test case %d failed, expect %d, got %d", i, test.expected, output)
		}
	}
}

func TestFindFirstDigit(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", -1},
		{"abc", -1},
		{"123", 0},
		{"abc123", 3},
	}
	for i, test := range tests {
		output := FindFirstDigit(test.input)
		if test.expected != output {
			t.Fatalf("Test case %d failed, expect %d, got %d", i, test.expected, output)
		}
	}
}

func TestFindFirstNonDigit(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", -1},
		{"123", -1},
		{"abc", 0},
		{"123abc", 3},
	}
	for i, test := range tests {
		output := FindFirstNonDigit(test.input)
		if test.expected != output {
			t.Fatalf("Test case %d failed, expect %d, got %d", i, test.expected, output)
		}
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"abc", "cba"},
		{"a", "a"},
		{"çınar", "ranıç"},
		{"    yağmur", "rumğay    "},
		{"επαγγελματίες", "ςείταμλεγγαπε"},
	}

	for i, test := range tests {
		output := Reverse(test.input)
		if test.expected != output {
			t.Fatalf("Test case %d failed, expect %s, got %s", i, test.expected, output)
		}
	}
}

func TestLongestCommonPrefix(t *testing.T) {
	tests := []struct {
		input1   string
		input2   string
		expected string
	}{
		{"", "a", ""},
		{"ab", "cd", ""},
		{"ab123", "abc456", "ab"},
	}
	for i, test := range tests {
		output := LongestCommonPrefix(test.input1, test.input2)
		if test.expected != output {
			t.Fatalf("Test case %d failed, expect %s, got %s", i, test.expected, output)
		}
	}
}

func BenchmarkBytesToString(b *testing.B) {
	b.StopTimer()
	var rawbytes = RandBytes(2048)
	b.StartTimer()
	var text string
	for i := 0; i < b.N; i++ {
		text = string(rawbytes)
	}
	text = text[:0]
}

func BenchmarkFastBytesToString(b *testing.B) {
	b.StopTimer()
	var rawbytes = RandBytes(2048)
	b.StartTimer()
	var text string
	for i := 0; i < b.N; i++ {
		text = BytesAsString(rawbytes)
	}
	text = text[:0]
}

func BenchmarkStringToBytes(b *testing.B) {
	b.StopTimer()
	var text = string(RandString(2048))
	b.StartTimer()
	var rawbytes []byte
	for i := 0; i < b.N; i++ {
		rawbytes = []byte(text)
	}
	rawbytes = rawbytes[:0]
}

func BenchmarkFastStringToBytes(b *testing.B) {
	b.StopTimer()
	var text = string(RandString(2048))
	b.StartTimer()
	var rawbytes []byte
	for i := 0; i < b.N; i++ {
		rawbytes = StringAsBytes(text)
	}
	rawbytes = rawbytes[:0]
}
