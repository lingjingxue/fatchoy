// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package treemap

// A Red-Black tree based map implementation.
// more details see [java.util.TreeMap](https://bit.ly/3rQWBYO)
type Map struct {
	root    *Entry
	size    int
	version int
}

func New() *Map {
	return &Map{}
}

// Returns the number of key-value mappings in this map.
func (m *Map) Size() int {
	return m.size
}

func (m *Map) IsEmpty() bool {
	return m.size == 0
}

// Return true if this map contains a mapping for the specified key
func (m *Map) Contains(key KeyType) bool {
	return m.getEntry(key) != nil
}

// Returns the value to which the specified key is mapped,
// or nil if this map contains no mapping for the key.
func (m *Map) Get(key KeyType) (interface{}, bool) {
	var p = m.getEntry(key)
	if p != nil {
		return p.value, true
	}
	return nil, false
}

// Returns the value to which the specified key is mapped,
// or `defaultValue` if this map contains no mapping for the key.
func (m *Map) GetOrDefault(key KeyType, defaultValue interface{}) interface{} {
	var p = m.getEntry(key)
	if p != nil {
		return p.value
	}
	return defaultValue
}

func (m *Map) FirstEntry() *Entry {
	return m.getFirstEntry()
}

// Returns the first key in the TreeMap (according to the key's order)
func (m *Map) FirstKey() KeyType {
	return key(m.getFirstEntry())
}

func (m *Map) LastEntry() *Entry {
	return m.getLastEntry()
}

// Returns the last key in the TreeMap (according to the key's order)
func (m *Map) LastKey() KeyType {
	return key(m.getLastEntry())
}

// Gets the entry corresponding to the specified key;
// if no such entry exists, returns the entry for the greatest key less than the specified key;
func (m *Map) FloorEntry(key KeyType) *Entry {
	return m.getFloorEntry(key)
}

// Gets the specified key, returns the greatest key less than the specified key if not exist.
func (m *Map) FloorKey(key KeyType) KeyType {
	var entry = m.getFloorEntry(key)
	if entry != nil {
		return entry.key
	}
	return nil
}

// Gets the entry corresponding to the specified key;
// returns the entry for the least key greater than the specified key if not exist.
func (m *Map) CeilingEntry(key KeyType) *Entry {
	return m.getCeilingEntry(key)
}

// Gets the specified key, return the least key greater than the specified key if not exist.
func (m *Map) CeilingKey(key KeyType) KeyType {
	var entry = m.getCeilingEntry(key)
	if entry != nil {
		return entry.key
	}
	return nil
}

// Gets the entry for the least key greater than the specified key
func (m *Map) HigherEntry(key KeyType) *Entry {
	return m.getHigherEntry(key)
}

// Returns the least key greater than the specified key
func (m *Map) HigherKey(key KeyType) KeyType {
	var entry = m.getHigherEntry(key)
	if entry != nil {
		return entry.key
	}
	return nil
}

// Performs the given action for each entry in this map until all entries
// have been processed or the action panic
func (m *Map) Foreach(action EntryAction) {
	var ver = m.version
	for e := m.getFirstEntry(); e != nil; e = successor(e) {
		action(e.key, e.value)
		if ver != m.version {
			panic("concurrent map modification")
		}
	}
}

// in-order traversal
func (m *Map) InOrderTraversal(action EntryAction) {
	inOrderTraversal(m.root, action)
}

// pre-order traversal
func (m *Map) PreOrderTraversal(action EntryAction) {
	preOrderTraversal(m.root, action)
}

// post-order traversal
func (m *Map) PostOrderTraversal(action EntryAction) {
	postOrderTraversal(m.root, action)
}

// Return list of all keys
func (m *Map) Keys() []KeyType {
	var keys = make([]KeyType, 0, m.size)
	for e := m.getFirstEntry(); e != nil; e = successor(e) {
		keys = append(keys, e.key)
	}
	return keys
}

