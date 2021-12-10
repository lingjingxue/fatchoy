// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package set

import (
	"strings"
	"testing"
)

func TestBitSet_Set(t *testing.T) {
	var bs = NewBitSet(100)
	for i := 0; i < 100; i += 2 {
		bs.Set(i)
	}
	if bs.bits[0] != 0b101010101010101010101010101010101010101010101010101010101010101 {
		t.Fatalf("bits[0] mismatch %b\n", bs.bits[0])
	}
	if bs.bits[1] != 0b10101010101010101010101010101010101 {
		t.Fatalf("bits[1] mismatch %b\n", bs.bits[1])
	}
	for i := 0; i < 100; i += 2 {
		if !bs.Test(i) {
			t.Fatalf("bit at index %d should be positive\n", i)
		}
	}
}

func TestBitSet_Flip(t *testing.T) {
	var bs = NewBitSet(100)
	for i := 0; i < 100; i++ {
		bs.Set(i)
	}
	for i := 1; i < 100; i += 2 {
		bs.Flip(i)
	}
	if bs.bits[0] != 0b101010101010101010101010101010101010101010101010101010101010101 {
		t.Fatalf("bits[0] mismatch %b\n", bs.bits[0])
	}
	if bs.bits[1] != 0b10101010101010101010101010101010101 {
		t.Fatalf("bits[1] mismatch %b\n", bs.bits[1])
	}
}

func TestBitSet_String(t *testing.T) {
	// empty bitset
	var bs = NewBitSet(200)
	var s = bs.String()
	var expected = strings.Repeat("0", 200)
	if s != expected {
		t.Fatalf("unexpect string output: %s\n", s)
	}

	// set 1 at each 20 gap
	for i := 0; i < bs.Size(); i += 20 {
		bs.Set(i)
	}
	expected = strings.Repeat("10000000000000000000", 10)
	s = bs.String()
	if s != expected {
		t.Fatalf("unexpect string output: %s\n", s)
	}

	// set all to 1
	for i := 0; i < bs.Size(); i++ {
		bs.Set(i)
	}
	expected = strings.Repeat("1", 200)
	s = bs.String()
	if s != expected {
		t.Fatalf("unexpect string output: %s\n", s)
	}
}

func TestBitSet_Count(t *testing.T) {
	var bs = NewBitSet(200)
	if n := bs.Count(); n != 0 {
		t.Fatalf("expect 0 but got %d\n", n)
	}
	for i := 0; i < 200; i += 3 {
		bs.Set(i)
	}
	if n := bs.Count(); n != 67 {
		t.Fatalf("expect 67 but got %d\n", n)
	}
	for i := 0; i < 200; i++ {
		bs.Set(i)
	}
	if n := bs.Count(); n != 200 {
		t.Fatalf("expect 200 but got %d\n", n)
	}
}
