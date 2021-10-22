// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
)

const (
	is64Bit = uint64(^uintptr(0)) == ^uint64(0)
)

var ErrBufferOutOfRange = errors.New("buffer out of range")

type Buffer struct {
	bytes.Buffer
}

func (b *Buffer) WriteBool(v bool) {
	var c byte = 0
	if v {
		c = 1
	}
	b.WriteByte(c)
}

func (b *Buffer) WriteUInt8(n uint8) {
	b.WriteByte(n)
}

func (b *Buffer) WriteInt8(n int8) {
	b.WriteByte(byte(n))
}

func (b *Buffer) WriteUint16(n uint16) {
	var tmp [2]byte
	binary.LittleEndian.PutUint16(tmp[:], n)
	b.Write(tmp[:])
}

func (b *Buffer) WriteInt16(n int16) {
	b.WriteUint16(uint16(n))
}

func (b *Buffer) WriteUint32(n uint32) {
	var tmp [4]byte
	binary.LittleEndian.PutUint32(tmp[:], n)
	b.Write(tmp[:])
}

func (b *Buffer) WriteInt32(n int32) {
	b.WriteUint32(uint32(n))
}

func (b *Buffer) WriteUint64(n uint64) {
	var tmp [8]byte
	binary.LittleEndian.PutUint64(tmp[:], n)
	b.Write(tmp[:])
}

func (b *Buffer) WriteInt64(n int64) {
	b.WriteUint64(uint64(n))
}

func (b *Buffer) WriteUint(n uint) {
	if is64Bit {
		b.WriteUint64(uint64(n))
	}
	b.WriteUint32(uint32(n))
}

func (b *Buffer) WriteInt(n int) {
	if is64Bit {
		b.WriteInt64(int64(n))
	}
	b.WriteInt32(int32(n))
}

func (b *Buffer) WriteFloat32(f float32) {
	var n = math.Float32bits(f)
	b.WriteUint32(n)
}

func (b *Buffer) WriteFloat64(f float64) {
	var n = math.Float64bits(f)
	b.WriteUint64(n)
}

func (b *Buffer) ReadBool() bool {
	return b.ReadInt8() != 0
}

func (b *Buffer) ReadUint8() uint8 {
	c, err := b.ReadByte()
	if err != nil {
		panic(err)
	}
	return c
}

func (b *Buffer) ReadInt8() int8 {
	var c = b.ReadUint8()
	return int8(c)
}

func (b *Buffer) ReadUint16() uint16 {
	var tmp [2]byte
	if _, err := b.Read(tmp[:]); err != nil {
		panic(err)
	}
	return binary.LittleEndian.Uint16(tmp[:])
}

func (b *Buffer) ReadInt16() int16 {
	var n = b.ReadUint16()
	return int16(n)
}

func (b *Buffer) ReadUint32() uint32 {
	var tmp [4]byte
	if _, err := b.Read(tmp[:]); err != nil {
		panic(err)
	}
	return binary.LittleEndian.Uint32(tmp[:])
}

func (b *Buffer) ReadInt32() int32 {
	var n = b.ReadUint32()
	return int32(n)
}

func (b *Buffer) ReadUint64() uint64 {
	var tmp [8]byte
	if _, err := b.Read(tmp[:]); err != nil {
		panic(err)
	}
	return binary.LittleEndian.Uint64(tmp[:])
}

func (b *Buffer) ReadInt64() int64 {
	var n = b.ReadUint64()
	return int64(n)
}

func (b *Buffer) ReadUint() uint {
	if is64Bit {
		return uint(b.ReadUint64())
	}
	return uint(b.ReadUint32())
}

func (b *Buffer) ReadInt() int {
	if is64Bit {
		return int(b.ReadInt64())
	}
	return int(b.ReadInt32())
}

func (b *Buffer) ReadFloat32() float32 {
	var n = b.ReadUint32()
	return math.Float32frombits(n)
}

func (b *Buffer) ReadFloat64() float64 {
	var n = b.ReadUint64()
	return math.Float64frombits(n)
}

func (b *Buffer) PeekBool() bool {
	return b.PeekInt8() != 0
}

func (b *Buffer) PeekUint8() uint8 {
	var data = b.Bytes()
	if len(data) < 1 {
		panic(ErrBufferOutOfRange)
	}
	return data[0]
}

func (b *Buffer) PeekInt8() int8 {
	return int8(b.PeekUint8())
}

func (b *Buffer) PeekUint16() uint16 {
	var data = b.Bytes()
	if len(data) < 2 {
		panic(ErrBufferOutOfRange)
	}
	return binary.LittleEndian.Uint16(data[:2])
}

func (b *Buffer) PeekInt16() int16 {
	return int16(b.PeekUint16())
}

func (b *Buffer) PeekUint32() uint32 {
	var data = b.Bytes()
	if len(data) < 4 {
		panic(ErrBufferOutOfRange)
	}
	return binary.LittleEndian.Uint32(data[:4])
}

func (b *Buffer) PeekInt32() int32 {
	return int32(b.PeekUint32())
}

func (b *Buffer) PeekUint64() uint64 {
	var data = b.Bytes()
	if len(data) < 8 {
		panic(ErrBufferOutOfRange)
	}
	return binary.LittleEndian.Uint64(data[:8])
}

func (b *Buffer) PeekInt64() int64 {
	return int64(b.PeekUint64())
}

func (b *Buffer) PeekUint() uint {
	if is64Bit {
		return uint(b.PeekUint64())
	}
	return uint(b.PeekUint32())
}

func (b *Buffer) PeekInt() int {
	if is64Bit {
		return int(b.PeekInt64())
	}
	return int(b.PeekInt32())
}

func (b *Buffer) PeekFloat32() float32 {
	var n = b.PeekUint32()
	return math.Float32frombits(n)
}

func (b *Buffer) PeekFloat64() float64 {
	var n = b.PeekUint64()
	return math.Float64frombits(n)
}
