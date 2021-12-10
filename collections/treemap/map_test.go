// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package treemap

import (
	"strconv"
	"testing"

	"qchen.fun/fatchoy/collections"
)

type Int int

func (n Int) CompareTo(other collections.Comparable) int {
	var n2 = other.(Int)
	if n < n2 {
		return -1
	} else if n > n2 {
		return 1
	}
	return 0
}

func TestMapExample(t *testing.T) {
	var m = New()
	for i := 0; i < 20; i++ {
		m.Put(Int(i), strconv.Itoa(i))
	}
	t.Logf("size: %d", m.Size())
	t.Logf("first: %v, last %v", m.FirstKey(), m.LastKey())

	for i := 0; i < 20; i++ {
		if i % 2 == 0 {
			m.Remove(Int(i))
		}
	}

	t.Logf("size: %d", m.Size())

	m.Foreach(func(key KeyType, val interface{}){
		t.Logf("%v => %v", key, val)
	})
}