// Return list of all values
func (m *Map) Values() []interface{} {
	var values = make([]interface{}, 0, m.size)
	for e := m.getFirstEntry(); e != nil; e = successor(e) {
		values = append(values, e.value)
	}
	return values
}

func (m *Map) Iterator() *EntryIterator {
	return NewEntryIterator(m, m.getFirstEntry())
}

func (m *Map) DescendingIterator() *DescendingEntryIterator {
	return NewKeyDescendingEntryIterator(m, m.getLastEntry())
}

func (m *Map) KeyIterator() *KeyIterator {
	return NewKeyIterator(m, m.getFirstEntry())
}

func (m *Map) DescendingKeyIterator() *DescendingKeyIterator {
	return NewDescendingKeyIterator(m, m.getLastEntry())
}

func (m *Map) ValueIterator() *ValueIterator {
	return NewValueIterator(m, m.getFirstEntry())
}

// Removes the mapping for this key from this TreeMap if present.
func (m *Map) Remove(key KeyType) bool {
	var p = m.getEntry(key)
	if p != nil {
		m.deleteEntry(p)
		return true
	}
	return false
}

// Removes all the mappings from this map.
func (m *Map) Clear() {
	m.size = 0
	m.root = nil
}

// Associates the specified value with the specified key in this map.
// If the map previously contained a mapping for the key, the old value is replaced.
func (m *Map) Put(key KeyType, value interface{}) interface{} {
	var t = m.root
	if t == nil {
		m.root = NewEntry(key, value, nil)
		m.size = 1
		m.version++
		return nil
	}
	var cmp int
	var parent *Entry
	for {
		parent = t
		cmp = key.CompareTo(t.key)
		if cmp < 0 {
			t = t.left
		} else if cmp > 0 {
			t = t.right
		} else {
			return t.SetValue(value)
		}
		if t == nil {
			break
		}
	}
	var e = NewEntry(key, value, parent)
	if cmp < 0 {
		parent.left = e
	} else {
		parent.right = e
	}
	m.fixAfterInsertion(e)
	m.size++
	m.version++
	return nil
}

// Returns the first Entry in the TreeMap (according to the key's order)
// Returns nil if the TreeMap is empty.
func (m *Map) getFirstEntry() *Entry {
	var p = m.root
	if p != nil {
		for p.left != nil {
			p = p.left
		}
	}
	return p
}

// Returns the last Entry in the TreeMap (according to the key's order)
// Returns nil if the TreeMap is empty.
func (m *Map) getLastEntry() *Entry {
	var p = m.root
	if p != nil {
		for p.right != nil {
			p = p.right
		}
	}
	return p
}

// Returns this map's entry for the given key,
// or nil if the map does not contain an entry for the key.
func (m *Map) getEntry(key KeyType) *Entry {
	var node = m.root
	for node != nil {
		var cmp = key.CompareTo(node.key)
		if cmp < 0 {
			node = node.left
		} else if cmp > 0 {
			node = node.right
		} else {
			return node
		}
	}
	return nil
}

// Gets the entry corresponding to the specified key;
// if no such entry exists, returns the entry for the least key greater than the specified key;
// if no such entry exists returns nil.
func (m *Map) getCeilingEntry(key KeyType) *Entry {
	var p = m.root
	for p != nil {
		var cmp = key.CompareTo(p.key)
		if cmp < 0 {
			if p.left != nil {
				p = p.left
			} else {
				return p
			}
		} else if cmp > 0 {
			if p.right != nil {
				p = p.right
			} else {
				var parent = p.parent
				var ch = p
				for parent != nil && ch == parent.right {
					ch = parent
					parent = parent.parent
				}
				return parent
			}
		} else {
			return p
		}
	}
	return nil
}

