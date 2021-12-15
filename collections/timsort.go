// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

import (
	"log"
	"math/bits"
	"runtime"
)

/**
 * A stable, adaptive, iterative mergesort that requires far fewer than
 * n lg(n) comparisons when running on partially sorted arrays, while
 * offering performance comparable to a traditional mergesort when run
 * on random arrays.  Like all proper mergesorts, this sort is stable and
 * runs O(n log n) time (worst case).  In the worst case, this sort requires
 * temporary storage space for n/2 object references; in the best case,
 * it requires only a small constant amount of space.
 *
 * This implementation was adapted from Tim Peters's list sort for
 * Python, which is described in detail here:
 *
 *   http://svn.python.org/projects/python/trunk/Objects/listsort.txt
 *
 * Tim's C code may be found here:
 *
 *   http://svn.python.org/projects/python/trunk/Objects/listobject.c
 *
 * The underlying techniques are described in this paper (and may have
 * even earlier origins):
 *
 *  "Optimistic Sorting and Information Theoretic Complexity"
 *  Peter McIlroy
 *  SODA (Fourth Annual ACM-SIAM Symposium on Discrete Algorithms),
 *  pp 467-474, Austin, Texas, 25-27 January 1993.
 *
 * While the API to this class consists solely of static methods, it is
 * (privately) instantiable; a TimSort instance holds the state of an ongoing
 * sort, assuming the input array is large enough to warrant the full-blown
 * TimSort. Small arrays are sorted in place, using a binary insertion sort.
 *
 * @author Josh Bloch
 */

const (
	// This is the minimum sized sequence that will be merged.  Shorter
	// sequences will be lengthened by calling binarySort.  If the entire
	// array is less than this length, no merges will be performed.
	//
	// This constant should be a power of two.  It was 64 in Tim Peter's C
	// implementation, but 32 was empirically determined to work better in
	// this implementation.  In the unlikely event that you set this constant
	// to be a number that's not a power of two, you'll need to change the
	//  #minRunLength computation.
	//
	// If you decrease this constant, you must change the stackLen
	// computation in the TimSort constructor, or you risk an
	// ArrayOutOfBounds exception.  See listsort.txt for a discussion
	// of the minimum stack length required as a function of the length
	// of the array being sorted and the minimum merge sequence length.
	//
	MIN_MERGE = 32

	// When we get into galloping mode, we stay there until both runs win less
	// often than MIN_GALLOP consecutive times.
	MIN_GALLOP = 7

	// Maximum initial size of tmp array, which is used for merging.
	// The array can grow to accommodate demand.
	INITIAL_TMP_STORAGE_LENGTH = 256
)

type TimSorter struct {

	// the array being sorted
	a []Comparable

	/**
	 * This controls when we get *into* galloping mode.  It is initialized
	 * to MIN_GALLOP.  The mergeLo and mergeHi methods nudge it higher for
	 * random data, and lower for highly structured data.
	 */
	minGallop int

	/**
	 * Temp storage for merges. A workspace array may optionally be
	 * provided in constructor, and if so will be used as long as it
	 * is big enough.
	 */
	tmp     []Comparable
	tmpBase int // base of tmp array slice
	tmpLen  int // length of tmp array slice

	/**
	 * A stack of pending runs yet to be merged.  Run i starts at
	 * address base[i] and extends for len[i] elements.  It's always
	 * true (so long as the indices are in bounds) that:
	 *
	 *     runBase[i] + runLen[i] == runBase[i + 1]
	 *
	 * so we could cut the storage for this, but it's a minor amount,
	 * and keeping all the info explicit simplifies the code.
	 */
	stackSize int // Number of pending runs on stack
	runBase   []int
	runLen    []int
}

/**
 * NewTimSorter Creates a TimSort instance to maintain the state of an ongoing sort.
 *
 * @param a the array to be sorted
 * @param work a workspace array (slice)
 * @param workBase origin of usable space in work array
 * @param workLen usable size of work array
 */
