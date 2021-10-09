// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package reflectutil

import (
	"bytes"
	"encoding/gob"
	"log"
	"reflect"
)

// 获取struct内所有field的值
func GetStructAllFieldValues(ptr interface{}) []interface{} {
	var value = reflect.ValueOf(ptr).Elem()
	var st = value.Type()
	var result = make([]interface{}, 0, st.NumField())
	for i := 0; i < st.NumField(); i++ {
		var field = value.Field(i)
		result = append(result, field.Interface())
	}
	return result
}

func GetStructFieldValues(ptr interface{}, except string) []interface{} {
	var value = reflect.ValueOf(ptr).Elem()
	var st = value.Type()
	var result = make([]interface{}, 0, st.NumField())
	for i := 0; i < st.NumField(); i++ {
		var field = value.Field(i)
		result = append(result, field.Interface())
	}
	return result
}

// 获取struct内指定field的值
func GetStructFieldValuesBy(ptr interface{}, names []string) []interface{} {
	var value = reflect.ValueOf(ptr).Elem()
	var st = value.Type()
	var result = make([]interface{}, 0, len(names))
	for _, fname := range names {
		_, ok := st.FieldByName(fname)
		if !ok {
			log.Panicf("field %s.%s not found", st.Name(), fname)
			return nil
		}
		field := value.FieldByName(fname)
		result = append(result, field.Interface())
	}
	return result
}

// 调用struct内指定名称的函数
func CallObjectMethod(ptr interface{}, method string, args ...interface{}) []interface{} {
	value := reflect.ValueOf(ptr)
	fn := value.MethodByName(method)
	var input []reflect.Value
	for _, arg := range args {
		input = append(input, reflect.ValueOf(arg))
	}
	output := fn.Call(input)
	var result []interface{}
	for _, out := range output {
		result = append(result, out.Interface())
	}
	return result
}

// 深拷贝src到dst，内部实现是使用gob编码再解码
func DeepCopy(src, dst interface{}) error {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(src); err != nil {
		return err
	}
	dec := gob.NewDecoder(buf)
	return dec.Decode(dst)
}
