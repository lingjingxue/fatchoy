// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package mathext

import (
	"math"
)

type (
	I8   int8
	I16  int16
	I32  int32
	I64  int64
	INT  int
	U8   uint8
	U16  uint16
	U32  uint32
	U64  uint64
	UINT uint
	F32  float32
	F64  float64
)

var (
	Int8    = new(I8)
	Int16   = new(I16)
	Int32   = new(I32)
	Int64   = new(I64)
	Int     = new(INT)
	UInt8   = new(U8)
	UInt16  = new(U16)
	UInt32  = new(U32)
	UInt64  = new(U64)
	UInt    = new(UINT)
	Float32 = new(F32)
	Float64 = new(F64)
)

func (I8) Abs(x int8) int8 {
	if x < 0 {
		return -x
	}
	if x == 0 {
		return 0
	}
	return x
}

func (I8) Max(x, y int8) int8 {
	if x < y {
		return y
	}
	return x
}

func (I8) Min(x, y int8) int8 {
	if y < x {
		return y
	}
	return x
}

func (a I8) Dim(x, y int8) int8 {
	return a.Max(x-y, 0)
}

/////////////////////////////////////////////

func (I16) Abs(x int16) int16 {
	if x < 0 {
		return -x
	}
	if x == 0 {
		return 0
	}
	return x
}

func (I16) Max(x, y int16) int16 {
	if x < y {
		return y
	}
	return x
}

func (I16) Min(x, y int16) int16 {
	if y < x {
		return y
	}
	return x
}

func (a I16) Dim(x, y int16) int16 {
	return a.Max(x-y, 0)
}

/////////////////////////////////////////////

func (I32) Abs(x int32) int32 {
	if x < 0 {
		return -x
	}
	if x == 0 {
		return 0
	}
	return x
}

func (I32) Max(x, y int32) int32 {
	if x < y {
		return y
	}
	return x
}

func (I32) Min(x, y int32) int32 {
	if y < x {
		return y
	}
	return x
}

func (a I32) Dim(x, y int32) int32 {
	return a.Max(x-y, 0)
}

/////////////////////////////////////////////

func (I64) Abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	if x == 0 {
		return 0
	}
	return x
}

func (I64) Max(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}

func (I64) Min(x, y int64) int64 {
	if y < x {
		return y
	}
	return x
}

func (a I64) Dim(x, y int64) int64 {
	return a.Max(x-y, 0)
}

/////////////////////////////////////////////

func (INT) Abs(x int) int {
	if x < 0 {
		return -x
	}
	if x == 0 {
		return 0
	}
	return x
}

func (INT) Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func (INT) Min(x, y int) int {
	if y < x {
		return y
	}
	return x
}

func (a INT) Dim(x, y int) int {
	return a.Max(x-y, 0)
}

/////////////////////////////////////////////

func (U8) Abs(x uint8) uint8 {
	if x < 0 {
		return -x
	}
	if x == 0 {
		return 0
	}
	return x
}

func (U8) Max(x, y uint8) uint8 {
	if x < y {
		return y
	}
	return x
}

func (U8) Min(x, y uint8) uint8 {
	if y < x {
		return y
	}
	return x
}

func (a U8) Dim(x, y uint8) uint8 {
	return a.Max(x-y, 0)
}

/////////////////////////////////////////////

func (U16) Abs(x uint16) uint16 {
	if x < 0 {
		return -x
	}
	if x == 0 {
		return 0
	}
	return x
}

func (U16) Max(x, y uint16) uint16 {
	if x < y {
		return y
	}
	return x
}

func (U16) Min(x, y uint16) uint16 {
	if y < x {
		return y
	}
	return x
}

func (a U16) Dim(x, y uint16) uint16 {
	return a.Max(x-y, 0)
}

/////////////////////////////////////////////

func (U32) Abs(x uint32) uint32 {
	if x < 0 {
		return -x
	}
	if x == 0 {
		return 0
	}
	return x
}

func (U32) Max(x, y uint32) uint32 {
	if x < y {
		return y
	}
	return x
}

func (U32) Min(x, y uint32) uint32 {
	if y < x {
		return y
	}
	return x
}

func (a U32) Dim(x, y uint32) uint32 {
	return a.Max(x-y, 0)
}

/////////////////////////////////////////////

func (U64) Abs(x uint64) uint64 {
	if x < 0 {
		return -x
	}
	if x == 0 {
		return 0
	}
	return x
}

func (U64) Max(x, y uint64) uint64 {
	if x < y {
		return y
	}
	return x
}

func (U64) Min(x, y uint64) uint64 {
	if y < x {
		return y
	}
	return x
}

func (a U64) Dim(x, y uint64) uint64 {
	return a.Max(x-y, 0)
}

/////////////////////////////////////////////

func (UINT) Abs(x uint) uint {
	if x < 0 {
		return -x
	}
	if x == 0 {
		return 0
	}
	return x
}

func (UINT) Max(x, y uint) uint {
	if x < y {
		return y
	}
	return x
}

func (UINT) Min(x, y uint) uint {
	if y < x {
		return y
	}
	return x
}

func (a UINT) Dim(x, y uint) uint {
	return a.Max(x-y, 0)
}

/////////////////////////////////////////////

func (F32) Abs(x float32) float32 {
	return float32(math.Abs(float64(x)))
}

func (F32) Max(x, y float32) float32 {
	return float32(math.Max(float64(x), float64(y)))
}

func (F32) Min(x, y float32) float32 {
	return float32(math.Min(float64(x), float64(y)))
}

func (a F32) Dim(x, y float32) float32 {
	return float32(math.Dim(float64(x), float64(y)))
}

func (a F32) SafeDiv(x, y float32) float32 {
	if y == 0 {
		return 0
	}
	return x / y
}

/////////////////////////////////////////////

func (F64) Abs(x float64) float64 {
	return math.Abs(x)
}

func (F64) Max(x, y float64) float64 {
	return math.Max(x, y)
}

func (F64) Min(x, y float64) float64 {
	return math.Min(x, y)
}

func (a F64) Dim(x, y float64) float64 {
	return math.Dim(x, y)
}

func (a F64) SafeDiv(x, y float64) float64 {
	if y == 0 {
		return 0
	}
	return x / y
}