func NewTimSorter(a, work []Comparable, workBase, workLen int) TimSorter {
	ts := TimSorter{
		a:         a,
		minGallop: MIN_GALLOP,
	}
	// Allocate temp storage (which may be increased later if necessary)
	var tlen = INITIAL_TMP_STORAGE_LENGTH
	if len(a) < 2*INITIAL_TMP_STORAGE_LENGTH {
		tlen = len(a) / 2
	}
	if work == nil || workLen < tlen || workBase+tlen > len(work) {
		ts.tmp = make([]Comparable, tlen)
		ts.tmpLen = tlen
	} else {
		ts.tmp = work
		ts.tmpBase = workBase
		ts.tmpLen = workLen
	}
	/*
	 * Allocate runs-to-be-merged stack (which cannot be expanded).  The
	 * stack length requirements are described in listsort.txt.  The C
	 * version always uses the same stack length (85), but this was
	 * measured to be too expensive when sorting "mid-sized" arrays (e.g.,
	 * 100 elements) in Java.  Therefore, we use smaller (but sufficiently
	 * large) stack lengths for smaller arrays.  The "magic numbers" in the
	 * computation below must be changed if MIN_MERGE is decreased.  See
	 * the MIN_MERGE declaration above for more information.
	 * The maximum value of 49 allows for an array up to length
	 * Integer.MAX_VALUE-4, if array is filled by the worst case stack size
	 * increasing scenario. More explanations are given in section 4 of:
	 * http://envisage-project.eu/wp-content/uploads/2015/02/sorting.pdf
	 */
	var stackLen = 49
	if len(a) < 120 {
		stackLen = 5
	} else if len(a) < 1542 {
		stackLen = 10
	} else if len(a) < 119151 {
		stackLen = 24
	}
	ts.runBase = make([]int, stackLen)
	ts.runLen = make([]int, stackLen)
	return ts
}

/**
 * TimSort sorts the given range, using the given workspace array slice
 * for temp storage when possible. This method is designed to be
 * invoked from public methods (in class Arrays) after performing
 * any necessary array bounds checks and expanding parameters into
 * the required forms.
 *
 * @param a the array to be sorted
 * @param lo the index of the first element, inclusive, to be sorted
 * @param hi the index of the last element, exclusive, to be sorted
 * @param work a workspace array (slice)
 * @param workBase origin of usable space in work array
 * @param workLen usable size of work array
 */
func TimSort2(a []Comparable, lo, hi int, work []Comparable, workBase, workLen int) {
	assert(a != nil && lo >= 0 && lo <= hi && hi <= len(a))
	var nRemaining = hi - lo
	if nRemaining < 2 {
		return // Arrays of size 0 and 1 are always sorted
	}

	// If array is small, do a "mini-TimSort" with no merges
	if nRemaining < MIN_MERGE {
		var initRunLen = countRunAndMakeAscending(a, lo, hi)
		BinarySort(a, lo, hi, lo+initRunLen)
		return
	}

	/**
	 * March over the array once, left to right, finding natural runs,
	 * extending short natural runs to minRun elements, and merging runs
	 * to maintain stack invariant.
	 */
	var ts = NewTimSorter(a, work, workBase, workLen)
	var minRun = minRunLength(nRemaining)
	for ok := true; ok; ok = nRemaining != 0 {
		// Identify next run
		var runLen = countRunAndMakeAscending(a, lo, hi)

		// If run is short, extend to min(minRun, nRemaining)
		if runLen < minRun {
			var force = minRun
			if nRemaining <= minRun {
				force = nRemaining
			}
			BinarySort(a, lo, lo+force, lo+runLen)
			runLen = force
		}

		// Push run onto pending-run stack, and maybe merge
		ts.pushRun(lo, runLen)
		ts.mergeCollapse()

		// Advance to find next run
		lo += runLen
		nRemaining -= runLen
	}
	// Merge all remaining runs to complete sort
	assert(lo == hi)
	ts.mergeForceCollapse()
	assert(ts.stackSize == 1)
}

func TimSort(a []Comparable) {
	TimSort2(a, 0, len(a), nil, 0, 0)
}

/**
 * Pushes the specified run onto the pending-run stack.
 *
 * @param runBase index of the first element in the run
 * @param runLen  the number of elements in the run
 */
func (ts *TimSorter) pushRun(runBase, runLen int) {
	ts.runBase[ts.stackSize] = runBase
	ts.runLen[ts.stackSize] = runLen
	ts.stackSize++
}

