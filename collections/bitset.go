// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/bits"
	"strings"
)

// fixed bitset
type BitSet struct {
	size int
	bits []uint64
}

func BitSetFrom(array []uint64, size int) *BitSet {
	if size <= 0 {
		size = len(array) * 64
	}
	return &BitSet{
		size: size,
		bits: array,
	}
}

func NewBitSet(bitsize int) *BitSet {
	var n = bitsize / 64
	if bitsize%64 > 0 {
		n++
	}
	return &BitSet{
		size: bitsize,
		bits: make([]uint64, n),
	}
}

// bit的数量
func (bs *BitSet) Size() int {
	return bs.size
}

// 设置bits[i]为1
func (bs *BitSet) Set(i int) bool {
	if i >= 0 && i < bs.size {
		var v = uint64(1) << (i % 64)
		bs.bits[i/64] |= v
		return true
	}
	return false
}

// 反转第i位
func (bs *BitSet) Flip(i int) bool {
	if i >= 0 && i < bs.size {
		bs.bits[i/64] ^= 1 << (i % 64)
		return true
	}
	return false
}

// 设置bits[i]为0
func (bs *BitSet) Clear(i int) bool {
	if i >= 0 && i < bs.size {
		var v = uint64(1) << (i % 64)
		bs.bits[i/64] &= ^v
		return true
	}
	return false
}

// 查看bits[i]是否为1
func (bs *BitSet) Test(i int) bool {
	if i >= 0 && i < bs.size {
		return bs.bits[i/64]&(1<<(i%64)) != 0
	}
	return false
}

// 清零所有位
func (bs *BitSet) ClearAll() {
	for i := 0; i < len(bs.bits); i++ {
		bs.bits[i] = 0
	}
}

// 置为1的位的数量
func (bs *BitSet) Count() int {
	var count = 0
	for i := 0; i < len(bs.bits); i++ {
		if bs.bits[i] > 0 {
			count += bits.OnesCount64(bs.bits[i])
		}
	}
	return count
}

func (bs BitSet) HashCode() string {
	h := md5.New()
	for i := 0; i < len(bs.bits); i++ {
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], bs.bits[i])
		h.Write(buf[:])
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (bs BitSet) FormattedString(width int) string {
	var sb strings.Builder
	var n = 0
	for i := 0; i < bs.size; i++ {
		if n%width == 0 {
			sb.WriteByte('\n')
		}
		n++
		if bs.Test(i) {
			sb.WriteByte('1')
		} else {
			sb.WriteByte('0')
		}
	}
	return sb.String()
}

func (bs BitSet) String() string {
	var sb strings.Builder
	sb.Grow(bs.size)
	for i := 0; i < bs.size; i++ {
		if bs.Test(i) {
			sb.WriteByte('1')
		} else {
			sb.WriteByte('0')
		}
	}
	return sb.String()
}
