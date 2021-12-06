// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

//go:build !ignore

package strutil

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
)

func ParseBool(s string) bool {
	switch len(s) {
	case 0:
		return false
	case 1:
		return s[0] == '1'
	case 2:
		return s == "on" || s == "ON"
	case 3:
		return s == "yes" || s == "YES"
	case 4:
		return s == "true" || s == "TRUE"
	default:
		b, err := strconv.ParseBool(s)
		if err != nil {
			log.Printf("ParseBool: cannot pasre %s to boolean: %v", s, err)
		}
		return b
	}
}

func ParseI8(s string) (int8, error) {
	n, err := ParseI32(s)
	if err != nil {
		return 0, err
	}
	if n > math.MaxInt8 || n < math.MinInt8 {
		return 0, fmt.Errorf("ParseI8: value %s out of range", s)
	}
	return int8(n), nil
}

func MustParseI8(s string) int8 {
	n := MustParseI32(s)
	if n > math.MaxInt8 || n < math.MinInt8 {
		log.Panicf("MustParseI8: value %s out of range", s)
		return 0
	}
	return int8(n)
}

func ParseU8(s string) (uint8, error) {
	n, err := ParseI32(s)
	if err != nil {
		return 0, err
	}
	if n > math.MaxUint8 || n < 0 {
		return 0, fmt.Errorf("ParseU8: value %s out of range", s)
	}
	return uint8(n), nil
}

func MustParseU8(s string) uint8 {
	n := MustParseI32(s)
	if n > math.MaxUint8 || n < 0 {
		log.Panicf("MustParseU8: value %s out of range", s)
	}
	return uint8(n)
}

func ParseI16(s string) (int16, error) {
	n, err := ParseI32(s)
	if err != nil {
		return 0, err
	}
	if n > math.MaxInt16 || n < math.MinInt16 {
		return 0, fmt.Errorf("ParseI16: value %s out of range", s)
	}
	return int16(n), nil
}

func MustParseI16(s string) int16 {
	n := MustParseI32(s)
	if n > math.MaxInt16 || n < math.MinInt16 {
		log.Panicf("MustParseI16: value %s out of range", s)
	}
	return int16(n)
}

func ParseU16(s string) (uint16, error) {
	n, err := ParseI32(s)
	if err != nil {
		return 0, err
	}
	if n > math.MaxUint16 || n < 0 {
		return 0, fmt.Errorf("ParseU16: value %s out of range", s)
	}
	return uint16(n), nil
}

func MustParseU16(s string) uint16 {
	n := MustParseI32(s)
	if n > math.MaxUint16 || n < 0 {
		log.Panicf("MustParseU16: value %s out of range", s)
		return 0
	}
	return uint16(n)
}

func ParseI32(s string) (int32, error) {
	if s == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(s, 10, 32)
	return int32(n), err
}

func MustParseI32(s string) int32 {
	n, err := ParseI32(s)
	if err != nil {
		log.Panicf("MustParseI32: cannot parse [%s] to int32: %v", s, err)
		return 0
	}
	return n
}

func ParseU32(s string) (uint32, error) {
	if s == "" {
		return 0, nil
	}
	n, err := strconv.ParseUint(s, 10, 32)
	return uint32(n), err
}

func MustParseU32(s string) uint32 {
	n, err := ParseU32(s)
	if err != nil {
		log.Panicf("MustParseU32: cannot parse [%s] to uint32: %v", s, err)
		return 0
	}
	return uint32(n)
}

func ParseI64(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.ParseInt(s, 10, 64)
}

func MustParseI64(s string) int64 {
	n, err := ParseI64(s)
	if err != nil {
		log.Panicf("MustParseI64: cannot parse [%s] to uint64: %v", s, err)
		return 0
	}
	return n
}

func ParseU64(s string) (uint64, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.ParseUint(s, 10, 64)
}

func MustParseU64(s string) uint64 {
	n, err := ParseU64(s)
	if err != nil {
		log.Panicf("MustParseU64: cannot parse [%s] to uint64: %v", s, err)
		return 0
	}
	return n
}

func ParseF32(s string) (float32, error) {
	f, err := ParseF64(s)
	if err != nil {
		return 0, err
	}
	if f > math.MaxFloat32 || f < math.SmallestNonzeroFloat32 {
		return 0, fmt.Errorf("ParseF32: value %s out of range", s)
	}
	return float32(f), nil
}

func MustParseF32(s string) float32 {
	f, err := ParseF32(s)
	if err != nil {
		log.Panicf("MustParseF32: cannot parse [%s] to float", s)
		return 0
	}
	return f
}

func ParseF64(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

func MustParseF64(s string) float64 {
	f, err := ParseF64(s)
	if err != nil {
		log.Panicf("MustParseF64: cannot parse [%s] to double: %v", s, err)
		return 0
	}
	return f
}

const KVPairQuote = '\''

func unquote(s string) string {
	n := len(s)
	if n >= 2 {
		if s[0] == KVPairQuote && s[n-1] == KVPairQuote {
			return s[1 : n-1]
		}
	}
	return s
}

// 解析字符串为K-V map，
// 示例：s = "a='x,y',c=z"
// 		ParseSepKeyValues(s,",","=") ==> {"a":"x,y", "c":"z"}
func ParseKeyValuePairs(text string, sep, equal byte) map[string]string {
	var result = make(map[string]string)
	var key string
	var inQuote = false
	var start = 0
	for i := 0; i < len(text); i++ {
		var ch = text[i]
		switch ch {
		case sep:
			if !inQuote {
				value := strings.TrimSpace(text[start:i])
				if key == "" {
					key = value
					value = ""
				}
				result[key] = unquote(value)
				key = ""
				start = i + 1
			}
		case equal:
			if !inQuote {
				key = strings.TrimSpace(text[start:i])
				start = i + 1
			}
		case KVPairQuote:
			inQuote = !inQuote
		}
	}
	if start < len(text) || key != "" {
		s := strings.TrimSpace(text[start:])
		if key == "" {
			key = s
			s = ""
		}
		result[key] = unquote(s)
	}
	return result
}
