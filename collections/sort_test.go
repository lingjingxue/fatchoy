// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

import (
	"fmt"
	"sort"
	"testing"
)

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