/**
 * Examines the stack of runs waiting to be merged and merges adjacent runs
 * until the stack invariants are reestablished:
 *
 *     1. runLen[i - 3] > runLen[i - 2] + runLen[i - 1]
 *     2. runLen[i - 2] > runLen[i - 1]
 *
 * This method is called each time a new run is pushed onto the stack,
 * so the invariants are guaranteed to hold for i < stackSize upon
 * entry to the method.
 *
 * Thanks to Stijn de Gouw, Jurriaan Rot, Frank S. de Boer,
 * Richard Bubel and Reiner Hahnle, this is fixed with respect to
 * the analysis in "On the Worst-Case Complexity of TimSort" by
 * Nicolas Auger, Vincent Jug, Cyril Nicaud, and Carine Pivoteau.
 */
func (ts *TimSorter) mergeCollapse() {
	for ts.stackSize > 1 {
		var n = ts.stackSize - 2
		if n > 0 && ts.runLen[n-1] <= ts.runLen[n]+ts.runLen[n+1] ||
			n > 1 && ts.runLen[n-2] <= ts.runLen[n]+ts.runLen[n-1] {
			if ts.runLen[n-1] < ts.runLen[n+1] {
				n--
			}
		} else if n < 0 || ts.runLen[n] > ts.runLen[n+1] {
			break // Invariant is established
		}
		ts.mergeAt(n)
	}
}

/**
 * Merges all runs on the stack until only one remains.  This method is
 * called once, to complete the sort.
 */
func (ts *TimSorter) mergeForceCollapse() {
	for ts.stackSize > 1 {
		var n = ts.stackSize - 2
		if n > 0 && ts.runLen[n-1] < ts.runLen[n+1] {
			n--
		}
		ts.mergeAt(n)
	}
}

/**
 * Merges the two runs at stack indices i and i+1.  Run i must be
 * the penultimate or antepenultimate run on the stack.  In other words,
 * i must be equal to stackSize-2 or stackSize-3.
 *
 * @param i stack index of the first of the two runs to merge
 */
func (ts *TimSorter) mergeAt(i int) {
	assert(ts.stackSize >= 2)
	assert(i >= 0)
	assert(i == ts.stackSize-2 || i == ts.stackSize-3)

	var base1 = ts.runBase[i]
	var len1 = ts.runLen[i]
	var base2 = ts.runBase[i+1]
	var len2 = ts.runLen[i+1]
	assert(len1 > 0 && len2 > 0)
	assert(base1+len1 == base2)

	/*
	 * Record the length of the combined runs; if i is the 3rd-last
	 * run now, also slide over the last run (which isn't involved
	 * in this merge).  The current run (i+1) goes away in any case.
	 */
	ts.runLen[i] = len1 + len2
	if i == ts.stackSize-3 {
		ts.runBase[i+1] = ts.runBase[i+2]
		ts.runLen[i+1] = ts.runLen[i+2]
	}
	ts.stackSize--

	/*
	 * Find where the first element of run2 goes in run1. Prior elements
	 * in run1 can be ignored (because they're already in place).
	 */
	var k = gallopRight(ts.a[base2], ts.a, base1, len1, 0)
	assert(k >= 0)
	base1 += k
	len1 -= k
	if len1 == 0 {
		return
	}

	/*
	 * Find where the last element of run1 goes in run2. Subsequent elements
	 * in run2 can be ignored (because they're already in place).
	 */
	len2 = gallopLeft(ts.a[base1+len1-1], ts.a, base2, len2, len2-1)
	assert(len2 >= 0)
	if len2 == 0 {
		return
	}

	// Merge remaining runs, using tmp array with min(len1, len2) elements
	if len1 <= len2 {
		ts.mergeLo(base1, len1, base2, len2)
	} else {
		ts.mergeHi(base1, len1, base2, len2)
	}
}

/**
 * Merges two adjacent runs in place, in a stable fashion.  The first
 * element of the first run must be greater than the first element of the
 * second run (a[base1] > a[base2]), and the last element of the first run
 * (a[base1 + len1-1]) must be greater than all elements of the second run.
 *
 * For performance, this method should be called only when len1 <= len2;
 * its twin, mergeHi should be called if len1 >= len2.  (Either method
 * may be called if len1 == len2.)
 *
 * @param base1 index of first element in first run to be merged
 * @param len1  length of first run to be merged (must be > 0)
 * @param base2 index of first element in second run to be merged
 *        (must be aBase + aLen)
 * @param len2  length of second run to be merged (must be > 0)
 */
