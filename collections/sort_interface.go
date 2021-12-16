// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

//
// 常用基本类型的sort.Interface wrapper
// hand-made generics，全手工打造
//

type Int8Slice []int8

func (x Int8Slice) Len() int           { return len(x) }
func (x Int8Slice) Less(i, j int) bool { return x[i] < x[j] }
func (x Int8Slice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Uint8Slice []uint8

func (x Uint8Slice) Len() int           { return len(x) }
func (x Uint8Slice) Less(i, j int) bool { return x[i] < x[j] }
func (x Uint8Slice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Int16Slice []uint16

func (x Int16Slice) Len() int           { return len(x) }
func (x Int16Slice) Less(i, j int) bool { return x[i] < x[j] }
func (x Int16Slice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Uint16Slice []uint16

func (x Uint16Slice) Len() int           { return len(x) }
func (x Uint16Slice) Less(i, j int) bool { return x[i] < x[j] }
func (x Uint16Slice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Int32Slice []int32

func (x Int32Slice) Len() int           { return len(x) }
func (x Int32Slice) Less(i, j int) bool { return x[i] < x[j] }
func (x Int32Slice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Uint32Slice []uint32

func (x Uint32Slice) Len() int           { return len(x) }
func (x Uint32Slice) Less(i, j int) bool { return x[i] < x[j] }
func (x Uint32Slice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type IntSlice []int

func (x IntSlice) Len() int           { return len(x) }
func (x IntSlice) Less(i, j int) bool { return x[i] < x[j] }
func (x IntSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type UintSlice []uint

func (x UintSlice) Len() int           { return len(x) }
func (x UintSlice) Less(i, j int) bool { return x[i] < x[j] }
func (x UintSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Int64Slice []int64

func (x Int64Slice) Len() int           { return len(x) }
func (x Int64Slice) Less(i, j int) bool { return x[i] < x[j] }
func (x Int64Slice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Uint64Slice []uint64

func (x Uint64Slice) Len() int           { return len(x) }
func (x Uint64Slice) Less(i, j int) bool { return x[i] < x[j] }
func (x Uint64Slice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Float32Slice []float32

func (x Float32Slice) Len() int           { return len(x) }
func (x Float32Slice) Less(i, j int) bool { return x[i] < x[j] }
func (x Float32Slice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Float64Slice []float64

func (x Float64Slice) Len() int           { return len(x) }
func (x Float64Slice) Less(i, j int) bool { return x[i] < x[j] }
func (x Float64Slice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type StringSlice []string

func (x StringSlice) Len() int           { return len(x) }
func (x StringSlice) Less(i, j int) bool { return x[i] < x[j] }
func (x StringSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
