// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

import (
	"testing"
)

func TestIDSetExample(t *testing.T) {
	var idset IDSet
	for i := 50; i < 100; i++ {
		idset = idset.Insert(int32(i))
		if i%7 == 0 {
			idset = idset.Delete(int32(i))
		}
	}
	for i := 1; i < 50; i++ {
		idset = idset.Insert(int32(i))
	}
	idset.Has(7)
}

func TestOrderedIDSetExample(t *testing.T) {
	var idset OrderedIDSet
	for i := 50; i < 100; i++ {
		idset = idset.Insert(int32(i))
		if i%7 == 0 {
			idset = idset.Delete(int32(i))
		}
	}
	for i := 1; i < 50; i++ {
		idset = idset.Insert(int32(i))
	}
	idset.Has(7)
}
