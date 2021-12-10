// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package treemap

import (
	"qchen.fun/fatchoy/collections"
)

const (
	RED   = 0
	BLACK = 1
)

type KeyType collections.Comparable

type Entry struct {
	left, right, parent *Entry
	key                 KeyType
	value               interface{}
	color               int8
}

func NewMapEntry(key KeyType, val interface{}, parent *Entry) *Entry {
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

func colorOf(p *Entry) int8 {
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

func setColor(p *Entry, color int8) {
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