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
	"strconv"
)

var (
	ErrNilViewContext = errors.New("nil view context")
	ErrNilExprNode    = errors.New("nil expr node")
	ErrNotValid       = errors.New("value not valid")

	zeroRValue reflect.Value
)

// 在`this`上，返回其`expr`对应的值
func EvalView(this interface{}, expr string) (result interface{}, err error) {
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
	var rv = reflect.ValueOf(this)
	rv, err = walkExpr(rv, node)
	if err == nil && rv.IsValid() {
		result = rv.Interface()
	} else {
		result = nil
	}
	return
}

// 在`this`上，设置v到对应`expr`
func EvalSet(this interface{}, expr string, v interface{}) error {
	if this == nil {
		if expr != "" {
			return ErrNilViewContext
		}
		return nil
	}
	var lhv = reflect.ValueOf(this)
	if lhv.Kind() == reflect.Ptr {
		lhv = lhv.Elem()
	}
	if !lhv.CanAddr() || !lhv.IsValid() {
		return ErrNotValid
	}
	var rhv = reflect.ValueOf(v)
	if !rhv.IsValid() {
		return ErrNotValid
	}

	node, err := parser.ParseExpr(expr)
	if err != nil {
		return err
	}
	switch n := node.(type) {
	case *ast.Ident:
		return setIdent(lhv, rhv, n)
	case *ast.SelectorExpr:
		return setSelectorExpr(lhv, rhv, n)
	case *ast.IndexExpr:
		return setIndexExpr(lhv, rhv, n)
	default:
		return fmt.Errorf("unexpected expr node %T", node)
	}
}

// 在`this`上， 删除对应`expr`的节点，只有map和slice可以有删除操作
func EvalRemove(this interface{}, expr string) error {
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return err
	}
	if node == nil {
		return ErrNilExprNode
	}
	var rv = reflect.ValueOf(this)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if !rv.CanAddr() || !rv.IsValid() {
		return ErrNotValid
	}
	switch n := node.(type) {
	case *ast.IndexExpr:
		return removeIndexExpr(rv, n)
	default:
		return fmt.Errorf("unexpected expr node %T", node)
	}
}

func walkExpr(rv reflect.Value, node ast.Expr) (reflect.Value, error) {
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if !rv.IsValid() {
		return zeroRValue, ErrNotValid
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
		i, err := strconv.Atoi(index.Value)
		if err != nil {
			return zeroRValue, fmt.Errorf("cannot index by key %s: %w", index.Value, err)
		}
		if i >= 0 && i < rv.Len() {
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

func setIdent(lhv, rhv reflect.Value, ident *ast.Ident) error {
	switch lhv.Kind() {
	case reflect.Struct:
		var field = lhv.FieldByName(ident.Name)
		if !field.IsValid() || !field.CanAddr() {
			return ErrNotValid
		}
		setEValue(field, rhv)
	case reflect.Map:
		if key, err := createMapKey(lhv, ident.Name); err != nil {
			return fmt.Errorf("cannot index map key %s: %w", ident.Name, err)
		} else {
			lhv.SetMapIndex(key, rhv)
		}
	default:
		return fmt.Errorf("unexpected kind %v with ident %s", lhv.Kind(), ident.Name)
	}
	return nil
}

func setSelectorExpr(lhv, rhv reflect.Value, expr *ast.SelectorExpr) error {
	if lhv.Kind() != reflect.Struct {
		return fmt.Errorf("unexpected kind %v with ident %s", lhv.Kind(), expr.Sel.Name)
	}
	rv, err := walkExpr(lhv, expr.X)
	if err != nil {
		return err
	}
	if !rv.CanAddr() {
		return ErrNotValid
	}
	var field = rv.FieldByName(expr.Sel.Name)
	if !field.IsValid() || !field.CanAddr() {
		return ErrNotValid
	}
	setEValue(field, rhv)
	return nil
}

func setIndexExpr(lhv, rhv reflect.Value, expr *ast.IndexExpr) error {
	index, ok := expr.Index.(*ast.BasicLit)
	if !ok {
		return fmt.Errorf("index is not literal")
	}
	rv, err := walkExpr(lhv, expr.X)
	if err != nil {
		return err
	}
	if !rv.CanAddr() {
		return ErrNotValid
	}
	switch rv.Kind() {
	case reflect.Slice:
		idx, err := strconv.Atoi(index.Value)
		if err != nil {
			return fmt.Errorf("cannot index by key %s: %w", index.Value, err)
		}
		setEValue(rv.Index(idx), rhv)

	case reflect.Map:
		if key, err := createMapKey(rv, index.Value); err != nil {
			return fmt.Errorf("cannot index map key %s: %w", index.Value, err)
		} else {
			rv.SetMapIndex(key, rhv)
		}
	default:
		return fmt.Errorf("unexpected kind %v with ident %s", lhv.Kind(), expr.Index)
	}
	return nil
}

func removeIndexExpr(val reflect.Value, expr *ast.IndexExpr) error {
	index, ok := expr.Index.(*ast.BasicLit)
	if !ok {
		return fmt.Errorf("index is not literal")
	}
	rv, err := walkExpr(val, expr.X)
	if err != nil {
		return err
	}
	if !rv.CanAddr() {
		return ErrNotValid
	}
	switch rv.Kind() {
	case reflect.Slice:
		idx, err := strconv.Atoi(index.Value)
		if err != nil {
			return fmt.Errorf("cannot index by key %s: %w", index.Value, err)
		}
		var sliceLen = rv.Len()
		if idx < 0 || idx >= sliceLen {
			return fmt.Errorf("slice index %s out of range", index.Value)
		}
		// 删除slice是通过先new一个新slice，然后把老的值赋到新slice里
		var newSlice = reflect.MakeSlice(rv.Type(), sliceLen-1, rv.Cap())
		var j = 0
		for i := 0; i < sliceLen; i++ {
			if i != idx {
				newSlice.Index(j).Set(rv.Index(i))
				j++
			}
		}
		rv.Set(newSlice)

	case reflect.Map:
		if key, err := createMapKey(rv, index.Value); err != nil {
			return fmt.Errorf("cannot index map key %s: %w", index.Value, err)
		} else {
			rv.SetMapIndex(key, reflect.Value{})
		}
	default:
		return fmt.Errorf("unexpected kind %v with ident %s", val.Kind(), expr.Index)
	}
	return nil
}

func createMapKey(rv reflect.Value, value string) (reflect.Value, error) {
	var keyType = rv.Type().Key()
	if !IsPrimitive(keyType.Kind()) {
		return zeroRValue, fmt.Errorf("cannot address map key %s %s", value, keyType.Name())
	}
	return CreatePrimitiveValue(keyType, value)
}

func setEValue(lhv, rhv reflect.Value) {
	ltype, rtype := lhv.Type(), rhv.Type()
	if rtype.ConvertibleTo(ltype) {
		v := rhv.Convert(ltype)
		lhv.Set(v)
	} else {
		lhv.Set(rhv) // raw set, may panic if type mismatch
	}
}
