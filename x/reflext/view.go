// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package reflext

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"reflect"
	"runtime/debug"
	"strconv"
)

var (
	ErrNilViewContext = errors.New("nil view context")
	ErrNilExprNode    = errors.New("nil expr node")

	zeroRValue reflect.Value
)

// 把`expr`挂在`this`上，显示值，如: expr=
func View(this interface{}, expr string) (result interface{}, err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("%v\n%s", v, debug.Stack())
		}
	}()

	if this == nil {
		if expr != "" {
			err = ErrNilViewContext
		}
		return
	}
	var node ast.Expr
	if node, err = parser.ParseExpr(expr); err != nil {
		return
	}
	fmt.Printf("expr: %s\n", expr)
	var rv = reflect.ValueOf(this)
	rv, err = walkExpr(rv, node)
	if err == nil && rv.IsValid() {
		result = rv.Interface()
	} else {
		result = nil
	}
	return
}

func walkExpr(rv reflect.Value, node ast.Expr) (reflect.Value, error) {
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if node == nil {
		return zeroRValue, ErrNilExprNode
	}
	switch n := node.(type) {
	case *ast.Ident:
		return evalIdent(rv, n)
	case *ast.SelectorExpr:
		return evalSelectorExpr(rv, n)
	case *ast.IndexExpr:
		return evalIndexExpr(rv, n)
	default:
		return zeroRValue, fmt.Errorf("unexpected expr node %T", node)
	}
}

// 常量只能是寻址struct的field和map的key
func evalIdent(val reflect.Value, ident *ast.Ident) (reflect.Value, error) {
	switch val.Kind() {
	case reflect.Struct:
		return val.FieldByName(ident.Name), nil
	case reflect.Map:
		return val.MapIndex(reflect.ValueOf(ident.Name)), nil
	default:
		return zeroRValue, fmt.Errorf("unexpected kind %v with ident %s", val.Kind(), ident.Name)
	}
}

func evalIndexExpr(val reflect.Value, expr *ast.IndexExpr) (reflect.Value, error) {
	index, ok := expr.Index.(*ast.BasicLit)
	if !ok {
		return zeroRValue, fmt.Errorf("index is not literal")
	}
	rv, err := walkExpr(val, expr.X)
	if err != nil {
		return zeroRValue, err
	}
	switch rv.Kind() {
	case reflect.Array, reflect.Slice, reflect.String:
		if i, err := strconv.Atoi(index.Value); err != nil {
			return zeroRValue, fmt.Errorf("cannot index by key %s: %w", index.Value, err)
		} else if i < rv.Len() {
			return rv.Index(i), nil
		}
	case reflect.Map:
		var keyType = rv.Type().Key()
		if !IsPrimitive(keyType.Kind()) {
			return zeroRValue, fmt.Errorf("cannot address map key %s %s", expr.Index, keyType.Name())
		}
		if key, err := CreatePrimitiveValue(keyType, index.Value); err != nil {
			return zeroRValue, fmt.Errorf("cannot index map key %s: %w", index.Value, err)
		} else {
			return rv.MapIndex(key), nil
		}
	default:
		return zeroRValue, fmt.Errorf("unexpected kind %v with ident %s", val.Kind(), expr.Index)
	}
	return zeroRValue, nil
}

func evalSelectorExpr(val reflect.Value, expr *ast.SelectorExpr) (reflect.Value, error) {
	if val.Kind() != reflect.Struct {
		return zeroRValue, fmt.Errorf("unexpected kind %v with ident %s", val.Kind(), expr.Sel.Name)
	}
	rv, err := walkExpr(val, expr.X)
	if err != nil {
		return zeroRValue, err
	}
	rv = rv.FieldByName(expr.Sel.Name)
	return rv, nil
}
