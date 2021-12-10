// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package treemap

type EntryIterator struct {
	owner           *Map
	next            *Entry
	lastReturned    *Entry
	expectedVersion int
}

func NewEntryIterator(m *Map, first *Entry) *EntryIterator {
	return &EntryIterator{
		owner:           m,
		next:            first,
		expectedVersion: m.version,
	}
}

func (it *EntryIterator) HasNext() bool {
	return it.next != nil
}

func (it *EntryIterator) nextEntry() *Entry {
	var e = it.next
	if e == nil {
		panic("EntryIterator: no such element")
	}
	if it.expectedVersion != it.owner.version {
		panic("EntryIterator: concurrent modification")
	}
	it.next = successor(e)
	it.lastReturned = e
	return e
}

func (it *EntryIterator) prevEntry() *Entry {
	var e = it.next
	if e == nil {
		panic("EntryIterator: no such element")
	}
	if it.expectedVersion != it.owner.version {
		panic("EntryIterator: concurrent modification")
	}
	it.next = predecessor(e)
	it.lastReturned = e
	return e
}

func (it *EntryIterator) Next() *Entry {
	return it.nextEntry()
}

// Removes from the underlying collection the last element returned
func (it *EntryIterator) Remove() {
	if it.lastReturned == nil {
		panic("EntryIterator: illegal state")
	}
	if it.expectedVersion != it.owner.version {
		panic("EntryIterator: concurrent modification")
	}
	if it.lastReturned.left != nil && it.lastReturned.right != nil {
		it.next = it.lastReturned
	}
	it.owner.deleteEntry(it.lastReturned)
	it.expectedVersion = it.owner.version
	it.lastReturned = nil
}

type DescendingEntryIterator struct {
	EntryIterator
}

func NewKeyDescendingEntryIterator(m *Map, first *Entry) *DescendingEntryIterator {
	return &DescendingEntryIterator{
		EntryIterator: EntryIterator{
			owner:           m,
			next:            first,
			expectedVersion: m.version,
		},
	}
}

func (it *DescendingEntryIterator) Next() *Entry {
	return it.prevEntry()
}

type KeyIterator struct {
	EntryIterator
}

func NewKeyIterator(m *Map, first *Entry) *KeyIterator {
	return &KeyIterator{
		EntryIterator: EntryIterator{
			owner:           m,
			next:            first,
			expectedVersion: m.version,
		},
	}
}

func (it *KeyIterator) Next() KeyType {
	return it.nextEntry().key
}

type DescendingKeyIterator struct {
	EntryIterator
}

func NewDescendingKeyIterator(m *Map, first *Entry) *DescendingKeyIterator {
	return &DescendingKeyIterator{
		EntryIterator: EntryIterator{
			owner:           m,
			next:            first,
			expectedVersion: m.version,
		},
	}
}

func (it *DescendingKeyIterator) Next() KeyType {
	return it.prevEntry().key
}

func (it *DescendingKeyIterator) Remove() {
	if it.lastReturned == nil {
		panic("DescendingKeyIterator: illegal state")
	}
	if it.expectedVersion != it.owner.version {
		panic("DescendingKeyIterator: concurrent modification")
	}
	it.owner.deleteEntry(it.lastReturned)
	it.lastReturned = nil
	it.expectedVersion = it.owner.version
}

type ValueIterator struct {
	EntryIterator
}

func NewValueIterator(m *Map, first *Entry) *ValueIterator {
	return &ValueIterator{
		EntryIterator: EntryIterator{
			owner:           m,
			next:            first,
			expectedVersion: m.version,
		},
	}
}

func (it *ValueIterator) Next() interface{} {
	return it.nextEntry().value
}
