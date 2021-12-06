// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package reflext

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
	"strings"
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
func castToFuncParams(fnType reflect.Type, args []string) ([]reflect.Value, error) {
	var params = make([]reflect.Value, 0, fnType.NumIn())
	for i := 0; i < len(args); i++ {
		value := reflect.New(fnType.In(i))
		if len(args[0]) > 0 {
			var dec = json.NewDecoder(bytes.NewReader([]byte(args[i])))
			dec.UseNumber()
			if err := dec.Decode(value.Interface()); err != nil {
				return nil, fmt.Errorf("unmarshal argument %d(%s): %w", i, args[i], err)
			}
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

func InvokeCallable(callable reflect.Value, args []string) (outResult, outErr interface{}, err error) {
	if !callable.IsValid() || callable.IsNil() {
		err = fmt.Errorf("method is not valid")
		return
	}
	// 传入的参数数量必须和函数的参数签名匹配
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

// 把一个函数调用解析为函数名和参数
func ParseCallExpr(expr string) (fnName string, params []string, outErr error) {
	node, err := parser.ParseExpr(expr)
	if err != nil {
		outErr = err
		return
	}
	// 只能是调用表达式
	call, ok := node.(*ast.CallExpr)
	if !ok {
		outErr = fmt.Errorf("only call expr allowed")
		return
	}
	// 函数名称只能是identifier
	fn, ok := call.Fun.(*ast.Ident)
	if !ok {
		outErr = fmt.Errorf("command is not identifier")
		return
	}
	fnName = fn.Name
	params = make([]string, 0, len(call.Args))

	// 把常量放入参数列表
	var putLiteralArg = func(literal *ast.BasicLit) error {
		var param = literal.Value
		switch literal.Kind {
		case token.STRING, token.CHAR:
			if param, err = strconv.Unquote(param); err != nil {
				return fmt.Errorf("cannot unquote: %w", err)
			}
			param = strings.TrimSpace(param)
		}
		params = append(params, param)
		return nil
	}
	// 参数只能是一元运算表达式或者常量，不支持嵌套和eval
	for i, arg := range call.Args {
		switch v := arg.(type) {
		case *ast.UnaryExpr:
			// 一元运算表达式只能是常量
			literal, ok := v.X.(*ast.BasicLit)
			if !ok {
				outErr = fmt.Errorf("argument %d is not literal", i)
				return
			}
			// 一元运算符号只接受【+-】
			switch v.Op {
			case token.SUB:
				literal.Value = "-" + literal.Value // 这里修改了ast里值
			case token.ADD:
			default:
				outErr = fmt.Errorf("unrecognized argument expression %T", v)
				return
			}
			if outErr = putLiteralArg(literal); outErr != nil {
				return
			}

		case *ast.BasicLit:
			if outErr = putLiteralArg(v); outErr != nil {
				return
			}
		}
	}
	return
}
