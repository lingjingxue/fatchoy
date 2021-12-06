// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

// 模仿java的丐版Comparable
type Comparable interface {

	// a.CompareTo(b) < 0 表明a < b
	// a.CompareTo(b) > 0 表明a > b
	// a.CompareTo(b) == 0 表明a == b
	//
	// 内部实现要符合结合律:
	// (a.compareTo(b) > 0 && b.compareTo(c) > 0) implies a.compareTo(c) > 0
	CompareTo(Comparable) int
}
