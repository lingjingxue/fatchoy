// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package reflext

import (
	"fmt"
	"reflect"

	"gopkg.in/qchencc/fatchoy.v1/x/strutil"
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

// 解析字符串的值到value
func CreatePrimitiveValue(rtype reflect.Type, s string) (reflect.Value, error) {
	var v = reflect.New(rtype).Elem()
	switch rtype.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		if n, err := strutil.ParseI64(s); err != nil {
			return zeroRValue, err
		} else {
			v.SetInt(n)
		}

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		if n, err := strutil.ParseU64(s); err != nil {
			return zeroRValue, err
		} else {
			v.SetUint(n)
		}

	case reflect.Float32, reflect.Float64:
		if f, err := strutil.ParseF64(s); err != nil {
			return zeroRValue, err
		} else {
			v.SetFloat(f)
		}

	case reflect.Bool:
		v.SetBool(strutil.ParseBool(s))

	case reflect.String:
		v.SetString(s)

	default:
		return zeroRValue, fmt.Errorf("unexpected kind %v", rtype.Kind())
	}
	return v, nil
}