func (ts *TimSorter) mergeLo(base1, len1, base2, len2 int) {
	assert(len1 > 0 && len2 > 0 && base1+len1 == base2)

	// Copy first run into temp array
	var a = ts.a // For performance
	var tmp = ts.ensureCapacity(len1)
	var cursor1 = ts.tmpBase // Indexes into tmp array
	var cursor2 = base2      // Indexes int a
	var dest = base1         // Indexes int a
	arraycopy(a, base1, tmp, cursor1, len1)

	// Move first element of second run and deal with degenerate cases
	a[dest] = a[cursor2]
	dest++
	cursor2++
	len2--
	if len2 == 0 {
		arraycopy(tmp, cursor1, a, dest, len1)
		return
	}
	if len1 == 1 {
		arraycopy(a, cursor2, a, dest, len2)
		a[dest+len2] = tmp[cursor1] // Last elt of run 1 to end of merge
		return
	}

	var minGallop = ts.minGallop //  "    "       "     "      "
outer:
	for {
		var count1 = 0 // Number of times in a row that first run won
		var count2 = 0 // Number of times in a row that second run won

		/*
		 * Do the straightforward thing until (if ever) one run starts
		 * winning consistently.
		 */
		for ok := true; ok; ok = (count1 | count2) < minGallop {
			assert(len1 > 1 && len2 > 0)
			if a[cursor2].CompareTo(tmp[cursor1]) < 0 {
				a[dest] = a[cursor2]
				dest++
				cursor2++
				count2++
				count1 = 0
				len2--
				if len2 == 0 {
					break outer
				}
			} else {
				a[dest] = tmp[cursor1]
				dest++
				cursor1++
				count1++
				count2 = 0
				len1--
				if len1 == 1 {
					break outer
				}
			}
		}

		/*
		 * One run is winning so consistently that galloping may be a
		 * huge win. So try that, and continue galloping until (if ever)
		 * neither run appears to be winning consistently anymore.
		 */
		for ok := true; ok; ok = count1 >= MIN_GALLOP || count2 >= MIN_GALLOP {
			assert(len1 > 1 && len2 > 0)
			count1 = gallopRight(a[cursor2], tmp, cursor1, len1, 0)
			if count1 != 0 {
				arraycopy(tmp, cursor1, a, dest, count1)
				dest += count1
				cursor1 += count1
				len1 -= count1
				if len1 <= 1 { // len1 == 1 || len1 == 0
					break outer
				}
			}
			a[dest] = a[cursor2]
			dest++
			cursor2++
			len2--
			if len2 == 0 {
				break outer
			}

			count2 = gallopLeft(tmp[cursor1], a, cursor2, len2, 0)
			if count2 != 0 {
				arraycopy(a, cursor2, a, dest, count2)
				dest += count2
				cursor2 += count2
				len2 -= count2
				if len2 == 0 {
					break outer
				}
			}
			a[dest] = tmp[cursor1]
			dest++
			cursor1++
			len1--
			if len1 == 1 {
				break outer
			}
			minGallop--
		}
		if minGallop < 0 {
			minGallop = 0
		}
		minGallop += 2 // Penalize for leaving gallop mode
	} // End of "outer" loop
	if minGallop < 1 {
		minGallop = 1
	}
	ts.minGallop = minGallop // Write back to field

	if len1 == 1 {
		assert(len2 > 0)
		arraycopy(a, cursor2, a, dest, len2)
		a[dest+len2] = tmp[cursor1] //  Last elt of run 1 to end of merge
	} else if len1 == 0 {
		panic("Comparison method violates its general contract!")
	} else {
		assert(len2 == 0)
		assert(len1 > 1)
		arraycopy(tmp, cursor1, a, dest, len1)
	}
}

/**
 * Like mergeLo, except that this method should be called only if
 * len1 >= len2; mergeLo should be called if len1 <= len2.  (Either method
 * may be called if len1 == len2.)
 *
 * @param base1 index of first element in first run to be merged
 * @param len1  length of first run to be merged (must be > 0)
 * @param base2 index of first element in second run to be merged
 *        (must be aBase + aLen)
 * @param len2  length of second run to be merged (must be > 0)
 */
