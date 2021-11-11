// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package reflext

import (
	"fmt"
	"image"
	"io"
	"testing"
)

type Cot struct {
}

func (c *Cot) TF11() {
	println("TF11")
}

func (c *Cot) TF12() error {
	println("TF12")
	return nil
}

func (c *Cot) TF13() error {
	println("TF13")
	return io.EOF
}

func (c *Cot) TF14() string {
	println("TF14")
	return "hello"
}

func (c *Cot) TF15() image.Point {
	println("TF15")
	return image.Point{X: 12, Y: 34}
}

func (c *Cot) TF21() (int, error) {
	println("TF21")
	return 123, nil
}

func (c *Cot) TF22() (string, error) {
	println("TF22")
	return "", io.EOF
}

func (c *Cot) TF23() (image.Point, error) {
	println("TF23")
	var pt = image.Point{X: 12, Y: 34}
	return pt, nil
}

func (c *Cot) TF24() (*image.Point, error) {
	println("TF24")
	var pt = &image.Point{X: 12, Y: 34}
	return pt, nil
}

func (c *Cot) TF25() (map[string]int, error) {
	println("TF25")
	var d = map[string]int{"X": 12, "Y": 34}
	return d, nil
}

func (c *Cot) TF31(n int, b bool, f float64, s string) {
	fmt.Printf("TF31: %d %v %f %s\n", n, b, f, s)
}

func (c *Cot) TF32(a []int, d map[string]int) {
	fmt.Printf("TF32: %v %v\n", a, d)
}

func (c *Cot) TF33(pt image.Point, rect *image.Rectangle) {
	fmt.Printf("TF33: %v %v\n", pt, rect)
}

func TestInvokeCallable(t *testing.T) {
	tests := []struct {
		method       string
		args         []string
		shouldHasErr bool
		result       string
	}{
		{"TF11", nil, false, ""},
		{"TF12", nil, false, ""},
		{"TF13", nil, true, "EOF"},
		{"TF14", nil, false, "hello"},
		{"TF15", nil, false, `{"X":12,"Y":34}`},
		{"TF21", nil, false, "123"},
		{"TF22", nil, true, "EOF"},
		{"TF23", nil, false, `{"X":12,"Y":34}`},
		{"TF24", nil, false, `{"X":12,"Y":34}`},
		{"TF25", nil, false, `{"X":12,"Y":34}`},
		{"TF31", []string{"1234", "false", "3.14", "\"hello\""}, false, ``},
		{"TF32", []string{"[12,34]", "{\"X\": 12, \"Y\": 34}"}, false, ``},
		{"TF33", []string{"{\"X\": 12, \"Y\": 34}", "{\"Min\":{\"X\":12,\"Y\":34}, \"Max\":{\"X\":56,\"Y\":78}}"}, false, ``},
	}

	var cot Cot
	var callables = EnumerateCallable(&cot)
	for _, tc := range tests {
		outResult, outErr, err := InvokeCallable(callables[tc.method], tc.args)
		if err != nil {
			t.Fatalf("%s: %v", tc.method, err)
		}
		if tc.shouldHasErr {
			if IsInterfaceNil(outErr) {
				t.Fatalf("%s should return error", tc.method)
			}
			er := outErr.(error)
			if er.Error() != tc.result {
				t.Fatalf("%s unexpected error", tc.method)
			}
			continue
		}
		if !IsInterfaceNil(outErr) {
			var s = fmt.Sprintf("%v", outResult)
			if s != tc.result {
				t.Fatalf("%s: %s != %s", tc.method, s, tc.result)
			}
		} else {
			if IsInterfaceNil(outErr) && tc.result == "" {
				continue
			}
			s := FormatToString(outResult)
			if s != tc.result {
				t.Fatalf("%s: %s != %s", tc.method, s, tc.result)
			}
		}
	}
}

func TestParseCallExpr(t *testing.T) {
	tests := []struct {
		expr string
		fn   string
		args []string
	}{
		{"FFF()", "FFF", nil},
		{"AAA(-123, 'a', `hello `)", "AAA", []string{"-123", "a", "hello"}},
		{`BBB("hello", "kitty", "{\"101\":2,\"102\":3}")`, "BBB", []string{"hello", "kitty", `{"101":2,"102":3}`}},
	}

	for _, tc := range tests {
		fnName, args, err := ParseCallExpr(tc.expr)
		if err != nil {
			t.Fatalf("%s: %v", tc.expr, err)
		}
		if len(args) != len(tc.args) {
			t.Fatalf("%s: output mismatch", tc.expr)
		}
		if fnName != tc.fn {
			t.Fatalf("%s: func name mismatch", tc.expr)
		}
		for i := 0; i < len(tc.args); i++ {
			if tc.args[i] != args[i] {
				t.Fatalf("%s: argument %d not equal", tc.expr, i)
			}
		}
	}
}
