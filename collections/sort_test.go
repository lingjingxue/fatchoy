// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

import (
	"fmt"
	"sort"
	"testing"
)

type Integer int

func (n Integer) CompareTo(o Comparable) int {
	var n2 = o.(Integer)
	if n < n2 {
		return -1
	} else if n2 > n {
		return 1
	}
	return 0
}

func TestInsertionSort(t *testing.T) {
	s := []int{5, 2, 6, 3, 1, 4} // unsorted
	InsertionSort(sort.IntSlice(s))
	fmt.Println(s)
	// Output: [1 2 3 4 5 6]
}

func TestHeapSort(t *testing.T) {
	s := []int{5, 2, 6, 3, 1, 4} // unsorted
	HeapSort(sort.IntSlice(s))
	fmt.Println(s)
	// Output: [1 2 3 4 5 6]
}

func TestTimSort(t *testing.T) {
	a := []Comparable{Integer(5), Integer(2), Integer(6), Integer(3), Integer(1), Integer(4)} // unsorted
	TimSort(a)
	fmt.Println(a)
	// Output: [1 2 3 4 5 6]
}