// Gets the entry corresponding to the specified key;
// if no such entry exists, returns the entry for the greatest key less than the specified key;
// if no such entry exists, returns nil.
func (m *Map) getFloorEntry(key KeyType) *Entry {
	var p = m.root
	for p != nil {
		var cmp = key.CompareTo(p.key)
		if cmp > 0 {
			if p.right != nil {
				p = p.right
			} else {
				return p
			}
		} else if cmp < 0 {
			if p.left != nil {
				p = p.left
			} else {
				var parent = p.parent
				var ch = p
				for parent != nil && ch == parent.left {
					ch = parent
					parent = parent.parent
				}
				return parent
			}
		} else {
			return p
		}

	}
	return nil
}

// Gets the entry for the least key greater than the specified key;
// if no such entry exists, returns the entry for the least key greater than the specified key;
// if no such entry exists returns nil.
func (m *Map) getHigherEntry(key KeyType) *Entry {
	var p = m.root
	for p != nil {
		var cmp = key.CompareTo(p.key)
		if cmp < 0 {
			if p.left != nil {
				p = p.left
			} else {
				return p
			}
		} else {
			if p.right != nil {
				p = p.right
			} else {
				var parent = p.parent
				var ch = p
				for parent != nil && ch == parent.right {
					ch = parent
					parent = parent.parent
				}
				return parent
			}
		}
	}
	return nil
}

// Returns the entry for the greatest key less than the specified key;
// if no such entry exists (i.e., the least key in the Tree is greater than the specified key), returns nil
func (m *Map) getLowerEntry(key KeyType) *Entry {
	var p = m.root
	for p != nil {
		var cmp = key.CompareTo(p.key)
		if cmp > 0 {
			if p.right != nil {
				p = p.right
			} else {
				return p
			}
		} else {
			if p.left != nil {
				p = p.left
			} else {
				var parent = p.parent
				var ch = p
				for parent != nil && ch == parent.left {
					ch = parent
					parent = parent.parent
				}
				return parent
			}
		}
	}
	return nil
}

/**
 * Balancing operations.
 *
 * Implementations of rebalancings during insertion and deletion are
 * slightly different than the CLR version.  Rather than using dummy
 * nil nodes, we use a set of accessors that deal properly with nil.  They
 * are used to avoid messiness surrounding nullness checks in the main
 * algorithms.
 */

func (m *Map) rotateLeft(p *Entry) {
	if p == nil {
		return
	}
	var r = p.right
	p.right = r.left
	if r.left != nil {
		r.left.parent = p
	}
	r.parent = p.parent
	if p.parent == nil {
		m.root = r
	} else if p.parent.left == p {
		p.parent.left = r
	} else {
		p.parent.right = r
	}
	r.left = p
	p.parent = r
}

func (m *Map) rotateRight(p *Entry) {
	if p == nil {
		return
	}
	var l = p.left
	p.left = l.right
	if l.right != nil {
		l.right.parent = p
	}
	l.parent = p.parent
	if p.parent == nil {
		m.root = l
	} else if p.parent.right == p {
		p.parent.right = l
	} else {
		p.parent.left = l
	}
	l.right = p
	p.parent = l
}

func (m *Map) fixAfterInsertion(x *Entry) {
	x.color = RED
	for x != nil && x != m.root && x.parent.color == RED {
		if parentOf(x) == leftOf(parentOf(parentOf(x))) {
			var y = rightOf(parentOf(parentOf(x)))
			if colorOf(y) == RED {
				setColor(parentOf(x), BLACK)
				setColor(y, BLACK)
				setColor(parentOf(parentOf(x)), RED)
				x = parentOf(parentOf(x))
			} else {
				if x == rightOf(parentOf(x)) {
					x = parentOf(x)
					m.rotateLeft(x)
				}
				setColor(parentOf(x), BLACK)
				setColor(parentOf(parentOf(x)), RED)
				m.rotateRight(parentOf(parentOf(x)))
			}
		} else {
			var y = leftOf(parentOf(parentOf(x)))
			if colorOf(y) == RED {
				setColor(parentOf(x), BLACK)
				setColor(y, BLACK)
				setColor(parentOf(parentOf(x)), RED)
				x = parentOf(parentOf(x))
			} else {
				if x == leftOf(parentOf(x)) {
					x = parentOf(x)
					m.rotateRight(x)
				}
				setColor(parentOf(x), BLACK)
				setColor(parentOf(parentOf(x)), RED)
				m.rotateLeft(parentOf(parentOf(x)))
			}
		}
	}
	m.root.color = BLACK
}

