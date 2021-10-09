// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package mathext

import (
	"math"
)

// code below taken from
// https://github.com/google/googletest/blob/master/googletest/include/gtest/internal/gtest-internal.h#L232

// How many ULP's (Units in the Last Place) we want to tolerate when
// comparing two numbers.  The larger the value, the more error we
// allow.  A 0 value means that two numbers must be exactly the same
// to be considered equal.
//
// The maximum error of a single floating-point operation is 0.5
// units in the last place.  On Intel CPU's, all floating-point
// calculations are done with 80-bit precision, while double has 64
// bits.  Therefore, 4 should be enough for ordinary use.
//
// See the following article for more details on ULP:
// http://randomascii.wordpress.com/2012/02/25/comparing-floating-point-numbers-2012-edition/
const kMaxUlps = 4

// Converts an integer from the sign-and-magnitude representation to
// the biased representation.  More precisely, let N be 2 to the
// power of (kBitCount - 1), an integer x is represented by the
// unsigned number x + N.
//
// For instance,
//
//   -N + 1 (the most negative number representable using
//          sign-and-magnitude) is represented by 1;
//   0      is represented by N; and
//   N - 1  (the biggest number representable using
//          sign-and-magnitude) is represented by 2N - 1.
//
// Read http://en.wikipedia.org/wiki/Signed_number_representations
// for more details on signed number representations.
func SignAndMagnitudeToBiasedFloat64(sam uint64) uint64 {
	const kSignBitMask uint64 = 0x8000000000000000
	if (kSignBitMask & sam) != 0 { // sam represents a negative number
		return ^sam + 1
	} else { // sam represents a positive number
		return kSignBitMask | sam
	}
}

func SignAndMagnitudeToBiasedFloat32(sam uint32) uint32 {
	const kSignBitMask uint32 = 0x80000000
	if (kSignBitMask & sam) != 0 { // sam represents a negative number
		return ^sam + 1
	} else { // sam represents a positive number
		return kSignBitMask | sam
	}
}

// Returns true if this number is at most kMaxUlps ULP's away from
// rhs.  In particular, this function:
//
//   - returns false if either number is (or both are) NAN.
//   - treats really large numbers as almost equal to infinity.
//   - thinks +0.0 and -0.0 are 0 DLP's apart.
func IsAlmostEqualFloat64(a, b float64) bool {
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	var biased1 = SignAndMagnitudeToBiasedFloat64(math.Float64bits(a))
	var biased2 = SignAndMagnitudeToBiasedFloat64(math.Float64bits(b))
	if biased1 >= biased2 {
		return (biased1 - biased2) <= kMaxUlps
	} else {
		return (biased2 - biased1) <= kMaxUlps
	}
}

func IsAlmostEqualFloat32(a, b float32) bool {
	if math.IsNaN(float64(a)) || math.IsNaN(float64(b)) {
		return false
	}
	var biased1 = SignAndMagnitudeToBiasedFloat32(math.Float32bits(a))
	var biased2 = SignAndMagnitudeToBiasedFloat32(math.Float32bits(b))
	if biased1 >= biased2 {
		return (biased1 - biased2) <= kMaxUlps
	} else {
		return (biased2 - biased1) <= kMaxUlps
	}
}
