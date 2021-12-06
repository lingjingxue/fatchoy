// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package strutil

var b62Alphabet = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var b62IndexTable = buildIndexTable(b62Alphabet)

// build alphabet index
func buildIndexTable(s []byte) map[byte]int64 {
	var table = make(map[byte]int64, len(s))
	for i := 0; i < len(s); i++ {
		table[s[i]] = int64(i)
	}
	return table
}

// 编码Base62
func EncodeBase62String(id int64) string {
	if id == 0 {
		return string(b62Alphabet[:1])
	}
	var short = make([]byte, 0, 12)
	for id > 0 {
		var rem = id % 62
		id = id / 62
		short = append(short, b62Alphabet[rem])
	}
	// reverse
	for i, j := 0, len(short)-1; i < j; i, j = i+1, j-1 {
		short[i], short[j] = short[j], short[i]
	}
	return string(short)
}

// 解码Base62
func DecodeBase62String(s string) int64 {
	var n int64
	for i := 0; i < len(s); i++ {
		n = (n * 62) + b62IndexTable[s[i]]
	}
	return n
}