func (m *Map) deleteEntry(p *Entry) {
	m.version++
	m.size--

	// If strictly internal, copy successor's element to p and then make p
	// point to successor.
	if p.left != nil && p.right != nil {
		var s = successor(p)
		p.key = s.key
		p.value = s.value
		p = s
	} // p has 2 children

	// Start fixup at replacement node, if it exists.
	var replacement = p.left
	if p.left == nil {
		replacement = p.right
	}

	if replacement != nil {
		// Link replacement to parent
		replacement.parent = p.parent
		if p.parent == nil {
			m.root = replacement
		} else if p == p.parent.left {
			p.parent.left = replacement
		} else {
			p.parent.right = replacement
		}

		// Null out links so they are OK to use by fixAfterDeletion.
		p.left = nil
		p.right = nil
		p.parent = nil

		// Fix replacement
		if p.color == BLACK {
			m.fixAfterDeletion(replacement)
		}
	} else if p.parent == nil { // return if we are the only node.
		m.root = nil
	} else { //  No children. Use self as phantom replacement and unlink.
		if p.color == BLACK {
			m.fixAfterDeletion(p)
		}
		if p.parent != nil {
			if p == p.parent.left {
				p.parent.left = nil
			} else if p == p.parent.right {
				p.parent.right = nil
			}
			p.parent = nil
		}
	}
}

func (m *Map) fixAfterDeletion(x *Entry) {
	for x != m.root && colorOf(x) == BLACK {
		if x == leftOf(parentOf(x)) {
			var sib = rightOf(parentOf(x))

			if colorOf(sib) == RED {
				setColor(sib, BLACK)
				setColor(parentOf(x), RED)
				m.rotateLeft(parentOf(x))
				sib = rightOf(parentOf(x))
			}

			if colorOf(leftOf(sib)) == BLACK &&
				colorOf(rightOf(sib)) == BLACK {
				setColor(sib, RED)
				x = parentOf(x)
			} else {
				if colorOf(rightOf(sib)) == BLACK {
					setColor(leftOf(sib), BLACK)
					setColor(sib, RED)
					m.rotateRight(sib)
					sib = rightOf(parentOf(x))
				}
				setColor(sib, colorOf(parentOf(x)))
				setColor(parentOf(x), BLACK)
				setColor(rightOf(sib), BLACK)
				m.rotateLeft(parentOf(x))
				x = m.root
			}
		} else { // symmetric
			var sib = leftOf(parentOf(x))

			if colorOf(sib) == RED {
				setColor(sib, BLACK)
				setColor(parentOf(x), RED)
				m.rotateRight(parentOf(x))
				sib = leftOf(parentOf(x))
			}

			if colorOf(rightOf(sib)) == BLACK &&
				colorOf(leftOf(sib)) == BLACK {
				setColor(sib, RED)
				x = parentOf(x)
			} else {
				if colorOf(leftOf(sib)) == BLACK {
					setColor(rightOf(sib), BLACK)
					setColor(sib, RED)
					m.rotateLeft(sib)
					sib = leftOf(parentOf(x))
				}
				setColor(sib, colorOf(parentOf(x)))
				setColor(parentOf(x), BLACK)
				setColor(leftOf(sib), BLACK)
				m.rotateRight(parentOf(x))
				x = m.root
			}
		}
	}
	setColor(x, BLACK)
}
