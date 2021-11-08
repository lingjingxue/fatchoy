// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package reflext

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// 方法不超过2个返回值
// 如果返回1个值，则返回的是result或者error
// 如果返回2个值，第1个是result，第2个是error
func validateFuncNumOut(fnType reflect.Type) bool {
	switch fnType.NumOut() {
	case 0, 1:
		return true
	case 2:
		outType := fnType.Out(1)
		return outType.Name() == "error"
	default:
		return false
	}
}

// 转换函数所有参数类型
func castToFuncParams(fnType reflect.Type, args []json.RawMessage) ([]reflect.Value, error) {
	var params = make([]reflect.Value, 0, fnType.NumIn())
	for i := 0; i < len(args); i++ {
		value := reflect.New(fnType.In(i))
		if err := json.Unmarshal(args[i], value.Interface()); err != nil {
			return nil, err
		}
		params = append(params, value.Elem())
	}
	return params, nil
}

func FormatToString(v interface{}) string {
	var rv = reflect.ValueOf(v)
	if IsPrimitive(rv.Kind()) {
		return fmt.Sprintf("%v", rv.Interface())
	}
	data, _ := json.Marshal(rv.Interface())
	return string(data)
}

// 方法不超过2个返回值
// 如果返回1个值，则返回的是result或者error
// 如果返回2个值，第1个是result，第2个是error
func retrieveOutput(fnType reflect.Type, output []reflect.Value) (result, outErr interface{}) {
	switch len(output) {
	case 0:
		return
	case 1:
		if fnType.Out(0).Name() == "error" {
			outErr = output[0].Interface()
		} else {
			result = output[0].Interface()
		}
	case 2:
		result = output[0].Interface()
		outErr = output[1].Interface()
	}
	return
}

// 枚举可调用函数
func EnumerateCallable(v interface{}) map[string]reflect.Value {
	var callables = map[string]reflect.Value{}
	var rv = reflect.ValueOf(v) // v需要传递为指针
	var rtype = rv.Type()
	for i := 0; i < rtype.NumMethod(); i++ {
		var method = rtype.Method(i)
		var fn = rv.Method(i)
		if !validateFuncNumOut(fn.Type()) {
			continue
		}
		if method.Name != "" && fn.IsValid() {
			callables[method.Name] = fn
		}
	}
	return callables
}

func InvokeCallable(callable reflect.Value, argsArrayJson string) (outResult, outErr interface{}, err error) {
	if !callable.IsValid() || callable.IsNil() {
		err = fmt.Errorf("method is not valid")
		return
	}
	// 传入的参数数量必须和函数的参数签名匹配
	var args []json.RawMessage
	if argsArrayJson != "" {
		if er := json.Unmarshal([]byte(argsArrayJson), &args); er != nil {
			err = fmt.Errorf("invalid arguments: %w", er)
			return
		}
	}
	var fnType = callable.Type()
	if len(args) != fnType.NumIn() {
		err = fmt.Errorf("method expect %d parameters but got %d", fnType.NumIn(), len(args))
		return
	}

	// 把参数转换为具体类型的值
	var fnArgs []reflect.Value
	if fnType.NumIn() > 0 {
		input, er := castToFuncParams(fnType, args)
		if er != nil {
			err = fmt.Errorf("cannot cast func parameters: %w", err)
			return
		}
		fnArgs = input
	}
	// 执行函数，打包返回值
	outputs := callable.Call(fnArgs)
	outResult, outErr = retrieveOutput(fnType, outputs)
	return
}
