// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package reflectutil

import (
	"fmt"
	"testing"
)

type AA struct {
	B int
	C bool
	D float64
	F string
}

func TestGetStructAllFieldValues(t *testing.T) {
	var a = &AA{123, false, 3.14, "ok"}
	result := GetStructAllFieldValues(a)
	fmt.Printf("%v\n", result)
}

func TestGetStructFieldValues(t *testing.T) {
	var a = &AA{123, false, 3.14, "ok"}
	result := GetStructFieldValues(a, "D")
	fmt.Printf("%v\n", result)
}

func TestGetStructFieldValuesBy(t *testing.T) {
	var a = &AA{123, false, 3.14, "ok"}
	result := GetStructFieldValuesBy(a, []string{"B", "C", "F"})
	fmt.Printf("%v\n", result)
}
