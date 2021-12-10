// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

//
// 基本数值类型的sort.Interface实现
//

type Int8Array []int8

func (x Int8Array) Len() int           { return len(x) }
func (x Int8Array) Less(i, j int) bool { return x[i] < x[j] }
func (x Int8Array) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Uint8Array []uint8

func (x Uint8Array) Len() int           { return len(x) }
func (x Uint8Array) Less(i, j int) bool { return x[i] < x[j] }
func (x Uint8Array) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Int16Array []uint16

func (x Int16Array) Len() int           { return len(x) }
func (x Int16Array) Less(i, j int) bool { return x[i] < x[j] }
func (x Int16Array) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Uint16Array []uint16

func (x Uint16Array) Len() int           { return len(x) }
func (x Uint16Array) Less(i, j int) bool { return x[i] < x[j] }
func (x Uint16Array) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Int32Array []int32

func (x Int32Array) Len() int           { return len(x) }
func (x Int32Array) Less(i, j int) bool { return x[i] < x[j] }
func (x Int32Array) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Uint32Array []uint32

func (x Uint32Array) Len() int           { return len(x) }
func (x Uint32Array) Less(i, j int) bool { return x[i] < x[j] }
func (x Uint32Array) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type IntArray []int

func (x IntArray) Len() int           { return len(x) }
func (x IntArray) Less(i, j int) bool { return x[i] < x[j] }
func (x IntArray) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type UintArray []uint

func (x UintArray) Len() int           { return len(x) }
func (x UintArray) Less(i, j int) bool { return x[i] < x[j] }
func (x UintArray) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Int64Array []int64

func (x Int64Array) Len() int           { return len(x) }
func (x Int64Array) Less(i, j int) bool { return x[i] < x[j] }
func (x Int64Array) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Uint64Array []uint64

func (x Uint64Array) Len() int           { return len(x) }
func (x Uint64Array) Less(i, j int) bool { return x[i] < x[j] }
func (x Uint64Array) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Float32Array []float32

func (x Float32Array) Len() int           { return len(x) }
func (x Float32Array) Less(i, j int) bool { return x[i] < x[j] }
func (x Float32Array) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Float64Array []float64

func (x Float64Array) Len() int           { return len(x) }
func (x Float64Array) Less(i, j int) bool { return x[i] < x[j] }
func (x Float64Array) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type StringArray []string

func (x StringArray) Len() int           { return len(x) }
func (x StringArray) Less(i, j int) bool { return x[i] < x[j] }
func (x StringArray) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
