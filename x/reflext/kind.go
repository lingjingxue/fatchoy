// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package reflext

import (
	"reflect"
)

// 整数
func IsInteger(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		return true
	}
	return false
}

func IsUnsigned(kind reflect.Kind) bool {
	switch kind {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		return true
	}
	return false
}

// 浮点数
func IsFloat(kind reflect.Kind) bool {
	switch kind {
	case reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

func IsNumber(kind reflect.Kind) bool {
	return IsInteger(kind) || IsFloat(kind) ||
		kind == reflect.Complex64 || kind == reflect.Complex128
}

// 基本类型
func IsPrimitive(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool, reflect.String:
		return true
	}
	return IsNumber(kind)
}

// interface是否是nil
func IsInterfaceNil(c interface{}) bool {
	if c == nil {
		return true
	}
	var rv = reflect.ValueOf(c)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Array, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return rv.IsNil()
	}
	return false
}
