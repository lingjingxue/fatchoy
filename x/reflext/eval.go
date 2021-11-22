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

//
type EvalContext struct {
	this interface{}
	expr string
}

func NewEvalContext(this interface{}) *EvalContext {
	return &EvalContext{
		this: this,
	}
}

func (c *EvalContext) walkExprGet(ctx reflect.Value, node ast.Expr) (reflect.Value, error) {
	if !ctx.IsValid() {
		return zeroRValue, ErrNotValid
	}
	if node == nil {
		return zeroRValue, ErrNilExprNode
	}
	//fmt.Printf("context: %T\n", ctx.Interface())
	switch n := node.(type) {
	case *ast.Ident:
		return c.evalIdentGet(ctx, n)
	case *ast.IndexExpr:
		return c.evalIndexGet(ctx, n)
	case *ast.CallExpr:
		return c.evalCall(ctx, n)
	case *ast.SelectorExpr:
		return c.evalSelectorGet(ctx, n)
	default:
		return zeroRValue, fmt.Errorf("unexpected expr node %T", node)
	}
}

// 常量只能是寻址struct的field和map的key
func (c *EvalContext) evalIdentGet(ctx reflect.Value, ident *ast.Ident) (reflect.Value, error) {
	if ctx.Kind() == reflect.Ptr {
		ctx = ctx.Elem()
	}
	switch ctx.Kind() {
	case reflect.Struct:
		return ctx.FieldByName(ident.Name), nil
	case reflect.Map:
		return ctx.MapIndex(reflect.ValueOf(ident.Name)), nil
	default:
		return zeroRValue, fmt.Errorf("unexpected kind %v with ident %s", ctx.Kind(), ident.Name)
	}
}

func createMapKey(rv reflect.Value, value string) (reflect.Value, error) {
	var keyType = rv.Type().Key()
	if !IsPrimitive(keyType.Kind()) {
		return zeroRValue, fmt.Errorf("cannot address map key %s %s", value, keyType.Name())
	}
	return CreatePrimitiveValue(keyType, value)
}

// 只有数组、切片、字符串和map支持`[]`操作符
func (c *EvalContext) evalIndexGet(ctx reflect.Value, expr *ast.IndexExpr) (reflect.Value, error) {
	index, ok := expr.Index.(*ast.BasicLit)
	if !ok {
		return zeroRValue, fmt.Errorf("index is not literal")
	}
	rv, err := c.walkExprGet(ctx, expr.X)
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
		key, err := createMapKey(rv, index.Value)
		if err != nil {
			return zeroRValue, err
		}
		return rv.MapIndex(key), nil
	default:
		return zeroRValue, fmt.Errorf("unexpected kind %v with ident %s", ctx.Kind(), expr.Index)
	}
	return zeroRValue, nil
}

func callGetter(fn reflect.Value, name string) (reflect.Value, error) {
	var method = fn.Type()
	if method.NumIn() != 0 || method.NumOut() != 1 {
		return zeroRValue, fmt.Errorf("method %s signature not match", name)
	}
	var output = fn.Call(nil)
	return output[0], nil
}

// 支持简单的单返回值函数调用
func (c *EvalContext) evalCall(ctx reflect.Value, call *ast.CallExpr) (reflect.Value, error) {
	if len(call.Args) > 0 || call.Ellipsis > 0 {
		return zeroRValue, fmt.Errorf("unexpected call expr")
	}
	switch expr := call.Fun.(type) {
	case *ast.Ident:
		var fn = ctx.MethodByName(expr.Name)
		if fn.IsValid() {
			return callGetter(fn, expr.Name)
		} else {
			return zeroRValue, fmt.Errorf("method %s not valid", expr.Name)
		}
	case *ast.SelectorExpr:
		var kind = ctx.Kind()
		if kind == reflect.Ptr {
			kind = ctx.Elem().Kind()
		}
		if kind != reflect.Struct {
			return zeroRValue, fmt.Errorf("unexpected kind %v with ident %s", ctx.Kind(), expr.Sel.Name)
		}
		val, err := c.walkExprGet(ctx, expr.X)
		if err != nil {
			return zeroRValue, err
		}
		var fn = val.MethodByName(expr.Sel.Name)
		if fn.IsValid() {
			return callGetter(fn, expr.Sel.Name)
		} else {
			return zeroRValue, fmt.Errorf("method %s not valid", expr.Sel.Name)
		}
	default:
		return zeroRValue, fmt.Errorf("unexpect call expr %T", call.Fun)
	}
}

// 选择表达式
func (c *EvalContext) evalSelectorGet(ctx reflect.Value, expr *ast.SelectorExpr) (reflect.Value, error) {
	var kind = ctx.Kind()
	if kind == reflect.Ptr {
		kind = ctx.Elem().Kind()
	}
	if kind != reflect.Struct {
		return zeroRValue, fmt.Errorf("unexpected kind %v with ident %s", ctx.Kind(), expr.Sel.Name)
	}
	val, err := c.walkExprGet(ctx, expr.X)
	if err != nil {
		return zeroRValue, err
	}
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	var rv = val.FieldByName(expr.Sel.Name)
	return rv, nil
}

