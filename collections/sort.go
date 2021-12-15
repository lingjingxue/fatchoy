// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

import (
	"sort"
)

// 插入排序
func InsertionSort(data sort.Interface) {
	var hi = data.Len()
	for i := 1; i < hi; i++ {
		for j := i; j > 0 && data.Less(j, j-1); j-- {
			data.Swap(j, j-1)
		}
	}
}

// siftDown implements the heap property on data[lo, hi).
// first is an offset into the array where the root of the heap lies.
func siftDown(data sort.Interface, lo, hi, first int) {
	root := lo
	for {
		child := 2*root + 1
		if child >= hi {
			break
		}
		if child+1 < hi && data.Less(first+child, first+child+1) {
			child++
		}
		if !data.Less(first+root, first+child) {
			return
		}
		data.Swap(first+root, first+child)
		root = child
	}
}

// 堆排序
func HeapSort(data sort.Interface) {
	first := 0
	lo := 0
	hi := data.Len()

	// Build heap with greatest element at top.
	for i := (hi - 1) / 2; i >= 0; i-- {
		siftDown(data, i, hi, first)
	}

	// Pop elements, largest first, into end of data.
	for i := hi - 1; i >= 0; i-- {
		data.Swap(first, first+i)
		siftDown(data, lo, i, first)
	}
}

/**
 * Sorts the specified portion of the specified array using a binary
 * insertion sort.  This is the best method for sorting small numbers
 * of elements.  It requires O(n log n) compares, but O(n^2) data
 * movement (worst case).
 *
 * If the initial part of the specified range is already sorted,
 * this method can take advantage of it: the method assumes that the
 * elements from index {@code lo}, inclusive, to {@code start},
 * exclusive are already sorted.
 *
 * @param a the array in which a range is to be sorted
 * @param lo the index of the first element in the range to be sorted
 * @param hi the index after the last element in the range to be sorted
 * @param start the index of the first element in the range that is
 *        not already known to be sorted ({@code lo <= start <= hi})
 */
func BinarySort(a []Comparable, lo, hi, start int) {
	assert(lo <= start || start <= hi)
	if start == lo {
		start++
	}
	for ; start < hi; start++ {
		var pivot = a[start]

		// Set left (and right) to the index where a[start] (pivot) belongs
		var left = lo
		var right = start
		assert(left <= right)
		/*
		 * Invariants:
		 *   pivot >= all in [lo, left).
		 *   pivot <  all in [right, start).
		 */
		for left < right {
			var mid = (left + right) / 2
			var cmp = pivot.CompareTo(a[mid])
			if cmp < 0 {
				right = mid
			} else {
				left = mid + 1
			}
		}
		assert(left == right)
		/*
		 * The invariants still hold: pivot >= all in [lo, left) and
		 * pivot < all in [left, start), so pivot belongs at left.  Note
		 * that if there are elements equal to pivot, left points to the
		 * first slot after them -- that's why this sort is stable.
		 * Slide elements over to make room for pivot.
		 */
		var n = start - left
		switch n {
		case 2:
			a[left+2] = a[left+1]
			fallthrough
		case 1:
			a[left+1] = a[left]
			break
		default:
			arraycopy(a, left, a, left+1, n)
		}
		a[left] = pivot
	}
}
