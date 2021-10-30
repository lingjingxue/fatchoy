// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package reflext

import (
	"image"
	"reflect"
	"testing"
)

type TT struct {
	A string
	B int
	C image.Rectangle
	D []int
	E []image.Point
	F map[int]string
	G map[float64]int
}

func TestView(t *testing.T) {
	var obj = TT{
		A: "hello",
		B: 1024,
		C: image.Rect(12, 34, 56, 78),
		D: []int{1, 2, 3, 4},
		E: []image.Point{image.Pt(12, 34)},
		F: map[int]string{12: "K"},
		G: map[float64]int{3.14: 314},
	}
	tests := []struct {
		expr   string
		hasErr bool
		result interface{}
	}{
		{"A", false, obj.A},
		{"C", false, obj.C},
		{"C.Max.X", false, obj.C.Max.X},
		{"D[1]", false, obj.D[1]},
		{"D[10]", false, nil},
		{"F[10]", false, nil},
		{"G[3.14]", false, 314},
		{`F["X"]`, true, nil},
		{"E[0].X", false, obj.E[0].X},
		{`F[12]`, false, obj.F[12]},
		{`F["KKK"]`, false, nil},
	}
	for _, tc := range tests {
		v, err := View(obj, tc.expr)
		if tc.hasErr {
			if err == nil {
				t.Fatalf("%v", err)
			}
			continue
		}
		t.Logf("%s: %v\n", tc.expr, v)
		if tc.result == nil && !IsInterfaceNil(v) {
			t.Fatalf("%v", err)
		} else if !reflect.DeepEqual(v, tc.result) {
			t.Fatalf("%v", err)
		}
	}
}
