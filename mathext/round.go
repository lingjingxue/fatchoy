// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package mathext

import "math"

// 四舍五入
func RoundHalf(v float64) int {
	return int(RoundFloat(v))
}

// https://github.com/montanaflynn/stats/blob/master/round.go
func RoundFloat(x float64) float64 {
	// If the float is not a number
	if math.IsNaN(x) {
		return math.NaN()
	}

	// Find out the actual sign and correct the input for later
	sign := 1.0
	if x < 0 {
		sign = -1
		x *= -1
	}

	// Get the actual decimal number as a fraction to be compared
	_, decimal := math.Modf(x)

	// If the decimal is less than .5 we round down otherwise up
	var rounded float64
	if decimal >= 0.5 {
		rounded = math.Ceil(x)
	} else {
		rounded = math.Floor(x)
	}

	// Finally we do the math to actually create a rounded number
	return rounded * sign
}
