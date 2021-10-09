// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package mathext

import (
	"testing"
)

func TestLazyLCGRand(t *testing.T) {
	var rng LCG
	for i := 0; i < 1000; i++ {
		rng.Rand()
	}
}

func TestSetLazyLCGSeed(t *testing.T) {
	var rng LCG
	rng.Seed(1234567890)
	for i := 0; i < 1000; i++ {
		rng.Rand()
	}
}

func TestRandInt(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := RandInt(0, 1000)
		if v < 0 {
			t.Fatalf("%v < 0", v)
		}
		if v > 1000 {
			t.Fatalf("%v > 1000", v)
		}
	}
}

func TestRandFloat(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := RandFloat(0, 1.0)
		if v < 0 {
			t.Fatalf("%v < 0.0", v)
		}
		if v > 1.0 {
			t.Fatalf("%v > 1.0", v)
		}
	}
}

func TestRangePerm(t *testing.T) {
	var appeared = make(map[int]bool, 1000)
	var list = RangePerm(0, 1000)
	for _, v := range list {
		if v < 0 {
			t.Fatalf("%v < 0", v)
		}
		if v > 1000 {
			t.Fatalf("%v > 1000", v)
		}
		if _, found := appeared[v]; found {
			t.Fatalf("duplicate %v", v)
		}
		appeared[v] = true
	}
}