func (ts *TimSorter) mergeHi(base1, len1, base2, len2 int) {
	assert(len1 > 0 && len2 > 0 && base1+len1 == base2)

	// Copy second run into temp array
	var a = ts.a // For performance
	var tmp = ts.ensureCapacity(len2)
	var tmpBase = ts.tmpBase
	arraycopy(a, base2, tmp, tmpBase, len2)

	var cursor1 = base1 + len1 - 1   // Indexes into a
	var cursor2 = tmpBase + len2 - 1 // Indexes into tmp array
	var dest = base2 + len2 - 1      // Indexes into a

	// Move last element of first run and deal with degenerate cases
	a[dest] = a[cursor1]
	dest--
	cursor1--
	len1--
	if len1 == 0 {
		arraycopy(tmp, tmpBase, a, dest-(len2-1), len2)
		return
	}
	if len2 == 1 {
		dest -= len1
		cursor1 -= len1
		arraycopy(a, cursor1+1, a, dest+1, len1)
		a[dest] = tmp[cursor2]
		return
	}
	//
	//	Comparator<? super T> c = this.c;  // Use local variable for performance
	var minGallop = ts.minGallop //  "    "       "     "      "
outer:
	for {
		var count1 = 0 // Number of times in a row that first run won
		var count2 = 0 // Number of times in a row that second run won

		/*
		 * Do the straightforward thing until (if ever) one run
		 * appears to win consistently.
		 */
		for ok := true; ok; ok = (count1 | count2) < minGallop {
			assert(len1 > 0 && len2 > 1)
			if tmp[cursor2].CompareTo(a[cursor1]) < 0 {
				a[dest] = a[cursor1]
				dest--
				cursor1--
				count1++
				count2 = 0
				len1--
				if len1 == 0 {
					break outer
				}
			} else {
				a[dest] = tmp[cursor2]
				dest--
				cursor2--
				count2++
				count1 = 0
				len2--
				if len2 == 1 {
					break outer
				}
			}
		}

		/*
		 * One run is winning so consistently that galloping may be a
		 * huge win. So try that, and continue galloping until (if ever)
		 * neither run appears to be winning consistently anymore.
		 */
		for ok := true; ok; ok = count1 >= MIN_GALLOP || count2 >= MIN_GALLOP {
			assert(len1 > 0 && len2 > 1)
			count1 = len1 - gallopRight(tmp[cursor2], a, base1, len1, len1-1)
			if count1 != 0 {
				dest -= count1
				cursor1 -= count1
				len1 -= count1
				arraycopy(a, cursor1+1, a, dest+1, count1)
				if len1 == 0 {
					break outer
				}
			}
			a[dest] = tmp[cursor2]
			dest--
			cursor2--
			len2--
			if len2 == 1 {
				break outer
			}

			count2 = len2 - gallopLeft(a[cursor1], tmp, tmpBase, len2, len2-1)
			if count2 != 0 {
				dest -= count2
				cursor2 -= count2
				len2 -= count2
				arraycopy(tmp, cursor2+1, a, dest+1, count2)
				if len2 <= 1 { // len2 == 1 || len2 == 0
					break outer
				}
			}
			a[dest] = a[cursor1]
			dest--
			cursor1--
			len1--
			if len1 == 0 {
				break outer
			}
			minGallop--
		}
		if minGallop < 0 {
			minGallop = 0
		}
		minGallop += 2 // Penalize for leaving gallop mode
	} // End of "outer" loop
	if minGallop < 1 {
		minGallop = 1
	}
	ts.minGallop = minGallop // Write back to field

	if len2 == 1 {
		assert(len1 > 0)
		dest -= len1
		cursor1 -= len1
		arraycopy(a, cursor1+1, a, dest+1, len1)
		a[dest] = tmp[cursor2] // Move first elt of run2 to front of merge
	} else if len2 == 0 {
		panic("Comparison method violates its general contract!")
	} else {
		assert(len1 == 0)
		assert(len2 > 0)
		arraycopy(tmp, tmpBase, a, dest-(len2-1), len2)
	}
}

func (ts *TimSorter) ensureCapacity(minCapacity int) []Comparable {
	if ts.tmpLen < minCapacity {
		// Compute smallest power of 2 > minCapacity
		var newSize = -1 >> bits.LeadingZeros64(uint64(minCapacity))
		newSize++

		if newSize < 0 { // Not bloody likely!
			newSize = minCapacity
		} else {
			if n := len(ts.a) >> 1; newSize > n {
				newSize = n
			}
		}

		var newArray = make([]Comparable, newSize)
		ts.tmp = newArray
		ts.tmpLen = newSize
		ts.tmpBase = 0
	}
	return ts.tmp
}

