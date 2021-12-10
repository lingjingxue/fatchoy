// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package treemap

import (
	"qchen.fun/fatchoy/collections"
)

type (
	KeyType     collections.Comparable
	EntryAction func(key KeyType, val interface{})
)

type Color int8

const (
	RED   Color = 0
	BLACK Color = 1
)

func (c Color) String() string {
	switch c {
	case RED:
		return "red"
	case BLACK:
		return "black"
	default:
		return "??"
	}
}

type Entry struct {
	left, right, parent *Entry
	key                 KeyType
	value               interface{}
	color               Color
}

func NewEntry(key KeyType, val interface{}, parent *Entry) *Entry {
	return &Entry{
		key:    key,
		value:  val,
		parent: parent,
		color:  BLACK,
	}
}

func (e *Entry) GetKey() KeyType {
	return e.key
}

func (e *Entry) GetValue() interface{} {
	return e.value
}

func (e *Entry) SetValue(val interface{}) interface{} {
	var old = e.value
	e.value = val
	return old
}

func (e *Entry) Equals(other *Entry) bool {
	if e == other {
		return true
	}
	return e.key == other.key && e.value == other.value
}

func key(e *Entry) KeyType {
	if e != nil {
		return e.key
	}
	return nil
}

func colorOf(p *Entry) Color {
	if p != nil {
		return p.color
	}
	return BLACK
}

func parentOf(p *Entry) *Entry {
	if p != nil {
		return p.parent
	}
	return nil
}

func setColor(p *Entry, color Color) {
	if p != nil {
		p.color = color
	}
}

func leftOf(p *Entry) *Entry {
	if p != nil {
		return p.left
	}
	return nil
}

func rightOf(p *Entry) *Entry {
	if p != nil {
		return p.right
	}
	return nil
}

// Returns the successor of the specified Entry, or null if no such.
func successor(t *Entry) *Entry {
	if t == nil {
		return nil
	} else if t.right != nil {
		var p = t.right
		for p.left != nil {
			p = p.left
		}
		return p
	} else {
		var p = t.parent
		var ch = t
		for p != nil && ch == p.right {
			ch = p
			p = p.parent
		}
		return p
	}
}

// Returns the predecessor of the specified Entry, or null if no such.
func predecessor(t *Entry) *Entry {
	if t == nil {
		return nil
	} else if t.left != nil {
		var p = t.left
		for p.right != nil {
			p = p.right
		}
		return p
	} else {
		var p = t.parent
		var ch = t
		for p != nil && ch == p.left {
			ch = p
			p = p.parent
		}
		return p
	}
}

// in-order traversal
func inOrderTraversal(entry *Entry, action EntryAction) {
	if entry == nil {
		return
	}
	inOrderTraversal(entry.left, action)
	action(entry.key, entry.value)
	inOrderTraversal(entry.right, action)
}

// pre-order traversal
func preOrderTraversal(entry *Entry, action EntryAction) {
	if entry == nil {
		return
	}
	action(entry.key, entry.value)
	preOrderTraversal(entry.left, action)
	preOrderTraversal(entry.right, action)
}

// post-order traversal
func postOrderTraversal(entry *Entry, action EntryAction) {
	if entry == nil {
		return
	}
	postOrderTraversal(entry.left, action)
	postOrderTraversal(entry.right, action)
	action(entry.key, entry.value)
}