// 在`this`上，返回其`expr`对应的值
func (c *EvalContext) View(expr string) (interface{}, error) {
	c.expr = expr
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return zeroRValue, err
	}
	rv, err := c.walkExprGet(reflect.ValueOf(c.this), node)
	if err == nil && rv.IsValid() {
		return rv.Interface(), nil
	}
	return nil, err
}

// set b to a
func setValueTo(a, b reflect.Value) error {
	ta, tb := a.Type(), b.Type()
	if !tb.ConvertibleTo(ta) {
		return fmt.Errorf("type %v not convertible to %v", tb.Kind(), ta.Kind())
	}
	v := b.Convert(tb)
	a.Set(v)
	return nil
}

// a.X = b
func (c *EvalContext) setIdent(lhv, rhv reflect.Value, ident *ast.Ident) error {
	var kind = lhv.Kind()
	if kind == reflect.Ptr {
		if lhv.Elem().Kind() == reflect.Struct {
			lhv = lhv.Elem()
		}
	}
	switch lhv.Kind() {
	case reflect.Struct:
		var field = lhv.FieldByName(ident.Name)
		if field.IsValid() && field.CanSet() {
			return setValueTo(field, rhv)
		}
		return ErrNotValid

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

// 数组或者map的下标赋值，a[X] = b
func (c *EvalContext) setIndexExpr(lhv, rhv reflect.Value, expr *ast.IndexExpr) error {
	index, ok := expr.Index.(*ast.BasicLit)
	if !ok {
		return fmt.Errorf("index is not literal")
	}
	rv, err := c.walkExprGet(lhv, expr.X)
	if err != nil {
		return err
	}
	if !rv.CanSet() {
		return ErrNotValid
	}
	switch rv.Kind() {
	case reflect.Slice:
		idx, err := strconv.Atoi(index.Value)
		if err != nil {
			return fmt.Errorf("cannot index by key %s: %w", index.Value, err)
		}
		return setValueTo(rv.Index(idx), rhv)

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

func (c *EvalContext) setSelector(lhv, rhv reflect.Value, expr *ast.SelectorExpr) error {
	if lhv.Kind() == reflect.Ptr {
		if lhv.Elem().Kind() == reflect.Struct {
			lhv = lhv.Elem()
		}
	}
	if lhv.Kind() != reflect.Struct {
		return fmt.Errorf("unexpected kind %v with ident %s", lhv.Kind(), expr.Sel.Name)
	}
	rv, err := c.walkExprGet(lhv, expr.X)
	if err != nil {
		return err
	}
	if !rv.CanSet() {
		return ErrNotValid
	}
	var field = rv.FieldByName(expr.Sel.Name)
	if !field.IsValid() || !field.CanAddr() {
		return ErrNotValid
	}
	return setValueTo(field, rhv)
}

// 在`this`上，设置v到对应`expr`
func (c *EvalContext) Set(expr string, v interface{}) error {
	c.expr = expr
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return err
	}
	var lhv = reflect.ValueOf(c.this)
	var rhv = reflect.ValueOf(v)
	switch n := node.(type) {
	case *ast.Ident:
		return c.setIdent(lhv, rhv, n)
	case *ast.IndexExpr:
		return c.setIndexExpr(lhv, rhv, n)
	case *ast.SelectorExpr:
		return c.setSelector(lhv, rhv, n)
	default:
		return fmt.Errorf("unexpected expr node %T", node)
	}
}

// 删除操作，array, slice, map
func (c *EvalContext) removeIndex(ctx reflect.Value, expr *ast.IndexExpr) error {
	index, ok := expr.Index.(*ast.BasicLit)
	if !ok {
		return fmt.Errorf("index is not literal")
	}
	rv, err := c.walkExprGet(ctx, expr.X)
	if err != nil {
		return err
	}
	if !rv.CanAddr() {
		return ErrNotValid
	}
	switch rv.Kind() {
	case reflect.Array:
		idx, err := strconv.Atoi(index.Value)
		if err != nil {
			return fmt.Errorf("cannot index by key %s: %w", index.Value, err)
		}
		var arrLen = rv.Len()
		if idx < 0 || idx >= arrLen {
			return fmt.Errorf("array index %s out of range", index.Value)
		}
		var dummy = reflect.New(rv.Elem().Type())
		rv.Index(idx).Set(dummy)

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
		return fmt.Errorf("unexpected kind %v with ident %s", ctx.Kind(), expr.Index)
	}
	return nil
}

// delete a[X]
func (c *EvalContext) Delete(expr string) error {
	c.expr = expr
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return err
	}
	if node == nil {
		return ErrNilExprNode
	}
	var rv = reflect.ValueOf(c.this)
	if !rv.CanAddr() || !rv.IsValid() {
		return ErrNotValid
	}
	switch n := node.(type) {
	case *ast.IndexExpr:
		return c.removeIndex(rv, n)
	default:
		return fmt.Errorf("unexpected expr node %T", node)
	}
}