/**
 * Returns the length of the run beginning at the specified position in
 * the specified array and reverses the run if it is descending (ensuring
 * that the run will always be ascending when the method returns).
 *
 * A run is the longest ascending sequence with:
 *
 *    a[lo] <= a[lo + 1] <= a[lo + 2] <= ...
 *
 * or the longest descending sequence with:
 *
 *    a[lo] >  a[lo + 1] >  a[lo + 2] >  ...
 *
 * For its intended use in a stable mergesort, the strictness of the
 * definition of "descending" is needed so that the call can safely
 * reverse a descending sequence without violating stability.
 *
 * @param a the array in which a run is to be counted and possibly reversed
 * @param lo index of the first element in the run
 * @param hi index after the last element that may be contained in the run.
 *        It is required that {@code lo < hi}.
 * @return  the length of the run beginning at the specified position in
 *          the specified array
 */
func countRunAndMakeAscending(a []Comparable, lo, hi int) int {
	assert(lo < hi)
	var runHi = lo + 1
	if runHi == hi {
		return 1
	}

	// Find end of run, and reverse range if descending
	var rh = runHi
	runHi++
	if a[rh].CompareTo(a[lo]) < 0 { // Descending
		for runHi < hi && a[runHi].CompareTo(a[runHi-1]) < 0 {
			runHi++
		}
		reverseRange(a, lo, runHi)
	} else { // Ascending
		for runHi < hi && a[runHi].CompareTo(a[runHi-1]) >= 0 {
			runHi++
		}
	}
	return runHi - lo
}

// Reverse the specified range of the specified array.
func reverseRange(a []Comparable, lo, hi int) {
	hi--
	for lo < hi {
		var t = a[lo]
		a[lo] = a[hi]
		a[hi] = t
		lo++
		hi--
	}
}

/**
 * Returns the minimum acceptable run length for an array of the specified
 * length. Natural runs shorter than this will be extended with
 * {@link #binarySort}.
 *
 * Roughly speaking, the computation is:
 *
 *  If n < MIN_MERGE, return n (it's too small to bother with fancy stuff).
 *  Else if n is an exact power of 2, return MIN_MERGE/2.
 *  Else return an int k, MIN_MERGE/2 <= k <= MIN_MERGE, such that n/k
 *   is close to, but strictly less than, an exact power of 2.
 *
 * For the rationale, see listsort.txt.
 */
func minRunLength(n int) int {
	assert(n >= 0)
	var r = 0 // Becomes 1 if any 1 bits are shifted off
	for n >= MIN_MERGE {
		r |= n & 1
		n >>= 1
	}
	return n + r
}

/**
 * Locates the position at which to insert the specified key into the
 * specified sorted range; if the range contains an element equal to key,
 * returns the index of the leftmost equal element.
 *
 * @param key the key whose insertion point to search for
 * @param a the array in which to search
 * @param base the index of the first element in the range
 * @param len the length of the range; must be > 0
 * @param hint the index at which to begin the search, 0 <= hint < n.
 *     The closer hint is to the result, the faster this method will run.
 *
 * @return the int k,  0 <= k <= n such that a[b + k - 1] < key <= a[b + k],
 *    pretending that a[b - 1] is minus infinity and a[b + n] is infinity.
 *    In other words, key belongs at index b + k; or in other words,
 *    the first k elements of a should precede key, and the last n - k
 *    should follow it.
 */
func gallopLeft(key Comparable, a []Comparable, base, len, hint int) int {
	assert(len > 0 && hint >= 0 && hint < len)
	var lastOfs = 0
	var ofs = 1
	if key.CompareTo(a[base+hint]) > 0 {
		// Gallop right until a[base+hint+lastOfs] < key <= a[base+hint+ofs]
		var maxOfs = len - hint
		for ofs < maxOfs && key.CompareTo(a[base+hint+ofs]) > 0 {
			lastOfs = ofs
			ofs = (ofs << 1) + 1
			if ofs <= 0 { // int overflow
				ofs = maxOfs
			}
		}
		if ofs > maxOfs {
			ofs = maxOfs
		}

		// Make offsets relative to base
		lastOfs += hint
		ofs += hint
	} else { // key <= a[base + hint]
		// Gallop left until a[base+hint-ofs] < key <= a[base+hint-lastOfs]
		var maxOfs = hint + 1
		for ofs < maxOfs && key.CompareTo(a[base+hint-ofs]) <= 0 {
			lastOfs = ofs
			ofs = (ofs << 1) + 1
			if ofs <= 0 { // int overflow
				ofs = maxOfs
			}
		}
		if ofs > maxOfs {
			ofs = maxOfs
		}

		// Make offsets relative to base
		var tmp = lastOfs
		lastOfs = hint - ofs
		ofs = hint - tmp
	}
	assert(-1 <= lastOfs && lastOfs < ofs && ofs <= len)

	/*
	 * Now a[base+lastOfs] < key <= a[base+ofs], so key belongs somewhere
	 * to the right of lastOfs but no farther right than ofs.  Do a binary
	 * search, with invariant a[base + lastOfs - 1] < key <= a[base + ofs].
	 */
	lastOfs++
	for lastOfs < ofs {
		var m = lastOfs + ((ofs - lastOfs) >> 1)

		if key.CompareTo(a[base+m]) > 0 {
			lastOfs = m + 1 // a[base + m] < key
		} else {
			ofs = m // key <= a[base + m]
		}
	}
	assert(lastOfs == ofs) // so a[base + ofs - 1] < key <= a[base + ofs]
	return ofs
}

