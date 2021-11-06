// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package reflext

import (
	"image"
	"reflect"
	"testing"
)

type TB struct {
	A string
	C image.Rectangle
	D []int16
	E map[int]string
}

type TA struct {
	A int32
	B TB
	C []image.Point
	D map[float64]string
}

func createTA() *TA {
	return &TA{
		B: TB{
			A: "hello",
			C: image.Rect(12, 34, 56, 78),
			D: []int16{12, 34, 56},
			E: map[int]string{12: "hello", 34: "world"},
		},
		A: 1234,
		C: []image.Point{image.Pt(12, 34)},
		D: map[float64]string{3.14: "PI"},
	}
}

func TestEvalView(t *testing.T) {
	var obj = createTA()
	tests := []struct {
		expr         string
		shouldHasErr bool
		result       interface{}
	}{
		{"B.A", false, obj.B.A},
		{"C", false, obj.C},
		{"B.C.Max.X", false, obj.B.C.Max.X},
		{"C[0].X", false, obj.C[0].X},
		{"D[3.14]", false, obj.D[3.14]},

		{"C[10]", false, nil},
		{"D[10]", false, nil},
		{"B.D[10]", false, nil},
		{"B.E[10]", false, nil},
		{`D["KKK"]`, true, nil},
		{`B.D["KKK"]`, true, nil},
	}
	for _, tc := range tests {
		v, err := EvalView(obj, tc.expr)
		if tc.shouldHasErr {
			if err == nil {
				t.Fatalf("%s: %v", tc.expr, err)
			}
			continue
		}
		t.Logf("%s: %v", tc.expr, err)
		if tc.result == nil && !IsInterfaceNil(v) {
			t.Fatalf("%s: %v", tc.expr, err)
		} else if !reflect.DeepEqual(v, tc.result) {
			t.Fatalf("%s: %v", tc.expr, err)
		}
	}
}

func TestEvalSet(t *testing.T) {
	var obj = createTA()
	tests := []struct {
		expr         string
		val          interface{}
		shouldHasErr bool
	}{
		//{"A", 5678, obj.A, false},
		//{"B.A", "hi", obj.B.A, false},
		//{"B.C.Max.X", 100, false},
		//{"C[0].X", 54321, false},
		//{"D[3.14]", "pi", false},
		{"D[1.68]", "ratio", false},
		{"D.B.E[100]", "100", false},
	}
	for _, tc := range tests {
		err := EvalSet(obj, tc.expr, tc.val)
		t.Logf("%s: %v", tc.expr, err)
		if tc.shouldHasErr {
			if err == nil {
				t.Fatalf("%s: %v", tc.expr, err)
			}
			continue
		}
		v, err := EvalView(obj, tc.expr)
		if err != nil {
			t.Fatalf("%s: %v", tc.expr, err)
		}
		if !reflect.DeepEqual(v, tc.val) {
			t.Fatalf("%s: %v", tc.expr, err)
		}
	}
}

func TestEvalRemove(t *testing.T) {
	var obj = createTA()
	tests := []struct {
		expr   string
		hasErr bool
	}{
		{"A", false},
		{"B.D[2]", false},
	}
	for _, tc := range tests {
		err := EvalRemove(obj, tc.expr)
		t.Logf("%s: %v", tc.expr, err)
		if tc.hasErr {
			if err == nil {
				t.Fatalf("%s: %v", tc.expr, err)
			}
		}
	}
}
