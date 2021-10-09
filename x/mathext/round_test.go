// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package mathext

import (
	"math"
	"testing"
)

var roundTests = [][2]float64{
	{-0.49999999999999994, -0.0}, // -0.5+epsilon
	{-0.5, -1},
	{-0.5000000000000001, -1}, // -0.5-epsilon
	{0, 0},
	{0.49999999999999994, 0}, // 0.5-epsilon
	{0.5, 1},
	{-0.0, -0.0},
	{0.5000000000000001, 1},  // 0.5+epsilon
	{1.390671161567e-309, 0}, // denormal
	{2.2517998136852485e+15, 2.251799813685249e+15}, // 1 bit fraction
	{4.503599627370497e+15, 4.503599627370497e+15},  // large integer
	{math.Inf(-1), math.Inf(-1)},
	{math.Inf(1), math.Inf(1)},
	{math.NaN(), math.NaN()},
}

func TestMathRoundHalf(t *testing.T) {
	for i, pair := range roundTests {
		var a = pair[0]
		var b = pair[1]
		var c = RoundFloat(a)
		if math.IsNaN(c) {
			if !math.IsNaN(b) {
				t.Fatalf("%d: %f => %f != %f", i+1, a, c, b)
			}
		} else {
			if c != b {
				t.Fatalf("%d: %f => %f != %f", i+1, a, c, b)
			}
		}
	}
}
