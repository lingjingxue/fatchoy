// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package geom

func MaxInt(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func MinInt(x, y int) int {
	if y < x {
		return y
	}
	return x
}

func AbsInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
