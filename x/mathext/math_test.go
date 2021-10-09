// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package mathext

import (
	"testing"
)

func TestMathAbs(t *testing.T) {
	if a := Int8.Abs(-11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Int16.Abs(-11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Int32.Abs(-11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Int64.Abs(-11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Int.Abs(-11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Float32.Abs(-11); a != 11.0 {
		t.Fatalf("%f != %f", a, 11.0)
	}
	if a := Float64.Abs(-11); a != 11.0 {
		t.Fatalf("%f != %f", a, 11.0)
	}
}

func TestMathMax(t *testing.T) {
	if a := Int8.Max(10, 11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Int16.Max(10, 11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Int32.Max(10, 11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Int64.Max(10, 11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Int.Max(10, 11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Float32.Max(10, 11); a != 11.0 {
		t.Fatalf("%f != %f", a, 11.0)
	}
	if a := Float64.Max(10, 11); a != 11.0 {
		t.Fatalf("%f != %f", a, 11.0)
	}
}

func TestMathMin(t *testing.T) {
	if a := Int8.Max(12, 11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Int16.Max(12, 11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Int32.Max(12, 11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Int64.Max(12, 11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Int.Max(12, 11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Float32.Max(12, 11); a != 11.0 {
		t.Fatalf("%f != %f", a, 11.0)
	}
	if a := Float64.Max(12, 11); a != 11.0 {
		t.Fatalf("%f != %f", a, 11.0)
	}
}

func TestMathDim(t *testing.T) {
	if a := Int8.Max(22, 11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Int16.Max(22, 11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Int32.Max(22, 11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Int64.Max(22, 11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Int.Max(22, 11); a != 11 {
		t.Fatalf("%d != %d", a, 11)
	}
	if a := Float32.Max(22, 11); a != 11.0 {
		t.Fatalf("%f != %f", a, 11.0)
	}
	if a := Float64.Max(22, 11); a != 11.0 {
		t.Fatalf("%f != %f", a, 11.0)
	}
}
