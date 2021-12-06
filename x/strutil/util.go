// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package strutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"unicode"
	"unsafe"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-=~!@#$%^&*()_+[]\\;',./{}|:<>?"

// 对[]byte的修改会影响到返回的string
func BytesAsString(b []byte) string {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := reflect.StringHeader{Data: bh.Data, Len: bh.Len}
	return *(*string)(unsafe.Pointer(&sh))
}

// 注意：修改返回的[]byte会引起panic
func StringAsBytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{Data: sh.Data, Len: sh.Len, Cap: sh.Len}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

// 随机长度的字符串
func RandString(length int) string {
	if length <= 0 {
		return ""
	}
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		idx := rand.Int() % len(alphabet)
		result[i] = alphabet[idx]
	}
	return string(result)
}

// 随机长度的字节数组
func RandBytes(length int) []byte {
	if length <= 0 {
		return nil
	}
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		ch := uint8(rand.Int31() % 0xFF)
		result[i] = ch
	}
	return result
}

// 在array中查找string
func FindStringInArray(a []string, x string) int {
	for i, v := range a {
		if v == x {
			return i
		}
	}
	return -1
}

// 查找第一个数字的位置
func FindFirstDigit(s string) int {
	for i, r := range s {
		if unicode.IsDigit(r) {
			return i
		}
	}
	return -1
}

// 查找第一个非数字的位置
func FindFirstNonDigit(s string) int {
	for i, r := range s {
		if !unicode.IsDigit(r) {
			return i
		}
	}
	return -1
}

// 反转字符串
func Reverse(str string) string {
	runes := []rune(str)
	l := len(runes)
	for i := 0; i < l/2; i++ {
		runes[i], runes[l-i-1] = runes[l-i-1], runes[i]
	}
	return string(runes)
}

// 打印容量大小
func PrettyBytes(n int64) string {
	if n < (1 << 10) {
		return fmt.Sprintf("%dB", n)
	} else if n < (1 << 20) {
		return fmt.Sprintf("%.2fKB", float64(n)/(1<<10))
	} else if n < (1 << 30) {
		return fmt.Sprintf("%.2fMB", float64(n)/(1<<20))
	} else if n < (1 << 40) {
		return fmt.Sprintf("%.2fGB", float64(n)/(1<<30))
	} else {
		return fmt.Sprintf("%.2fTB", float64(n)/(1<<40))
	}
}

// 字符串最长共同前缀
func LongestCommonPrefix(s1, s2 string) string {
	if s1 == "" || s2 == "" {
		return ""
	}
	i := 0
	for i < len(s1) && i < len(s2) {
		if s1[i] == s2[i] {
			i++
			continue
		}
		break
	}
	return s1[:i]
}

// 使用bigint序列化大整数
func UnmarshalJSON(data []byte, v interface{}) error {
	if len(data) > 0 {
		var dec = json.NewDecoder(bytes.NewReader(data))
		dec.UseNumber()
		return dec.Decode(v)
	}
	return nil
}
