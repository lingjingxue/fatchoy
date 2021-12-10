// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package treemap

import (
	"fmt"
	"strings"
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

func createTreeMap() *Map {
	var m = New()
	m.Put(Int(5), "e")
	m.Put(Int(6), "f")
	m.Put(Int(7), "g")
	m.Put(Int(3), "c")
	m.Put(Int(4), "d")
	m.Put(Int(1), "x")
	m.Put(Int(2), "b")
	m.Put(Int(1), "a") //overwrite
	return m
}

func mapKeysText(m *Map) string {
	var sb strings.Builder
	for _, key := range m.Keys() {
		fmt.Fprintf(&sb, "%v", key)
	}
	return sb.String()
}

func mapValuesText(m *Map) string {
	var sb strings.Builder
	for _, val := range m.Values() {
		fmt.Fprintf(&sb, "%v", val)
	}
	return sb.String()
}

func checkMapKeyValue(t *testing.T, m *Map, keyS, valueS string, size int) {
	if actualValue := m.Size(); actualValue != size {
		t.Errorf("Got %v expected %v", actualValue, size)
	}
	if actualValue, expectedValue := mapKeysText(m), keyS; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}
	if actualValue, expectedValue := mapValuesText(m), valueS; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}
}

func TestTreeMapPut(t *testing.T) {
	var m = createTreeMap()

	checkMapKeyValue(t, m, "1234567", "abcdefg", 7)

	tests := []struct {
		key   KeyType
		value interface{}
		found bool
	}{
		{Int(1), "a", true},
		{Int(2), "b", true},
		{Int(3), "c", true},
		{Int(4), "d", true},
		{Int(5), "e", true},
		{Int(6), "f", true},
		{Int(7), "g", true},
		{Int(8), nil, false},
	}
	for _, tc := range tests {
		// retrievals
		actualValue, actualFound := m.Get(tc.key)
		if actualValue != tc.value || actualFound != tc.found {
			t.Errorf("Got %v expected %v", actualValue, tc.value)
		}
	}
}

func TestTreeMapRemove(t *testing.T) {
	var m = createTreeMap()
	for i := 5; i <= 8; i++ {
		m.Remove(Int(i))
	}
	m.Remove(Int(5)) // remove again

	checkMapKeyValue(t, m, "1234", "abcd", 4)

	tests := []struct {
		key   KeyType
		value interface{}
		found bool
	}{
		{Int(1), "a", true},
		{Int(2), "b", true},
		{Int(3), "c", true},
		{Int(4), "d", true},
		{Int(5), nil, false},
		{Int(6), nil, false},
		{Int(7), nil, false},
		{Int(8), nil, false},
	}
	for _, tc := range tests {
		// retrievals
		actualValue, actualFound := m.Get(tc.key)
		if actualValue != tc.value || actualFound != tc.found {
			t.Errorf("Got %v expected %v", actualValue, tc.value)
		}
	}

	m.Clear()
	checkMapKeyValue(t, m, "", "", 0)
}

func TestTreeMapFirstLast(t *testing.T) {
	var m Map
	if actualValue := m.FirstKey(); actualValue != nil {
		t.Errorf("Got %v expected %v", actualValue, nil)
	}
	if actualValue := m.LastKey(); actualValue != nil {
		t.Errorf("Got %v expected %v", actualValue, nil)
	}

	m.Put(Int(1), "a")
	m.Put(Int(5), "e")
	m.Put(Int(6), "f")
	m.Put(Int(7), "g")
	m.Put(Int(3), "c")
	m.Put(Int(4), "d")
	m.Put(Int(1), "x") // overwrite
	m.Put(Int(2), "b")

	firstKey, lastKey := m.FirstKey(), m.LastKey()
	firstVal, _ := m.Get(firstKey)
	lastVal, _ := m.Get(lastKey)

	if actualValue, expectedValue := fmt.Sprintf("%v", firstKey), "1"; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}
	if actualValue, expectedValue := fmt.Sprintf("%v", firstVal), "x"; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}

	if actualValue, expectedValue := fmt.Sprintf("%v", lastKey), "7"; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}
	if actualValue, expectedValue := fmt.Sprintf("%v", lastVal), "g"; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}
}

func TestTreeMapCeilingAndFloor(t *testing.T) {
	var m Map

	if entry := m.FloorEntry(Int(0)); entry != nil {
		t.Errorf("Got %v expected %v", entry, "<nil>")
	}
	if entry := m.CeilingEntry(Int(0)); entry != nil {
		t.Errorf("Got %v expected %v", entry, "<nil>")
	}

	m.Put(Int(5), "e")
	m.Put(Int(6), "f")
	m.Put(Int(7), "g")
	m.Put(Int(3), "c")
	m.Put(Int(4), "d")
	m.Put(Int(1), "x")
	m.Put(Int(2), "b")

	if node := m.FloorEntry(Int(4)); node.GetKey() != Int(4) {
		t.Errorf("Got %v expected %v", node.GetKey(), 4)
	}
	if node := m.FloorEntry(Int(0)); node != nil {
		t.Errorf("Got %v expected %v", node.GetKey(), "<nil>")
	}

	if node := m.CeilingEntry(Int(4)); node.GetKey() != Int(4) {
		t.Errorf("Got %v expected %v", node.GetKey(), 4)
	}
	if node := m.CeilingEntry(Int(8)); node != nil {
		t.Errorf("Got %v expected %v", node.GetKey(), "<nil>")
	}
}

func createTreeMap2() *Map {
	var m = New()
	m.Put(Int(5), "e")
	m.Put(Int(6), "f")
	m.Put(Int(7), "g")
	m.Put(Int(3), "c")
	m.Put(Int(4), "d")
	m.Put(Int(1), "x")
	m.Put(Int(2), "b")
	m.Put(Int(1), "a") //overwrite

	// │   ┌── 7
	// └── 6
	//     │   ┌── 5
	//     └── 4
	//         │   ┌── 3
	//         └── 2
	//             └── 1
	return m
}

func TestTreeMapIterator(t *testing.T) {
	var m = createTreeMap2()

	var count = 0
	var sb1 strings.Builder
	var sb2 strings.Builder

	var iter = m.Iterator()
	for iter.HasNext() {
		count++
		var entry = iter.Next()
		fmt.Fprintf(&sb1, "%v", entry.GetKey())
		fmt.Fprintf(&sb2, "%v", entry.GetValue())
	}
	if actualValue, expectedValue := sb1.String(), "1234567"; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}
	if actualValue, expectedValue := sb2.String(), "abcdefg"; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}
	if actualValue, expectedValue := count, m.Size(); actualValue != expectedValue {
		t.Errorf("Size different. Got %v expected %v", actualValue, expectedValue)
	}
}

func TestTreeMapDescendingIterator(t *testing.T) {
	var m = createTreeMap2()

	var count = 0
	var sb1 strings.Builder
	var sb2 strings.Builder

	var iter = m.DescendingIterator()
	for iter.HasNext() {
		count++
		var entry = iter.Next()
		fmt.Fprintf(&sb1, "%v", entry.GetKey())
		fmt.Fprintf(&sb2, "%v", entry.GetValue())
	}
	if actualValue, expectedValue := sb1.String(), "7654321"; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}
	if actualValue, expectedValue := sb2.String(), "gfedcba"; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}
	if actualValue, expectedValue := count, m.Size(); actualValue != expectedValue {
		t.Errorf("Size different. Got %v expected %v", actualValue, expectedValue)
	}
}