/**
 * Like gallopLeft, except that if the range contains an element equal to
 * key, gallopRight returns the index after the rightmost equal element.
 *
 * @param key the key whose insertion point to search for
 * @param a the array in which to search
 * @param base the index of the first element in the range
 * @param len the length of the range; must be > 0
 * @param hint the index at which to begin the search, 0 <= hint < n.
 *     The closer hint is to the result, the faster this method will run.
 *
 * @return the int k,  0 <= k <= n such that a[b + k - 1] <= key < a[b + k]
 */
func gallopRight(key Comparable, a []Comparable, base, len, hint int) int {
	assert(len > 0 && hint >= 0 && hint < len)

	var ofs = 1
	var lastOfs = 0
	if key.CompareTo(a[base+hint]) < 0 {
		// Gallop left until a[b+hint - ofs] <= key < a[b+hint - lastOfs]
		var maxOfs = hint + 1
		for ofs < maxOfs && key.CompareTo(a[base+hint-ofs]) < 0 {
			lastOfs = ofs
			ofs = (ofs << 1) + 1
			if ofs <= 0 { // int overflow
				ofs = maxOfs
			}
		}
		if ofs > maxOfs {
			ofs = maxOfs
		}

		// Make offsets relative to b
		var tmp = lastOfs
		lastOfs = hint - ofs
		ofs = hint - tmp
	} else { // a[b + hint] <= key
		// Gallop right until a[b+hint + lastOfs] <= key < a[b+hint + ofs]
		var maxOfs = len - hint
		for ofs < maxOfs && key.CompareTo(a[base+hint+ofs]) >= 0 {
			lastOfs = ofs
			ofs = (ofs << 1) + 1
			if ofs <= 0 { // int overflow
				ofs = maxOfs
			}
		}
		if ofs > maxOfs {
			ofs = maxOfs
		}

		// Make offsets relative to b
		lastOfs += hint
		ofs += hint
	}
	assert(-1 <= lastOfs && lastOfs < ofs && ofs <= len)

	/*
	 * Now a[b + lastOfs] <= key < a[b + ofs], so key belongs somewhere to
	 * the right of lastOfs but no farther right than ofs.  Do a binary
	 * search, with invariant a[b + lastOfs - 1] <= key < a[b + ofs].
	 */
	lastOfs++
	for lastOfs < ofs {
		var m = lastOfs + ((ofs - lastOfs) >> 1)
		if key.CompareTo(a[base+m]) < 0 {
			ofs = m // key < a[b + m]
		} else {
			lastOfs = m + 1 // a[b + m] <= key
		}
	}
	assert(lastOfs == ofs) // so a[b + ofs - 1] <= key < a[b + ofs]
	return ofs
}

func arraycopy(src []Comparable, srcPos int, dest []Comparable, destPos, length int) {
	for i := 0; i < length; i++ {
		src[srcPos] = dest[destPos]
		srcPos++
		destPos++
	}
}

func assert(cond bool) {
	if !cond {
		pc := make([]uintptr, 1)
		numFrames := runtime.Callers(2, pc)
		if numFrames < 1 {
			log.Printf("assertion failed\n")
			return
		}
		frame, _ := runtime.CallersFrames(pc).Next()
		log.Printf("%s(%d) assertion failed", frame.File, frame.Line)
	}
}
